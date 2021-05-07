// Copyright 2021 Workrise Technologies Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	"syscall"
	"time"

	"github.com/elazarl/goproxy"
	"github.com/spf13/viper"
	"golang.org/x/term"

	"github.com/rigup/ephemeral-iam/internal/appconfig"
	util "github.com/rigup/ephemeral-iam/internal/eiamutil"
	errorsutil "github.com/rigup/ephemeral-iam/internal/errors"
	"github.com/rigup/ephemeral-iam/internal/gcpclient"
)

var (
	certCache = make(map[string]*tls.Certificate)
	certLock  = &sync.Mutex{}

	wg sync.WaitGroup

	funcHTTPSHandler = func(host string, ctx *goproxy.ProxyCtx) (*goproxy.ConnectAction, string) {
		return goproxy.MitmConnect, host
	}
)

// StartProxyServer spins up the proxy that replaces the gcloud auth token.
func StartProxyServer(
	accessToken,
	reason,
	svcAcct,
	project string,
	expirationDate time.Time,
	defaultCluster map[string]string,
) error {
	if err := checkProxyCertificate(); err != nil {
		return err
	}

	srv, err := createProxy(accessToken, reason)
	if err != nil {
		return err
	}

	// Catch interrupts to gracefully shutdown the proxy and restore the gcloud config.
	idleConnsClosed := make(chan struct{})
	sigint := make(chan os.Signal, 1)
	go func() {
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		// An interrupt signal was received, shutdown the proxy server.
		if err := srv.Shutdown(context.Background()); err != nil {
			util.Logger.WithError(err).Error("failed to properly shut down proxy server")
		}
		close(idleConnsClosed)
		util.Logger.Info("Stopping auth proxy and restoring gcloud config")
		errorsutil.CheckRevertGcloudConfigError(gcpclient.UnsetGcloudProxy())
		os.Exit(0)
	}()

	proxyServerExit := &sync.WaitGroup{}
	proxyServerExit.Add(1)
	go func() {
		defer proxyServerExit.Done()
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			util.Logger.WithError(err).Fatal("failed to start the auth proxy")
		}
		<-idleConnsClosed
	}()

	sessionLength := time.Until(expirationDate)
	sessionEnd := time.Now().Add(sessionLength).Format(time.RFC1123)
	util.Logger.Infof("Starting auth proxy. Privileged session will last until %s", sessionEnd)

	wg.Add(1)
	var oldState *term.State
	// TODO: Instead of handling errors in the startShell function, handle them here.
	go startShell(svcAcct, accessToken, expirationDate.Format(time.RFC3339Nano), defaultCluster, &oldState)

	// Shut down the auth proxy when the user exits the sub-shell.
	go func() {
		wg.Wait()
		sigint <- syscall.SIGINT
	}()

	time.Sleep(sessionLength)

	if err := term.Restore(int(os.Stdin.Fd()), oldState); err != nil {
		return errorsutil.New("Failed to restore original shell", err)
	}

	util.Logger.Info("Privileged session expired, stopping auth proxy and restoring gcloud config")
	if err := srv.Shutdown(context.Background()); err != nil {
		return errorsutil.New("Failed to properly shut down proxy server", err)
	}
	errorsutil.CheckRevertGcloudConfigError(gcpclient.UnsetGcloudProxy())
	return nil
}

func createProxy(accessToken, reason string) (*http.Server, error) {
	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = viper.GetBool(appconfig.AuthProxyVerbose)

	// Create log file.
	timestamp := time.Now().Format("20060102150405")
	logFilename := filepath.Join(viper.GetString(appconfig.AuthProxyLogDir), fmt.Sprintf("%s_auth_proxy.log", timestamp))
	logFile, err := os.OpenFile(logFilename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o666)
	if err != nil {
		return nil, errorsutil.New("Failed to create log file", err)
	}
	defer logFile.Close()

	// Set auth proxy to log to file.
	proxy.Logger = log.New(logFile, "", log.LstdFlags)
	util.Logger.Infof("Writing auth proxy logs to %s\n", logFilename)

	certFile, keyFile := viper.GetString(appconfig.AuthProxyCertFile), viper.GetString(appconfig.AuthProxyKeyFile)
	if err := setCa(certFile, keyFile); err != nil {
		util.Logger.Error("Failed to set proxy certificate authority")
		return nil, err
	}

	proxy.OnRequest().HandleConnect(goproxy.FuncHttpsHandler(funcHTTPSHandler))

	proxy.OnRequest().DoFunc(func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		r.Header.Set("authorization", fmt.Sprintf("Bearer %s", accessToken))
		r.Header.Set("X-Goog-Request-Reason", reason)
		return r, nil
	})

	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", viper.GetString(appconfig.AuthProxyAddress), viper.GetString(appconfig.AuthProxyPort)),
		Handler: proxy,
	}
	return srv, nil
}
