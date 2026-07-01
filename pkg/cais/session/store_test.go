package session

import "testing"

func TestMemoryStore_CreateGetDelete(t *testing.T) {
	store := NewMemoryStore()

	token, err := store.Create(42)
	if err != nil {
		t.Fatal(err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	id, ok := store.Get(token)
	if !ok || id != 42 {
		t.Fatalf("Get() = (%d, %v), want (42, true)", id, ok)
	}

	store.Delete(token)
	if _, ok := store.Get(token); ok {
		t.Error("session should be deleted")
	}
}

func TestMemoryStore_Get_UnknownToken(t *testing.T) {
	store := NewMemoryStore()
	if _, ok := store.Get("nope"); ok {
		t.Error("unknown token should not be found")
	}
}
