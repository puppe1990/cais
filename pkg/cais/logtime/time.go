package logtime

import "time"

// Format returns a Rails-style timestamp: 2006-01-02 15:04:05 -0700
func Format(t time.Time) string {
	return t.Format("2006-01-02 15:04:05 -0700")
}

func Now() string {
	return Format(time.Now())
}
