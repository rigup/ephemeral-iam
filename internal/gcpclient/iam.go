/*
Copyright © 2021 Jesse Somerville

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package gcpclient

import (
	"context"
	"fmt"

	credentials "cloud.google.com/go/iam/credentials/apiv1"
	"github.com/golang/protobuf/ptypes/duration"
	"golang.org/x/oauth2/google"
	gcp "google.golang.org/api/container/v1"
	"google.golang.org/api/iam/v1"
	credentialspb "google.golang.org/genproto/googleapis/iam/credentials/v1"
)

var sessionDuration int64 = 600

// GenerateTemporaryAccessToken generates short-lived credentials for the given service account
func GenerateTemporaryAccessToken(serviceAccountEmail string, client *credentials.IamCredentialsClient) (*credentialspb.GenerateAccessTokenResponse, error) {

	sessionDuration := &duration.Duration{
		Seconds: sessionDuration, // Expire after 10 minutes
	}

	req := credentialspb.GenerateAccessTokenRequest{
		Name:     fmt.Sprintf("projects/-/serviceAccounts/%s", serviceAccountEmail),
		Lifetime: sessionDuration,
		Scope: []string{
			gcp.CloudPlatformScope,
			"https://www.googleapis.com/auth/userinfo.email",
		},
	}

	ctx := context.Background()
	resp, err := client.GenerateAccessToken(ctx, &req)
	if err != nil {
		return nil, fmt.Errorf("Failed to generate GCP access token for service account %s: %v", serviceAccountEmail, err)
	}
	return resp, nil
}

// GetServiceAccounts fetches each of the service accounts that the authenticated
// user can impersonate in the active project.
func GetServiceAccounts(project string) ([]*iam.ServiceAccount, error) {
	ctx := context.Background()

	iamService, err := getIamService()
	if err != nil {
		return []*iam.ServiceAccount{}, err
	}

	projectResource := fmt.Sprintf("projects/%s", project)
	req := iamService.Projects.ServiceAccounts.List(projectResource)

	var serviceAccounts []*iam.ServiceAccount

	if err := req.Pages(ctx, func(page *iam.ListServiceAccountsResponse) error {
		for _, serviceAccount := range page.Accounts {
			serviceAccounts = append(serviceAccounts, serviceAccount)
		}
		return nil
	}); err != nil {
		return []*iam.ServiceAccount{}, fmt.Errorf("An error occured while fetching service accounts: %v", err)
	}
	return serviceAccounts, nil
}

// CanImpersonate checks if a given service account can be impersonated by the
// authenticated user.
func CanImpersonate(project, serviceAccountEmail string) (bool, error) {

	permissions := &iam.TestIamPermissionsRequest{
		Permissions: []string{"iam.serviceAccounts.getAccessToken"},
	}

	iamService, err := getIamService()
	if err != nil {
		return false, err
	}

	saResource := fmt.Sprintf("projects/%s/serviceAccounts/%s", project, serviceAccountEmail)

	testIamPermReq := iamService.Projects.ServiceAccounts.TestIamPermissions(saResource, permissions)

	resp, err := testIamPermReq.Do()
	if err != nil {
		return false, err
	}

	for _, permission := range resp.Permissions {
		if permission == "iam.serviceAccounts.getAccessToken" {
			return true, nil
		}
	}
	return false, nil
}

func getIamService() (*iam.Service, error) {
	ctx := context.Background()

	c, err := google.DefaultClient(ctx, iam.CloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("Failed to create IAM DefaultClient: %v", err)
	}

	iamService, err := iam.New(c)
	if err != nil {
		return nil, fmt.Errorf("Failed to create IAM API Service: %v", err)
	}
	return iamService, nil
}
