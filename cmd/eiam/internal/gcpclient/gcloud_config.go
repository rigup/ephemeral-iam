package gcpclient

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/jessesomerville/ephemeral-iam/cmd/eiam/internal/appconfig"
)

// ConfigureGcloudProxy configures the current gcloud configuration to use the auth proxy
func ConfigureGcloudProxy() error {
	if err := exec.Command("gcloud", "config", "set", "proxy/address", appconfig.Config.AuthProxy.ProxyAddress).Run(); err != nil {
		return err
	}
	if err := exec.Command("gcloud", "config", "set", "proxy/port", appconfig.Config.AuthProxy.ProxyPort).Run(); err != nil {
		return err
	}
	if err := exec.Command("gcloud", "config", "set", "proxy/type", "http").Run(); err != nil {
		return err
	}
	if err := exec.Command("gcloud", "config", "set", "core/custom_ca_certs_file", appconfig.CertFile).Run(); err != nil {
		return err
	}
	return nil
}

// UnsetGcloudProxy restores the auth proxy changes made to the gcloud config
func UnsetGcloudProxy() error {
	if err := exec.Command("gcloud", "config", "unset", "proxy/address").Run(); err != nil {
		return err
	}
	if err := exec.Command("gcloud", "config", "unset", "proxy/port").Run(); err != nil {
		return err
	}
	if err := exec.Command("gcloud", "config", "unset", "proxy/type").Run(); err != nil {
		return err
	}
	if err := exec.Command("gcloud", "config", "unset", "core/custom_ca_certs_file").Run(); err != nil {
		return err
	}
	return nil
}

// GetCurrentProject get the active project from the gcloud config
func GetCurrentProject() (string, error) {
	out, err := exec.Command("gcloud", "config", "get-value", "project").Output()
	if err != nil {
		return "", fmt.Errorf("Failed to get active project from config: %v", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// GetCurrentRegion get the active region from the gcloud config
func GetCurrentRegion() (string, error) {
	out, err := exec.Command("gcloud", "config", "get-value", "compute/region").Output()
	if err != nil {
		return "", fmt.Errorf("Failed to get active region config: %v", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// GetCurrentZone get the active zone from the gcloud config
func GetCurrentZone() (string, error) {
	out, err := exec.Command("gcloud", "config", "get-value", "compute/zone").Output()
	if err != nil {
		return "", fmt.Errorf("Failed to get active zone config: %v", err)
	}
	return strings.TrimSpace(string(out)), nil
}
