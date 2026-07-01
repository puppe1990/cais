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
