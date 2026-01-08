//go:build e2e

package benchmarks

import (
	"context"
	"log"
	"net"
	"testing"

	desc "github.com/casnerano/protoc-gen-go-guard/e2e/grpc/pb/benchmarks"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"
)

const bufferSize = 1024 * 1024

func testContextWithSubject(roles []string) context.Context {
	md := metadata.MD{}
	md.Append("authenticated", "1")
	md.Append("roles", roles...)

	return metadata.NewOutgoingContext(context.Background(), md)
}

func startServer(server *grpc.Server) *bufconn.Listener {
	listener := bufconn.Listen(bufferSize)

	go func() {
		if err := server.Serve(listener); err != nil {
			log.Printf("Failed serve at %v: %v", listener.Addr(), err)
		}
	}()

	return listener
}

func getClientConn(listener *bufconn.Listener) *grpc.ClientConn {
	conn, _ := grpc.NewClient(
		"passthrough://bufnet",
		grpc.WithContextDialer(func(_ context.Context, _ string) (net.Conn, error) {
			return listener.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	return conn
}

func benchmark(b *testing.B, server *grpc.Server, ctx context.Context, fn func(client desc.BenchmarksClient, ctx context.Context) error) {
	listener := startServer(server)
	defer func() {
		server.GracefulStop()
		_ = listener.Close()
	}()

	conn := getClientConn(listener)
	defer func() {
		_ = conn.Close()
	}()

	client := desc.NewBenchmarksClient(conn)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := fn(client, ctx); err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}
