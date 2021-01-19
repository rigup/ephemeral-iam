package gcpclient

import (
	"context"
	"fmt"

	credentials "cloud.google.com/go/iam/credentials/apiv1"
	"emperror.dev/errors"
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
		return nil, errors.WrapIfWithDetails(err, "Failed to generate GCP access token for service account", "service account", serviceAccountEmail)
	}
	return resp, nil
}

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
		return []*iam.ServiceAccount{}, errors.WrapIf(err, "An error occured while fetching service accounts")
	}
	return serviceAccounts, nil
}

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
		return nil, errors.WrapIf(err, "Failed to create IAM DefaultClient")
	}

	iamService, err := iam.New(c)
	if err != nil {
		return nil, errors.WrapIf(err, "Failed to create IAM API Service")
	}
	return iamService, nil
}
