package appconfig

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kirsle/configdir"
	"github.com/spf13/viper"

	util "github.com/jessesomerville/ephemeral-iam/cmd/eiam/internal/eiamutil"
)

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

	if _, err := os.Stat(CertFile); os.IsNotExist(err) {
		if err := GenerateCerts(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	util.NewLogger()
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
	viper.SetDefault("authproxy.proxyaddress", "127.0.0.1")
	viper.SetDefault("authproxy.proxyport", "8084")
	viper.SetDefault("authproxy.verbose", false)
	viper.SetDefault("authproxy.writetofile", false)
	viper.SetDefault("authproxy.logdir", filepath.Join(getConfigDir(), "log"))
	viper.SetDefault("logging.format", "text")
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.disableleveltruncation", true)
	viper.SetDefault("logging.padleveltext", true)

	if err := viper.SafeWriteConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileAlreadyExistsError); !ok {
			fmt.Fprintf(os.Stderr, "Failed to write config file %s/config.yml: %v", getConfigDir(), err)
			os.Exit(1)
		}
	}
}
