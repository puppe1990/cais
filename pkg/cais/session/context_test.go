package session

import (
	"net/http/httptest"
	"testing"
)

func TestUserID_Context(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	if _, ok := UserID(req); ok {
		t.Error("expected no user before WithUserID")
	}

	req = WithUserID(req, 7)
	id, ok := UserID(req)
	if !ok || id != 7 {
		t.Fatalf("UserID() = (%d, %v), want (7, true)", id, ok)
	}
}
