package gcpclient

import (
	"fmt"

	"golang.org/x/net/context"
	crm "google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/iam/v1"
	"google.golang.org/api/option"
	"google.golang.org/api/pubsub/v1"
	"google.golang.org/api/storage/v1"

	util "github.com/jessesomerville/ephemeral-iam/cmd/eiam/internal/eiamutil"
)

var ctx = context.Background()

// QueryTestablePermissionsOnResource gets the testable permissions on a resource
// Modified from https://github.com/salrashid123/gcp_iam/blob/main/query/main.go#L71-L108
func QueryTestablePermissionsOnResource(resource string) ([]string, error) {
	iamService, err := iam.NewService(ctx)
	if err != nil {
		return []string{}, fmt.Errorf("Failed to create Cloud IAM SDK Client: %v", err)
	}
	permissionsService := iam.NewPermissionsService(iamService)

	util.Logger.Debugf("Fetching testable permissions on %s\n", resource)

	var permsToTest []string
	nextPageToken := ""
	for {
		ps, err := permissionsService.QueryTestablePermissions(&iam.QueryTestablePermissionsRequest{
			FullResourceName: resource,
			PageToken:        nextPageToken,
		}).Do()
		if err != nil {
			return []string{}, fmt.Errorf("Failed to query testable permissions on %s: %v", resource, err)
		}

		for _, perm := range ps.Permissions {
			util.Logger.Debugf("Adding testable permission: %s", perm.Name)
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
func QueryComputeInstancePermissions(permsToTest []string, project, zone, instance string) ([]string, error) {
	computeService, err := compute.NewService(ctx)
	if err != nil {
		return []string{}, fmt.Errorf("Failed to create Compute SDK Client: %v", err)
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
		return []string{}, fmt.Errorf("Failed to test IAM permissions on instance %s: %v", instance, err)
	}

	return resp.Permissions, nil
}

// QueryProjectPermissions gets the authenticated members permissions on a project
// Modified from https://github.com/salrashid123/gcp_iam/blob/main/query/main.go#L534-L575
func QueryProjectPermissions(permsToTest []string, project string) ([]string, error) {
	crmService, err := crm.NewService(ctx)
	if err != nil {
		return []string{}, fmt.Errorf("Failed to create Cloud Resource Manager SDK Client: %v", err)
	}
	crmProjService := crm.NewProjectsService(crmService)

	// TestIamPermissions accepts a max of 100 permissions at a time so we split them into chunks
	var chunked [][]string
	pageSize := (len(permsToTest) + 50 - 1) / 50
	for i := 0; i < len(permsToTest); i += pageSize {
		end := i + pageSize

		if end > len(permsToTest) {
			end = len(permsToTest)
		}
		chunked = append(chunked, permsToTest[i:end])
	}

	var userPermissions []string
	for _, permSet := range chunked {
		resp, err := crmProjService.TestIamPermissions(project, &crm.TestIamPermissionsRequest{
			Permissions: permSet,
		}).Do()
		if err != nil {
			return []string{}, fmt.Errorf("Failed to test IAM permissions on project %s: %v", project, err)
		}
		userPermissions = append(userPermissions, resp.Permissions...)
	}

	return userPermissions, nil
}

// QueryPubSubPermissions gets the authenticated members permissions on a PubSub topic
func QueryPubSubPermissions(permsToTest []string, project, topic, serviceAccountEmail string) ([]string, error) {
	var pubsubService *pubsub.Service
	if serviceAccountEmail != "" {
		if svc, err := pubsub.NewService(ctx, option.ImpersonateCredentials(serviceAccountEmail)); err == nil {
			pubsubService = svc
		} else {
			return []string{}, fmt.Errorf("Failed to create PubSub SDK Client with service account %s: %v", serviceAccountEmail, err)
		}
	} else {
		if svc, err := pubsub.NewService(ctx); err == nil {
			pubsubService = svc
		} else {
			return []string{}, fmt.Errorf("Failed to create PubSub SDK Client: %v", err)
		}
	}

	topicsService := pubsub.NewProjectsTopicsService(pubsubService)

	resource := fmt.Sprintf("projects/%s/topics/%s", project, topic)
	resp, err := topicsService.TestIamPermissions(resource, &pubsub.TestIamPermissionsRequest{
		Permissions: permsToTest,
	}).Do()
	if err != nil {
		return []string{}, fmt.Errorf("Failed to test IAM permissions on PubSub topic %s: %v", resource, err)
	}

	return resp.Permissions, nil
}

// QueryServiceAccountPermissions gets the authenticated members permissions on a service account
// Modified from https://github.com/salrashid123/gcp_iam/blob/main/query/main.go#L150-L173
func QueryServiceAccountPermissions(permsToTest []string, project, email string) ([]string, error) {
	iamService, err := iam.NewService(ctx)
	if err != nil {
		return []string{}, fmt.Errorf("Failed to create Cloud IAM SDK Client: %v", err)
	}
	saIamService := iam.NewProjectsServiceAccountsService(iamService)

	resource := fmt.Sprintf("projects/%s/serviceAccounts/%s", project, email)
	resp, err := saIamService.TestIamPermissions(resource, &iam.TestIamPermissionsRequest{
		Permissions: permsToTest,
	}).Do()
	if err != nil {
		return []string{}, fmt.Errorf("Failed to test IAM permissions on %s: %v", resource, err)
	}

	return resp.Permissions, nil
}

// QueryStorageBucketPermissions gets the authenticated members permissions on a storage bucket
// Modified from https://github.com/salrashid123/gcp_iam/blob/main/query/main.go#L313-L338
func QueryStorageBucketPermissions(permsToTest []string, bucket string) ([]string, error) {
	storageService, err := storage.NewService(ctx)
	if err != nil {
		return []string{}, fmt.Errorf("Failed to create Cloud Storage SDK Client: %v", err)
	}

	permsToTest = remove(permsToTest, []string{
		"resourcemanager.resourceTagBindings.create",
		"resourcemanager.resourceTagBindings.delete",
		"resourcemanager.resourceTagBindings.list",
	})

	resp, err := storageService.Buckets.TestIamPermissions(bucket, permsToTest).Do()
	if err != nil {
		return []string{}, fmt.Errorf("Failed to test IAM permissions on bucket %s: %v", bucket, err)
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
