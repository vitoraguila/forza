package forza

import (
	"testing"
)

func TestNewFunction_Empty(t *testing.T) {
	shape := NewFunction()
	if len(shape) != 0 {
		t.Errorf("expected empty shape, got %d entries", len(shape))
	}
}

func TestNewFunction_SingleProperty(t *testing.T) {
	shape := NewFunction(
		WithProperty("name", "the user name", true),
	)

	if len(shape) != 1 {
		t.Fatalf("expected 1 property, got %d", len(shape))
	}

	prop, ok := shape["name"]
	if !ok {
		t.Fatal("expected 'name' property to exist")
	}
	if prop.Description != "the user name" {
		t.Errorf("expected description 'the user name', got %q", prop.Description)
	}
	if !prop.Required {
		t.Error("expected property to be required")
	}
}

func TestNewFunction_MultipleProperties(t *testing.T) {
	shape := NewFunction(
		WithProperty("name", "the user name", true),
		WithProperty("age", "the user age", false),
		WithProperty("email", "the user email", true),
	)

	if len(shape) != 3 {
		t.Fatalf("expected 3 properties, got %d", len(shape))
	}

	tests := []struct {
		name     string
		desc     string
		required bool
	}{
		{"name", "the user name", true},
		{"age", "the user age", false},
		{"email", "the user email", true},
	}

	for _, tt := range tests {
		prop, ok := shape[tt.name]
		if !ok {
			t.Errorf("expected %q property to exist", tt.name)
			continue
		}
		if prop.Description != tt.desc {
			t.Errorf("%s: expected description %q, got %q", tt.name, tt.desc, prop.Description)
		}
		if prop.Required != tt.required {
			t.Errorf("%s: expected required=%v, got %v", tt.name, tt.required, prop.Required)
		}
	}
}

func TestWithProperty_Overwrite(t *testing.T) {
	shape := NewFunction(
		WithProperty("name", "first description", true),
		WithProperty("name", "second description", false),
	)

	if len(shape) != 1 {
		t.Fatalf("expected 1 property after overwrite, got %d", len(shape))
	}

	prop := shape["name"]
	if prop.Description != "second description" {
		t.Errorf("expected overwritten description, got %q", prop.Description)
	}
	if prop.Required {
		t.Error("expected required to be overwritten to false")
	}
}
