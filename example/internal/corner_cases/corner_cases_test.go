package corner_cases

import (
	"context"
	"net"

	"github.com/casnerano/protoc-gen-go-rbac/pkg/interceptor"
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
			interceptor.RbacUnary(
				testAuthContextResolver(),
				interceptor.WithPolicies(testPolicies()),
			),
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

//func testRegisterServers(server *grpc.Server) {
//	desc.RegisterDefaultRulesServer(server, &DefaultRulesServer{})
//	desc.RegisterEmptyServiceRulesServer(server, &EmptyServiceRulesServer{})
//	desc.RegisterEmptyMethodRulesServer(server, &EmptyMethodRulesServer{})
//	desc.RegisterEmptyServiceAndMethodRulesServer(server, &EmptyServiceAndMethodRulesServer{})
//
//	desc.RegisterInheritAndOverrideOneServer(server, &InheritAndOverrideOneServer{})
//	desc.RegisterInheritAndOverrideTwoServer(server, &InheritAndOverrideTwoServer{})
//
//	desc.RegisterMixedTypesAccessServer(server, &MixedTypesAccessServer{})
//
//	desc.RegisterPolicyBasedAccessServer(server, &PolicyBasedAccessServer{})
//	desc.RegisterRoleBasedAccessServer(server, &RoleBasedAccessServer{})
//}

func testAuthContextResolver() interceptor.AuthContextResolver {
	return func(ctx context.Context) (*interceptor.AuthContext, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return &interceptor.AuthContext{}, nil
		}

		authContext := interceptor.AuthContext{}

		if authorization, exists := md["authorization"]; exists && len(authorization) > 0 {
			authContext.Authenticated = true
		}

		if roles, exists := md["roles"]; exists && len(roles) > 0 {
			authContext.Roles = roles
		}

		return &authContext, nil
	}
}

func testPolicies() interceptor.Policies {
	return interceptor.Policies{
		"positive-policy-1": func(ctx context.Context, authContext *interceptor.AuthContext, request interface{}) (bool, error) {
			return true, nil
		},
		"positive-policy-2": func(ctx context.Context, authContext *interceptor.AuthContext, request interface{}) (bool, error) {
			return true, nil
		},
		"negative-policy-1": func(ctx context.Context, authContext *interceptor.AuthContext, request interface{}) (bool, error) {
			return false, nil
		},
	}
}

func testContextWithMetadata(token string, roles ...string) context.Context {
	md := metadata.MD{}
	md.Append("authorization", token)

	if len(roles) > 0 {
		md.Append("roles", roles...)
	}

	return metadata.NewOutgoingContext(context.Background(), md)
}
