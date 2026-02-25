package main

import (
	"strings"
	"testing"
)

func TestDefaultFolderBru(t *testing.T) {
	tests := []struct {
		name       string
		folderName string
		contains   []string
	}{
	{
		name:       "basic folder",
		folderName: "api",
		contains: []string{
			"meta {",
			"name: api",
		},
	},
	{
		name:       "folder with special chars",
		folderName: "users-v2",
		contains: []string{
			"meta {",
			"name: users-v2",
		},
	},
	{
		name:       "empty folder name",
		folderName: "",
		contains: []string{
			"meta {",
			"name:",
		},
	},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DefaultFolderBru(tt.folderName)

			for _, want := range tt.contains {
				if !strings.Contains(got, want) {
					t.Errorf("DefaultFolderBru(%q) = %q, should contain %q", tt.folderName, got, want)
				}
			}
		})
	}
}

func TestDefaultFolderBruStructure(t *testing.T) {
	got := DefaultFolderBru("test")

	// Should have meta block
	metaIdx := strings.Index(got, "meta {")

	if metaIdx == -1 {
		t.Error("DefaultFolderBru() should contain 'meta {'")
	}

	// Should NOT have headers block (moved to collection)
	headersIdx := strings.Index(got, "headers {")
	if headersIdx != -1 {
		t.Error("DefaultFolderBru() should not contain 'headers {' (moved to collection)")
	}
}

func TestDefaultFolderBruNoHeaders(t *testing.T) {
	got := DefaultFolderBru("test")

	// Should NOT have Cookie or Authorization headers (moved to collection)
	if strings.Contains(got, "Cookie:") {
		t.Error("DefaultFolderBru() should not contain Cookie header (moved to collection)")
	}
	if strings.Contains(got, "Authorization:") {
		t.Error("DefaultFolderBru() should not contain Authorization header (moved to collection)")
	}
}
