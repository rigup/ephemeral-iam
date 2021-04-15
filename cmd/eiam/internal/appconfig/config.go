package appconfig

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/kirsle/configdir"
	"github.com/spf13/viper"
)

var (
	configDir string
	once      sync.Once
)

// GetConfigDir returns the directory to use for the ephemeral-iam configurations
func GetConfigDir() string {
	once.Do(func() {
		configDir = getConfigDir()
	})
	return configDir
}

func getConfigDir() string {
	configPath := configdir.LocalConfig("ephemeral-iam")

	// Check to ensure that the path is user-specific instead of global
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("failed to get user home directory: %v", err)
	}
	if !strings.HasPrefix(configPath, userHomeDir) {
		if runtime.GOOS == "linux" {
			configPath = path.Join(userHomeDir, ".config/ephemeral-iam")
		} else if runtime.GOOS == "darwin" {
			configPath = path.Join(userHomeDir, configPath)
		} else {
			log.Fatalf("%s is not a recognized OS. Supported OS are 'linux' and 'darwin'", runtime.GOOS)
		}
	}

	if err := configdir.MakePath(configPath); err != nil {
		log.Fatalf("failed to get default configuration path: %v", err)
	}
	return configPath
}

func initConfig() {
	viper.SetDefault("authproxy.proxyaddress", "127.0.0.1")
	viper.SetDefault("authproxy.proxyport", "8084")
	viper.SetDefault("authproxy.verbose", false)
	viper.SetDefault("authproxy.logdir", filepath.Join(GetConfigDir(), "log"))
	viper.SetDefault("authproxy.certfile", filepath.Join(GetConfigDir(), "server.pem"))
	viper.SetDefault("authproxy.keyfile", filepath.Join(GetConfigDir(), "server.key"))
	viper.SetDefault("logging.format", "text")
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.disableleveltruncation", true)
	viper.SetDefault("logging.padleveltext", true)

	if err := viper.SafeWriteConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileAlreadyExistsError); !ok {
			fmt.Fprintf(os.Stderr, "failed to write config file %s/config.yml: %v", GetConfigDir(), err)
			os.Exit(1)
		}
	}
}
