package cli

import (
	"strings"
	"testing"
)

func TestParseFields_rejectsUnknownType(t *testing.T) {
	_, err := parseFields("title:strng")
	if err == nil {
		t.Fatal("expected error for unknown field type")
	}
	if !strings.Contains(err.Error(), "strng") {
		t.Errorf("error = %v, want mention of unknown type", err)
	}
}

func TestParseFields_optionalNullableSQL(t *testing.T) {
	fields, err := parseFields("title:string,notes:text?")
	if err != nil {
		t.Fatal(err)
	}
	if len(fields) != 2 {
		t.Fatalf("len = %d", len(fields))
	}
	notes := fields[1]
	if notes.Required {
		t.Error("notes should be optional")
	}
	if strings.Contains(notes.SQLType, "NOT NULL") {
		t.Errorf("optional notes SQLType = %q, want nullable", notes.SQLType)
	}
	if notes.GoType != "*string" {
		t.Errorf("optional notes GoType = %q, want *string", notes.GoType)
	}
}

func TestParseFields_referencesType(t *testing.T) {
	fields, err := parseFields("title:string,category_id:references")
	if err != nil {
		t.Fatal(err)
	}
	if len(fields) != 2 {
		t.Fatalf("len = %d, want 2", len(fields))
	}
	ref := fields[1]
	if ref.Name != "category_id" {
		t.Errorf("Name = %q, want category_id", ref.Name)
	}
	if ref.GoType != "int64" {
		t.Errorf("GoType = %q, want int64", ref.GoType)
	}
	if ref.Widget != "select" {
		t.Errorf("Widget = %q, want select", ref.Widget)
	}
	if ref.RefTable != "categories" {
		t.Errorf("RefTable = %q, want categories", ref.RefTable)
	}
	if ref.RefPascal != "Category" {
		t.Errorf("RefPascal = %q, want Category", ref.RefPascal)
	}
	if !strings.Contains(ref.SQLType, "REFERENCES categories(id)") {
		t.Errorf("SQLType = %q, want FK to categories", ref.SQLType)
	}
}

func TestParseFields_referencesOptional(t *testing.T) {
	fields, err := parseFields("category_id:references?")
	if err != nil {
		t.Fatal(err)
	}
	if fields[0].Required {
		t.Error("expected optional references field")
	}
	if fields[0].GoType != "*int64" {
		t.Errorf("GoType = %q, want *int64", fields[0].GoType)
	}
	if strings.Contains(fields[0].SQLType, "NOT NULL") {
		t.Errorf("SQLType = %q, want nullable", fields[0].SQLType)
	}
}

func TestParseFields_belongsToAlias(t *testing.T) {
	fields, err := parseFields("category:belongs_to")
	if err != nil {
		t.Fatal(err)
	}
	if fields[0].Name != "category_id" {
		t.Errorf("Name = %q, want category_id", fields[0].Name)
	}
	if fields[0].RefTable != "categories" {
		t.Errorf("RefTable = %q, want categories", fields[0].RefTable)
	}
}

func TestParseFields_referencesRequiresIDSuffix(t *testing.T) {
	_, err := parseFields("category:references")
	if err == nil {
		t.Fatal("expected error for references without _id suffix")
	}
}

func TestParseFields_optionalNullableInt(t *testing.T) {
	fields, err := parseFields("qty:int?")
	if err != nil {
		t.Fatal(err)
	}
	if fields[0].GoType != "*int64" {
		t.Errorf("GoType = %q, want *int64", fields[0].GoType)
	}
	if strings.Contains(fields[0].SQLType, "NOT NULL") {
		t.Errorf("SQLType = %q, want nullable", fields[0].SQLType)
	}
}

func TestParseFields(t *testing.T) {
	fields, err := parseFields("title:string,url:url,notes:text?")
	if err != nil {
		t.Fatal(err)
	}
	if len(fields) != 3 {
		t.Fatalf("len = %d", len(fields))
	}
	if fields[2].Required {
		t.Error("notes should be optional")
	}
}

func TestParseFields_DateType(t *testing.T) {
	fields, err := parseFields("title:string,due_date:date")
	if err != nil {
		t.Fatal(err)
	}
	if len(fields) != 2 {
		t.Fatalf("len = %d", len(fields))
	}
	if fields[1].GoType != "string" {
		t.Errorf("date GoType = %q, want string", fields[1].GoType)
	}
	if fields[1].HTMLType != "date" {
		t.Errorf("date HTMLType = %q, want date", fields[1].HTMLType)
	}
	if fields[1].SQLType != "TEXT NOT NULL" {
		t.Errorf("date SQLType = %q", fields[1].SQLType)
	}
}
