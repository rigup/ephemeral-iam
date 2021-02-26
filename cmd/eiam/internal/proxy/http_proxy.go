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

	"github.com/elazarl/goproxy"
	"github.com/spf13/viper"
	credentialspb "google.golang.org/genproto/googleapis/iam/credentials/v1"

	"github.com/jessesomerville/ephemeral-iam/cmd/eiam/internal/appconfig"
	util "github.com/jessesomerville/ephemeral-iam/cmd/eiam/internal/eiamutil"
	"github.com/jessesomerville/ephemeral-iam/cmd/eiam/internal/gcpclient"
	"github.com/jessesomerville/ephemeral-iam/cmd/eiam/internal/shell"
)

var (
	certCache = make(map[string]*tls.Certificate)
	certLock  = &sync.Mutex{}
)

// StartProxyServer spins up the proxy that replaces the gcloud auth token
func StartProxyServer(privilegedAccessToken *credentialspb.GenerateAccessTokenResponse, reason, svcAcct string) (retErr error) {
	accessToken := privilegedAccessToken.GetAccessToken()
	expirationDate := privilegedAccessToken.GetExpireTime().AsTime()

	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = viper.GetBool("authproxy.verbose")

	if viper.GetBool("authproxy.writetofile") {
		_, err := os.Stat(viper.GetString("authproxy.logdir"))
		if os.IsNotExist(err) {
			if err := os.MkdirAll(viper.GetString("authproxy.logdir"), 0o755); err != nil {
				return fmt.Errorf("Failed to create proxy log directory: %v", err)
			}
		}
		// Create log file
		timestamp := time.Now().Format("20060102150405")
		logFilename := filepath.Join(viper.GetString("authproxy.logdir"), fmt.Sprintf("%s_auth_proxy.log", timestamp))
		logFile, err := os.OpenFile(logFilename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o666)
		if err != nil {
			return fmt.Errorf("Failed to create log file: %v", err)
		}

		// I'm not calling logFile.Close() because it gets invoked too early

		// Set auth proxy to log to file
		proxy.Logger = log.New(logFile, "", log.LstdFlags)
		util.Logger.Infof("Writing auth proxy logs to %s\n", logFilename)
	}

	setCa(appconfig.CertFile, appconfig.KeyFile)

	proxy.OnRequest().HandleConnect(goproxy.FuncHttpsHandler(proxyConnectHandle))

	proxy.OnRequest().DoFunc(func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		r.Header.Set("authorization", fmt.Sprintf("Bearer %s", accessToken))
		r.Header.Set("X-Goog-Request-Reason", reason)
		return r, nil
	})

	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", viper.GetString("authproxy.proxyaddress"), viper.GetString("authproxy.proxyport")),
		Handler: proxy,
	}

	proxyServerExitDone := &sync.WaitGroup{}
	proxyServerExitDone.Add(1)

	// Catch interrupts to gracefully shutdown the proxy and restore the gcloud config
	idleConnsClosed := make(chan struct{})
	sigint := make(chan os.Signal, 1)
	go func() {
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		// TODO: Don't think I'm handling these errors correctly
		// An interrupt signal was received, shutdown the proxy server
		if err := srv.Shutdown(context.Background()); err != nil {
			retErr = fmt.Errorf("Failed to properly shut down proxy server: %v", err)
		}
		close(idleConnsClosed)
		util.Logger.Info("Stopping auth proxy and restoring gcloud config")
		if err := gcpclient.UnsetGcloudProxy(); err != nil {
			retErr = fmt.Errorf("Failed to reset gcloud configuration: %v", err)
		}
		os.Exit(0)
	}()

	sessionLength := time.Until(expirationDate)
	expiresOn := time.Now().Add(sessionLength)

	util.Logger.Info("Starting auth proxy")
	go func() {
		defer proxyServerExitDone.Done()
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			util.Logger.Fatal("Error starting or closing auth proxy")
		}
		<-idleConnsClosed
	}()

	util.Logger.Info("Privileged session will last until ", expiresOn.Format(time.RFC1123))
	util.Logger.Warn("Press CTRL+C to quit privileged session")

	// Drop the user into a interactive shell until the session expires
loop:
	for expired := time.After(sessionLength); ; {
		select {
		case <-expired:
			break loop
		default:
			if err := shell.CommandPrompt(sigint, svcAcct); err != nil {
				return err
			}
		}
	}

	util.Logger.Info("Privileged session expired, stopping auth proxy and restoring gcloud config")
	if err := srv.Shutdown(context.Background()); err != nil {
		retErr = fmt.Errorf("Failed to properly shut down proxy server: %v", err)
	}
	if err := gcpclient.UnsetGcloudProxy(); err != nil {
		retErr = fmt.Errorf("Failed to reset gcloud configuration: %v", err)
	}
	return nil
}

func proxyConnectHandle(host string, ctx *goproxy.ProxyCtx) (*goproxy.ConnectAction, string) {
	return goproxy.MitmConnect, host
}
