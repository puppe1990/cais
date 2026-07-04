package cli

import "testing"

func TestParseBuildFlags_defaults(t *testing.T) {
	opts := parseBuildFlags(nil)
	if opts.GOOS != "" || opts.GOARCH != "" {
		t.Fatalf("expected empty GOOS/GOARCH, got %+v", opts)
	}
	if opts.Output != serverBin {
		t.Errorf("Output = %q, want %q", opts.Output, serverBin)
	}
}

func TestParseBuildFlags_crossCompile(t *testing.T) {
	opts := parseBuildFlags([]string{"--os", "linux", "--arch", "amd64", "-o", "bin/server-linux"})
	if opts.GOOS != "linux" || opts.GOARCH != "amd64" {
		t.Fatalf("got %+v", opts)
	}
	if opts.Output != "bin/server-linux" {
		t.Errorf("Output = %q", opts.Output)
	}
}
