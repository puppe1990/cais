package pagination

import "testing"

func TestNew_normalizesPageAndPerPage(t *testing.T) {
	m := New(0, 0, 100)
	if m.Page != 1 {
		t.Errorf("Page = %d, want 1", m.Page)
	}
	if m.PerPage != 25 {
		t.Errorf("PerPage = %d, want 25", m.PerPage)
	}
	if !m.HasNext {
		t.Error("expected HasNext")
	}
	if m.HasPrev {
		t.Error("unexpected HasPrev on first page")
	}
}

func TestNew_lastPage(t *testing.T) {
	m := New(4, 25, 100)
	if m.HasNext {
		t.Error("unexpected HasNext on last page")
	}
	if !m.HasPrev {
		t.Error("expected HasPrev")
	}
	if m.PrevPage != 3 || m.NextPage != 5 {
		t.Errorf("PrevPage=%d NextPage=%d", m.PrevPage, m.NextPage)
	}
}

func TestOffset(t *testing.T) {
	if got := Offset(3, 25); got != 50 {
		t.Errorf("Offset = %d, want 50", got)
	}
	if got := Offset(0, 0); got != 0 {
		t.Errorf("Offset(0,0) = %d, want 0", got)
	}
}
