/*
Copyright © 2021 Jesse Somerville <jssomerville2@gmail.com>

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
package cmd

import (
	"strconv"

	"emperror.dev/emperror"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	proxyAddress        string
	proxyPort           string
	verboseProxyLog     string
	writeProxyLogToFile string
	proxylogDir         string
	logFormat           string
	logLevel            string
)

// editConfigCmd represents the editConfig command
var editConfigCmd = &cobra.Command{
	Use:   "editConfig",
	Short: "Edit configuration values",
	Long: `
	Example:
		gcp-iam-escalate editConfig --writeProxyLogToFile true`,
	Run: func(cmd *cobra.Command, args []string) {
		verbose, err := strconv.ParseBool(verboseProxyLog)
		emperror.Panic(err)
		writeToFile, err := strconv.ParseBool(writeProxyLogToFile)
		emperror.Panic(err)

		viper.Set("AuthProxy.ProxyAddress", proxyAddress)
		viper.Set("AuthProxy.ProxyPort", proxyPort)
		viper.Set("AuthProxy.Verbose", verbose)
		viper.Set("AuthProxy.WriteToFile", writeToFile)
		viper.Set("AuthProxy.LogDir", proxylogDir)

		viper.Set("Logging.Format", logFormat)
		viper.Set("Logging.Level", logLevel)

		viper.WriteConfig()

		logger.Info("Config successfully updated")
	},
}

func init() {
	rootCmd.AddCommand(editConfigCmd)
	editConfigCmd.Flags().StringVar(&proxyAddress, "proxyAddress", config.AuthProxy.ProxyAddress, "The address to the auth proxy. You shouldn't need to update this")
	editConfigCmd.Flags().StringVar(&proxyPort, "proxyPort", config.AuthProxy.ProxyPort, "The port to run the auth proxy on")
	editConfigCmd.Flags().StringVar(&verboseProxyLog, "verboseProxyLog", strconv.FormatBool(config.AuthProxy.Verbose), "Enables verbose logging output from the auth proxy")
	editConfigCmd.Flags().StringVar(&writeProxyLogToFile, "writeProxyLogToFile", strconv.FormatBool(config.AuthProxy.WriteToFile), "Enables writing auth proxy logs to a log file")
	editConfigCmd.Flags().StringVar(&proxylogDir, "proxylogDir", config.AuthProxy.LogDir, "The directory to write auth proxy logs to")
	editConfigCmd.Flags().StringVar(&logFormat, "logFormat", config.Logging.Format, "The format for the console logs. Can be either 'json' or 'text'")
	editConfigCmd.Flags().StringVar(&logLevel, "logLevel", config.Logging.Level, "The logging level to write to the console")
}
