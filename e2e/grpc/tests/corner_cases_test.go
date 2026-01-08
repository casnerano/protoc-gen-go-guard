//go:build e2e

package tests

import (
	"context"
	"net"

	"github.com/casnerano/protoc-gen-go-guard/pkg/interceptor"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"
)

type CornerCasesServerTestSuite struct {
	suite.Suite

	listener *bufconn.Listener
	server   *grpc.Server
}

func (g *CornerCasesServerTestSuite) SetupSuite() {
	const bufferSize = 1024 * 1024
	g.listener = bufconn.Listen(bufferSize)

	g.server = grpc.NewServer(
		grpc.UnaryInterceptor(
			interceptor.New(
				testSubjectResolver(),
				interceptor.WithPolicies(testPolicies()),
			).Unary(),
		),
	)
}

func (g *CornerCasesServerTestSuite) TearDownSuite() {
	if g.server != nil {
		g.server.GracefulStop()
	}

	if g.listener != nil {
		_ = g.listener.Close()
	}
}

func (g *CornerCasesServerTestSuite) StartServer() {
	go func() {
		if err := g.server.Serve(g.listener); err != nil {
			g.T().Logf("Failed serve at %v: %v", g.listener.Addr(), err)
		}
	}()
}

func (g *CornerCasesServerTestSuite) GetClientConn() (*grpc.ClientConn, error) {
	return grpc.NewClient(
		"passthrough://bufnet",
		grpc.WithContextDialer(func(_ context.Context, _ string) (net.Conn, error) {
			return g.listener.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
}

func testSubjectResolver() interceptor.SubjectResolver {
	return func(ctx context.Context) (*interceptor.Subject, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, nil
		}

		if _, authenticated := md["authenticated"]; !authenticated {
			return nil, nil
		}

		subject := interceptor.Subject{}
		if roles, exists := md["roles"]; exists {
			subject.Roles = roles
		}

		return &subject, nil
	}
}

func testContextWithSubject(subject interceptor.Subject) context.Context {
	md := metadata.MD{}
	md.Append("authenticated", "1")
	md.Append("roles", subject.Roles...)

	return metadata.NewOutgoingContext(context.Background(), md)
}

func testPolicies() interceptor.Policies {
	return interceptor.Policies{
		"positive-policy-1": func(ctx context.Context, input *interceptor.Input) (bool, error) {
			return true, nil
		},
		"positive-policy-2": func(ctx context.Context, input *interceptor.Input) (bool, error) {
			return true, nil
		},
		"negative-policy-1": func(ctx context.Context, input *interceptor.Input) (bool, error) {
			return false, nil
		},
	}
}
