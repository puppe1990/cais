package store

import (
	"testing"

	"github.com/puppe1990/cais/pkg/cais/passwordreset"
	"github.com/puppe1990/cais/pkg/cais/session"
)

func TestStore_CreatePasswordResetToken_andFind(t *testing.T) {
	s, err := NewSQLiteStore(":memory:", "development")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })

	user, err := s.FindUserByEmail("demo@example.com")
	if err != nil {
		t.Fatal(err)
	}

	token, err := s.CreatePasswordResetToken(user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	gotID, ok := s.FindPasswordResetUserID(token)
	if !ok || gotID != user.ID {
		t.Fatalf("FindPasswordResetUserID = (%d, %v), want (%d, true)", gotID, ok, user.ID)
	}
}

func TestStore_FindPasswordResetUserID_rejectsUnknown(t *testing.T) {
	s := newTestStore(t)

	if _, ok := s.FindPasswordResetUserID("missing"); ok {
		t.Fatal("expected false for unknown token")
	}
}

func TestStore_ResetPasswordWithToken_updatesPasswordAndRevokesSessions(t *testing.T) {
	s, err := NewSQLiteStore(":memory:", "development")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })

	user, err := s.FindUserByEmail("demo@example.com")
	if err != nil {
		t.Fatal(err)
	}

	token, err := s.CreatePasswordResetToken(user.ID)
	if err != nil {
		t.Fatal(err)
	}

	sess := s.Sessions()
	sessionToken, err := sess.Create(user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := sess.Get(sessionToken); !ok {
		t.Fatal("session should exist before reset")
	}

	newHash, err := session.HashPassword("new-password-123")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.ResetPasswordWithToken(token, newHash); err != nil {
		t.Fatal(err)
	}

	updated, err := s.FindUserByEmail("demo@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if !session.VerifyPassword(updated.PasswordHash, "new-password-123") {
		t.Fatal("password was not updated")
	}
	if _, ok := s.FindPasswordResetUserID(token); ok {
		t.Fatal("token should be consumed")
	}
	if _, ok := sess.Get(sessionToken); ok {
		t.Fatal("sessions should be revoked after reset")
	}
}

func TestStore_CreatePasswordResetToken_invalidatesPrevious(t *testing.T) {
	s, err := NewSQLiteStore(":memory:", "development")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })

	user, err := s.FindUserByEmail("demo@example.com")
	if err != nil {
		t.Fatal(err)
	}

	first, err := s.CreatePasswordResetToken(user.ID)
	if err != nil {
		t.Fatal(err)
	}
	second, err := s.CreatePasswordResetToken(user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if first == second {
		t.Fatal("expected different tokens")
	}
	if _, ok := s.FindPasswordResetUserID(first); ok {
		t.Fatal("first token should be invalidated")
	}
	if _, ok := s.FindPasswordResetUserID(second); !ok {
		t.Fatal("second token should be valid")
	}
}

func TestStore_ResetPasswordWithToken_rejectsUnknownToken(t *testing.T) {
	s := newTestStore(t)

	hash, err := session.HashPassword("new-password-123")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.ResetPasswordWithToken("not-a-real-token", hash); err == nil {
		t.Fatal("expected error for unknown token")
	}
}

func TestStore_ResetPasswordWithToken_rejectsExpired(t *testing.T) {
	s, err := NewSQLiteStore(":memory:", "development")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })

	user, err := s.FindUserByEmail("demo@example.com")
	if err != nil {
		t.Fatal(err)
	}

	token, err := passwordreset.NewToken()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.DB().Exec(
		"INSERT INTO password_reset_tokens (token, user_id, expires_at) VALUES (?, ?, datetime('now', '-1 hour'))",
		token, user.ID,
	); err != nil {
		t.Fatal(err)
	}

	hash, err := session.HashPassword("new-password-123")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.ResetPasswordWithToken(token, hash); err == nil {
		t.Fatal("expected error for expired token")
	}
}
