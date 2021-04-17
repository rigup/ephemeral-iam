package errors

import (
	"fmt"
	"regexp"
	"strings"

	util "github.com/jessesomerville/ephemeral-iam/internal/eiamutil"
)

var invalidCommandErrMsg = regexp.MustCompile(`unknown command "[\S]+" for "[a-z\-]+"`)

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
