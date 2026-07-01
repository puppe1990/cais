package validate

import (
	"testing"
)

func TestRequired_MissingField(t *testing.T) {
	err := Required(map[string]string{"title": ""}, "title")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRequired_AllPresent(t *testing.T) {
	err := Required(map[string]string{"title": "Hi", "url": "https://x.com"}, "title", "url")
	if err != nil {
		t.Fatal(err)
	}
}

func TestURL_Valid(t *testing.T) {
	if err := URL("https://example.com"); err != nil {
		t.Fatal(err)
	}
}

func TestURL_Invalid(t *testing.T) {
	if err := URL("not-a-url"); err == nil {
		t.Fatal("expected error")
	}
}
