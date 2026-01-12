// Package plugin implements the protoc plugin that generates Go guard definitions
// from protobuf service and method options.
//
// For each .proto file containing gRPC services annotated with guard rules,
// it produces a corresponding .guard.go file that:
//   - declares a *guard.Service struct containing the access rules
//     (both service-level and method-level),
//   - adds a GuardService() method to the Unimplemented<ServiceName>Server
//     to allow runtime access to these rules.
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

// TemplateData holds all the information passed to the Go template
// that generates the .guard.go output file.
type TemplateData struct {
	Meta     Meta             // Build and tooling metadata.
	File     File             // Information about the source .proto file.
	Services []*guard.Service // Guard rules for all gRPC services in the file.
}

// Meta contains version information used in the generated file header.
type Meta struct {
	ModuleVersion string // Version of protoc-gen-go-guard.
	ProtocVersion string // Version of the protoc compiler used.
}

// File describes the input .proto file being processed.
type File struct {
	Name    string // Base name of the generated Go file.
	Package string // Go package name of the generated code.
	Source  string // Path to the original .proto source file.
}

// Execute processes protobuf files passed by protoc and generates .guard.go files.
// It inspects service and method options for guard rules,
// converts them into internal guard structures,
// and renders Go code that embeds these rules and exposes them via a GuardService() method
// on the corresponding gRPC server base type.
//
// The output file is named <prefix>.guard.go and placed in the same package
// as the generated gRPC code.
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
		}

		filename := file.GeneratedFilenamePrefix + outputFileSuffix
		if err = tmpl.Execute(plugin.NewGeneratedFile(filename, file.GoImportPath), templateData); err != nil {
			return fmt.Errorf("failed execute template: %w", err)
		}
	}

	return nil
}

// collectServices converts protobuf service descriptors into internal guard.Service structs,
// extracting explicitly defined service-level rules and method-level rules.
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

// collectMethods gathers method-level access rules from protobuf method options
// and returns a map keyed by method name.
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

// extractRule translates a protobuf-defined Rule message into the internal guard.Rule representation.
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
					Roles:       roleBased.Roles,
					Requirement: guard.RequirementAtLeastOne,
				}

				if roleBased.Requirement != nil {
					authenticatedAccess.RoleBased.Requirement = guard.Requirement(*roleBased.Requirement)
				}
			}

			if policyBased := mode.AuthenticatedAccess.PolicyBased; policyBased != nil {
				authenticatedAccess.PolicyBased = &guard.PolicyBased{
					Policies:    policyBased.Policies,
					Requirement: guard.RequirementAtLeastOne,
				}

				if policyBased.Requirement != nil {
					authenticatedAccess.PolicyBased.Requirement = guard.Requirement(*policyBased.Requirement)
				}
			}

			rule.AuthenticatedAccess = authenticatedAccess
		}
	}

	return &rule
}

// parseTemplate loads and parses the embedded Go template used to generate .guard.go files.
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
