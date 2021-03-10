package appconfig

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/kirsle/configdir"
	"github.com/spf13/viper"

	credentials "cloud.google.com/go/iam/credentials/apiv1"
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

	checkADCExists()
}

func getConfigDir() string {
	configPath := configdir.LocalConfig("ephemeral-iam")

	// Check to ensure that the path is user-specific instead of global
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Failed to get user home directory: %v", err)
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
		log.Fatalf("Failed to get default configuration path: %v", err)
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

func checkADCExists() {
	ctx := context.Background()
	_, err := credentials.NewIamCredentialsClient(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "could not find default credentials") {
			util.Logger.Warn("No Application Default Credentials were found, attempting to generate them")
			util.Logger.Info("A browser window will be opened and prompt you to authenticate to Google. Please follow the instructions in the browser")
			cmd := exec.Command("gcloud", "auth", "application-default", "login")
			if err := cmd.Run(); err != nil {
				util.Logger.Fatal("Unable to create application default credentials, please run `gcloud auth application-default login` to resolve this issue")
			}
			util.Logger.Info("Application default credentials were successfully created")
		} else {
			util.Logger.Fatal("Failed to check if application default credentials exist")
		}
	}
}
