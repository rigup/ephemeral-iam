package appconfig

import (
	"os"
	"path/filepath"

	"emperror.dev/emperror"
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
	CertFile = filepath.Join(getConfigDir(), "server.pem")
	KeyFile  = filepath.Join(getConfigDir(), "server.key")
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
			emperror.Panic(err)
		}
	}

	emperror.Panic(viper.Unmarshal(&Config))

	if _, err := os.Stat(CertFile); os.IsNotExist(err) {
		GenerateCerts()
	}
}

func getConfigDir() string {
	configPath := configdir.LocalConfig("ephemeral-iam")
	emperror.Panic(configdir.MakePath(configPath))
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
			emperror.Panic(err)
		}
	}
}

// func generateProxyCerts() {
// 	_, currFile, _, ok := runtime.Caller(0)
// 	if !ok {
// 		emperror.Panic(errors.New("Failed to get package directory: No caller information"))
// 	}

// 	if err := exec.Command(filepath.Join(currFile, "../../scripts/generate_proxy_certs.sh")).Run(); err != nil {
// 		emperror.Panic(err)
// 	}
// }
