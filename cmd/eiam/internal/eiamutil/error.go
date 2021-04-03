package eiamutil

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/lithammer/dedent"
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
		return fmt.Sprintf("failed to create %s SDK client with service account %s: %s", e.ResourceType, e.ServiceAccount, e.Err)
	}
	return fmt.Sprintf("failed to create %s SDK client: %s", e.ResourceType, e.Err)
}

func CheckRevertGcloudConfigError(err error) {
	if err != nil {
		Logger.WithError(err).Error("failed to revert gcloud configuration")
		Logger.Warn("Please run the following command to manually fix this issue:")
		fmt.Println(dedent.Dedent(`
			gcloud config unset proxy/address \
			&& gcloud config unset proxy/port \
			&& gcloud config unset proxy/type \
			&& gcloud config unset core/custom_ca_certs_file
		`))
	}
}

// CheckError handles simple error handling
func CheckError(err error) {
	if err != nil {
		if strings.Contains(err.Error(), "could not find default credentials") {
			Logger.Fatal("No Application Default Credentials were found. Please run the following command to remediate this issue:\n\n  $ gcloud auth application-default login\n\n")
		} else if invalidCommandErrMsg.MatchString(err.Error()) {
			Logger.Errorf("%v", err)
		} else {
			Logger.WithError(err).Error("eiam crashed due to an unhandled error")
		}
	}
}
