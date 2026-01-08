// go:build e2e

package benchmarks

import (
	"context"
	"testing"

	desc "github.com/casnerano/protoc-gen-go-guard/e2e/grpc/pb/benchmarks"
	services "github.com/casnerano/protoc-gen-go-guard/e2e/grpc/services/benchmarks"
	"google.golang.org/protobuf/types/known/emptypb"
)

func Benchmark_Simple_PublicAccessMethod(b *testing.B) {
	ctx := context.Background()
	server := services.NewSimpleServer()

	benchmark(b, server, ctx, func(client desc.BenchmarksClient, ctx context.Context) error {
		_, err := client.PublicAccessMethod(ctx, &emptypb.Empty{})
		return err
	})
}

func Benchmark_Simple_AuthenticatedAccessMethod(b *testing.B) {
	ctx := testContextWithSubject(nil)
	server := services.NewSimpleServer()

	benchmark(b, server, ctx, func(client desc.BenchmarksClient, ctx context.Context) error {
		_, err := client.AuthenticatedAccessMethod(ctx, &emptypb.Empty{})
		return err
	})
}

func Benchmark_Simple_RoleBasedAccessMethod(b *testing.B) {
	ctx := testContextWithSubject([]string{"first", "second", "third"})
	server := services.NewSimpleServer()

	benchmark(b, server, ctx, func(client desc.BenchmarksClient, ctx context.Context) error {
		_, err := client.RoleBasedAccessMethod(ctx, &emptypb.Empty{})
		return err
	})
}
