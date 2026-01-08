package corner_cases

import (
	"context"

	desc "github.com/casnerano/protoc-gen-go-guard/e2e/grpc/pb/corner_cases"
	"google.golang.org/protobuf/types/known/emptypb"
)

type DefaultRulesServer struct {
	desc.UnimplementedDefaultRulesServer
}

func (d *DefaultRulesServer) GetOne(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

type EmptyServiceRulesServer struct {
	desc.UnimplementedEmptyServiceRulesServer
}

func (e *EmptyServiceRulesServer) GetOne(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

type EmptyMethodRulesServer struct {
	desc.UnimplementedEmptyMethodRulesServer
}

func (e *EmptyMethodRulesServer) GetOne(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

type EmptyServiceAndMethodRulesServer struct {
	desc.UnimplementedEmptyServiceAndMethodRulesServer
}

func (e *EmptyServiceAndMethodRulesServer) GetOne(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}
