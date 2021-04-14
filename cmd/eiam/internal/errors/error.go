package errors

import (
	"fmt"
	"regexp"
	"strings"

	"google.golang.org/api/googleapi"

	util "github.com/jessesomerville/ephemeral-iam/cmd/eiam/internal/eiamutil"
)

var (
	invalidCommandErrMsg = regexp.MustCompile(`unknown command "[\S]+" for "[a-z\-]+"`)

	googleErrorCodes = map[int]string{
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
)

// SDKClientCreateError is used for errors caused when attempting to
// create a GCP SDK client/service.
type SDKClientCreateError struct {
	Err            error
	ResourceType   string
	ServiceAccount string
}

func (e *SDKClientCreateError) Error() string {
	if e.ServiceAccount != "" {
		return fmt.Sprintf("failed to create %s SDK client with service account %s: %v", e.ResourceType, e.ServiceAccount, e.Err)
	}
	return fmt.Sprintf("failed to create %s SDK client: %v", e.ResourceType, e.Err)
}

func CheckRevertGcloudConfigError(err error) {
	if err != nil {
		util.Logger.WithError(err).Error("failed to revert gcloud configuration")
		util.Logger.Warn("Please run the following command to manually fix this issue:")
		fmt.Println(`
    gcloud config unset proxy/address \
	  && gcloud config unset proxy/port \
	  && gcloud config unset proxy/type \
	  && gcloud config unset core/custom_ca_certs_file
		`)
	}
}

// CheckError handles simple error handling
func CheckError(err error) {
	if err != nil {
		if checkGoogleAPIError(err) {
			return
		} else if checkGoogleRPCError(err) {
			return
		}

		if strings.Contains(err.Error(), "could not find default credentials") {
			util.Logger.Fatal("No Application Default Credentials were found. Please run the following command to remediate this issue:\n\n  $ gcloud auth application-default login\n\n")
		} else if invalidCommandErrMsg.MatchString(err.Error()) {
			util.Logger.Errorf("%v", err)
		} else {
			util.Logger.WithError(err).Error("An error occurred")
		}
	}
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
