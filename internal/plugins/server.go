package plugins

import (
	"context"

	pb "github.com/rigup/ephemeral-iam/internal/plugins/proto"
)

type GRPCServer struct {
	Impl EIAMPlugin
	pb.UnimplementedEIAMPluginServer
}

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

func (m *GRPCServer) Run(ctx context.Context, args *pb.Args) (*pb.Empty, error) {
	return &pb.Empty{}, m.Impl.Run(args.Arg)
}
