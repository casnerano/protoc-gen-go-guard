package plugin

import (
	"embed"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/casnerano/protoc-gen-go-guard/pkg/guard"
	desc "github.com/casnerano/protoc-gen-go-guard/proto"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

const (
	outputFileSuffix = ".guard.go"
)

//go:embed plugin.go.tmpl
var templateFS embed.FS

type TemplateData struct {
	Meta     Meta
	File     File
	Services []*guard.Service

	Test string
}

type Meta struct {
	ModuleVersion string
	ProtocVersion string
}

type File struct {
	Name    string
	Package string
	Source  string
}

func Execute(plugin *protogen.Plugin) error {
	tmpl, err := parseTemplate()
	if err != nil {
		return err
	}

	for _, file := range plugin.Files {
		if !file.Generate {
			continue
		}

		services := collectServices(file.Services)
		if len(services) == 0 {
			continue
		}

		templateData := TemplateData{
			Meta: Meta{
				ProtocVersion: func() string {
					if ver := plugin.Request.CompilerVersion; ver != nil {
						return fmt.Sprintf("v%d.%d.%d", ver.Major, ver.Minor, ver.Patch)
					}
					return "(unknown)"
				}(),
				ModuleVersion: "(unknown)",
			},
			File: File{
				Name:    filepath.Base(file.GeneratedFilenamePrefix),
				Package: string(file.GoPackageName),
				Source:  file.Desc.Path(),
			},
			Services: services,
			Test:     fmt.Sprintf("%#v", services),
		}

		filename := file.GeneratedFilenamePrefix + outputFileSuffix
		if err = tmpl.Execute(plugin.NewGeneratedFile(filename, file.GoImportPath), templateData); err != nil {
			return fmt.Errorf("failed execute template: %w", err)
		}
	}

	return nil
}

func collectServices(protoServices []*protogen.Service) []*guard.Service {
	var services []*guard.Service
	for _, protoService := range protoServices {
		service := guard.Service{
			Name: string(protoService.Desc.Name()),
		}

		if options := protoService.Desc.Options().(*descriptorpb.ServiceOptions); options != nil {
			if pbRules, ok := proto.GetExtension(options, desc.E_ServiceRules).([]*desc.Rule); ok {
				service.Rules = make([]*guard.Rule, 0, len(pbRules))
				for _, pbRule := range pbRules {
					service.Rules = append(service.Rules, extractRule(pbRule))
				}
			}
		}

		if methods := collectMethods(protoService.Methods); len(methods) > 0 {
			service.Methods = methods
		}

		if len(service.Rules) == 0 && len(service.Methods) == 0 {
			continue
		}

		services = append(services, &service)
	}

	return services
}

func collectMethods(protoMethods []*protogen.Method) map[string]*guard.Method {
	methods := make(map[string]*guard.Method)

	for _, protoMethod := range protoMethods {
		if options := protoMethod.Desc.Options().(*descriptorpb.MethodOptions); options != nil {
			if pbRules, ok := proto.GetExtension(options, desc.E_MethodRules).([]*desc.Rule); ok {
				guardRules := make([]*guard.Rule, 0, len(pbRules))
				for _, pbRule := range pbRules {
					guardRules = append(guardRules, extractRule(pbRule))
				}

				methods[string(protoMethod.Desc.Name())] = &guard.Method{
					Rules: guardRules,
				}
			}
		}
	}

	return methods
}

func extractRule(pbRule *desc.Rule) *guard.Rule {
	rule := guard.Rule{}

	switch mode := pbRule.Mode.(type) {
	case *desc.Rule_AllowPublic:
		rule.AllowPublic = &mode.AllowPublic

	case *desc.Rule_RequireAuthentication:
		rule.RequireAuthentication = &mode.RequireAuthentication

	case *desc.Rule_AuthenticatedAccess:
		if mode.AuthenticatedAccess != nil {
			authenticatedAccess := &guard.AuthenticatedAccess{}

			if roleBased := mode.AuthenticatedAccess.RoleBased; roleBased != nil {
				authenticatedAccess.RoleBased = &guard.RoleBased{
					Roles: roleBased.Roles,
					Match: guard.MatchAtLeastOne,
				}

				if roleBased.Match != nil {
					authenticatedAccess.RoleBased.Match = guard.Match(*roleBased.Match)
				}
			}

			if policyBased := mode.AuthenticatedAccess.PolicyBased; policyBased != nil {
				authenticatedAccess.PolicyBased = &guard.PolicyBased{
					Policies: policyBased.Policies,
					Match:    guard.MatchAtLeastOne,
				}

				if policyBased.Match != nil {
					authenticatedAccess.PolicyBased.Match = guard.Match(*policyBased.Match)
				}
			}

			rule.AuthenticatedAccess = authenticatedAccess
		}
	}

	return &rule
}

func parseTemplate() (*template.Template, error) {
	templateContent, err := templateFS.ReadFile("plugin.go.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to read template: %w", err)
	}

	tmpl := template.New("plugin.guard").
		Funcs(template.FuncMap{
			"toLower": strings.ToLower,
		})

	return tmpl.Parse(string(templateContent))
}
