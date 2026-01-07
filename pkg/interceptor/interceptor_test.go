package interceptor

import (
	"github.com/casnerano/protoc-gen-go-guard/pkg/guard"
)

type mockGuardServiceProvider struct {
	service *guard.Service
}

func (m mockGuardServiceProvider) GuardService() *guard.Service {
	return m.service
}
