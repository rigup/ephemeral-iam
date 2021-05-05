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
