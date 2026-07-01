package jobs

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// CronMatches reports whether expr (5-field cron) matches the given instant.
// Supports *, N, and */N for each field (minute hour dom month dow).
func CronMatches(expr string, t time.Time) (bool, error) {
	fields := strings.Fields(expr)
	if len(fields) != 5 {
		return false, fmt.Errorf("cron: want 5 fields, got %d", len(fields))
	}
	utc := t.UTC()
	checks := []struct {
		field string
		value int
		max   int
	}{
		{fields[0], utc.Minute(), 59},
		{fields[1], utc.Hour(), 23},
		{fields[2], utc.Day(), 31},
		{fields[3], int(utc.Month()), 12},
		{fields[4], int(utc.Weekday()), 6},
	}
	for _, c := range checks {
		ok, err := cronFieldMatches(c.field, c.value, c.max)
		if err != nil {
			return false, err
		}
		if !ok {
			return false, nil
		}
	}
	return true, nil
}

// ShouldRunRecurring returns true when cron matches now and the task did not
// already run in the same UTC minute.
func ShouldRunRecurring(expr string, lastRun *time.Time, now time.Time) (bool, error) {
	match, err := CronMatches(expr, now)
	if err != nil || !match {
		return match, err
	}
	if lastRun == nil {
		return true, nil
	}
	return lastRun.UTC().Truncate(time.Minute).Before(now.UTC().Truncate(time.Minute)), nil
}

// ValidateCron checks that expr is a valid 5-field cron expression.
func ValidateCron(expr string) error {
	_, err := CronMatches(expr, time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC))
	return err
}

func cronFieldMatches(field string, value int, max int) (bool, error) {
	if field == "*" {
		return true, nil
	}
	if strings.HasPrefix(field, "*/") {
		step, err := strconv.Atoi(field[2:])
		if err != nil || step < 1 {
			return false, fmt.Errorf("cron: invalid step %q", field)
		}
		return value%step == 0, nil
	}
	n, err := strconv.Atoi(field)
	if err != nil {
		return false, fmt.Errorf("cron: invalid field %q", field)
	}
	if n < 0 || n > max {
		return false, fmt.Errorf("cron: %q out of range 0-%d", field, max)
	}
	return n == value, nil
}
