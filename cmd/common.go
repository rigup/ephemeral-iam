package cmd

import (
	"emperror.dev/emperror"
	"github.com/sirupsen/logrus"

	"github.com/jessesomerville/gcp-iam-escalate/appconfig"
	"github.com/jessesomerville/gcp-iam-escalate/errorhandler"
	"github.com/jessesomerville/gcp-iam-escalate/loghandler"
)

var config = &appconfig.Config

var logger *logrus.Logger
var errorHandler emperror.Handler

func init() {
	logger = loghandler.GetLogger(&config.Logging)
	errorHandler = errorhandler.GetErrorHandler(logger)
}
