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
	LogDir       string
	Verbose      bool
	WriteToFile  bool
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
