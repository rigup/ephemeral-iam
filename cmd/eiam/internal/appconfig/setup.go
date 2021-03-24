package appconfig

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	credentials "cloud.google.com/go/iam/credentials/apiv1"
	"github.com/manifoldco/promptui"
	"github.com/spf13/viper"

	util "github.com/jessesomerville/ephemeral-iam/cmd/eiam/internal/eiamutil"
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
			fmt.Fprintf(os.Stderr, "Failed to read config file %s/config.yml: %v", GetConfigDir(), err)
			os.Exit(1)
		}
	}

	util.NewLogger()

	allConfigKeys := viper.AllKeys()
	if !util.Contains(allConfigKeys, "binarypaths.gcloud") && !util.Contains(allConfigKeys, "binarypaths.kubectl") {
		util.CheckError(checkDependencies())
	}
	checkADCExists()

	if err := checkProxyCertificate(); err != nil {
		util.Logger.Fatal(err)
	}

	if err := createLogDir(); err != nil {
		util.Logger.Fatal(err)
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

func checkADCExists() {
	ctx := context.Background()
	_, err := credentials.NewIamCredentialsClient(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "could not find default credentials") {
			util.Logger.Warn("No Application Default Credentials were found, attempting to generate them\n")

			cmd := exec.Command(viper.GetString("binarypaths.gcloud"), "auth", "application-default", "login", "--no-launch-browser")
			cmd.Stderr = os.Stderr
			cmd.Stdout = os.Stdout
			cmd.Stdin = os.Stdin
			if err := cmd.Run(); err != nil {
				util.Logger.Fatal("Unable to create application default credentials. Please run the following command to remediate this issue: \n\n  $ gcloud auth application-default login\n\n")
			}
			fmt.Println()
			util.Logger.Info("Application default credentials were successfully created")
		} else {
			fmt.Println()
			util.Logger.Fatalf("Failed to check if application default credentials exist: %v", err)
		}
	} else if adcPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"); adcPath != "" {
		util.Logger.Warnf("The GOOGLE_APPLICATION_CREDENTIALS environment variable is set:\n\tADC Path: %s\n\n", adcPath)
		if err := checkADCIdentity(adcPath); err != nil {
			util.Logger.Fatal(err)
		}
	}
}

func checkADCIdentity(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read ADC file %s: %v", path, err)
	}

	var adcMap map[string]interface{}
	if err := json.Unmarshal(data, &adcMap); err != nil {
		return fmt.Errorf("failed to unmarshal ADC file %s: %v", path, err)
	}

	if email, ok := adcMap["client_email"]; ok {
		account, err := gcpclient.CheckActiveAccountSet()
		if err != nil {
			return fmt.Errorf("failed to get account from gcloud config: %v", err)
		}
		util.Logger.Warnf("API calls made by eiam will not be authenticated as your default account:\n\tAccount Set:     %s\n\tDefault Account: %s\n\n", email, account)

		prompt := promptui.Prompt{
			Label:     fmt.Sprintf("Authenticate as %s", email),
			IsConfirm: true,
		}

		if _, err := prompt.Run(); err != nil {
			fmt.Print("\n\n")
			util.Logger.Info("Attempting to reconfigure eiam's authenticated account")
			os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "")
			checkADCExists()
			util.Logger.Infof("Success. You should now be authenticated as %s", account)
		}
	}

	return nil
}

func checkProxyCertificate() error {
	certFile := viper.GetString("authproxy.certfile")
	keyFile := viper.GetString("authproxy.keyfile")
	if certFile == "" || keyFile == "" {
		if keyFile == "" {
			util.Logger.Debug("Setting authproxy.keyfile")
			viper.Set("authproxy.keyfile", filepath.Join(GetConfigDir(), "server.key"))
			keyFile = viper.GetString("authproxy.keyfile")
		}
		if certFile == "" {
			util.Logger.Debug("Setting authproxy.certfile")
			viper.Set("authproxy.certfile", filepath.Join(GetConfigDir(), "server.pem"))
			certFile = viper.GetString("authproxy.certfile")
		}
		if err := viper.WriteConfig(); err != nil {
			return err
		}
	}

	_, certErr := os.Stat(certFile)
	_, keyErr := os.Stat(keyFile)
	certExists := !os.IsNotExist(certErr)
	keyExists := !os.IsNotExist(keyErr)

	if certExists != keyExists { // Check if only one of either the key or the cert exist
		util.Logger.Warn("Either the auth proxy cert or key is missing. Regenerating both")
		if err := GenerateCerts(); err != nil {
			return err
		}
	} else if !certExists { // Check if neither files exist
		if err := GenerateCerts(); err != nil {
			return err
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
