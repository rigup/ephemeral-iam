package proxy

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/creack/pty"
	"github.com/google/uuid"
	"golang.org/x/term"

	"github.com/jessesomerville/ephemeral-iam/cmd/eiam/internal/appconfig"
	util "github.com/jessesomerville/ephemeral-iam/cmd/eiam/internal/eiamutil"
)

func startShell(svcAcct string, defaultCluster map[string]string, oldState **term.State) {
	tmpKubeConfig, err := createTempFile()
	if err != nil {
		util.Logger.WithError(err).Fatal("failed to create privileged kubeconfig")
	}
	defer os.Remove(tmpKubeConfig.Name()) // Remove tmpKubeConfig after priv session ends

	// Create the kubeconfig entry for the privileged service account
	c := exec.Command("gcloud", "container", "clusters", "get-credentials", defaultCluster["name"], "--zone", defaultCluster["location"])

	// Copy environment variables from user, set PS1 prompt, and set the KUBECONFIG env var
	c.Env = append(c.Env, os.Environ()...)
	c.Env = append(c.Env, buildPrompt(svcAcct))
	c.Env = append(c.Env, fmt.Sprintf("KUBECONFIG=%s", tmpKubeConfig.Name()))

	if err := c.Run(); err != nil {
		util.Logger.WithError(err).Errorf("failed to fetch credentials for cluster %s", defaultCluster["name"])
	} else {
		util.Logger.Infof("kubectl is now authenticated as %s", svcAcct)
	}

	// Create the shell command and copy the environment variables from the previous command
	shellCmd := exec.Command("bash")
	shellCmd.Env = c.Env

	util.Logger.Warn("Enter `exit` or press CTRL+D to quit privileged session")

	// Start the pty sub-shell
	ptmx, err := pty.Start(shellCmd)
	if err != nil {
		util.Logger.WithError(err).Fatal("failed to start privileged sub-shell")
	}
	defer func() {
		if err := ptmx.Close(); err != nil {
			util.Logger.WithError(err).Fatal("failed to close privileged sub-shell")
		}
	}()

	// Resize the pty shell when the user's terminal is resized
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)
	go func() {
		for range ch {
			if err := pty.InheritSize(os.Stdin, ptmx); err != nil {
				util.Logger.WithError(err).Fatal("failed to resize pty")
			}
		}
	}()
	ch <- syscall.SIGWINCH

	// Save the state of the current shell so it can be restored later
	if *oldState, err = term.MakeRaw(int(os.Stdin.Fd())); err != nil {
		util.Logger.WithError(err).Fatal("failed to save state of current shell")
	}
	defer func() {
		if err := term.Restore(int(os.Stdin.Fd()), *oldState); err != nil {
			util.Logger.WithError(err).Fatal("failed to restore original shell")
		}
	}()

	// Send user input to the sub-shell
	go func() {
		if _, err := io.Copy(ptmx, os.Stdin); err != nil {
			util.Logger.WithError(err).Error("failed to send user input to the sub-shell")
		}
	}()

	// Write the output from the sub-shell to stdout
	if _, err := io.Copy(os.Stdout, ptmx); err != nil {
		// On some linux systems, this error is thrown when CTRL-D is received
		if serr, ok := err.(*fs.PathError); ok {
			if serr.Path == "/dev/ptmx" {
				wg.Done()
				return
			}
		} else {
			util.Logger.WithError(err).Error("failed to write the output from the sub-shell to stdout")
		}
	}
	wg.Done()
}

func buildPrompt(svcAcct string) string {
	yellow := "\\[\\e[33m\\]"
	green := "\\[\\e[36m\\]"
	endColor := "\\[\\e[m\\]"
	return fmt.Sprintf("PS1=\n[%s%s%s]\n[%seiam%s] > ", yellow, svcAcct, endColor, green, endColor)
}

func createTempFile() (*os.File, error) {
	configDir := appconfig.GetConfigDir()
	tmpFileName := uuid.New().String()
	tmpKubeConfig, err := os.CreateTemp(configDir, tmpFileName)
	if err != nil {
		return nil, err
	}
	return tmpKubeConfig, nil
}
