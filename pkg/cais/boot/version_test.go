package boot

import (
	"strings"
	"testing"
)

func TestCaisVersion(t *testing.T) {
	v := CaisVersion()
	if strings.TrimSpace(v) == "" {
		t.Fatal("CaisVersion() returned empty string")
	}
}
