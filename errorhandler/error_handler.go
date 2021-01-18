package errorhandler

import (
	"sync"

	"emperror.dev/emperror"
	logrushandler "emperror.dev/handler/logrus"
	"github.com/sirupsen/logrus"
)

var errorHandler emperror.Handler
var once sync.Once

// GetErrorHandler gets the error handler instance
func GetErrorHandler(logger *logrus.Logger) emperror.Handler {
	once.Do(func() {
		errorHandler = newErrorHandler(logger)
	})
	return errorHandler
}

func newErrorHandler(logger *logrus.Logger) emperror.Handler {
	handler := logrushandler.New(logger)
	return handler
}
