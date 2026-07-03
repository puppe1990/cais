package boot

import (
	"runtime/debug"
	"strings"
	"testing"
)

func TestCaisVersion(t *testing.T) {
	v := CaisVersion()
	if strings.TrimSpace(v) == "" {
		t.Fatal("CaisVersion() returned empty string")
	}
}

func TestVersionFrom_nil(t *testing.T) {
	if got := versionFrom(nil); got != "dev" {
		t.Fatalf("versionFrom(nil) = %q, want dev", got)
	}
}

func TestVersionFrom_mainModuleRelease(t *testing.T) {
	got := versionFrom(&debug.BuildInfo{
		Main: debug.Module{Path: modulePath, Version: "v1.2.3"},
	})
	if got != "1.2.3" {
		t.Fatalf("versionFrom(main release) = %q, want 1.2.3", got)
	}
}

func TestVersionFrom_mainModuleWithoutVPrefix(t *testing.T) {
	got := versionFrom(&debug.BuildInfo{
		Main: debug.Module{Path: modulePath, Version: "2.0.0"},
	})
	if got != "2.0.0" {
		t.Fatalf("versionFrom(main no v) = %q, want 2.0.0", got)
	}
}

func TestVersionFrom_mainDevelFallsBackToDeps(t *testing.T) {
	got := versionFrom(&debug.BuildInfo{
		Main: debug.Module{Path: modulePath, Version: "(devel)"},
		Deps: []*debug.Module{{Path: modulePath, Version: "v0.9.0"}},
	})
	if got != "0.9.0" {
		t.Fatalf("versionFrom(devel + dep) = %q, want 0.9.0", got)
	}
}

func TestVersionFrom_mainEmptyVersionFallsBackToDeps(t *testing.T) {
	got := versionFrom(&debug.BuildInfo{
		Main: debug.Module{Path: modulePath, Version: ""},
		Deps: []*debug.Module{{Path: "other/mod", Version: "v9.9.9"}, {Path: modulePath, Version: "v0.4.1"}},
	})
	if got != "0.4.1" {
		t.Fatalf("versionFrom(empty main + dep) = %q, want 0.4.1", got)
	}
}

func TestVersionFrom_dependencyOnly(t *testing.T) {
	got := versionFrom(&debug.BuildInfo{
		Main: debug.Module{Path: "github.com/acme/myapp", Version: "(devel)"},
		Deps: []*debug.Module{{Path: modulePath, Version: "v0.8.2"}},
	})
	if got != "0.8.2" {
		t.Fatalf("versionFrom(dep only) = %q, want 0.8.2", got)
	}
}

func TestVersionFrom_noMatchReturnsDev(t *testing.T) {
	got := versionFrom(&debug.BuildInfo{
		Main: debug.Module{Path: "github.com/acme/myapp", Version: "v1.0.0"},
		Deps: []*debug.Module{{Path: "other/mod", Version: "v0.1.0"}},
	})
	if got != "dev" {
		t.Fatalf("versionFrom(no match) = %q, want dev", got)
	}
}
