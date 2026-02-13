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
				"headers {",
				"Cookie: {{cook}}",
				"Authorization: Bearer {{token}}",
			},
		},
		{
			name:       "folder with special chars",
			folderName: "users-v2",
			contains: []string{
				"meta {",
				"name: users-v2",
				"headers {",
			},
		},
		{
			name:       "empty folder name",
			folderName: "",
			contains: []string{
				"meta {",
				"name:",
				"headers {",
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

	// Should have meta block before headers block
	metaIdx := strings.Index(got, "meta {")
	headersIdx := strings.Index(got, "headers {")

	if metaIdx == -1 {
		t.Error("DefaultFolderBru() should contain 'meta {'")
	}
	if headersIdx == -1 {
		t.Error("DefaultFolderBru() should contain 'headers {'")
	}
	if metaIdx > headersIdx {
		t.Error("DefaultFolderBru() meta block should come before headers block")
	}
}

func TestDefaultFolderBruHeaders(t *testing.T) {
	got := DefaultFolderBru("test")

	// Should have both Cookie and Authorization headers
	if !strings.Contains(got, "Cookie:") {
		t.Error("DefaultFolderBru() should contain Cookie header")
	}
	if !strings.Contains(got, "Authorization:") {
		t.Error("DefaultFolderBru() should contain Authorization header")
	}

	// Should use template variables
	if !strings.Contains(got, "{{cook}}") {
		t.Error("DefaultFolderBru() should use {{cook}} template variable")
	}
	if !strings.Contains(got, "{{token}}") {
		t.Error("DefaultFolderBru() should use {{token}} template variable")
	}

	// Authorization should use Bearer scheme
	if !strings.Contains(got, "Bearer {{token}}") {
		t.Error("DefaultFolderBru() should use Bearer scheme for Authorization")
	}
}

func TestDefaultFolderBruNewline(t *testing.T) {
	got := DefaultFolderBru("test")

	// Should have a newline between meta and headers blocks
	// The structure is: meta { ... }\n\nheaders { ... }
	if !strings.Contains(got, "}\n\nheaders") {
		t.Error("DefaultFolderBru() should have empty line between meta and headers blocks")
	}
}
