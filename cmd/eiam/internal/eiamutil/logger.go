package eiamutil

import (
	"fmt"
	"os"

	rt "github.com/banzaicloud/logrus-runtime-formatter"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Logger is the global logging instance
var Logger *logrus.Logger

// NewLogger instantiates a new logging instance
func NewLogger() *logrus.Logger {
	logger := logrus.New()

	level, err := logrus.ParseLevel(viper.GetString("logging.level"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create logger instance: %v", err)
		os.Exit(1)
	}

	logger.Level = level
	logger.Out = os.Stderr

	switch viper.GetString("logging.format") {
	case "json":
		logger.Formatter = new(logrus.JSONFormatter)
	// The 'debug' formatter will include the filename, function, and line number
	// that a log entry is written from
	case "debug":
		logger.Formatter = &rt.Formatter{
			ChildFormatter: &logrus.TextFormatter{
				DisableLevelTruncation: viper.GetBool("logging.disableleveltruncation"),
				DisableQuote:           true,
				DisableTimestamp:       true,
				PadLevelText:           viper.GetBool("logging.padleveltext"),
			},
			Line: true,
		}
	default:
		logger.Formatter = &logrus.TextFormatter{
			DisableLevelTruncation: viper.GetBool("logging.disableleveltruncation"),
			DisableQuote:           true,
			DisableTimestamp:       true,
			PadLevelText:           viper.GetBool("logging.padleveltext"),
		}
	}

	return logger
}
