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
		"logging.disableleveltruncation",
		"logging.padleveltext",
	}
)

var configInfo = dedent.Dedent(`
		┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┳━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
		┃ Key                            ┃ Description                                 ┃
		┡━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━╇━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┩
		│ authproxy.certfile             │ The path to the auth proxy's TLS            │
		│                                │ certificate                                 │
		├────────────────────────────────┼─────────────────────────────────────────────┤
		│ authproxy.keyfile              │ The path to the auth proxy's x509 key       │
		├────────────────────────────────┼─────────────────────────────────────────────┤
		│ authproxy.logdir               │ The directory that auth proxy logs will be  │
		│                                │ written to                                  │
		├────────────────────────────────┼─────────────────────────────────────────────┤
		│ authproxy.proxyaddress         │ The address that the auth proxy is hosted   │
		│                                │ on                                          │
		├────────────────────────────────┼─────────────────────────────────────────────┤
		│ authproxy.proxyport            │ The port that the auth proxy runs on        │
		├────────────────────────────────┼─────────────────────────────────────────────┤
		│ authproxy.verbose              │ When set to 'true', verbose output for      │
		│                                │ proxy logs will be enabled                  │
		├────────────────────────────────┼─────────────────────────────────────────────┤
		│ binarypaths.gcloud             │ The path to the gcloud binary on your       │
		│                                │ filesystem                                  │
		├────────────────────────────────┼─────────────────────────────────────────────┤
		│ binarypaths.kubectl            │ The path to the kubectl binary on your      │
		│                                │ filesystem                                  │
		├────────────────────────────────┼─────────────────────────────────────────────┤
		│ logging.format                 │ The format for which to write console logs  │
		│                                │ Can be either 'json' or 'text'              │
		├────────────────────────────────┼─────────────────────────────────────────────┤
		│ logging.level                  │ The logging level to write to the console   │
		│                                │ Can be one of 'trace', 'debug', 'info',     │
		│                                │ 'warn', 'error', 'fatal', or 'panic'        │
		├────────────────────────────────┼─────────────────────────────────────────────┤
		│ logging.disableleveltruncation │ When set to 'true', the level indicator for │
		│                                │ logs will not be trucated                   │
		├────────────────────────────────┼─────────────────────────────────────────────┤
		│ logging.padleveltext           │ When set to 'true', output logs will align  │
		│                                │ evenly with their output level indicator    │
		└────────────────────────────────┴─────────────────────────────────────────────┘
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
