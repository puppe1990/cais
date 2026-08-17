package dotenv

import "testing"

func TestParse_keyValueAndComments(t *testing.T) {
	got := Parse([]byte("LOCALE=pt\n# comment\n\nAPP_URL=http://127.0.0.1:8081\n"))
	if got["LOCALE"] != "pt" {
		t.Fatalf("LOCALE = %q, want pt", got["LOCALE"])
	}
	if got["APP_URL"] != "http://127.0.0.1:8081" {
		t.Fatalf("APP_URL = %q", got["APP_URL"])
	}
	if _, ok := got["# comment"]; ok {
		t.Fatal("comment line should be ignored")
	}
}

func TestParse_trimsSpacesAndSkipsMalformed(t *testing.T) {
	got := Parse([]byte("  ENV = production  \nnot-a-pair\n"))
	if got["ENV"] != "production" {
		t.Fatalf("ENV = %q, want production", got["ENV"])
	}
	if len(got) != 1 {
		t.Fatalf("got %d keys, want 1: %v", len(got), got)
	}
}
