package console

import "testing"

func TestHistory_AddAndList(t *testing.T) {
	h := NewHistory()
	h.Add("store.Ping()")
	h.Add("sql SELECT 1")
	h.Add("exit")

	got := h.List()
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2 (exit excluded)", len(got))
	}
	if got[0] != "store.Ping()" || got[1] != "sql SELECT 1" {
		t.Fatalf("list = %#v", got)
	}
}

func TestHistory_At(t *testing.T) {
	h := NewHistory()
	h.Add("first")
	h.Add("second")

	if h.At(1) != "first" {
		t.Fatalf("At(1) = %q", h.At(1))
	}
	if h.At(2) != "second" {
		t.Fatalf("At(2) = %q", h.At(2))
	}
	if h.At(99) != "" {
		t.Fatalf("At(99) should be empty")
	}
}

func TestHistory_Last(t *testing.T) {
	h := NewHistory()
	if h.Last() != "" {
		t.Fatalf("Last() on empty = %q, want empty", h.Last())
	}

	h.Add("first")
	h.Add("second")
	if h.Last() != "second" {
		t.Fatalf("Last() = %q, want second", h.Last())
	}
}

func TestHistory_Lines(t *testing.T) {
	h := NewHistory()
	h.Add("one")
	h.Add("two")

	lines := h.Lines()
	if len(lines) != 2 || lines[0] != "one" || lines[1] != "two" {
		t.Fatalf("Lines() = %#v", lines)
	}
}
