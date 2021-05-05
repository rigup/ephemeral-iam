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

package plugins

import (
	"context"

	pb "github.com/rigup/ephemeral-iam/internal/plugins/proto"
)

type GRPCClient struct {
	Client pb.EIAMPluginClient
}

func (m *GRPCClient) GetInfo() (name, desc, version string, err error) {
	resp, err := m.Client.GetInfo(context.Background(), &pb.Empty{})
	if err != nil {
		return "", "", "", err
	}
	return resp.Name, resp.Description, resp.Version, nil
}

func (m *GRPCClient) Run(args []string) error {
	_, err := m.Client.Run(context.Background(), &pb.Args{
		Arg: args,
	})
	if err != nil {
		return err
	}
	return nil
}
