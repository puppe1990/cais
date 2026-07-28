package cli

import "testing"

func TestParseSemverCore(t *testing.T) {
	cases := []struct {
		in          string
		maj, min, p int
		ok          bool
	}{
		{"0.8.0", 0, 8, 0, true},
		{"v0.7.0", 0, 7, 0, true},
		{"0.6.1-0.20260706022352-08aa1182c3f5", 0, 6, 1, true},
		{"dev", 0, 0, 0, false},
		{"", 0, 0, 0, false},
		{"(devel)", 0, 0, 0, false},
	}
	for _, tc := range cases {
		got := parseSemverCore(tc.in)
		if got.OK != tc.ok || got.Major != tc.maj || got.Minor != tc.min || got.Patch != tc.p {
			t.Errorf("parseSemverCore(%q) = %+v, want maj=%d min=%d p=%d ok=%v",
				tc.in, got, tc.maj, tc.min, tc.p, tc.ok)
		}
	}
}

func TestCompareSemverCore(t *testing.T) {
	a := parseSemverCore("0.6.1")
	b := parseSemverCore("0.7.0")
	if compareSemverCore(a, b) != -1 {
		t.Error("0.6.1 should be < 0.7.0")
	}
	if compareSemverCore(b, a) != 1 {
		t.Error("0.7.0 should be > 0.6.1")
	}
	if compareSemverCore(a, a) != 0 {
		t.Error("equal versions")
	}
	if compareSemverCore(parseSemverCore("0.8.0"), parseSemverCore(minViteWatchVersion)) != 0 {
		t.Error("0.8.0 should equal minViteWatchVersion")
	}
}
