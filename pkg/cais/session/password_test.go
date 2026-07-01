package session

import "testing"

func TestHashPassword_AndVerify(t *testing.T) {
	hash, err := HashPassword("secret-pass")
	if err != nil {
		t.Fatal(err)
	}
	if hash == "secret-pass" {
		t.Error("hash should not equal plaintext")
	}
	if !VerifyPassword(hash, "secret-pass") {
		t.Error("expected password to verify")
	}
	if VerifyPassword(hash, "wrong") {
		t.Error("wrong password should not verify")
	}
}
