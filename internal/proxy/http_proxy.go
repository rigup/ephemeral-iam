package proxy

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"time"

	"emperror.dev/emperror"
	"github.com/elazarl/goproxy"
	credentialspb "google.golang.org/genproto/googleapis/iam/credentials/v1"

	"github.com/jessesomerville/ephemeral-iam/internal/appconfig"
	"github.com/jessesomerville/ephemeral-iam/internal/gcpclient"
)

var (
	certCache = make(map[string]*tls.Certificate)
	certLock  = &sync.Mutex{}
)

// StartProxyServer spins up the proxy that replaces the gcloud auth token
func StartProxyServer(privilegedAccessToken *credentialspb.GenerateAccessTokenResponse) {

	accessToken := privilegedAccessToken.GetAccessToken()
	expirationDate := privilegedAccessToken.GetExpireTime().AsTime()

	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = config.AuthProxy.Verbose
	if config.AuthProxy.WriteToFile {
		_, err := os.Stat(config.AuthProxy.LogDir)
		if os.IsNotExist(err) {
			emperror.Panic(os.MkdirAll(config.AuthProxy.LogDir, 0755))
		}
		// Create log file
		timestamp := time.Now().Format("20060102150405")
		logFilename := filepath.Join(config.AuthProxy.LogDir, fmt.Sprintf("%s_auth_proxy.log", timestamp))
		logFile, err := os.OpenFile(logFilename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		emperror.Panic(err)

		// I'm not calling logFile.Close() because it gets invoked too early

		// Set auth proxy to log to file
		proxy.Logger = log.New(logFile, "", log.LstdFlags)
		proxy.Logger.Printf("test")
		logger.Info("Writing auth proxy logs to ", logFilename)
	}

	setCa(appconfig.CertFile, appconfig.KeyFile)

	proxy.OnRequest().HandleConnect(goproxy.FuncHttpsHandler(proxyConnectHandle))

	proxy.OnRequest().DoFunc(func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		r.Header.Set("authorization", fmt.Sprintf("Bearer %s", accessToken))
		return r, nil
	})

	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", config.AuthProxy.ProxyAddress, config.AuthProxy.ProxyPort),
		Handler: proxy,
	}

	proxyServerExitDone := &sync.WaitGroup{}
	proxyServerExitDone.Add(1)

	// Catch interrupts to gracefully shutdown the proxy and restore the gcloud config
	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		// An interrupt signal was recieved, shutdown the proxy server
		emperror.Panic(srv.Shutdown(context.Background()))
		close(idleConnsClosed)
		fmt.Println()
		logger.Info("Stopping auth proxy and restoring gcloud config")
		emperror.Panic(gcpclient.UnsetGcloudProxy())
		os.Exit(0)
	}()

	sessionLength := time.Until(expirationDate)
	expiresOn := time.Now().Add(sessionLength)

	logger.Info("Starting auth proxy")
	go func() {
		defer proxyServerExitDone.Done()
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			logger.Fatal("Error starting or closing auth proxy")
		}
		<-idleConnsClosed
	}()

	logger.Info("Privileged session will last until ", expiresOn.Format(time.RFC1123))
	logger.Warn("Press CTRL+C to quit privileged session")

	// Wait until the token expires
	time.Sleep(sessionLength)

	logger.Info("Privileged session expired, stopping auth proxy and restoring gcloud config")
	emperror.Panic(srv.Shutdown(context.Background()))
	emperror.Panic(gcpclient.UnsetGcloudProxy())
}

func proxyConnectHandle(host string, ctx *goproxy.ProxyCtx) (*goproxy.ConnectAction, string) {
	return goproxy.MitmConnect, host
}
