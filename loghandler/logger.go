package loghandler

import (
	"sync"

	"emperror.dev/emperror"
	"github.com/sirupsen/logrus"

	"github.com/jessesomerville/gcp-iam-escalate/appconfig"
)

var logger *logrus.Logger
var once sync.Once

// GetLogger returns the output log instance
func GetLogger(config *appconfig.LogConfig) *logrus.Logger {
	once.Do(func() {
		logger = newLogger(config)
	})
	return logger
}

func newLogger(config *appconfig.LogConfig) *logrus.Logger {
	logger := logrus.New()

	level, err := logrus.ParseLevel(config.Level)
	emperror.Panic(err)

	logger.Level = level

	switch config.Format {
	case "json":
		logger.Formatter = new(logrus.JSONFormatter)

	default:
		logger.Formatter = &logrus.TextFormatter{
			DisableLevelTruncation: config.DisableLevelTrucation,
			PadLevelText:           config.PadLevelText,
		}
	}

	return logger
}
