package cais

import (
	"net"
	"testing"
)

func TestResolvePort_UsesNextWhenBusy(t *testing.T) {
	probe, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}
	_, port, err := net.SplitHostPort(probe.Addr().String())
	_ = probe.Close()
	if err != nil {
		t.Fatal(err)
	}

	preferred := ":" + port
	ln, err := net.Listen("tcp", preferred)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ln.Close() })
	resolved, shifted, err := ResolvePort(preferred, "development")
	if err != nil {
		t.Fatal(err)
	}
	if !shifted {
		t.Fatal("expected shifted=true")
	}
	if resolved == preferred {
		t.Fatalf("resolved = %q, want different from %q", resolved, preferred)
	}

	check, err := net.Listen("tcp", "127.0.0.1"+resolved)
	if err != nil {
		t.Fatalf("resolved port not free: %v", err)
	}
	_ = check.Close()
}

func TestResolvePort_KeepsPreferredInProduction(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ln.Close() })

	_, port, err := net.SplitHostPort(ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}

	preferred := ":" + port
	resolved, shifted, err := ResolvePort(preferred, "production")
	if err != nil {
		t.Fatal(err)
	}
	if shifted {
		t.Fatal("production should not auto-shift ports")
	}
	if resolved != preferred {
		t.Fatalf("resolved = %q, want %q", resolved, preferred)
	}
}

func TestResolvePort_StrictFailsWhenBusy(t *testing.T) {
	t.Setenv("PORT_STRICT", "1")

	probe, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}
	_, port, err := net.SplitHostPort(probe.Addr().String())
	_ = probe.Close()
	if err != nil {
		t.Fatal(err)
	}

	preferred := ":" + port
	ln, err := net.Listen("tcp", preferred)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ln.Close() })

	_, shifted, err := ResolvePort(preferred, "development")
	if err == nil {
		t.Fatal("expected error when PORT_STRICT=1 and port busy")
	}
	if shifted {
		t.Fatal("expected shifted=false on strict failure")
	}
}

func TestPortBusy_trueWhenListening(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = ln.Close() }()
	if !PortBusy(ln.Addr().String()) {
		t.Fatal("PortBusy should be true for bound address")
	}
}

func TestPortBusy_falseWhenFree(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	_ = ln.Close()
	if PortBusy(addr) {
		t.Fatal("PortBusy should be false after listener closed")
	}
}

func TestResolvePort_UnchangedWhenFree(t *testing.T) {
	resolved, shifted, err := ResolvePort(":0", "development")
	if err != nil {
		t.Fatal(err)
	}
	if shifted {
		t.Fatal("expected shifted=false for free port")
	}
	if resolved != ":0" {
		t.Fatalf("resolved = %q, want :0", resolved)
	}
}
