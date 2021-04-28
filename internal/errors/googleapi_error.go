package errors

import (
	"fmt"

	util "github.com/jessesomerville/ephemeral-iam/internal/eiamutil"
	"google.golang.org/api/googleapi"
)

// See https://cloud.google.com/apis/design/errors#handling_errors
var googleErrorCodes = map[int]string{
	400: "Invalid argument",
	401: "Invalid authentication credentials",
	403: "Permission denied",
	404: "Resource not found",
	409: "Resource conflict",
	429: "Quota limit exceeded",
	499: "Request cancelled by client",
	500: "Internal server error",
	501: "Unimplemented method",
	503: "Server unavailable",
	504: "Server deadline exceeded",
}

func checkGoogleAPIError(serr *EiamError) *EiamError {
	err := serr.Err
	if gerr, ok := err.(*googleapi.Error); ok {
		errStatusMsg, ok := googleErrorCodes[gerr.Code]
		if !ok {
			errStatusMsg = "Unknown error"
		}
		errMsg := gerr.Message
		if len(errMsg) == 0 {
			// TODO Check if message can be parsed from body
			errMsg = gerr.Body
		}
		return &EiamError{
			Log: util.Logger.WithField("error", errMsg),
			Msg: fmt.Sprintf("[Google API Error] %s", errStatusMsg),
			Err: err,
		}
	}
	return nil
}
