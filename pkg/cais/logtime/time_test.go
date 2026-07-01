package logtime

import (
	"regexp"
	"testing"
	"time"
)

func TestFormat_IncludesDateTimeAndZone(t *testing.T) {
	ts := time.Date(2026, 7, 1, 11, 22, 16, 0, time.FixedZone("BRT", -3*3600))
	got := Format(ts)
	re := regexp.MustCompile(`^2026-07-01 11:22:16 -0300$`)
	if !re.MatchString(got) {
		t.Fatalf("Format() = %q", got)
	}
}
