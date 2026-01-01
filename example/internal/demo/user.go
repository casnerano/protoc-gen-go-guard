package demo

import (
	"context"

	desc "github.com/casnerano/protoc-gen-go-guard/example/pb/demo"
	"google.golang.org/protobuf/types/known/emptypb"
)

type UserServer struct {
	desc.UnimplementedUserServer
}

func (s *UserServer) GetPublicProfile(_ context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (s *UserServer) GetPrivateProfile(_ context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (s *UserServer) UpdateProfileStatus(_ context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (s *UserServer) DeleteProfile(_ context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}
