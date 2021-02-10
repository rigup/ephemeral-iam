/*
Copyright © 2021 Jesse Somerville

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package gcpclient

import (
	"os/exec"
	"strings"

	"github.com/jessesomerville/ephemeral-iam/internal/appconfig"
)

// ConfigureGcloudProxy configures the current gcloud configuration to use the auth proxy
func ConfigureGcloudProxy() error {

	if err := exec.Command("gcloud", "config", "set", "proxy/address", config.AuthProxy.ProxyAddress).Run(); err != nil {
		return err
	}
	if err := exec.Command("gcloud", "config", "set", "proxy/port", config.AuthProxy.ProxyPort).Run(); err != nil {
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

// GetCurrentProject gets the active GCP project from the gcloud config
func GetCurrentProject() (string, error) {
	out, err := exec.Command("gcloud", "config", "get-value", "project").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
