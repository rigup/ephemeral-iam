package eiamutil

import "strings"

// CheckError handles simple error handling
func CheckError(err error) {
	if err != nil {
		if strings.Contains(err.Error(), "could not find default credentials") {
			Logger.Fatal("No Application Default Credentials were found. Please run the following command to remediate this issue: \n\n  $ gcloud auth application-default login")
		}
		Logger.Fatalln(err)
	}
}
