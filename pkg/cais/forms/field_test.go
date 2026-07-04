package forms

import (
	"strings"
	"testing"
)

func TestMakeField_setsErrorFromMap(t *testing.T) {
	f := MakeField("email", "Email", "bad@", "email", true, map[string]string{"email": "invalid email"})
	if f.Error != "invalid email" {
		t.Errorf("Error = %q, want invalid email", f.Error)
	}
	if f.Name != "email" || f.Label != "Email" || f.Value != "bad@" || f.Type != "email" || !f.Required {
		t.Errorf("unexpected field: %+v", f)
	}
}

func TestFieldInput_rendersTextInputAndError(t *testing.T) {
	html := string(FieldInput(MakeField("name", "Name", "Ada", "text", true, map[string]string{"name": "required"})))
	if !strings.Contains(html, `name="name"`) {
		t.Error("missing name attribute")
	}
	if !strings.Contains(html, `value="Ada"`) {
		t.Error("missing value")
	}
	if !strings.Contains(html, "required") {
		t.Error("missing required attribute")
	}
	if !strings.Contains(html, "required") && !strings.Contains(html, "Name") {
		t.Error("missing label")
	}
	if !strings.Contains(html, "required") || !strings.Contains(html, `class="text-red-600`) {
		t.Error("missing error paragraph")
	}
}

func TestFieldInput_textarea(t *testing.T) {
	f := MakeField("notes", "Notes", "hello", "textarea", false, nil)
	html := string(FieldInput(f))
	if !strings.Contains(html, "<textarea") || !strings.Contains(html, "hello</textarea>") {
		t.Errorf("expected textarea: %s", html)
	}
}

func TestFieldInput_float(t *testing.T) {
	html := string(FieldInput(MakeField("lat", "Latitude", "-25.42", "float", true, nil)))
	if !strings.Contains(html, `type="number"`) || !strings.Contains(html, `step="any"`) {
		t.Errorf("expected float number input: %s", html)
	}
	if !strings.Contains(html, `value="-25.42"`) {
		t.Error("missing float value")
	}
}

func TestFieldInput_checkbox(t *testing.T) {
	f := FieldData{Name: "active", Label: "Active", Type: "checkbox", Value: "true"}
	html := string(FieldInput(f))
	if !strings.Contains(html, `type="checkbox"`) || !strings.Contains(html, "checked") {
		t.Errorf("expected checked checkbox: %s", html)
	}
}

func TestFieldInput_escapesValue(t *testing.T) {
	html := string(FieldInput(MakeField("x", "X", `"><script>`, "text", false, nil)))
	if strings.Contains(html, "<script>") {
		t.Errorf("value not escaped: %s", html)
	}
}

func TestFieldPassword_rendersToggleButton(t *testing.T) {
	html := string(FieldPassword(MakeField("password", "Password", "", "password", true, nil)))
	for _, needle := range []string{
		`type="password"`,
		`name="password"`,
		`data-cais-password-toggle`,
		`data-cais-password-icon="show"`,
		`data-cais-password-icon="hide"`,
		`padding-right:2.5rem`,
		`data-cais-password-icon="hide"`,
	} {
		if !strings.Contains(html, needle) {
			t.Errorf("FieldPassword missing %q:\n%s", needle, html)
		}
	}
}

func TestFieldPassword_rendersFieldError(t *testing.T) {
	html := string(FieldPassword(MakeField("password", "Password", "", "password", true, map[string]string{
		"password": "too short",
	})))
	if !strings.Contains(html, "too short") {
		t.Errorf("expected error text: %s", html)
	}
}

func TestFuncs_registersFieldHelpers(t *testing.T) {
	funcs := Funcs()
	for _, name := range []string{"makeField", "fieldInput", "fieldPassword"} {
		if _, ok := funcs[name]; !ok {
			t.Errorf("%s missing from Funcs()", name)
		}
	}
}
