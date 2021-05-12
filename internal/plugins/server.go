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

// GRPCServer is the implementation of the go-plugin gRPC server.
type GRPCServer struct {
	pb.UnimplementedEIAMPluginServer
	Impl EIAMPlugin
}

// GetInfo is the gRPC method that is called to get metadata about a plugin.
func (m *GRPCServer) GetInfo(ctx context.Context, req *pb.Empty) (*pb.PluginInfo, error) {
	name, desc, version, err := m.Impl.GetInfo()
	if err != nil {
		return nil, err
	}
	pi := &pb.PluginInfo{
		Name:        name,
		Description: desc,
		Version:     version,
	}
	return pi, nil
}

// Run is the gRPC method that is called to invoke a plugin's root command.
func (m *GRPCServer) Run(ctx context.Context, args *pb.Empty) (*pb.Empty, error) {
	return &pb.Empty{}, m.Impl.Run()
}
