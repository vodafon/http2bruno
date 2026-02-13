package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDoCollection(t *testing.T) {
	tests := []struct {
		name       string
		collection string
		wantErr    bool
		errContain string
	}{
		{
			name:       "empty collection name returns error",
			collection: "",
			wantErr:    true,
			errContain: "collection name is required",
		},
		{
			name:       "whitespace only collection name returns error",
			collection: "   ",
			wantErr:    true,
			errContain: "collection name is required",
		},
		{
			name:       "valid collection name succeeds",
			collection: "test-api",
			wantErr:    false,
		},
		{
			name:       "collection with dots",
			collection: "api.example.com",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory and change to it
			tmpDir := t.TempDir()
			origDir, _ := os.Getwd()
			os.Chdir(tmpDir)
			defer os.Chdir(origDir)

			err := DoCollection(tt.collection)

			if (err != nil) != tt.wantErr {
				t.Errorf("DoCollection() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				if tt.errContain != "" && !strings.Contains(err.Error(), tt.errContain) {
					t.Errorf("DoCollection() error = %q, should contain %q", err.Error(), tt.errContain)
				}
				return
			}

			// Verify directory was created
			collDir := filepath.Join(tmpDir, strings.TrimSpace(tt.collection))
			if _, err := os.Stat(collDir); os.IsNotExist(err) {
				t.Errorf("DoCollection() did not create directory %q", collDir)
			}

			// Verify collection.bru was created
			collBru := filepath.Join(collDir, "collection.bru")
			if _, err := os.Stat(collBru); os.IsNotExist(err) {
				t.Errorf("DoCollection() did not create %q", collBru)
			}

			// Verify bruno.json was created
			brunoJSON := filepath.Join(collDir, "bruno.json")
			if _, err := os.Stat(brunoJSON); os.IsNotExist(err) {
				t.Errorf("DoCollection() did not create %q", brunoJSON)
			}

			// Verify environments directory was created
			envDir := filepath.Join(collDir, "environments")
			if _, err := os.Stat(envDir); os.IsNotExist(err) {
				t.Errorf("DoCollection() did not create %q", envDir)
			}

			// Verify base.bru was created in environments
			baseBru := filepath.Join(envDir, "base.bru")
			if _, err := os.Stat(baseBru); os.IsNotExist(err) {
				t.Errorf("DoCollection() did not create %q", baseBru)
			}
		})
	}
}

func TestDoCollectionFileContents(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	collName := "my-test-api"
	err := DoCollection(collName)
	if err != nil {
		t.Fatalf("DoCollection() error = %v", err)
	}

	// Check bruno.json content
	brunoJSON := filepath.Join(tmpDir, collName, "bruno.json")
	data, err := os.ReadFile(brunoJSON)
	if err != nil {
		t.Fatalf("Failed to read bruno.json: %v", err)
	}

	jsonStr := string(data)
	if !strings.Contains(jsonStr, `"name": "my-test-api"`) {
		t.Error("bruno.json should contain collection name")
	}
	if !strings.Contains(jsonStr, `"version": "1"`) {
		t.Error("bruno.json should contain version")
	}

	// Check collection.bru content
	collBru := filepath.Join(tmpDir, collName, "collection.bru")
	data, err = os.ReadFile(collBru)
	if err != nil {
		t.Fatalf("Failed to read collection.bru: %v", err)
	}

	bruStr := string(data)
	if !strings.Contains(bruStr, "headers {") {
		t.Error("collection.bru should contain headers block")
	}
	if !strings.Contains(bruStr, "User-Agent: {{ua}}") {
		t.Error("collection.bru should contain User-Agent header")
	}

	// Check base.bru content
	baseBru := filepath.Join(tmpDir, collName, "environments", "base.bru")
	data, err = os.ReadFile(baseBru)
	if err != nil {
		t.Fatalf("Failed to read base.bru: %v", err)
	}

	envStr := string(data)
	if !strings.Contains(envStr, "vars {") {
		t.Error("base.bru should contain vars block")
	}
	if !strings.Contains(envStr, "host:") {
		t.Error("base.bru should contain host variable")
	}
	if !strings.Contains(envStr, "proto:") {
		t.Error("base.bru should contain proto variable")
	}
}

func TestDoFolder(t *testing.T) {
	tests := []struct {
		name       string
		folder     string
		wantErr    bool
		errContain string
	}{
		{
			name:       "empty folder name returns error",
			folder:     "",
			wantErr:    true,
			errContain: "folder name is required",
		},
		{
			name:       "whitespace only folder name returns error",
			folder:     "   ",
			wantErr:    true,
			errContain: "folder name is required",
		},
		{
			name:    "valid folder name succeeds",
			folder:  "users",
			wantErr: false,
		},
		{
			name:    "nested folder path",
			folder:  "api/v1/users",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			err := DoFolder(tt.folder, tmpDir)

			if (err != nil) != tt.wantErr {
				t.Errorf("DoFolder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				if tt.errContain != "" && !strings.Contains(err.Error(), tt.errContain) {
					t.Errorf("DoFolder() error = %q, should contain %q", err.Error(), tt.errContain)
				}
				return
			}

			// Verify directory was created
			folderDir := filepath.Join(tmpDir, strings.TrimSpace(tt.folder))
			if _, err := os.Stat(folderDir); os.IsNotExist(err) {
				t.Errorf("DoFolder() did not create directory %q", folderDir)
			}

			// Verify folder.bru was created
			folderBru := filepath.Join(folderDir, "folder.bru")
			if _, err := os.Stat(folderBru); os.IsNotExist(err) {
				t.Errorf("DoFolder() did not create %q", folderBru)
			}
		})
	}
}

func TestDoFolderFileContents(t *testing.T) {
	tmpDir := t.TempDir()

	folderName := "my-folder"
	err := DoFolder(folderName, tmpDir)
	if err != nil {
		t.Fatalf("DoFolder() error = %v", err)
	}

	// Check folder.bru content
	folderBru := filepath.Join(tmpDir, folderName, "folder.bru")
	data, err := os.ReadFile(folderBru)
	if err != nil {
		t.Fatalf("Failed to read folder.bru: %v", err)
	}

	bruStr := string(data)
	if !strings.Contains(bruStr, "meta {") {
		t.Error("folder.bru should contain meta block")
	}
	if !strings.Contains(bruStr, "name: my-folder") {
		t.Error("folder.bru should contain folder name in meta")
	}
	if !strings.Contains(bruStr, "headers {") {
		t.Error("folder.bru should contain headers block")
	}
}

func TestDoStructure(t *testing.T) {
	tests := []struct {
		name       string
		collection string
		folder     string
		wantErr    bool
	}{
		{
			name:       "empty collection returns error",
			collection: "",
			folder:     "users",
			wantErr:    true,
		},
		{
			name:       "collection only (no folder)",
			collection: "test-api",
			folder:     "",
			wantErr:    false,
		},
		{
			name:       "collection with folder",
			collection: "test-api",
			folder:     "users",
			wantErr:    false,
		},
		{
			name:       "collection with whitespace folder (treated as empty)",
			collection: "test-api",
			folder:     "   ",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			origDir, _ := os.Getwd()
			os.Chdir(tmpDir)
			defer os.Chdir(origDir)

			err := DoStructure(tt.collection, tt.folder)

			if (err != nil) != tt.wantErr {
				t.Errorf("DoStructure() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return
			}

			// Verify collection was created
			collDir := filepath.Join(tmpDir, strings.TrimSpace(tt.collection))
			if _, err := os.Stat(collDir); os.IsNotExist(err) {
				t.Errorf("DoStructure() did not create collection directory %q", collDir)
			}

			// If folder specified, verify it was created
			folder := strings.TrimSpace(tt.folder)
			if folder != "" {
				folderDir := filepath.Join(collDir, folder)
				if _, err := os.Stat(folderDir); os.IsNotExist(err) {
					t.Errorf("DoStructure() did not create folder directory %q", folderDir)
				}

				folderBru := filepath.Join(folderDir, "folder.bru")
				if _, err := os.Stat(folderBru); os.IsNotExist(err) {
					t.Errorf("DoStructure() did not create folder.bru in %q", folderDir)
				}
			}
		})
	}
}

func TestDoCollectionIdempotent(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	// Create collection twice - should not error (MkdirAll is idempotent)
	err := DoCollection("test-api")
	if err != nil {
		t.Fatalf("First DoCollection() error = %v", err)
	}

	// Second call should also succeed (files will be overwritten)
	err = DoCollection("test-api")
	if err != nil {
		t.Fatalf("Second DoCollection() error = %v", err)
	}
}
