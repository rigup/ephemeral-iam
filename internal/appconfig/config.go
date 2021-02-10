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
package appconfig

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kirsle/configdir"
	"github.com/spf13/viper"
)

// Configuration is the top level struct for the config file
type Configuration struct {
	AuthProxy ProxyConfig
	Logging   LogConfig
}

// ProxyConfig is the struct for the auth proxy configuration
type ProxyConfig struct {
	ProxyAddress string
	ProxyPort    string
	Verbose      bool
	WriteToFile  bool
	LogDir       string
}

// LogConfig is the struct for the logging configuration
type LogConfig struct {
	Format                string
	Level                 string
	DisableLevelTrucation bool
	PadLevelText          bool
}

// Config is the global configuration instance
var Config Configuration

var (
	// CertFile is the filepath pointing to the TLS cert
	CertFile = filepath.Join(getConfigDir(), "server.pem")
	// KeyFile is the filepath pointing to the TLS key
	KeyFile = filepath.Join(getConfigDir(), "server.key")
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

	if err := viper.Unmarshal(&Config); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to unmarshal config %s/config.yml: %v", getConfigDir(), err)
		os.Exit(1)
	}

	if _, err := os.Stat(CertFile); os.IsNotExist(err) {
		if err := GenerateCerts(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}

func getConfigDir() string {
	configPath := configdir.LocalConfig("ephemeral-iam")
	if err := configdir.MakePath(configPath); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get default configuration path: %v", err)
		os.Exit(1)
	}
	return configPath
}

func initConfig() {

	viper.SetDefault("AuthProxy.ProxyAddress", "127.0.0.1")
	viper.SetDefault("AuthProxy.ProxyPort", "8084")
	viper.SetDefault("AuthProxy.Verbose", false)
	viper.SetDefault("AuthProxy.WriteToFile", false)
	viper.SetDefault("AuthProxy.LogDir", filepath.Join(getConfigDir(), "log"))

	viper.SetDefault("Logging.Format", "text")
	viper.SetDefault("Logging.Level", "info")
	viper.SetDefault("Logging.DisableLevelTruncation", true)
	viper.SetDefault("Logging.PadLevelText", true)

	if err := viper.SafeWriteConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileAlreadyExistsError); !ok {
			fmt.Fprintf(os.Stderr, "Failed to write config file %s/config.yml: %v", getConfigDir(), err)
			os.Exit(1)
		}
	}
}
