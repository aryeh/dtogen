package parser

import (
	"testing"
)

func TestParse(t *testing.T) {
	// Point to the testdata directory relative to where the test runs
	// Since we are in internal/parser, we need to go up two levels to find testdata
	// But packages.Load with a relative path works from the CWD or relative to the module.
	// Let's try using the absolute path or module path if possible, but for simplicity in this env:
	// We will run the test from the project root.

	info, err := Parse("../../testdata", "User")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if info.Name != "User" {
		t.Errorf("Expected struct name User, got %s", info.Name)
	}

	expectedFields := map[string]string{
		"ID":        "int",
		"Username":  "string",
		"Email":     "string",
		"IsActive":  "bool",
		"CreatedAt": "string",
	}

	if len(info.Fields) != len(expectedFields) {
		t.Errorf("Expected %d fields, got %d", len(expectedFields), len(info.Fields))
	}

	for _, f := range info.Fields {
		expectedType, ok := expectedFields[f.Name]
		if !ok {
			t.Errorf("Unexpected field found: %s", f.Name)
			continue
		}
		if f.Type != expectedType {
			t.Errorf("Field %s: expected type %s, got %s", f.Name, expectedType, f.Type)
		}
	}
}
