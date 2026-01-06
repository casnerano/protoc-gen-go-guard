package corner_cases

import (
	"context"
	"testing"

	desc "github.com/casnerano/protoc-gen-go-guard/example/pb/corner_cases"
	"github.com/casnerano/protoc-gen-go-guard/pkg/interceptor"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type InheritAndOverrideOneServerTestSuite struct {
	CornerCasesServerTestSuite

	client desc.InheritAndOverrideOneClient
}

func (s *InheritAndOverrideOneServerTestSuite) SetupSuite() {
	s.CornerCasesServerTestSuite.SetupSuite()
	desc.RegisterInheritAndOverrideOneServer(s.server, &InheritAndOverrideOneServer{})
	s.StartServer()

	client, err := s.GetClientConn()
	s.Require().NoError(err, "Failed to dial test server")

	s.client = desc.NewInheritAndOverrideOneClient(client)
}

func (s *InheritAndOverrideOneServerTestSuite) TestInheritedMethod() {
	testCases := []struct {
		name      string
		context   context.Context
		canAccess bool
	}{
		{
			name:      "access denied without token",
			context:   context.Background(),
			canAccess: false,
		},
		{
			name:      "access denied with token",
			context:   testContextWithSubject(interceptor.Subject{}),
			canAccess: false,
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			_, err := s.client.InheritedMethod(tt.context, &emptypb.Empty{})
			if tt.canAccess {
				s.NoError(err)
			} else {
				s.Equal(codes.PermissionDenied, status.Code(err))
			}
		})
	}
}

func (s *InheritAndOverrideOneServerTestSuite) TestOverriddenMethod() {
	testCases := []struct {
		name      string
		context   context.Context
		canAccess bool
	}{
		{
			name:      "access allowed without token",
			context:   context.Background(),
			canAccess: true,
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			_, err := s.client.OverriddenMethod(tt.context, &emptypb.Empty{})
			if tt.canAccess {
				s.NoError(err)
			} else {
				s.Equal(codes.PermissionDenied, status.Code(err))
			}
		})
	}
}

type InheritAndOverrideTwoServerTestSuite struct {
	CornerCasesServerTestSuite

	client desc.InheritAndOverrideTwoClient
}

func (s *InheritAndOverrideTwoServerTestSuite) SetupSuite() {
	s.CornerCasesServerTestSuite.SetupSuite()
	desc.RegisterInheritAndOverrideTwoServer(s.server, &InheritAndOverrideTwoServer{})
	s.StartServer()

	client, err := s.GetClientConn()
	s.Require().NoError(err, "Failed to dial test server")

	s.client = desc.NewInheritAndOverrideTwoClient(client)
}

func (s *InheritAndOverrideTwoServerTestSuite) TestInheritedMethod() {
	testCases := []struct {
		name      string
		context   context.Context
		canAccess bool
	}{
		{
			name:      "access allowed without token",
			context:   context.Background(),
			canAccess: true,
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			_, err := s.client.InheritedMethod(tt.context, &emptypb.Empty{})
			if tt.canAccess {
				s.NoError(err)
			} else {
				s.Equal(codes.PermissionDenied, status.Code(err))
			}
		})
	}
}

func (s *InheritAndOverrideTwoServerTestSuite) TestOverriddenMethod() {
	testCases := []struct {
		name      string
		context   context.Context
		canAccess bool
	}{
		{
			name:      "access denied without token",
			context:   context.Background(),
			canAccess: false,
		},
		{
			name:      "access denied with token",
			context:   testContextWithSubject(interceptor.Subject{}),
			canAccess: false,
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			_, err := s.client.OverriddenMethod(tt.context, &emptypb.Empty{})
			if tt.canAccess {
				s.NoError(err)
			} else {
				s.Equal(codes.PermissionDenied, status.Code(err))
			}
		})
	}
}

func TestInheritAndOverrideOneServer(t *testing.T) {
	suite.Run(t, new(InheritAndOverrideOneServerTestSuite))
}

func TestInheritAndOverrideTwoServer(t *testing.T) {
	suite.Run(t, new(InheritAndOverrideTwoServerTestSuite))
}
