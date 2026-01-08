//go:build e2e

package corner_cases

import (
	"context"
	"testing"

	desc "github.com/casnerano/protoc-gen-go-guard/e2e/grpc/pb/corner_cases"
	services "github.com/casnerano/protoc-gen-go-guard/e2e/grpc/services/corner_cases"
	"github.com/casnerano/protoc-gen-go-guard/pkg/interceptor"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type DefaultRulesServerTestSuite struct {
	CornerCasesServerTestSuite

	client desc.DefaultRulesClient
}

func (s *DefaultRulesServerTestSuite) SetupSuite() {
	s.CornerCasesServerTestSuite.SetupSuite()
	desc.RegisterDefaultRulesServer(s.server, &services.DefaultRulesServer{})
	s.StartServer()

	client, err := s.GetClientConn()
	s.Require().NoError(err, "Failed to dial test server")

	s.client = desc.NewDefaultRulesClient(client)
}

func (s *DefaultRulesServerTestSuite) TestGetOne() {
	testCases := []struct {
		name      string
		context   context.Context
		canAccess bool
	}{
		{
			name:      "access denied for unauthenticated",
			context:   context.Background(),
			canAccess: false,
		},
		{
			name:      "access denied for authenticated",
			context:   testContextWithSubject(interceptor.Subject{}),
			canAccess: false,
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			_, err := s.client.GetOne(tt.context, &emptypb.Empty{})
			if tt.canAccess {
				s.NoError(err)
			} else {
				s.Equal(codes.PermissionDenied.String(), status.Code(err).String())
			}
		})
	}
}

type EmptyServiceRulesServerTestSuite struct {
	CornerCasesServerTestSuite

	client desc.EmptyServiceRulesClient
}

func (s *EmptyServiceRulesServerTestSuite) SetupSuite() {
	s.CornerCasesServerTestSuite.SetupSuite()
	desc.RegisterEmptyServiceRulesServer(s.server, &services.EmptyServiceRulesServer{})
	s.StartServer()

	client, err := s.GetClientConn()
	s.Require().NoError(err, "Failed to dial test server")

	s.client = desc.NewEmptyServiceRulesClient(client)
}

func (s *EmptyServiceRulesServerTestSuite) TestGetOne() {
	testCases := []struct {
		name      string
		context   context.Context
		canAccess bool
	}{
		{
			name:      "access denied for unauthenticated",
			context:   context.Background(),
			canAccess: false,
		},
		{
			name:      "access denied for authenticated",
			context:   testContextWithSubject(interceptor.Subject{}),
			canAccess: false,
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			_, err := s.client.GetOne(tt.context, &emptypb.Empty{})
			if tt.canAccess {
				s.NoError(err)
			} else {
				s.Equal(codes.PermissionDenied.String(), status.Code(err).String())
			}
		})
	}
}

type EmptyMethodRulesServerTestSuite struct {
	CornerCasesServerTestSuite

	client desc.EmptyMethodRulesClient
}

func (s *EmptyMethodRulesServerTestSuite) SetupSuite() {
	s.CornerCasesServerTestSuite.SetupSuite()
	desc.RegisterEmptyMethodRulesServer(s.server, &services.EmptyMethodRulesServer{})
	s.StartServer()

	client, err := s.GetClientConn()
	s.Require().NoError(err, "Failed to dial test server")

	s.client = desc.NewEmptyMethodRulesClient(client)
}

func (s *EmptyMethodRulesServerTestSuite) TestGetOne() {
	testCases := []struct {
		name      string
		context   context.Context
		canAccess bool
	}{
		{
			name:      "access denied for unauthenticated",
			context:   context.Background(),
			canAccess: false,
		},
		{
			name:      "access denied for authenticated",
			context:   testContextWithSubject(interceptor.Subject{}),
			canAccess: false,
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			_, err := s.client.GetOne(tt.context, &emptypb.Empty{})
			if tt.canAccess {
				s.NoError(err)
			} else {
				s.Equal(codes.PermissionDenied.String(), status.Code(err).String())
			}
		})
	}
}

type EmptyServiceAndMethodRulesServerTestSuite struct {
	CornerCasesServerTestSuite

	client desc.EmptyServiceAndMethodRulesClient
}

func (s *EmptyServiceAndMethodRulesServerTestSuite) SetupSuite() {
	s.CornerCasesServerTestSuite.SetupSuite()
	desc.RegisterEmptyServiceAndMethodRulesServer(s.server, &services.EmptyServiceAndMethodRulesServer{})
	s.StartServer()

	client, err := s.GetClientConn()
	s.Require().NoError(err, "Failed to dial test server")

	s.client = desc.NewEmptyServiceAndMethodRulesClient(client)
}

func (s *EmptyServiceAndMethodRulesServerTestSuite) TestGetOne() {
	testCases := []struct {
		name      string
		context   context.Context
		canAccess bool
	}{
		{
			name:      "access denied for unauthenticated",
			context:   context.Background(),
			canAccess: false,
		},
		{
			name:      "access denied for authenticated",
			context:   testContextWithSubject(interceptor.Subject{}),
			canAccess: false,
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			_, err := s.client.GetOne(tt.context, &emptypb.Empty{})
			if tt.canAccess {
				s.NoError(err)
			} else {
				s.Equal(codes.PermissionDenied.String(), status.Code(err).String())
			}
		})
	}
}

func TestDefaultRulesServer(t *testing.T) {
	suite.Run(t, new(DefaultRulesServerTestSuite))
}

func TestEmptyServiceRulesServer(t *testing.T) {
	suite.Run(t, new(EmptyServiceRulesServerTestSuite))
}

func TestEmptyMethodRulesServer(t *testing.T) {
	suite.Run(t, new(EmptyMethodRulesServerTestSuite))
}

func TestEmptyServiceAndMethodRulesServer(t *testing.T) {
	suite.Run(t, new(EmptyServiceAndMethodRulesServerTestSuite))
}
