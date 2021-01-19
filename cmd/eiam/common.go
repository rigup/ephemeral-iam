package main

import (
	"emperror.dev/emperror"
	"github.com/sirupsen/logrus"

	"github.com/jessesomerville/ephemeral-iam/internal/appconfig"
	"github.com/jessesomerville/ephemeral-iam/internal/errorhandler"
	"github.com/jessesomerville/ephemeral-iam/internal/loghandler"
)

var config = &appconfig.Config

var logger *logrus.Logger
var errorHandler emperror.Handler

func init() {
	logger = loghandler.GetLogger(&config.Logging)
	errorHandler = errorhandler.GetErrorHandler(logger)
}
