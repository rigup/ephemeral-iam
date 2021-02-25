package eiamutil

import (
	"fmt"
	"os/exec"

	"github.com/spf13/viper"
)

// CheckDependencies ensures that the prequisites for running `eiam` are met
func CheckDependencies() error {
	gcloudPath, err := checkCommandExists("gcloud")
	if err != nil {
		return err
	}
	kubectlPath, err := checkCommandExists("kubectl")
	if err != nil {
		return err
	}
	viper.Set("binarypaths.gcloud", gcloudPath)
	viper.Set("binarypaths.kubectl", kubectlPath)
	return nil
}

func checkCommandExists(command string) (string, error) {
	path, err := exec.LookPath(command)
	if err != nil {
		return "", fmt.Errorf("Failed to find %s binary path: %v", command, err)
	}
	Logger.Debugf("Found binary %s at %s\n", command, path)
	return path, nil
}
