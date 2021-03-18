package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/lithammer/dedent"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	util "github.com/jessesomerville/ephemeral-iam/cmd/eiam/internal/eiamutil"
)

var (
	LoggingLevels    = []string{"trace", "debug", "info", "warn", "error", "fatal", "panic"}
	LoggingFormats   = []string{"text", "json"}
	BoolConfigFields = []string{
		"authproxy.verbose",
		"authproxy.writetofile",
		"logging.disableleveltruncation",
		"logging.padleveltext",
	}
)

var configInfo = dedent.Dedent(`
		┏━━━━━━━━━━━━━━━━━━━━━━━━┳━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
		┃ Key                    ┃ Description                                         ┃
		┡━━━━━━━━━━━━━━━━━━━━━━━━╇━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┩
		│ authproxy.proxyaddress │ The address to the auth proxy. You shouldn't need   │
		│                        │ to update this                                      │
		├────────────────────────┼─────────────────────────────────────────────────────┤
		│ authproxy.proxyport    │ The port to run the auth proxy on                   │
		├────────────────────────┼─────────────────────────────────────────────────────┤
		│ authproxy.verbose      │ Enables verbose logging output from the auth proxy  │
		├────────────────────────┼─────────────────────────────────────────────────────┤
		│ authproxy.writetofile  │ Enables writing auth proxy logs to a log file       │
		├────────────────────────┼─────────────────────────────────────────────────────┤
		│ authproxy.logdir       │ The directory to write auth proxy logs to           │
		├────────────────────────┼─────────────────────────────────────────────────────┤
		│ binarypaths.gcloud     │ The absolute path to the gcloud binary              │
		├────────────────────────┼─────────────────────────────────────────────────────┤
		│ binarypaths.kubectl    │ The absolute path to the kubectl binary             │
		├────────────────────────┼─────────────────────────────────────────────────────┤
		│ logging.format         │ The format for the console logs.                    │
		│                        │ Can be either 'json' or 'text'                      │
		├────────────────────────┼─────────────────────────────────────────────────────┤
		│ logging.level          │ The logging level to write to the console.          │
		│                        │ Can be one of "trace", "debug", "info", "warn",     │
		│                        │ "error", "fatal", "panic"                           │
		└────────────────────────┴─────────────────────────────────────────────────────┘
`)

func newCmdConfig() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration values",
	}

	cmd.AddCommand(newCmdConfigPrint())
	cmd.AddCommand(newCmdConfigView())
	cmd.AddCommand(newCmdConfigSet())
	cmd.AddCommand(newCmdConfigInfo())

	return cmd
}

func newCmdConfigPrint() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "print",
		Short: "Print the current configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			configFile := viper.ConfigFileUsed()
			data, err := ioutil.ReadFile(configFile)
			if err != nil {
				return err
			}
			fmt.Printf("\n%s\n", string(data))
			return nil
		},
	}
	return cmd
}

func newCmdConfigInfo() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info",
		Short: "Print information about config fields",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(configInfo)
		},
	}
	return cmd
}

func newCmdConfigView() *cobra.Command {
	cmd := &cobra.Command{
		Use:       "view",
		Short:     "View the value of a provided config item",
		Args:      cobra.ExactValidArgs(1),
		ValidArgs: viper.AllKeys(),
		Run: func(cmd *cobra.Command, args []string) {
			val := viper.Get(args[0])
			util.Logger.Infof("%s: %v\n", args[0], val)
		},
	}
	return cmd
}

func newCmdConfigSet() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set",
		Short: "Set the value of a provided config item",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return errors.New("requires both a config key and a new value")
			}

			if !util.Contains(viper.AllKeys(), args[0]) {
				return fmt.Errorf("invalid config key %s", args[0])
			}

			if args[0] == "logging.level" {
				if !util.Contains(LoggingLevels, args[1]) {
					return fmt.Errorf("logging level must be one of %v", LoggingLevels)
				}
			} else if args[0] == "logging.format" {
				if !util.Contains(LoggingFormats, args[1]) {
					return fmt.Errorf("logging format must be one of %v", LoggingFormats)
				}
			} else if util.Contains(BoolConfigFields, args[0]) {
				_, err := strconv.ParseBool(args[1])
				if err != nil {
					return fmt.Errorf("the %s value must be either true or false", args[0])
				}
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			oldVal := viper.Get(args[0])

			if util.Contains(BoolConfigFields, args[0]) {
				newValue, _ := strconv.ParseBool(args[1])
				viper.Set(args[0], newValue)
			} else {
				viper.Set(args[0], args[1])
			}
			switch args[0] {
			case "logging.level":
				level, err := logrus.ParseLevel(args[1])
				if err != nil {
					return err
				}
				util.Logger.Level = level
			case "logging.format":
				switch args[1] {
				case "json":
					util.Logger.Formatter = new(logrus.JSONFormatter)

				default:
					util.Logger.Formatter = &logrus.TextFormatter{
						DisableLevelTruncation: viper.GetBool("logging.disableleveltruncation"),
						PadLevelText:           viper.GetBool("logging.padleveltext"),
						DisableTimestamp:       true,
					}
				}

			}
			util.CheckError(viper.WriteConfig())
			util.Logger.Infof("Updated %s from %v to %s", args[0], oldVal, args[1])
			return nil
		},
	}
	return cmd
}
