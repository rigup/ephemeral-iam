package appconfig

import (
	"fmt"
	"os/exec"

	util "github.com/jessesomerville/ephemeral-iam/cmd/eiam/internal/eiamutil"
	"github.com/spf13/viper"
)

func init() {
	allConfigKeys := viper.AllKeys()
	if !util.Contains(allConfigKeys, "binarypaths.gcloud") && !util.Contains(allConfigKeys, "binarypaths.kubectl") {
		util.CheckError(checkDependencies())
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
	viper.WriteConfig()
	return nil
}

// CheckCommandExists tries to find the location of a given binary
func CheckCommandExists(command string) (string, error) {
	path, err := exec.LookPath(command)
	if err != nil {
		return "", fmt.Errorf("Failed to find %s binary path: %v", command, err)
	}
	util.Logger.Debugf("Found binary %s at %s\n", command, path)
	return path, nil
}
