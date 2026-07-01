package validate

import "testing"

func TestFieldErrors_AddAndFirst(t *testing.T) {
	var errs FieldErrors
	errs.Add("email", "email is invalid")
	errs.Add("name", "name is required")

	if errs.First() != "email is invalid" {
		t.Errorf("First() = %q", errs.First())
	}
	if errs.Has("email") != true {
		t.Error("Has(email) = false, want true")
	}
	if len(errs) != 2 {
		t.Fatalf("len = %d, want 2", len(errs))
	}
}
