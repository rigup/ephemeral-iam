package appconfig

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	credentials "cloud.google.com/go/iam/credentials/apiv1"
	"github.com/spf13/viper"

	util "github.com/jessesomerville/ephemeral-iam/cmd/eiam/internal/eiamutil"
)

func init() {
	viper.SetConfigName("config")
	viper.AddConfigPath(getConfigDir())
	viper.AutomaticEnv()
	viper.SetConfigType("yml")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			initConfig()
		} else {
			fmt.Fprintf(os.Stderr, "Failed to read config file %s/config.yml: %v", getConfigDir(), err)
			os.Exit(1)
		}
	}

	util.NewLogger()

	allConfigKeys := viper.AllKeys()
	if !util.Contains(allConfigKeys, "binarypaths.gcloud") && !util.Contains(allConfigKeys, "binarypaths.kubectl") {
		util.CheckError(checkDependencies())
	}
	// gcpclient.CheckActiveAccountSet()
	checkADCExists()

	if _, err := os.Stat(CertFile); os.IsNotExist(err) {
		if err := GenerateCerts(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}

// CheckDependencies ensures that the prequisites for running `eiam` are met
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
	}
}
