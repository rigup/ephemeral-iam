package proxy

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/creack/pty"
	"github.com/elazarl/goproxy"
	"github.com/google/uuid"
	"github.com/lithammer/dedent"
	"github.com/spf13/viper"
	"golang.org/x/term"
	credentialspb "google.golang.org/genproto/googleapis/iam/credentials/v1"

	"github.com/jessesomerville/ephemeral-iam/cmd/eiam/internal/appconfig"
	util "github.com/jessesomerville/ephemeral-iam/cmd/eiam/internal/eiamutil"
	"github.com/jessesomerville/ephemeral-iam/cmd/eiam/internal/gcpclient"
)

var (
	certCache = make(map[string]*tls.Certificate)
	certLock  = &sync.Mutex{}

	wg sync.WaitGroup
)

// StartProxyServer spins up the proxy that replaces the gcloud auth token
func StartProxyServer(privilegedAccessToken *credentialspb.GenerateAccessTokenResponse, reason, svcAcct string) (retErr error) {
	accessToken := privilegedAccessToken.GetAccessToken()
	expirationDate := privilegedAccessToken.GetExpireTime().AsTime()
	sessionLength := time.Until(expirationDate)

	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = viper.GetBool("authproxy.verbose")

	if viper.GetBool("authproxy.writetofile") {
		_, err := os.Stat(viper.GetString("authproxy.logdir"))
		if os.IsNotExist(err) {
			if err := os.MkdirAll(viper.GetString("authproxy.logdir"), 0o755); err != nil {
				return fmt.Errorf("failed to create proxy log directory: %v", err)
			}
		} else if err != nil {
			return fmt.Errorf("failed to find proxy log dir %s: %v", viper.GetString("authproxy.logdir"), err)
		}
		// Create log file
		timestamp := time.Now().Format("20060102150405")
		logFilename := filepath.Join(viper.GetString("authproxy.logdir"), fmt.Sprintf("%s_auth_proxy.log", timestamp))
		logFile, err := os.OpenFile(logFilename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o666)
		if err != nil {
			return fmt.Errorf("failed to create log file: %v", err)
		}

		// I'm not calling logFile.Close() because it gets invoked too early

		// Set auth proxy to log to file
		proxy.Logger = log.New(logFile, "", log.LstdFlags)
		util.Logger.Infof("Writing auth proxy logs to %s\n", logFilename)
	}

	util.CheckError(setCa(viper.GetString("authproxy.certfile"), viper.GetString("authproxy.keyfile")))

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
			retErr = fmt.Errorf("failed to properly shut down proxy server: %v", err)
		}
		close(idleConnsClosed)
		util.Logger.Info("Stopping auth proxy and restoring gcloud config")
		if err := gcpclient.UnsetGcloudProxy(); err != nil {
			retErr = fmt.Errorf("failed to reset gcloud configuration: %v", err)
		}
		os.Exit(0)
	}()

	util.Logger.Info("Starting auth proxy")
	util.Logger.Infof("Privileged session will last until %s", time.Now().Add(sessionLength).Format(time.RFC1123))
	util.Logger.Warn("Enter `exit` or press CTRL+D to quit privileged session")

	go func() {
		defer proxyServerExitDone.Done()
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			util.Logger.Fatal("Error starting or closing auth proxy")
		}
		<-idleConnsClosed
	}()

	tmpKubeConfig, err := createTempKubeconfig()
	if err != nil {
		return err
	}
	defer os.Remove(tmpKubeConfig.Name()) // Remove tmpKubeConfig after priv session ends

	wg.Add(1)
	var oldState *term.State
	go func() {
		c := exec.Command("bash")
		c.Env = append(c.Env, os.Environ()...)
		c.Env = append(c.Env, buildPrompt(svcAcct))
		c.Env = append(c.Env, fmt.Sprintf("KUBECONFIG=%s", tmpKubeConfig.Name()))

		ptmx, err := pty.Start(c)
		if err != nil {
			util.Logger.Fatalf("Failed to start privileged sub-shell: %v", err)
		}
		defer func() {
			if err := ptmx.Close(); err != nil {
				util.Logger.Fatalf("Failed to close sub-shell file descriptor: %v", err)
			}
		}()

		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGWINCH)
		go func() {
			for range ch {
				if err := pty.InheritSize(os.Stdin, ptmx); err != nil {
					util.Logger.Fatalf("Error resizing pty: %s", err)
				}
			}
		}()
		ch <- syscall.SIGWINCH

		oldState, err = term.MakeRaw(int(os.Stdin.Fd()))
		if err != nil {
			util.Logger.Fatalf("Failed to set sub-shell to raw mode: %v", err)
		}
		defer func() {
			if err := term.Restore(int(os.Stdin.Fd()), oldState); err != nil {
				util.Logger.Fatalf("Failed to restore original shell: %v", err)
			}
		}()

		go func() {
			if _, err := io.Copy(ptmx, os.Stdin); err != nil {
				util.Logger.Errorf("failed to copy stdin to the pty: %v", err)
			}
		}()
		if _, err := io.Copy(os.Stdout, ptmx); err != nil {
			// On some linux systems, this error is thrown when CTRL-D is received
			if serr, ok := err.(*fs.PathError); ok {
				if serr.Path == "/dev/ptmx" {
					wg.Done()
					return
				}
			} else {
				util.Logger.Errorf("failed to copy the pty to stdout: %v", err)
			}
		}
		wg.Done()
	}()

	// Shut down the auth proxy when the user exits the sub-shell
	go func() {
		wg.Wait()
		sigint <- syscall.SIGINT
	}()

	time.Sleep(sessionLength)

	if err := term.Restore(int(os.Stdin.Fd()), oldState); err != nil {
		util.Logger.Fatalf("Failed to restore original shell: %v", err)
	}
	fmt.Println()

	util.Logger.Info("Privileged session expired, stopping auth proxy and restoring gcloud config")
	if err := srv.Shutdown(context.Background()); err != nil {
		return fmt.Errorf("failed to properly shut down proxy server: %v", err)
	}
	if err := gcpclient.UnsetGcloudProxy(); err != nil {
		util.Logger.Warn("Failed to revert gcloud configuration! Please run the following command to manually fix this issue:")
		fmt.Println(dedent.Dedent(`
			gcloud config unset proxy/address \
			  && gcloud config unset proxy/port \
			  && gcloud config unset proxy/type \
			  && gcloud config unset core/custom_ca_certs_file
		`))
		retErr = fmt.Errorf("failed to reset gcloud configuration: %v", err)
	}
	return nil
}

func proxyConnectHandle(host string, ctx *goproxy.ProxyCtx) (*goproxy.ConnectAction, string) {
	return goproxy.MitmConnect, host
}

func buildPrompt(svcAcct string) string {
	yellow := "\\[\\e[33m\\]"
	green := "\\[\\e[36m\\]"
	endColor := "\\[\\e[m\\]"
	return fmt.Sprintf("PS1=\n[%s%s%s]\n[%seiam%s] > ", yellow, svcAcct, endColor, green, endColor)
}

func createTempKubeconfig() (*os.File, error) {
	configDir := appconfig.GetConfigDir()
	tmpFileName := uuid.New().String()
	tmpKubeConfig, err := os.CreateTemp(configDir, tmpFileName)
	if err != nil {
		return nil, fmt.Errorf("failed to create tmp kubeconfig: %v", err)
	}
	return tmpKubeConfig, nil
}
