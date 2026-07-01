package forms

import (
	"strings"
	"testing"
)

func TestFieldSelect_rendersOptions(t *testing.T) {
	got := string(FieldSelect(SelectFieldData{
		Name:     "category_id",
		Label:    "Category",
		Value:    "2",
		Required: true,
		Options: []SelectOption{
			{Value: "1", Label: "Books"},
			{Value: "2", Label: "Music"},
		},
	}))
	if !strings.Contains(got, `<select`) {
		t.Errorf("missing select: %s", got)
	}
	if !strings.Contains(got, `name="category_id"`) {
		t.Errorf("missing name: %s", got)
	}
	if !strings.Contains(got, `value="2" selected`) {
		t.Errorf("missing selected option: %s", got)
	}
	if !strings.Contains(got, ">Books<") || !strings.Contains(got, ">Music<") {
		t.Errorf("missing option labels: %s", got)
	}
}

func TestFieldSelect_requiredPlaceholder(t *testing.T) {
	got := string(FieldSelect(SelectFieldData{
		Name:     "category_id",
		Label:    "Category",
		Required: true,
		Options:  []SelectOption{{Value: "1", Label: "Books"}},
	}))
	if !strings.Contains(got, "Select Category") {
		t.Errorf("missing placeholder: %s", got)
	}
}
