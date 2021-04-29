// Copyright 2021 Workrise Technologies Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package errors

import (
	"fmt"

	"github.com/sirupsen/logrus"

	util "github.com/rigup/ephemeral-iam/internal/eiamutil"
)

// EiamError represents a generic ephemeral-iam error.
type EiamError struct {
	Err error
	Log *logrus.Entry
	Msg string
}

func (e EiamError) Error() string {
	errStr, err := e.Log.String()
	if err != nil {
		return e.Err.Error()
	}
	return errStr
}

// CheckError is the top-level error handler.
func CheckError(err error) {
	if err != nil {
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
