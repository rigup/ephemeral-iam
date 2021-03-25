package gcpclient

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"strings"
	"sync"

	"github.com/lithammer/dedent"
	"github.com/manifoldco/promptui"
	"github.com/spf13/viper"
	"gopkg.in/ini.v1"

	util "github.com/jessesomerville/ephemeral-iam/cmd/eiam/internal/eiamutil"
)

var (
	once           sync.Once
	gcloudConfig   *ini.File
	pathToConfig   string
	initialProjVal string
)

func readGcloudConfigFromFile() error {
	usr, err := user.Current()
	if err != nil {
		return fmt.Errorf("failed to get current system user: %v", err)
	}
	configDir := path.Join(usr.HomeDir, ".config", "gcloud")

	activeConfig, err := getActiveConfig(configDir)
	if err != nil {
		return err
	}

	configName := fmt.Sprintf("config_%s", activeConfig)
	pathToConfig = path.Join(configDir, "configurations", configName)

	gcloudConfig, err = ini.Load(pathToConfig)
	if err != nil {
		return fmt.Errorf("failed to parse gcloud config %s: %v", pathToConfig, err)
	}
	initialProjVal = gcloudConfig.Section("core").Key("project").String()
	return nil
}

func getActiveConfig(configDir string) (string, error) {
	activeConfigFile := path.Join(configDir, "active_config")
	if _, err := os.Stat(activeConfigFile); os.IsNotExist(err) {
		util.Logger.Warn("No active gcloud config is set. Attempting to set one")
		configurationsDir := path.Join(configDir, "configurations")
		if _, err := os.Stat(configurationsDir); os.IsNotExist(err) {
			if err := os.Mkdir(configurationsDir, 0o755); err != nil {
				return "", err
			}
			defaultConfig := path.Join(configurationsDir, "config_default")
			if _, err := os.Create(defaultConfig); err != nil {
				return "", err
			}
		}

		activeConfig, err := setActiveConfig(configurationsDir, activeConfigFile)
		if err != nil {
			return "", err
		}
		return activeConfig, nil
	}

	configFromFile, err := ioutil.ReadFile(activeConfigFile)
	if err != nil {
		return "", fmt.Errorf("failed to get active gcloud config: %v", err)
	}
	return string(configFromFile), nil
}

func setActiveConfig(configsDir, activeConfigFile string) (string, error) {
	var configName string
	if configs, err := os.ReadDir(configsDir); err != nil {
		return "", fmt.Errorf("failed to get active gcloud config: %v", err)
	} else if len(configs) == 0 {
		return "", errors.New("there are no existing gcloud configurations")
	} else if len(configs) == 1 {
		configName = strings.Split(configs[0].Name(), "_")[1]
	} else {
		var configNames []string
		for _, name := range configs {
			configNames = append(configNames, strings.Split(name.Name(), "_")[1])
		}
		chosenConfig, err := promptForConfigToSet(configNames)
		if err != nil {
			return "", errors.New("failed to select active config")
		}
		configName = chosenConfig
	}

	fd, err := os.Create(activeConfigFile)
	if err != nil {
		return "", fmt.Errorf("failed to set active gcloud config: %v", err)
	}
	defer fd.Close()

	util.Logger.Infof("Setting active gcloud config to %s", configName)
	if _, err := fd.Write([]byte(configName)); err != nil {
		return "", err
	}
	return configName, nil
}

func promptForConfigToSet(configs []string) (string, error) {
	prompt := promptui.Select{
		Label: "Select which config to use",
		Items: configs,
	}

	_, result, err := prompt.Run()
	if err != nil {
		return "", err
	}
	return result, nil
}

func getGcloudConfig() (configErr error) {
	once.Do(func() {
		configErr = readGcloudConfigFromFile()
	})
	return configErr
}

// ConfigureGcloudProxy configures the current gcloud configuration to use the auth proxy
func ConfigureGcloudProxy(project string) error {
	if err := getGcloudConfig(); err != nil {
		return err
	}

	gcloudConfig.Section("proxy").Key("address").SetValue(viper.GetString("authproxy.proxyaddress"))
	gcloudConfig.Section("proxy").Key("port").SetValue(viper.GetString("authproxy.proxyport"))
	gcloudConfig.Section("proxy").Key("type").SetValue("http")
	gcloudConfig.Section("core").Key("custom_ca_certs_file").SetValue(viper.GetString("authproxy.certfile"))
	// If the user specified a project flag, set it in the gcloud config
	if project != "" {
		gcloudConfig.Section("core").Key("project").SetValue(project)
	}
	if err := gcloudConfig.SaveTo(pathToConfig); err != nil {
		return fmt.Errorf("failed to update gcloud configuration: %v", err)
	}
	return nil
}

// UnsetGcloudProxy restores the auth proxy changes made to the gcloud config
func UnsetGcloudProxy() error {
	if err := getGcloudConfig(); err != nil {
		return err
	}

	gcloudConfig.Section("proxy").DeleteKey("address")
	gcloudConfig.Section("proxy").DeleteKey("port")
	gcloudConfig.Section("proxy").DeleteKey("type")
	gcloudConfig.Section("core").DeleteKey("custom_ca_certs_file")
	gcloudConfig.Section("core").Key("project").SetValue(initialProjVal)
	if err := gcloudConfig.SaveTo(pathToConfig); err != nil {
		return fmt.Errorf("failed to revert gcloud configuration: %v", err)
	}
	return nil
}

// CheckActiveAccountSet ensures that the current gcloud config has an active account value
// TODO: This is not currently being used
func CheckActiveAccountSet() (string, error) {
	if err := getGcloudConfig(); err != nil {
		return "", err
	}
	if acct := gcloudConfig.Section("core").Key("account").String(); acct == "" {
		return "", fmt.Errorf(dedent.Dedent(`no active account set for gcloud. please run:
		
		  $ gcloud auth login
		  
		to obtain new credentials.  If you have already logged in with a different account:
		
		  $ gcloud config set account ACCOUNT
		  
	 	to select an already authenticated account to use.`))
	} else {
		return acct, nil
	}
}

// GetCurrentProject get the active project from the gcloud config
func GetCurrentProject() (string, error) {
	if err := getGcloudConfig(); err != nil {
		return "", err
	}
	return gcloudConfig.Section("core").Key("project").String(), nil
}

// GetCurrentRegion get the active region from the gcloud config
func GetCurrentRegion() (string, error) {
	if err := getGcloudConfig(); err != nil {
		return "", err
	}
	return gcloudConfig.Section("compute").Key("region").String(), nil
}

// GetCurrentZone get the active zone from the gcloud config
func GetCurrentZone() (string, error) {
	if err := getGcloudConfig(); err != nil {
		return "", err
	}
	return gcloudConfig.Section("compute").Key("zone").String(), nil
}
