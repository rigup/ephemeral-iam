/*
Copyright © 2021 Jesse Somerville

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
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
	credentialspb "google.golang.org/genproto/googleapis/iam/credentials/v1"

	"github.com/jessesomerville/ephemeral-iam/internal/appconfig"
	"github.com/jessesomerville/ephemeral-iam/internal/gcpclient"
)

var (
	certCache = make(map[string]*tls.Certificate)
	certLock  = &sync.Mutex{}
)

// StartProxyServer spins up the proxy that replaces the gcloud auth token
func StartProxyServer(privilegedAccessToken *credentialspb.GenerateAccessTokenResponse, reason string) (retErr error) {

	accessToken := privilegedAccessToken.GetAccessToken()
	expirationDate := privilegedAccessToken.GetExpireTime().AsTime()

	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = config.AuthProxy.Verbose
	if config.AuthProxy.WriteToFile {
		_, err := os.Stat(config.AuthProxy.LogDir)
		if os.IsNotExist(err) {
			if err := os.MkdirAll(config.AuthProxy.LogDir, 0755); err != nil {
				return fmt.Errorf("Failed to create proxy log directory: %v", err)
			}
		}
		// Create log file
		timestamp := time.Now().Format("20060102150405")
		logFilename := filepath.Join(config.AuthProxy.LogDir, fmt.Sprintf("%s_auth_proxy.log", timestamp))
		logFile, err := os.OpenFile(logFilename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			return fmt.Errorf("Failed to create log file: %v", err)
		}

		// I'm not calling logFile.Close() because it gets invoked too early

		// Set auth proxy to log to file
		proxy.Logger = log.New(logFile, "", log.LstdFlags)
		logger.Infof("Writing auth proxy logs to %s\n", logFilename)
	}

	setCa(appconfig.CertFile, appconfig.KeyFile)

	proxy.OnRequest().HandleConnect(goproxy.FuncHttpsHandler(proxyConnectHandle))

	proxy.OnRequest().DoFunc(func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		r.Header.Set("authorization", fmt.Sprintf("Bearer %s", accessToken))
		r.Header.Set("X-Goog-Request-Reason", reason)
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

		// TODO: Don't think I'm handling these errors correctly
		// An interrupt signal was recieved, shutdown the proxy server
		if err := srv.Shutdown(context.Background()); err != nil {
			retErr = fmt.Errorf("Failed to properly shut down proxy server: %v", err)
		}
		close(idleConnsClosed)
		fmt.Println()
		logger.Info("Stopping auth proxy and restoring gcloud config")
		if err := gcpclient.UnsetGcloudProxy(); err != nil {
			retErr = fmt.Errorf("Failed to reset gcloud configuration: %v", err)
		}
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
