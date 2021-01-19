package proxy

import (
	"emperror.dev/emperror"
	"github.com/jessesomerville/gcp-iam-escalate/appconfig"
	"github.com/jessesomerville/gcp-iam-escalate/errorhandler"
	"github.com/jessesomerville/gcp-iam-escalate/loghandler"
	"github.com/sirupsen/logrus"
)

var config = &appconfig.Config

var logger *logrus.Logger
var errorHandler emperror.Handler

func init() {
	logger = loghandler.GetLogger(&config.Logging)
	errorHandler = errorhandler.GetErrorHandler(logger)
}
