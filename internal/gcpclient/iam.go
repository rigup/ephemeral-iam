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
	"sync"

	"github.com/golang/protobuf/ptypes/duration"
	"google.golang.org/api/iam/v1"
	credentialspb "google.golang.org/genproto/googleapis/iam/credentials/v1"

	util "github.com/rigup/ephemeral-iam/internal/eiamutil"
	errorsutil "github.com/rigup/ephemeral-iam/internal/errors"
)

var (
	sessionDuration int64 = 600
	ctx                   = context.Background()

	wg sync.WaitGroup
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

// CanImpersonate checks if a given service account can be impersonated by the
// authenticated user.
func CanImpersonate(project, serviceAccountEmail string) (bool, error) {
	resource := fmt.Sprintf("//iam.googleapis.com/projects/%s/serviceAccounts/%s", project, serviceAccountEmail)
	testablePerms, err := QueryTestablePermissionsOnResource(resource)
	if err != nil {
		return false, err
	}

	perms, err := QueryServiceAccountPermissions(testablePerms, project, serviceAccountEmail)
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

// FetchAvailableServiceAccounts gets a list of service accounts that the user can impersonate.
func FetchAvailableServiceAccounts(project string) ([]*iam.ServiceAccount, error) {
	util.Logger.Infof("Using current project: %s", project)

	serviceAccounts, err := getServiceAccounts(project)
	if err != nil {
		return nil, err
	}
	util.Logger.Infof("Checking %d service accounts in %s", len(serviceAccounts), project)

	wg.Add(len(serviceAccounts))

	var availableSAs []*iam.ServiceAccount
	for _, svcAcct := range serviceAccounts {
		go func(serviceAccount *iam.ServiceAccount) {
			hasAccess, err := CanImpersonate(project, serviceAccount.Email)
			if err != nil {
				util.Logger.Errorf("error checking IAM permissions: %v", err)
			} else if hasAccess {
				availableSAs = append(availableSAs, serviceAccount)
			}
			wg.Done()
		}(svcAcct)
	}
	wg.Wait()

	return availableSAs, nil
}

func getServiceAccounts(project string) ([]*iam.ServiceAccount, error) {
	iamService, err := iam.NewService(context.Background())
	if err != nil {
		return nil, errorsutil.NewSDKError("Cloud IAM", "", err)
	}
	svcAcctClient := iam.NewProjectsServiceAccountsService(iamService)

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
