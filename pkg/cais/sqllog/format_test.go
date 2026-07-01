package sqllog

import "testing"

func TestOperationLabel(t *testing.T) {
	tests := []struct {
		query string
		want  string
	}{
		{"SELECT * FROM users WHERE email = ?", "User Load"},
		{"select id from exercises", "Exercise Load"},
		{"INSERT INTO sessions (token, user_id) VALUES (?, ?)", "Session Create"},
		{"UPDATE users SET display_name = ? WHERE id = ?", "User Update"},
		{"DELETE FROM sessions WHERE token = ?", "Session Destroy"},
		{"SELECT COUNT(*) FROM routines", "Routine Load"},
	}

	for _, tc := range tests {
		if got := operationLabel(tc.query); got != tc.want {
			t.Errorf("operationLabel(%q) = %q, want %q", tc.query, got, tc.want)
		}
	}
}

func TestFormatArgs(t *testing.T) {
	got := formatArgs([]any{"demo@pulsefit.local", int64(1)})
	if got != `["demo@pulsefit.local", 1]` {
		t.Fatalf("formatArgs = %q", got)
	}
}
