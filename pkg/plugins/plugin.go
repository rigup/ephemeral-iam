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

package eiamplugin

import (
	"context"

	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"

	"github.com/rigup/ephemeral-iam/internal/plugins"
	pb "github.com/rigup/ephemeral-iam/internal/plugins/proto"
)

var Handshake = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "EIAM_PLUGIN",
	MagicCookieValue: "dab75867-cde1-41fc-8416-818b718e4d62",
}

// Command is the implementation of plugin.GRPCPlugin that allows it to be
// served and consumed.
type Command struct {
	plugin.Plugin
	Impl plugins.EIAMPlugin
}

func (p *Command) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	pb.RegisterEIAMPluginServer(s, &plugins.GRPCServer{Impl: p.Impl})
	return nil
}

func (p *Command) GRPCClient(
	ctx context.Context,
	broker *plugin.GRPCBroker,
	c *grpc.ClientConn,
) (interface{}, error) {
	return &plugins.GRPCClient{Client: pb.NewEIAMPluginClient(c)}, nil
}
