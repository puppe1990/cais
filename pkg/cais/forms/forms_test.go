package forms

import (
	"strings"
	"testing"
)

func TestCSRFTokenField_rendersHiddenInput(t *testing.T) {
	got := string(CSRFTokenField("secret-token"))
	if !strings.Contains(got, `type="hidden"`) {
		t.Errorf("CSRFTokenField missing hidden input: %s", got)
	}
	if !strings.Contains(got, `name="csrf_token"`) {
		t.Errorf("CSRFTokenField missing field name: %s", got)
	}
	if !strings.Contains(got, `value="secret-token"`) {
		t.Errorf("CSRFTokenField missing token value: %s", got)
	}
}

func TestCSRFTokenField_escapesValue(t *testing.T) {
	got := string(CSRFTokenField(`"><script>`))
	if strings.Contains(got, "<script>") {
		t.Errorf("CSRFTokenField not escaped: %s", got)
	}
}

func TestFieldError_withError(t *testing.T) {
	errs := map[string]string{"email": "invalid email"}
	if got := FieldError(errs, "email"); got != "invalid email" {
		t.Errorf("FieldError() = %q, want %q", got, "invalid email")
	}
}

func TestFieldError_withoutError(t *testing.T) {
	errs := map[string]string{"email": "invalid email"}
	if got := FieldError(errs, "name"); got != "" {
		t.Errorf("FieldError() = %q, want empty", got)
	}
}

func TestFieldError_nilMap(t *testing.T) {
	if got := FieldError(nil, "email"); got != "" {
		t.Errorf("FieldError(nil) = %q, want empty", got)
	}
}

func TestFuncs_registersHelpers(t *testing.T) {
	funcs := Funcs()
	if _, ok := funcs["csrfField"]; !ok {
		t.Error("csrfField missing from Funcs()")
	}
	if _, ok := funcs["fieldError"]; !ok {
		t.Error("fieldError missing from Funcs()")
	}
}
