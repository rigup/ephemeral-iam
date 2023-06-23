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
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"syscall"

	"github.com/creack/pty"
	"github.com/google/uuid"
	"golang.org/x/term"
	"k8s.io/apimachinery/pkg/runtime"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	clientcmdapilatest "k8s.io/client-go/tools/clientcmd/api/latest"

	"github.com/rigup/ephemeral-iam/internal/appconfig"
	util "github.com/rigup/ephemeral-iam/internal/eiamutil"
	errorsutil "github.com/rigup/ephemeral-iam/internal/errors"
)

func startShell(svcAcct, accessToken, expiry string, defaultCluster map[string]string, oldState **term.State) {
	tmpKubeConfig, err := createTempKubeConfig()
	if err != nil {
		util.Logger.WithError(err).Fatal("failed to create temp kubeconfig")
	}
	defer os.Remove(tmpKubeConfig.Name()) // Remove tmpKubeConfig after priv session ends.

	// Copy environment variables from user, set PS1 prompt, and set the KUBECONFIG env var.
	cmdEnv := append(
		os.Environ(),
		buildPrompt(svcAcct),
		fmt.Sprintf("KUBECONFIG=%s", tmpKubeConfig.Name()),
		// The Terraform provider can source this config from this environment variable.
		fmt.Sprintf("KUBE_CONFIG_PATH=%s", tmpKubeConfig.Name()),
		// The Terraform provider can source this and use it as the access token.
		fmt.Sprintf("GOOGLE_OAUTH_ACCESS_TOKEN=%s", accessToken),
	)

	if len(defaultCluster) > 0 {
		// Create the kubeconfig entry for the privileged service account.
		c := exec.Command( //nolint:gosec  // This would just get you code exec on your own computer
			"gcloud", "container", "clusters", "get-credentials", defaultCluster["name"],
			"--zone", defaultCluster["location"],
		)
		c.Env = cmdEnv
		errOut := bytes.Buffer{}
		c.Stderr = &errOut

		if err = c.Run(); err != nil {
			util.Logger.Errorf(errOut.String())
		} else {
			util.Logger.Infof("kubectl is now authenticated as %s", svcAcct)
		}
	}

	if err = writeCredsToKubeConfig(tmpKubeConfig, accessToken, expiry); err != nil {
		util.Logger.WithError(err).Fatal("failed to write credentials to temp kubeconfig")
	}

	// Create the shell command and copy the environment variables from the previous command.
	shellCmd := exec.Command("bash")
	shellCmd.Env = cmdEnv

	util.Logger.Warn("Enter `exit` or press CTRL+D to quit privileged session")

	// Start the pty sub-shell.
	ptmx, err := pty.Start(shellCmd)
	if err != nil {
		util.Logger.WithError(err).Fatal("failed to start privileged sub-shell")
	}
	defer func() {
		if err = ptmx.Close(); err != nil {
			util.Logger.WithError(err).Fatal("failed to close privileged sub-shell")
		}
	}()

	// Resize the pty shell when the user's terminal is resized.
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)
	go func() {
		for range ch {
			if err = pty.InheritSize(os.Stdin, ptmx); err != nil {
				util.Logger.WithError(err).Fatal("failed to resize pty")
			}
		}
	}()
	ch <- syscall.SIGWINCH

	// Save the state of the current shell so it can be restored later.
	if *oldState, err = term.MakeRaw(int(os.Stdin.Fd())); err != nil {
		util.Logger.WithError(err).Fatal("failed to save state of current shell")
	}
	defer func() {
		if err := term.Restore(int(os.Stdin.Fd()), *oldState); err != nil {
			util.Logger.WithError(err).Fatal("failed to restore original shell")
		}
	}()

	// Send user input to the sub-shell.
	go func() {
		if _, err := io.Copy(ptmx, os.Stdin); err != nil {
			util.Logger.WithError(err).Error("failed to send user input to the sub-shell")
		}
	}()

	// Write the output from the sub-shell to stdout.
	if _, err := io.Copy(os.Stdout, ptmx); err != nil {
		// On some linux systems, this error is thrown when CTRL-D is received.
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

func createTempKubeConfig() (*os.File, error) {
	kubeConfigDir := path.Join(appconfig.GetConfigDir(), "tmp_kube_config")
	tmpFileName := uuid.New().String()
	tmpKubeConfig, err := os.CreateTemp(kubeConfigDir, tmpFileName)
	if err != nil {
		return nil, err
	}
	return tmpKubeConfig, nil
}

func writeCredsToKubeConfig(tmpKubeConfig *os.File, accessToken, expiry string) error {
	// Read the tmpKubeConfig into a client-go config object.
	config := clientcmdapi.NewConfig()
	configBytes, err := ioutil.ReadFile(tmpKubeConfig.Name())
	if err != nil {
		return errorsutil.New("Failed to read generated tmp kubeconfig", err)
	}
	if err = runtime.DecodeInto(clientcmdapilatest.Codec, configBytes, config); err != nil {
		return errorsutil.New("Failed to deserialize generated tmp kubeconfig", err)
	}

	// There should only be one, this is an efficient way of getting it.
	for _, authInfo := range config.AuthInfos {
		// Write the service account's token to the temp kubeconfig.
		authInfo.AuthProvider.Config["access-token"] = accessToken
		authInfo.AuthProvider.Config["expiry"] = expiry
	}

	// Serialize the updated config and write it back to the file.
	newConfigBytes, err := runtime.Encode(clientcmdapilatest.Codec, config)
	if err != nil {
		return errorsutil.New("Failed to serialize updated tmp kubeconfig", err)
	}
	if _, err := tmpKubeConfig.Write(newConfigBytes); err != nil {
		return errorsutil.New("Failed to write updated tmp kubeconfig", err)
	}
	return nil
}
