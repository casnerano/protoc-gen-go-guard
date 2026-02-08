package main

import (
	"fmt"
	"runtime/debug"

	guardplugin "github.com/casnerano/protoc-gen-go-guard/internal/plugin"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
)

var version string

func main() {
	protogen.Options{}.Run(func(plugin *protogen.Plugin) error {
		plugin.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)

		meta := guardplugin.Meta{
			ProtocVersion: getProtocVersion(plugin),
			PluginVersion: getPluginVersion(),
		}

		return guardplugin.Execute(plugin, meta)
	})
}

func getPluginVersion() string {
	if version != "" {
		return version
	}

	if info, ok := debug.ReadBuildInfo(); ok {
		return info.Main.Version
	}

	return "(devel)"
}

func getProtocVersion(plugin *protogen.Plugin) string {
	if ver := plugin.Request.CompilerVersion; ver != nil {
		return fmt.Sprintf("v%d.%d.%d", *ver.Major, *ver.Minor, *ver.Patch)
	}

	return "(unknown)"
}
