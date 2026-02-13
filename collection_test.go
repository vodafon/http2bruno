package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestDefaultBrunoJSON(t *testing.T) {
	tests := []struct {
		name         string
		collectionNm string
		checkFields  map[string]interface{}
	}{
		{
			name:         "basic collection",
			collectionNm: "my-api",
			checkFields: map[string]interface{}{
				"version": "1",
				"name":    "my-api",
				"type":    "collection",
			},
		},
		{
			name:         "collection with special chars",
			collectionNm: "api.example.com",
			checkFields: map[string]interface{}{
				"version": "1",
				"name":    "api.example.com",
				"type":    "collection",
			},
		},
		{
			name:         "empty name",
			collectionNm: "",
			checkFields: map[string]interface{}{
				"version": "1",
				"name":    "",
				"type":    "collection",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DefaultBrunoJSON(tt.collectionNm)

			// Parse as JSON
			var result map[string]interface{}
			if err := json.Unmarshal([]byte(got), &result); err != nil {
				t.Fatalf("DefaultBrunoJSON() returned invalid JSON: %v", err)
			}

			// Check expected fields
			for k, v := range tt.checkFields {
				if result[k] != v {
					t.Errorf("DefaultBrunoJSON() field %q = %v, want %v", k, result[k], v)
				}
			}

			// Check ignore array
			ignoreField, ok := result["ignore"]
			if !ok {
				t.Errorf("DefaultBrunoJSON() missing 'ignore' field")
				return
			}

			ignoreArr, ok := ignoreField.([]interface{})
			if !ok {
				t.Errorf("DefaultBrunoJSON() 'ignore' field is not an array")
				return
			}

			// Check that node_modules and .git are in ignore
			ignoreStrs := make([]string, len(ignoreArr))
			for i, v := range ignoreArr {
				ignoreStrs[i] = v.(string)
			}

			hasNodeModules := false
			hasGit := false
			for _, v := range ignoreStrs {
				if v == "node_modules" {
					hasNodeModules = true
				}
				if v == ".git" {
					hasGit = true
				}
			}

			if !hasNodeModules {
				t.Error("DefaultBrunoJSON() ignore array should contain 'node_modules'")
			}
			if !hasGit {
				t.Error("DefaultBrunoJSON() ignore array should contain '.git'")
			}
		})
	}
}

func TestDefaultBrunoJSONFormat(t *testing.T) {
	got := DefaultBrunoJSON("test")

	// Should be pretty-printed (indented)
	if !strings.Contains(got, "\n") {
		t.Error("DefaultBrunoJSON() should be pretty-printed with newlines")
	}

	// Should start with {
	if !strings.HasPrefix(got, "{") {
		t.Errorf("DefaultBrunoJSON() should start with '{', got %q", got[:10])
	}

	// Should end with }
	if !strings.HasSuffix(got, "}") {
		t.Errorf("DefaultBrunoJSON() should end with '}', got %q", got[len(got)-10:])
	}
}

func TestDefaultCollectionBru(t *testing.T) {
	got := DefaultCollectionBru()

	// Should contain headers block
	if !strings.Contains(got, "headers {") {
		t.Error("DefaultCollectionBru() should contain 'headers {'")
	}

	// Should contain User-Agent with template variable
	if !strings.Contains(got, "User-Agent: {{ua}}") {
		t.Error("DefaultCollectionBru() should contain 'User-Agent: {{ua}}'")
	}

	// Should end with closing brace
	if !strings.HasSuffix(got, "}\n") {
		t.Error("DefaultCollectionBru() should end with '}\\n'")
	}
}

func TestBrunoJSONStruct(t *testing.T) {
	// Test the BrunoJSON struct directly
	bj := BrunoJSON{
		Version: "1",
		Name:    "test-collection",
		Type:    "collection",
		Ignore:  []string{"node_modules", ".git", "custom"},
	}

	data, err := json.Marshal(bj)
	if err != nil {
		t.Fatalf("Failed to marshal BrunoJSON: %v", err)
	}

	// Unmarshal and verify
	var result BrunoJSON
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal BrunoJSON: %v", err)
	}

	if result.Version != bj.Version {
		t.Errorf("Version = %q, want %q", result.Version, bj.Version)
	}
	if result.Name != bj.Name {
		t.Errorf("Name = %q, want %q", result.Name, bj.Name)
	}
	if result.Type != bj.Type {
		t.Errorf("Type = %q, want %q", result.Type, bj.Type)
	}
	if len(result.Ignore) != len(bj.Ignore) {
		t.Errorf("Ignore length = %d, want %d", len(result.Ignore), len(bj.Ignore))
	}
}

func TestBrunoJSONTags(t *testing.T) {
	// Verify JSON field names are correct
	bj := BrunoJSON{
		Version: "1",
		Name:    "test",
		Type:    "collection",
		Ignore:  []string{},
	}

	data, _ := json.Marshal(bj)
	jsonStr := string(data)

	// Check that JSON uses correct field names
	expectedFields := []string{`"version"`, `"name"`, `"type"`, `"ignore"`}
	for _, field := range expectedFields {
		if !strings.Contains(jsonStr, field) {
			t.Errorf("JSON should contain field %s, got %s", field, jsonStr)
		}
	}
}
