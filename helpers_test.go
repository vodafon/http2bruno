package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNameBlockMap(t *testing.T) {
	tests := []struct {
		name     string
		blockNm  string
		heads    map[string]string
		contains []string
		isEmpty  bool
	}{
		{
			name:    "empty map returns empty string",
			blockNm: "headers",
			heads:   map[string]string{},
			isEmpty: true,
		},
		{
			name:    "single entry",
			blockNm: "headers",
			heads:   map[string]string{"Content-Type": "application/json"},
			contains: []string{
				"headers {",
				"Content-Type: application/json",
				"}",
			},
		},
		{
			name:    "multiple entries",
			blockNm: "meta",
			heads: map[string]string{
				"name": "test",
				"type": "http",
			},
			contains: []string{
				"meta {",
				"name: test",
				"type: http",
				"}",
			},
		},
		{
			name:    "vars block",
			blockNm: "vars",
			heads:   map[string]string{"host": "example.com"},
			contains: []string{
				"vars {",
				"host: example.com",
				"}",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NameBlockMap(tt.blockNm, tt.heads)

			if tt.isEmpty {
				if got != "" {
					t.Errorf("NameBlockMap() = %q, want empty string", got)
				}
				return
			}

			for _, want := range tt.contains {
				if !strings.Contains(got, want) {
					t.Errorf("NameBlockMap() = %q, should contain %q", got, want)
				}
			}
		})
	}
}

func TestNameBlockStrings(t *testing.T) {
	tests := []struct {
		name     string
		blockNm  string
		list     []string
		contains []string
		isEmpty  bool
	}{
		{
			name:    "empty list returns empty string",
			blockNm: "docs",
			list:    []string{},
			isEmpty: true,
		},
		{
			name:    "single item",
			blockNm: "docs",
			list:    []string{"- [ ] item1"},
			contains: []string{
				"docs {",
				"- [ ] item1",
				"}",
			},
		},
		{
			name:    "multiple items",
			blockNm: "docs",
			list:    []string{"- [ ] methods", "- [ ] params", "- [ ] headers"},
			contains: []string{
				"docs {",
				"- [ ] methods",
				"- [ ] params",
				"- [ ] headers",
				"}",
			},
		},
		{
			name:    "json block",
			blockNm: "json",
			list:    []string{`{"key": "value"}`},
			contains: []string{
				"json {",
				`{"key": "value"}`,
				"}",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NameBlockStrings(tt.blockNm, tt.list)

			if tt.isEmpty {
				if got != "" {
					t.Errorf("NameBlockStrings() = %q, want empty string", got)
				}
				return
			}

			for _, want := range tt.contains {
				if !strings.Contains(got, want) {
					t.Errorf("NameBlockStrings() = %q, should contain %q", got, want)
				}
			}
		})
	}
}

func TestBlockMap(t *testing.T) {
	tests := []struct {
		name     string
		m        map[string]string
		contains []string
		isEmpty  bool
	}{
		{
			name:    "empty map returns empty string",
			m:       map[string]string{},
			isEmpty: true,
		},
		{
			name: "single entry with indentation",
			m:    map[string]string{"key": "value"},
			contains: []string{
				"  key: value",
			},
		},
		{
			name: "multiple entries",
			m: map[string]string{
				"host":  "example.com",
				"proto": "https",
			},
			contains: []string{
				"  host: example.com",
				"  proto: https",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BlockMap(tt.m)

			if tt.isEmpty {
				if got != "" {
					t.Errorf("BlockMap() = %q, want empty string", got)
				}
				return
			}

			for _, want := range tt.contains {
				if !strings.Contains(got, want) {
					t.Errorf("BlockMap() = %q, should contain %q", got, want)
				}
			}
		})
	}
}

func TestBlockStrings(t *testing.T) {
	tests := []struct {
		name     string
		m        []string
		contains []string
		isEmpty  bool
	}{
		{
			name:    "empty list returns empty string",
			m:       []string{},
			isEmpty: true,
		},
		{
			name: "single item with indentation",
			m:    []string{"item1"},
			contains: []string{
				"  item1",
			},
		},
		{
			name: "multiple items",
			m:    []string{"line1", "line2", "line3"},
			contains: []string{
				"  line1",
				"  line2",
				"  line3",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BlockStrings(tt.m)

			if tt.isEmpty {
				if got != "" {
					t.Errorf("BlockStrings() = %q, want empty string", got)
				}
				return
			}

			for _, want := range tt.contains {
				if !strings.Contains(got, want) {
					t.Errorf("BlockStrings() = %q, should contain %q", got, want)
				}
			}
		})
	}
}

func TestDirFilesCount(t *testing.T) {
	tests := []struct {
		name        string
		setupFiles  []string // files to create
		setupDirs   []string // subdirectories to create
		expected    int
		nonExistent bool // test non-existent directory
	}{
		{
			name:       "empty directory",
			setupFiles: []string{},
			setupDirs:  []string{},
			expected:   0,
		},
		{
			name:       "single file",
			setupFiles: []string{"file1.txt"},
			setupDirs:  []string{},
			expected:   1,
		},
		{
			name:       "multiple files",
			setupFiles: []string{"file1.txt", "file2.bru", "file3.json"},
			setupDirs:  []string{},
			expected:   3,
		},
		{
			name:       "files with subdirectories - should not count subdirs",
			setupFiles: []string{"file1.txt", "file2.txt"},
			setupDirs:  []string{"subdir1", "subdir2"},
			expected:   2,
		},
		{
			name:       "only subdirectories",
			setupFiles: []string{},
			setupDirs:  []string{"subdir1", "subdir2", "subdir3"},
			expected:   0,
		},
		{
			name:        "non-existent directory returns 0",
			nonExistent: true,
			expected:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var dir string

			if tt.nonExistent {
				dir = "/non/existent/path/that/does/not/exist"
			} else {
				// Create temp directory
				tmpDir := t.TempDir()
				dir = tmpDir

				// Create files
				for _, f := range tt.setupFiles {
					fp := filepath.Join(tmpDir, f)
					if err := os.WriteFile(fp, []byte("test"), 0644); err != nil {
						t.Fatalf("failed to create file %q: %v", f, err)
					}
				}

				// Create subdirectories
				for _, d := range tt.setupDirs {
					dp := filepath.Join(tmpDir, d)
					if err := os.MkdirAll(dp, 0755); err != nil {
						t.Fatalf("failed to create directory %q: %v", d, err)
					}
				}
			}

			got := DirFilesCount(dir)
			if got != tt.expected {
				t.Errorf("DirFilesCount(%q) = %d, want %d", dir, got, tt.expected)
			}
		})
	}
}

func TestBlockMapIndentation(t *testing.T) {
	// Verify that BlockMap uses exactly 2 spaces for indentation
	m := map[string]string{"key": "value"}
	got := BlockMap(m)

	if !strings.HasPrefix(got, "  ") {
		t.Errorf("BlockMap() should start with 2 spaces, got %q", got)
	}
}

func TestBlockStringsIndentation(t *testing.T) {
	// Verify that BlockStrings uses exactly 2 spaces for indentation
	list := []string{"item"}
	got := BlockStrings(list)

	if !strings.HasPrefix(got, "  ") {
		t.Errorf("BlockStrings() should start with 2 spaces, got %q", got)
	}
}

func TestNameBlockMapFormat(t *testing.T) {
	// Verify the complete format of NameBlockMap
	m := map[string]string{"key": "value"}
	got := NameBlockMap("test", m)

	// Should start with "test {"
	if !strings.HasPrefix(got, "test {") {
		t.Errorf("NameBlockMap() should start with 'test {', got %q", got)
	}

	// Should end with "}\n"
	if !strings.HasSuffix(got, "}\n") {
		t.Errorf("NameBlockMap() should end with '}\\n', got %q", got)
	}
}

func TestNameBlockStringsFormat(t *testing.T) {
	// Verify the complete format of NameBlockStrings
	list := []string{"item"}
	got := NameBlockStrings("test", list)

	// Should start with "test {"
	if !strings.HasPrefix(got, "test {") {
		t.Errorf("NameBlockStrings() should start with 'test {', got %q", got)
	}

	// Should end with "}\n"
	if !strings.HasSuffix(got, "}\n") {
		t.Errorf("NameBlockStrings() should end with '}\\n', got %q", got)
	}
}
