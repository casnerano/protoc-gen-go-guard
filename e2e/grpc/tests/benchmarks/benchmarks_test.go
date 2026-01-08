package benchmarks

import (
	"context"

	"google.golang.org/grpc/metadata"
)

const bufferSize = 1024 * 1024

func testContextWithSubject(roles []string) context.Context {
	md := metadata.MD{}
	md.Append("authenticated", "1")
	md.Append("roles", roles...)

	return metadata.NewOutgoingContext(context.Background(), md)
}
