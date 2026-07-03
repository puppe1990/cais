package boot

import (
	"runtime/debug"
	"strings"
)

const modulePath = "github.com/puppe1990/cais"

func CaisVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "dev"
	}
	return versionFrom(info)
}

func versionFrom(info *debug.BuildInfo) string {
	if info == nil {
		return "dev"
	}
	if info.Main.Path == modulePath {
		v := strings.TrimPrefix(info.Main.Version, "v")
		if v != "" && v != "(devel)" {
			return v
		}
	}
	for _, dep := range info.Deps {
		if dep.Path == modulePath {
			return strings.TrimPrefix(dep.Version, "v")
		}
	}
	return "dev"
}
