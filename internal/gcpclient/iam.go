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

package gcpclient

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/ptypes/duration"
	"google.golang.org/api/iam/v1"
	"google.golang.org/api/option"
	credentialspb "google.golang.org/genproto/googleapis/iam/credentials/v1"

	util "github.com/rigup/ephemeral-iam/internal/eiamutil"
	errorsutil "github.com/rigup/ephemeral-iam/internal/errors"
	queryiam "github.com/rigup/ephemeral-iam/internal/gcpclient/query_iam"
)

var (
	sessionDuration int64 = 600
	ctx                   = context.Background()
)

// GenerateTemporaryAccessToken generates short-lived credentials for the given service account.
func GenerateTemporaryAccessToken(svcAcct, reason string) (*credentialspb.GenerateAccessTokenResponse, error) {
	client, err := ClientWithReason(reason)
	if err != nil {
		return nil, err
	}

	sessionDuration := &duration.Duration{
		Seconds: sessionDuration, // Expire after 10 minutes.
	}

	req := credentialspb.GenerateAccessTokenRequest{
		Name:     fmt.Sprintf("projects/-/serviceAccounts/%s", svcAcct),
		Lifetime: sessionDuration,
		Scope: []string{
			iam.CloudPlatformScope,
			"https://www.googleapis.com/auth/userinfo.email",
		},
	}

	resp, err := client.GenerateAccessToken(ctx, &req)
	if err != nil {
		util.Logger.Errorf("Failed to generate GCP access token for service account %s", svcAcct)
		return nil, err
	}
	return resp, nil
}

// GetServiceAccounts fetches each of the service accounts that the authenticated
// user can impersonate in the active project.
func GetServiceAccounts(project, reason string) ([]*iam.ServiceAccount, error) {
	svcAcctClient, err := newServiceAccountClient(reason)
	if err != nil {
		return nil, err
	}

	projectResource := fmt.Sprintf("projects/%s", project)
	req := svcAcctClient.List(projectResource)

	var serviceAccounts []*iam.ServiceAccount

	if err := req.Pages(ctx, func(page *iam.ListServiceAccountsResponse) error {
		serviceAccounts = append(serviceAccounts, page.Accounts...)
		return nil
	}); err != nil {
		util.Logger.Error("Failed to list service accounts")
		return []*iam.ServiceAccount{}, err
	}
	return serviceAccounts, nil
}

// CanImpersonate checks if a given service account can be impersonated by the
// authenticated user.
func CanImpersonate(project, serviceAccountEmail, reason string) (bool, error) {
	resource := fmt.Sprintf("//iam.googleapis.com/projects/%s/serviceAccounts/%s", project, serviceAccountEmail)
	testablePerms, err := queryiam.QueryTestablePermissionsOnResource(resource)
	if err != nil {
		return false, err
	}

	perms, err := queryiam.QueryServiceAccountPermissions(testablePerms, project, serviceAccountEmail)
	if err != nil {
		return false, err
	}

	for _, permission := range perms {
		if permission == "iam.serviceAccounts.getAccessToken" {
			return true, nil
		}
	}
	return false, nil
}

func newServiceAccountClient(reason string) (*iam.ProjectsServiceAccountsService, error) {
	iamService, err := iam.NewService(context.Background(), option.WithRequestReason(reason))
	if err != nil {
		return nil, &errorsutil.SDKClientCreateError{Err: err, ResourceType: "Cloud IAM"}
	}

	return iam.NewProjectsServiceAccountsService(iamService), nil
}
