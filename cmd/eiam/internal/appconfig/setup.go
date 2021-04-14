package appconfig

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/spf13/viper"
	"google.golang.org/api/oauth2/v1"

	util "github.com/jessesomerville/ephemeral-iam/cmd/eiam/internal/eiamutil"
	errorsutil "github.com/jessesomerville/ephemeral-iam/cmd/eiam/internal/errors"
	"github.com/jessesomerville/ephemeral-iam/cmd/eiam/internal/gcpclient"
)

func init() {
	viper.SetConfigName("config")
	viper.AddConfigPath(GetConfigDir())
	viper.AutomaticEnv()
	viper.SetConfigType("yml")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			initConfig()
		} else {
			fmt.Fprintf(os.Stderr, "failed to read config file %s/config.yml: %v", GetConfigDir(), err)
			os.Exit(1)
		}
	}

	util.Logger = util.NewLogger()

	allConfigKeys := viper.AllKeys()
	if !util.Contains(allConfigKeys, "binarypaths.gcloud") && !util.Contains(allConfigKeys, "binarypaths.kubectl") {
		errorsutil.CheckError(checkDependencies())
	}

	if err := checkValidADCExists(); err != nil {
		util.Logger.WithError(err).Fatal("setup error")
	}

	if err := createLogDir(); err != nil {
		util.Logger.WithError(err).Fatal("setup error")
	}
	if err := createTempKubeConfigDir(); err != nil {
		util.Logger.WithError(err).Fatal("setup error")
	}
}

// checkDependencies ensures that the prequisites for running `eiam` are met
func checkDependencies() error {
	gcloudPath, err := CheckCommandExists("gcloud")
	if err != nil {
		return err
	}
	kubectlPath, err := CheckCommandExists("kubectl")
	if err != nil {
		return err
	}
	viper.Set("binarypaths.gcloud", gcloudPath)
	viper.Set("binarypaths.kubectl", kubectlPath)
	if err := viper.WriteConfig(); err != nil {
		return err
	}
	return nil
}

// CheckCommandExists tries to find the location of a given binary
func CheckCommandExists(command string) (string, error) {
	path, err := exec.LookPath(command)
	if err != nil {
		return "", err
	}
	util.Logger.Debugf("Found binary %s at %s\n", command, path)
	return path, nil
}

// checkValidADCExists checks that application default credentials exist, that
// they are valid, and that they are for the correct user
func checkValidADCExists() error {
	ctx := context.Background()
	oauth2Service, err := oauth2.NewService(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "could not find default credentials") {
			util.Logger.Warn("No Application Default Credentials were found, attempting to generate them\n")

			cmd := exec.Command(viper.GetString("binarypaths.gcloud"), "auth", "application-default", "login", "--no-launch-browser")
			cmd.Stderr = os.Stderr
			cmd.Stdout = os.Stdout
			cmd.Stdin = os.Stdin
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("unable to create application default credentials: %v", err)
			}
			fmt.Println()
			util.Logger.Info("Application default credentials were successfully created")
		} else {
			return fmt.Errorf("failed to check if application default credentials exist: %v", err)
		}
	} else {
		util.Logger.Debug("Checking validity of application default credentials")
		tokenInfo, err := oauth2Service.Tokeninfo().Do()
		if err != nil {
			return fmt.Errorf("failed to parse OAuth2 Token: %v", err)
		}

		return checkADCIdentity(tokenInfo.Email)
	}
	return nil
}

func checkADCIdentity(tokenEmail string) error {
	account, err := gcpclient.CheckActiveAccountSet()
	if err != nil {
		return err
	}

	util.Logger.Debugf("OAuth 2.0 Token Email: %s", tokenEmail)
	if account != tokenEmail {
		util.Logger.Warnf("API calls made by eiam will not be authenticated as your default account:\n\tAccount Set:     %s\n\tDefault Account: %s\n\n", tokenEmail, account)

		prompt := promptui.Prompt{
			Label:     fmt.Sprintf("Authenticate as %s", tokenEmail),
			IsConfirm: true,
		}

		if ans, err := prompt.Run(); err != nil && strings.ToLower(ans) == "y" {
			fmt.Print("\n\n")
			util.Logger.Info("Attempting to reconfigure eiam's authenticated account")
			os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "")
			if err := checkValidADCExists(); err != nil {
				util.Logger.Fatal(err)
			}
			util.Logger.Infof("Success. You should now be authenticated as %s", account)
		}
	}

	return nil
}

func createLogDir() error {
	logDir := viper.GetString("authproxy.logdir")
	_, err := os.Stat(logDir)
	if os.IsNotExist(err) {
		util.Logger.Debugf("Creating log directory: %s", logDir)
		if err := os.MkdirAll(viper.GetString("authproxy.logdir"), 0o755); err != nil {
			return fmt.Errorf("failed to create proxy log directory %s: %v", logDir, err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to find proxy log dir %s: %v", logDir, err)
	}
	return nil
}

func createTempKubeConfigDir() error {
	configDir := GetConfigDir()
	kubeConfigDir := path.Join(configDir, "tmp_kube_config")
	_, err := os.Stat(kubeConfigDir)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(kubeConfigDir, 0o755); err != nil {
			return fmt.Errorf("failed to create temp kubeconfig directory: %v", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to find temp kubeconfig dir %s: %v", kubeConfigDir, err)
	}
	// Clear any leftover kubeconfigs from improper shutdowns
	if err := os.RemoveAll(kubeConfigDir); err != nil {
		return fmt.Errorf("failed to clear old kubeconfigs: %v", err)
	}
	if err := os.MkdirAll(kubeConfigDir, 0o755); err != nil {
		return fmt.Errorf("failed to recreate temp kubeconfig directory: %v", err)
	}
	return nil
}
