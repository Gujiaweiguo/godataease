package service

import (
	"strings"
	"testing"
)

func TestInferSQLVariableDeType(t *testing.T) {
	cases := []struct {
		name     string
		types    []string
		expected int
	}{
		{name: "datetime", types: []string{"DATETIME"}, expected: 1},
		{name: "double", types: []string{"DOUBLE"}, expected: 3},
		{name: "bigint", types: []string{"BIGINT"}, expected: 2},
		{name: "text", types: []string{"TEXT"}, expected: 0},
	}

	for _, tc := range cases {
		if actual := inferSQLVariableDeType(tc.types); actual != tc.expected {
			t.Fatalf("%s expected %d, got %d", tc.name, tc.expected, actual)
		}
	}
}

func TestValidatePreviewSQL(t *testing.T) {
	if err := validatePreviewSQL("SELECT * FROM core_dataset_group"); err != nil {
		t.Fatalf("expected select sql valid, got error: %v", err)
	}

	if err := validatePreviewSQL("INSERT INTO x VALUES (1)"); err == nil {
		t.Fatal("expected insert sql to be rejected")
	}

	if err := validatePreviewSQL("SELECT 1; SELECT 2"); err == nil {
		t.Fatal("expected multi statement sql to be rejected")
	}
}

func TestParseFilterFieldIDs(t *testing.T) {
	ids := parseFilterFieldIDs(" 1,2,2,abc,0,-3, 4 ")
	if len(ids) != 3 {
		t.Fatalf("expected 3 ids, got %d", len(ids))
	}
	if ids[0] != 1 || ids[1] != 2 || ids[2] != 4 {
		t.Fatalf("unexpected ids: %#v", ids)
	}
}

func TestExtractFilterValues(t *testing.T) {
	values := extractFilterValues([]interface{}{"  A  ", "", nil, "A", 100, " 100 "})
	if len(values) != 2 {
		t.Fatalf("expected 2 values, got %d", len(values))
	}
	if values[0] != "A" || values[1] != "100" {
		t.Fatalf("unexpected values: %#v", values)
	}
}

func TestNormalizeEnumValueScientific(t *testing.T) {
	deType := 3
	normalized := normalizeEnumValue("1.23E+3", &deType)
	if normalized == "" {
		t.Fatal("expected non-empty normalized value")
	}
	if strings.ContainsAny(strings.ToUpper(normalized), "E") {
		t.Fatalf("expected non scientific notation, got %s", normalized)
	}
}
