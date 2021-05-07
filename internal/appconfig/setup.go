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

package appconfig

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/spf13/viper"
	"google.golang.org/api/oauth2/v1"

	util "github.com/rigup/ephemeral-iam/internal/eiamutil"
	errorsutil "github.com/rigup/ephemeral-iam/internal/errors"
	"github.com/rigup/ephemeral-iam/internal/gcpclient"
)

// Setup ensures that the prequisites for running ephemeral-iam are met.
func Setup() error {
	if err := checkValidADCExists(); err != nil {
		return err
	}
	if err := createLogDir(); err != nil {
		return err
	}
	if err := createTempKubeConfigDir(); err != nil {
		return err
	}
	if err := createPluginDir(); err != nil {
		return err
	}
	return nil
}

// checkValidADCExists checks that application default credentials exist, that
// they are valid, and that they are for the correct user.
func checkValidADCExists() error {
	ctx := context.Background()
	oauth2Service, err := oauth2.NewService(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "could not find default credentials") {
			util.Logger.Warn("No Application Default Credentials were found, attempting to generate them\n")

			gcloud := viper.GetString("binarypaths.gcloud")
			cmd := exec.Command(gcloud, "auth", "application-default", "login", "--no-launch-browser")
			cmd.Stderr = os.Stderr
			cmd.Stdout = os.Stdout
			cmd.Stdin = os.Stdin
			if err = cmd.Run(); err != nil {
				return errorsutil.New("Failed to create application default credentials", err)
			}
			fmt.Println()
			util.Logger.Info("Application default credentials were successfully created")
		} else {
			return errorsutil.New("Failed to check if application default credentials exist", err)
		}
	} else {
		util.Logger.Debug("Checking validity of application default credentials")
		tokenInfo, err := oauth2Service.Tokeninfo().Do()
		if err != nil {
			return errorsutil.New("Failed to parse OAuth token", err)
		}

		return checkADCIdentity(tokenInfo.Email)
	}
	return nil
}

// checkADCIdentity checks the active account set in the users gcloud config
// against the identity associated with the application default credentials.
func checkADCIdentity(tokenEmail string) error {
	account, err := gcpclient.CheckActiveAccountSet()
	if err != nil {
		return err
	}

	util.Logger.Debugf("OAuth 2.0 Token Email: %s", tokenEmail)
	if account != tokenEmail {
		util.Logger.Warnf(`API calls made by eiam will not be authenticated as your default account:
		  Account Set:     %s
		  Default Account: %s`, tokenEmail, account)

		prompt := promptui.Prompt{
			Label:     fmt.Sprintf("Authenticate as %s", tokenEmail),
			IsConfirm: true,
		}

		if ans, err := prompt.Run(); err != nil {
			if strings.EqualFold(ans, "y") {
				fmt.Print("\n\n")
				util.Logger.Info("Attempting to reconfigure eiam's authenticated account")
				os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "")
				if err := checkValidADCExists(); err != nil {
					util.Logger.Fatal(err)
				}
				util.Logger.Infof("Success. You should now be authenticated as %s", account)
			}
		}
	}
	return nil
}

// createLogDir creates the directory to store log files.
func createLogDir() error {
	logDir := viper.GetString(AuthProxyLogDir)
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		util.Logger.Debugf("Creating log directory: %s", logDir)
		if err = os.MkdirAll(viper.GetString(AuthProxyLogDir), 0o755); err != nil {
			return errorsutil.New(fmt.Sprintf("Failed to create proxy log directory %s", logDir), err)
		}
	} else if err != nil {
		return errorsutil.New(fmt.Sprintf("Failed to find proxy log directory %s", logDir), err)
	}
	return nil
}

// createTempKubeConfigDir creates the directory to hold the temporary kubeconfigs
// used in the assume-privileges command.
func createTempKubeConfigDir() error {
	configDir := GetConfigDir()
	kubeConfigDir := path.Join(configDir, "tmp_kube_config")
	_, err := os.Stat(kubeConfigDir)
	if os.IsNotExist(err) {
		if err = os.MkdirAll(kubeConfigDir, 0o755); err != nil {
			return fmt.Errorf("failed to create temp kubeconfig directory: %v", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to find temp kubeconfig dir %s: %v", kubeConfigDir, err)
	}
	// Clear any leftover kubeconfigs from improper shutdowns.
	if err := os.RemoveAll(kubeConfigDir); err != nil {
		return fmt.Errorf("failed to clear old kubeconfigs: %v", err)
	}
	if err := os.MkdirAll(kubeConfigDir, 0o755); err != nil {
		return fmt.Errorf("failed to recreate temp kubeconfig directory: %v", err)
	}
	return nil
}

func createPluginDir() error {
	pluginDir := filepath.Join(GetConfigDir(), "plugins")
	if _, err := os.Stat(pluginDir); os.IsNotExist(err) {
		util.Logger.Debugf("Creating plugin directory: %s", pluginDir)
		if err = os.MkdirAll(pluginDir, 0o755); err != nil {
			return errorsutil.New(fmt.Sprintf("failed to create plugin directory: %s", pluginDir), err)
		}
	} else if err != nil {
		return errorsutil.New(fmt.Sprintf("failed to find plugin directory: %s", pluginDir), err)
	}
	return nil
}
