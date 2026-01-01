package demo

import (
	"context"

	desc "github.com/casnerano/protoc-gen-go-guard/example/pb/demo"
	"google.golang.org/protobuf/types/known/emptypb"
)

type AuthServer struct {
	desc.UnimplementedAuthServer
}

func (s *AuthServer) Register(_ context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (s *AuthServer) Login(_ context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (s *AuthServer) Logout(_ context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}
