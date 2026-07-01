package jobs

import (
	"testing"
	"time"
)

func TestCronMatches_dailyAt3am(t *testing.T) {
	tm := time.Date(2026, 7, 1, 3, 0, 0, 0, time.UTC)
	ok, err := CronMatches("0 3 * * *", tm)
	if err != nil || !ok {
		t.Fatalf("match = %v, err = %v", ok, err)
	}
	tm = tm.Add(time.Hour)
	ok, err = CronMatches("0 3 * * *", tm)
	if err != nil || ok {
		t.Fatalf("4am should not match, ok=%v err=%v", ok, err)
	}
}

func TestCronMatches_every15Minutes(t *testing.T) {
	for _, min := range []int{0, 15, 30, 45} {
		tm := time.Date(2026, 7, 1, 12, min, 0, 0, time.UTC)
		ok, err := CronMatches("*/15 * * * *", tm)
		if err != nil || !ok {
			t.Fatalf("minute %d: ok=%v err=%v", min, ok, err)
		}
	}
	tm := time.Date(2026, 7, 1, 12, 7, 0, 0, time.UTC)
	ok, _ := CronMatches("*/15 * * * *", tm)
	if ok {
		t.Fatal("minute 7 should not match */15")
	}
}

func TestShouldRunRecurring_sameMinute(t *testing.T) {
	now := time.Date(2026, 7, 1, 3, 0, 0, 0, time.UTC)
	last := now
	ok, err := ShouldRunRecurring("0 3 * * *", &last, now)
	if err != nil || ok {
		t.Fatalf("should not rerun same minute: ok=%v err=%v", ok, err)
	}
}
