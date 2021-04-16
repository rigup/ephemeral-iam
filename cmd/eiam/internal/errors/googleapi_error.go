package errors

import (
	util "github.com/jessesomerville/ephemeral-iam/cmd/eiam/internal/eiamutil"
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

func checkGoogleAPIError(err error) bool {
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
		util.Logger.Errorf("[Google API Error] %s: %s", errStatusMsg, errMsg)
		return true
	}
	return false
}
