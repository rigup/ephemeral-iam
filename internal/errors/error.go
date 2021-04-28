package errors

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"

	util "github.com/rigup/ephemeral-iam/internal/eiamutil"
)

type EiamError struct {
	Err error
	Log *logrus.Entry
	Msg string
}

func (e EiamError) Error() string {
	if errStr, err := e.Log.String(); err != nil {
		return e.Err.Error()
	} else {
		return errStr
	}
}

// CheckError is the top-level error handler
func CheckError(err error) {
	if err != nil {
		if strings.Contains(err.Error(), "could not find default credentials") {
			util.Logger.Fatal("No Application Default Credentials were found. Please run the following command to remediate this issue:\n\n  $ gcloud auth application-default login\n\n")
		}

		if serr, ok := err.(EiamError); ok {
			if googleErr := checkGoogleAPIError(&serr); googleErr != nil {
				serr = *googleErr
			} else if grpcError := checkGoogleRPCError(&serr); grpcError != nil {
				serr = *grpcError
			}
			serr.Log.Error(serr.Msg)
			util.Logger.Exit(1)
		}
		util.Logger.Fatal(err)
	}
}

// CheckRevertGcloudConfigError prompts the user to manually reset their gcloud
// config if eiam failed to do so.
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
