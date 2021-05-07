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

	crm "google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/iam/v1"
	"google.golang.org/api/option"
	"google.golang.org/api/pubsub/v1"
	"google.golang.org/api/storage/v1"

	util "github.com/rigup/ephemeral-iam/internal/eiamutil"
	errorsutil "github.com/rigup/ephemeral-iam/internal/errors"
)

var (
	ctx = context.Background()

	wg sync.WaitGroup
)

// QueryTestablePermissionsOnResource gets the testable permissions on a resource
// Modified from https://github.com/salrashid123/gcp_iam/blob/main/query/main.go#L71-L108
func QueryTestablePermissionsOnResource(resource string) ([]string, error) {
	iamService, err := iam.NewService(ctx)
	if err != nil {
		return []string{}, errorsutil.NewSDKError("Cloud IAM", "", err)
	}
	permissionsService := iam.NewPermissionsService(iamService)

	util.Logger.Debugf("Fetching testable permissions on %s\n", resource)

	var permsToTest []string
	nextPageToken := ""
	for {
		ps, err := permissionsService.QueryTestablePermissions(&iam.QueryTestablePermissionsRequest{
			FullResourceName: resource,
			PageToken:        nextPageToken,
			PageSize:         1000,
		}).Do()
		if err != nil {
			return []string{}, errorsutil.New(fmt.Sprintf("Failed to get testable permissions for %s", resource), err)
		}

		for _, perm := range ps.Permissions {
			permsToTest = append(permsToTest, perm.Name)
		}

		nextPageToken = ps.NextPageToken
		if nextPageToken == "" {
			break
		}
	}
	return permsToTest, nil
}

// QueryComputeInstancePermissions gets the authenticated members permissions on a compute instance
// Modified from https://github.com/salrashid123/gcp_iam/blob/main/query/main.go#L351-L371
func QueryComputeInstancePermissions(
	permsToTest []string,
	project,
	zone,
	instance,
	svcAcct,
	reason string,
) ([]string, error) {
	var computeService *compute.Service
	if svcAcct != "" {
		clientOptions := []option.ClientOption{
			option.ImpersonateCredentials(svcAcct),
			option.WithRequestReason(reason),
		}
		if svc, err := compute.NewService(ctx, clientOptions...); err == nil {
			computeService = svc
		} else {
			return []string{}, errorsutil.NewSDKError("Compute", svcAcct, err)
		}
	} else {
		if svc, err := compute.NewService(ctx); err == nil {
			computeService = svc
		} else {
			return []string{}, errorsutil.NewSDKError("Compute", "", err)
		}
	}

	permsToTest = remove(permsToTest, []string{
		"resourcemanager.resourceTagBindings.create",
		"resourcemanager.resourceTagBindings.delete",
		"resourcemanager.resourceTagBindings.list",
	})

	resp, err := computeService.Instances.TestIamPermissions(project, zone, instance, &compute.TestPermissionsRequest{
		Permissions: permsToTest,
	}).Do()
	if err != nil {
		return []string{}, errorsutil.EiamError{
			Log: util.Logger.WithError(err),
			Msg: fmt.Sprintf(
				"Failed to query permissions on resource projects/%s/zones/%s/instances/%s",
				project,
				zone,
				instance,
			),
			Err: err,
		}
	}

	return resp.Permissions, nil
}

// QueryProjectPermissions gets the authenticated members permissions on a project
// Modified from https://github.com/salrashid123/gcp_iam/blob/main/query/main.go#L534-L575
func QueryProjectPermissions(permsToTest []string, project, svcAcct, reason string) (perms []string, err error) {
	var crmService *crm.Service
	if svcAcct != "" {
		clientOptions := []option.ClientOption{
			option.ImpersonateCredentials(svcAcct),
			option.WithRequestReason(reason),
		}
		if svc, err := crm.NewService(ctx, clientOptions...); err == nil {
			crmService = svc
		} else {
			return []string{}, errorsutil.NewSDKError("Cloud Resource Manager", svcAcct, err)
		}
	} else {
		if svc, err := crm.NewService(ctx); err == nil {
			crmService = svc
		} else {
			return []string{}, errorsutil.NewSDKError("Cloud Resource Manager", "", err)
		}
	}
	crmProjService := crm.NewProjectsService(crmService)

	// TestIamPermissions accepts a max of 100 permissions at a time so we split them into chunks.
	var chunked [][]string
	numOfChunks := len(permsToTest) / 100
	for i := 0; i < numOfChunks; i++ {
		start := i * 100
		end := start + 100
		chunked = append(chunked, permsToTest[start:end])
	}
	rem := len(permsToTest) % 100
	chunked = append(chunked, permsToTest[len(permsToTest)-rem:])

	wg.Add(len(chunked))

	var userPermissions []string
	for _, permSet := range chunked {
		go func(permissions []string, granted *[]string) {
			resp, err := crmProjService.TestIamPermissions(project, &crm.TestIamPermissionsRequest{
				Permissions: permissions,
			}).Do()
			if err != nil {
				util.Logger.Errorf("Failed to query permissions on projects/%s", project)
				return
			}
			*granted = append(*granted, resp.Permissions...)
			wg.Done()
		}(permSet, &userPermissions)
	}
	// Wait until each of the go routines have finished before returning.
	wg.Wait()

	return userPermissions, nil
}

// QueryPubSubPermissions gets the authenticated members permissions on a PubSub topic.
func QueryPubSubPermissions(permsToTest []string, project, topic, svcAcct, reason string) ([]string, error) {
	var pubsubService *pubsub.Service
	if svcAcct != "" {
		clientOptions := []option.ClientOption{
			option.ImpersonateCredentials(svcAcct),
			option.WithRequestReason(reason),
		}
		if svc, err := pubsub.NewService(ctx, clientOptions...); err == nil {
			pubsubService = svc
		} else {
			return []string{}, errorsutil.NewSDKError("PubSub", svcAcct, err)
		}
	} else {
		if svc, err := pubsub.NewService(ctx); err == nil {
			pubsubService = svc
		} else {
			return []string{}, errorsutil.NewSDKError("PubSub", "", err)
		}
	}

	topicsService := pubsub.NewProjectsTopicsService(pubsubService)

	resource := fmt.Sprintf("projects/%s/topics/%s", project, topic)
	resp, err := topicsService.TestIamPermissions(resource, &pubsub.TestIamPermissionsRequest{
		Permissions: permsToTest,
	}).Do()
	if err != nil {
		return []string{}, errorsutil.New(fmt.Sprintf("Failed to query permissions on %s", resource), err)
	}

	return resp.Permissions, nil
}

// QueryServiceAccountPermissions gets the authenticated members permissions on a service account
// Modified from https://github.com/salrashid123/gcp_iam/blob/main/query/main.go#L150-L173
func QueryServiceAccountPermissions(permsToTest []string, project, email string) ([]string, error) {
	iamService, err := iam.NewService(ctx)
	if err != nil {
		return []string{}, errorsutil.NewSDKError("Cloud IAM", "", err)
	}
	saIamService := iam.NewProjectsServiceAccountsService(iamService)

	resource := fmt.Sprintf("projects/%s/serviceAccounts/%s", project, email)
	resp, err := saIamService.TestIamPermissions(resource, &iam.TestIamPermissionsRequest{
		Permissions: permsToTest,
	}).Do()
	if err != nil {
		return []string{}, err
	}

	return resp.Permissions, nil
}

// QueryStorageBucketPermissions gets the authenticated members permissions on a storage bucket
// Modified from https://github.com/salrashid123/gcp_iam/blob/main/query/main.go#L313-L338
func QueryStorageBucketPermissions(permsToTest []string, bucket, svcAcct, reason string) ([]string, error) {
	var storageService *storage.Service
	if svcAcct != "" {
		clientOptions := []option.ClientOption{
			option.ImpersonateCredentials(svcAcct),
			option.WithRequestReason(reason),
		}
		if svc, err := storage.NewService(ctx, clientOptions...); err == nil {
			storageService = svc
		} else {
			return []string{}, errorsutil.NewSDKError("Cloud Storage", svcAcct, err)
		}
	} else {
		if svc, err := storage.NewService(ctx); err == nil {
			storageService = svc
		} else {
			return []string{}, errorsutil.NewSDKError("Cloud Storage", "", err)
		}
	}

	permsToTest = remove(permsToTest, []string{
		"resourcemanager.resourceTagBindings.create",
		"resourcemanager.resourceTagBindings.delete",
		"resourcemanager.resourceTagBindings.list",
	})

	resp, err := storageService.Buckets.TestIamPermissions(bucket, permsToTest).Do()
	if err != nil {
		return []string{}, errorsutil.New(fmt.Sprintf("Failed to query permissions on storage bucket %s", bucket), err)
	}
	return resp.Permissions, nil
}

func remove(perms, remove []string) []string {
	rmap := make(map[string]struct{}, len(remove))
	for _, perm := range remove {
		rmap[perm] = struct{}{}
	}

	n := 0
	for _, perm := range perms {
		if _, found := rmap[perm]; !found {
			perms[n] = perm
			n++
		}
	}
	return perms[:n]
}
