package cli

import "testing"

func TestToPlural(t *testing.T) {
	tests := map[string]string{
		"dish":     "dishes",
		"recipe":   "recipes",
		"category": "categories",
		"task":     "tasks",
		"class":    "classes",
		"box":      "boxes",
		"bookmark": "bookmarks",
		"recipes":  "recipes",
		"status":   "statuses",
	}
	for in, want := range tests {
		if got := toPlural(in); got != want {
			t.Errorf("toPlural(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestDataForResource_PluralPascal(t *testing.T) {
	data := dataForResource("dish")
	if data.Plural != "dishes" {
		t.Errorf("Plural = %q, want dishes", data.Plural)
	}
	if data.PluralPascal != "Dishes" {
		t.Errorf("PluralPascal = %q, want Dishes", data.PluralPascal)
	}
}

func TestNames(t *testing.T) {
	data := dataForHandler("user_settings")
	if data.Pascal != "UserSettings" {
		t.Errorf("Pascal = %q", data.Pascal)
	}
	if data.Snake != "user_settings" {
		t.Errorf("Snake = %q", data.Snake)
	}
}
