package jobsui

import "testing"

func TestTrunc_isRuneSafe(t *testing.T) {
	got := trunc("ação extra", 4)
	if got != "ação…" {
		t.Fatalf("trunc = %q, want ação…", got)
	}
}
