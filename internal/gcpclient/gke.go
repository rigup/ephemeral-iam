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

	container "cloud.google.com/go/container/apiv1"
	"google.golang.org/api/option"
	containerpb "google.golang.org/genproto/googleapis/container/v1"

	util "github.com/rigup/ephemeral-iam/internal/eiamutil"
	errorsutil "github.com/rigup/ephemeral-iam/internal/errors"
)

func GetClusters(project, reason string) ([]map[string]string, error) {
	gkeClient, err := container.NewClusterManagerClient(context.Background(), option.WithRequestReason(reason))
	if err != nil {
		return []map[string]string{}, &errorsutil.SDKClientCreateError{Err: err, ResourceType: "Container"}
	}

	listClustersReq := &containerpb.ListClustersRequest{
		Parent: fmt.Sprintf("projects/%s/locations/-", project),
	}

	resp, err := gkeClient.ListClusters(ctx, listClustersReq)
	if err != nil {
		util.Logger.Error("Failed to list GKE clusters")
		return []map[string]string{}, err
	}
	clusterNames := []map[string]string{}
	for _, cluster := range resp.Clusters {
		clusterNames = append(clusterNames, map[string]string{"name": cluster.Name, "location": cluster.Location})
	}
	return clusterNames, nil
}
