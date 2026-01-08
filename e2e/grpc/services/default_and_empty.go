package services

import (
	"context"

	"github.com/casnerano/protoc-gen-go-guard/e2e/grpc/pb/corner_cases"
	"google.golang.org/protobuf/types/known/emptypb"
)

type DefaultRulesServer struct {
	corner_cases.UnimplementedDefaultRulesServer
}

func (d *DefaultRulesServer) GetOne(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

type EmptyServiceRulesServer struct {
	corner_cases.UnimplementedEmptyServiceRulesServer
}

func (e *EmptyServiceRulesServer) GetOne(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

type EmptyMethodRulesServer struct {
	corner_cases.UnimplementedEmptyMethodRulesServer
}

func (e *EmptyMethodRulesServer) GetOne(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

type EmptyServiceAndMethodRulesServer struct {
	corner_cases.UnimplementedEmptyServiceAndMethodRulesServer
}

func (e *EmptyServiceAndMethodRulesServer) GetOne(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}
