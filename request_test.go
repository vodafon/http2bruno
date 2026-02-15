package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestBodyTypeName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"formUrlEncoded converts to form-urlencoded", "formUrlEncoded", "form-urlencoded"},
		{"multipartForm converts to multipart-form", "multipartForm", "multipart-form"},
		{"json returns json", "json", "json"},
		{"xml returns xml", "xml", "xml"},
		{"text returns text", "text", "text"},
		{"none returns none", "none", "none"},
		{"unknown type returns itself", "unknown", "unknown"},
		{"empty string returns empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BodyTypeName(tt.input)
			if got != tt.expected {
				t.Errorf("BodyTypeName(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestBodyTypeFromContentType(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		expected    string
		wantErr     bool
	}{
		{"empty content type returns none", "", "none", false},
		{"application/json returns json", "application/json", "json", false},
		{"application/json with charset", "application/json; charset=utf-8", "json", false},
		{"application/xml returns xml", "application/xml", "xml", false},
		{"text/xml returns xml", "text/xml", "xml", false},
		{"text/plain returns text", "text/plain", "text", false},
		{"multipart/form-data returns multipartForm", "multipart/form-data", "multipartForm", false},
		{"multipart/form-data with boundary", "multipart/form-data; boundary=----WebKitFormBoundary", "multipartForm", false},
		{"application/x-www-form-urlencoded returns formUrlEncoded", "application/x-www-form-urlencoded", "formUrlEncoded", false},
		{"unsupported content type returns error", "application/octet-stream", "", true},
		{"invalid content type returns error", "invalid/;/type", "", true},
		{"text/html unsupported", "text/html", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BodyTypeFromContentType(tt.contentType)
			if (err != nil) != tt.wantErr {
				t.Errorf("BodyTypeFromContentType(%q) error = %v, wantErr %v", tt.contentType, err, tt.wantErr)
				return
			}
			if got != tt.expected {
				t.Errorf("BodyTypeFromContentType(%q) = %q, want %q", tt.contentType, got, tt.expected)
			}
		})
	}
}

func TestParseBodyUrlEncoded(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		expected map[string]string
	}{
		{
			name:     "empty body returns empty map",
			body:     "",
			expected: map[string]string{},
		},
		{
			name:     "single key-value pair",
			body:     "key=value",
			expected: map[string]string{"key": "value"},
		},
		{
			name:     "multiple key-value pairs",
			body:     "key1=value1&key2=value2",
			expected: map[string]string{"key1": "value1", "key2": "value2"},
		},
		{
			name:     "URL encoded values",
			body:     "name=John%20Doe&email=john%40example.com",
			expected: map[string]string{"name": "John Doe", "email": "john@example.com"},
		},
		{
			name:     "key without value",
			body:     "key",
			expected: map[string]string{"key": ""},
		},
		{
			name:     "key with empty value",
			body:     "key=",
			expected: map[string]string{"key": ""},
		},
		{
			name:     "multiple keys some without values",
			body:     "key1=value1&key2&key3=value3",
			expected: map[string]string{"key1": "value1", "key2": "", "key3": "value3"},
		},
		{
			name:     "value with equals sign",
			body:     "equation=1%2B1=2",
			expected: map[string]string{"equation": "1+1=2"},
		},
		{
			name:     "special characters encoded",
			body:     "data=%7B%22id%22%3A1%7D",
			expected: map[string]string{"data": `{"id":1}`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseBodyUrlEncoded(tt.body)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("ParseBodyUrlEncoded(%q) = %v, want %v", tt.body, got, tt.expected)
			}
		})
	}
}

func TestExtractBoundary(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		expected string
	}{
		{
			name:     "boundary with CRLF line ending",
			body:     "--boundary123\r\nContent-Disposition: form-data; name=\"field\"\r\n\r\nvalue",
			expected: "boundary123",
		},
		{
			name:     "boundary with LF line ending",
			body:     "--boundary456\r\nContent-Disposition: form-data; name=\"field\"\r\n\r\nvalue",
			expected: "boundary456",
		},
		{
			name:     "WebKit style boundary",
			body:     "----WebKitFormBoundary7MA4YWxkTrZu0gW\r\nContent-Disposition: form-data; name=\"file\"",
			expected: "--WebKitFormBoundary7MA4YWxkTrZu0gW",
		},
		{
			name:     "no boundary prefix returns empty",
			body:     "Content-Disposition: form-data",
			expected: "",
		},
		{
			name:     "empty body returns empty",
			body:     "",
			expected: "",
		},
		{
			name:     "only dashes",
			body:     "--",
			expected: "",
		},
		{
			name:     "boundary with special chars",
			body:     "--abc-123_XYZ\r\ndata",
			expected: "abc-123_XYZ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractBoundary(tt.body)
			if got != tt.expected {
				t.Errorf("extractBoundary() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestParseBodyMultipartForm(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		expected map[string]string
	}{
		{
			name:     "empty body returns empty map",
			body:     "",
			expected: map[string]string{},
		},
		{
			name: "single field",
			body: "--boundary\r\n" +
				"Content-Disposition: form-data; name=\"field1\"\r\n\r\n" +
				"value1\r\n" +
				"--boundary--\r\n",
			expected: map[string]string{"field1": "value1"},
		},
		{
			name: "multiple fields",
			body: "--boundary\r\n" +
				"Content-Disposition: form-data; name=\"field1\"\r\n\r\n" +
				"value1\r\n" +
				"--boundary\r\n" +
				"Content-Disposition: form-data; name=\"field2\"\r\n\r\n" +
				"value2\r\n" +
				"--boundary--\r\n",
			expected: map[string]string{"field1": "value1", "field2": "value2"},
		},
		{
			name: "field with special characters in value",
			body: "--boundary\r\n" +
				"Content-Disposition: form-data; name=\"json\"\r\n\r\n" +
				"{\"key\": \"value\"}\r\n" +
				"--boundary--\r\n",
			expected: map[string]string{"json": "{\"key\": \"value\"}"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseBodyMultipartForm(tt.body)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("ParseBodyMultipartForm() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestPathToName(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "simple path",
			path:     "api/users",
			expected: "api-users",
		},
		{
			name:     "path with single variable",
			path:     "api/users/{{user_id}}",
			expected: "api-users-USER_ID",
		},
		{
			name:     "path with multiple variables",
			path:     "api/{{org}}/users/{{user_id}}",
			expected: "api-ORG-users-USER_ID",
		},
		{
			name:     "path with simple variable",
			path:     "api/items/{{id}}",
			expected: "api-items-ID",
		},
		{
			name:     "empty path",
			path:     "",
			expected: "",
		},
		{
			name:     "single segment",
			path:     "api",
			expected: "api",
		},
		{
			name:     "path with numbers",
			path:     "api/v2/users",
			expected: "api-v2-users",
		},
		{
			name:     "nested path",
			path:     "api/users/posts/comments",
			expected: "api-users-posts-comments",
		},
		{
			name:     "variable with underscore",
			path:     "api/{{account_name}}/settings",
			expected: "api-ACCOUNT_NAME-settings",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pathToName(tt.path)
			if got != tt.expected {
				t.Errorf("pathToName(%q) = %q, want %q", tt.path, got, tt.expected)
			}
		})
	}
}

func TestParseRawRequest(t *testing.T) {
	tests := []struct {
		name       string
		rawRequest string
		wantMethod string
		wantPath   string
		wantHost   string
		wantErr    bool
	}{
		{
			name:       "simple GET request",
			rawRequest: "GET /api/users HTTP/1.1\r\nHost: example.com\r\n\r\n",
			wantMethod: "GET",
			wantPath:   "/api/users",
			wantHost:   "example.com",
			wantErr:    false,
		},
		{
			name:       "POST request with headers",
			rawRequest: "POST /api/users HTTP/1.1\r\nHost: example.com\r\nContent-Type: application/json\r\n\r\n",
			wantMethod: "POST",
			wantPath:   "/api/users",
			wantHost:   "example.com",
			wantErr:    false,
		},
		{
			name:       "request with comment lines",
			rawRequest: "# This is a comment\n# Another comment\nGET /api/data HTTP/1.1\r\nHost: test.com\r\n\r\n",
			wantMethod: "GET",
			wantPath:   "/api/data",
			wantHost:   "test.com",
			wantErr:    false,
		},
		{
			name:       "request with multiple comment lines",
			rawRequest: "# Comment 1\n# Comment 2\n# Comment 3\nPUT /api/resource HTTP/1.1\r\nHost: api.example.com\r\n\r\n",
			wantMethod: "PUT",
			wantPath:   "/api/resource",
			wantHost:   "api.example.com",
			wantErr:    false,
		},
		{
			name:       "DELETE request",
			rawRequest: "DELETE /api/users/123 HTTP/1.1\r\nHost: example.com\r\n\r\n",
			wantMethod: "DELETE",
			wantPath:   "/api/users/123",
			wantHost:   "example.com",
			wantErr:    false,
		},
		{
			name:       "request with query string",
			rawRequest: "GET /api/search?q=test&page=1 HTTP/1.1\r\nHost: example.com\r\n\r\n",
			wantMethod: "GET",
			wantPath:   "/api/search",
			wantHost:   "example.com",
			wantErr:    false,
		},
		{
			name:       "invalid request",
			rawRequest: "INVALID REQUEST",
			wantMethod: "",
			wantPath:   "",
			wantHost:   "",
			wantErr:    true,
		},
		{
			name:       "empty request",
			rawRequest: "",
			wantMethod: "",
			wantPath:   "",
			wantHost:   "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := ParseRawRequest([]byte(tt.rawRequest))
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRawRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if req.Method != tt.wantMethod {
				t.Errorf("ParseRawRequest() method = %q, want %q", req.Method, tt.wantMethod)
			}
			if req.URL.Path != tt.wantPath {
				t.Errorf("ParseRawRequest() path = %q, want %q", req.URL.Path, tt.wantPath)
			}
			if req.Host != tt.wantHost {
				t.Errorf("ParseRawRequest() host = %q, want %q", req.Host, tt.wantHost)
			}
		})
	}
}

func TestFindRequestFolder(t *testing.T) {
	tests := []struct {
		name            string
		setupDirs       []string // directories to create
		basedir         string
		path            string
		expectedDir     string
		expectedRemains string
	}{
		{
			name:            "empty path returns basedir",
			setupDirs:       []string{},
			basedir:         "",
			path:            "",
			expectedDir:     ".",
			expectedRemains: "",
		},
		{
			name:            "no matching folder returns basedir with full path",
			setupDirs:       []string{},
			basedir:         "",
			path:            "api/users/123",
			expectedDir:     ".",
			expectedRemains: "api/users/123",
		},
		{
			name:            "matching single folder",
			setupDirs:       []string{"api"},
			basedir:         "",
			path:            "api/users/123",
			expectedDir:     "api",
			expectedRemains: "users/123",
		},
		{
			name:            "matching nested folders",
			setupDirs:       []string{"api", "api/users"},
			basedir:         "",
			path:            "api/users/123",
			expectedDir:     "api/users",
			expectedRemains: "123",
		},
		{
			name:            "deep nested match",
			setupDirs:       []string{"api", "api/v1", "api/v1/users"},
			basedir:         "",
			path:            "api/v1/users/posts/comments",
			expectedDir:     "api/v1/users",
			expectedRemains: "posts/comments",
		},
		{
			name:            "path with leading slash",
			setupDirs:       []string{"api"},
			basedir:         "",
			path:            "/api/users",
			expectedDir:     "api",
			expectedRemains: "users",
		},
		{
			name:            "path with trailing slash",
			setupDirs:       []string{"api"},
			basedir:         "",
			path:            "api/users/",
			expectedDir:     "api",
			expectedRemains: "users",
		},
		{
			name:            "full path exists as folder",
			setupDirs:       []string{"api", "api/users", "api/users/list"},
			basedir:         "",
			path:            "api/users/list",
			expectedDir:     "api/users/list",
			expectedRemains: "",
		},

		// Tests with basedir as a specific subdirectory (not current dir)
		{
			name:            "basedir is subdirectory with no matching path",
			setupDirs:       []string{"project"},
			basedir:         "project",
			path:            "api/users/123",
			expectedDir:     "project",
			expectedRemains: "api/users/123",
		},
		{
			name:            "basedir is subdirectory with matching subfolder",
			setupDirs:       []string{"project", "project/api"},
			basedir:         "project",
			path:            "api/users/123",
			expectedDir:     "project/api",
			expectedRemains: "users/123",
		},
		{
			name:            "basedir is deeply nested",
			setupDirs:       []string{"a", "a/b", "a/b/c"},
			basedir:         "a/b/c",
			path:            "api/users",
			expectedDir:     "a/b/c",
			expectedRemains: "api/users",
		},
		{
			name:            "basedir name matches first path segment",
			setupDirs:       []string{"api"},
			basedir:         "api",
			path:            "api/users/123",
			expectedDir:     "api",
			expectedRemains: "users/123",
		},
		{
			name:            "basedir name matches first segment with existing subfolder",
			setupDirs:       []string{"api", "api/users"},
			basedir:         "api",
			path:            "api/users/123",
			expectedDir:     "api/users",
			expectedRemains: "123",
		},
		{
			name:            "basedir with nested matching structure",
			setupDirs:       []string{"collection", "collection/api", "collection/api/v1"},
			basedir:         "collection",
			path:            "api/v1/users/posts",
			expectedDir:     "collection/api/v1",
			expectedRemains: "users/posts",
		},
		{
			name:            "basedir with partial nested match",
			setupDirs:       []string{"collection", "collection/api"},
			basedir:         "collection",
			path:            "api/v1/users",
			expectedDir:     "collection/api",
			expectedRemains: "v1/users",
		},
		{
			name:            "basedir is subdirectory with deep nested match",
			setupDirs:       []string{"myproject", "myproject/api", "myproject/api/v1", "myproject/api/v1/users"},
			basedir:         "myproject",
			path:            "api/v1/users/123/posts",
			expectedDir:     "myproject/api/v1/users",
			expectedRemains: "123/posts",
		},
		{
			name:            "basedir name matches but no subfolder exists",
			setupDirs:       []string{"users"},
			basedir:         "users",
			path:            "users/123/profile",
			expectedDir:     "users",
			expectedRemains: "123/profile",
		},
		{
			name:            "basedir with empty path",
			setupDirs:       []string{"project"},
			basedir:         "project",
			path:            "",
			expectedDir:     "project",
			expectedRemains: "",
		},
		{
			name:            "basedir with path that has no common segments",
			setupDirs:       []string{"frontend", "frontend/components"},
			basedir:         "frontend",
			path:            "api/backend/services",
			expectedDir:     "frontend",
			expectedRemains: "api/backend/services",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory
			tmpDir := t.TempDir()

			// Setup directories
			for _, dir := range tt.setupDirs {
				err := os.MkdirAll(filepath.Join(tmpDir, dir), 0o755)
				if err != nil {
					t.Fatalf("failed to create test directory %q: %v", dir, err)
				}
			}

			// Determine basedir
			basedir := tt.basedir
			if basedir == "" {
				basedir = tmpDir
			} else {
				basedir = filepath.Join(tmpDir, basedir)
			}

			// Determine expected directory
			expectedDir := tt.expectedDir
			if expectedDir == "." {
				expectedDir = tmpDir
			} else {
				expectedDir = filepath.Join(tmpDir, expectedDir)
			}

			gotDir, gotRemains := findRequestFolder(basedir, tt.path)

			if gotDir != expectedDir {
				t.Errorf("findRequestFolder() dir = %q, want %q", gotDir, expectedDir)
			}
			if gotRemains != tt.expectedRemains {
				t.Errorf("findRequestFolder() remains = %q, want %q", gotRemains, tt.expectedRemains)
			}
		})
	}
}

func TestRequestBodyBlock(t *testing.T) {
	tests := []struct {
		name     string
		rd       RequestData
		contains []string // strings that should be in the output
	}{
		{
			name: "json body",
			rd: RequestData{
				BodyType: "json",
				Body:     `{"key": "value"}`,
			},
			contains: []string{"json {", `{"key": "value"}`},
		},
		{
			name: "xml body",
			rd: RequestData{
				BodyType: "xml",
				Body:     "<root><item>value</item></root>",
			},
			contains: []string{"xml {", "<root><item>value</item></root>"},
		},
		{
			name: "text body",
			rd: RequestData{
				BodyType: "text",
				Body:     "plain text content",
			},
			contains: []string{"text {", "plain text content"},
		},
		{
			name: "formUrlEncoded body",
			rd: RequestData{
				BodyType: "formUrlEncoded",
				Body:     "key1=value1&key2=value2",
			},
			contains: []string{"body:form-urlencoded {", "key1: value1", "key2: value2"},
		},
		{
			name: "multipartForm body",
			rd: RequestData{
				BodyType: "multipartForm",
				Body: "--boundary\r\n" +
					"Content-Disposition: form-data; name=\"field1\"\r\n\r\n" +
					"value1\r\n" +
					"--boundary--\r\n",
			},
			contains: []string{"body:multipart-form {", "field1: value1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RequestBodyBlock(tt.rd)
			for _, want := range tt.contains {
				if !containsString(got, want) {
					t.Errorf("RequestBodyBlock() = %q, should contain %q", got, want)
				}
			}
		})
	}
}

func TestRequestContent(t *testing.T) {
	tests := []struct {
		name     string
		rd       RequestData
		contains []string
	}{
		{
			name: "basic GET request",
			rd: RequestData{
				Name:       "test-request",
				FilesCount: 0,
				Method:     "GET",
				Path:       "/api/users",
				BodyType:   "none",
				Env: &BrunoEnv{
					Vars: map[string]string{
						"proto": "https",
						"host":  "example.com",
					},
				},
			},
			contains: []string{
				"meta {",
				"name: test-request",
				"seq: 1",
				"type: http",
				"get {",
				"url: {{proto}}://{{host}}/api/users",
				"body: none",
				"auth: none",
				"settings {",
				"encodeUrl: false",
				"docs {",
			},
		},
		{
			name: "POST request with JSON body",
			rd: RequestData{
				Name:       "create-user",
				FilesCount: 5,
				Method:     "POST",
				Path:       "/api/users",
				BodyType:   "json",
				Body:       `{"name": "John"}`,
				Env: &BrunoEnv{
					Vars: map[string]string{
						"proto": "https",
						"host":  "api.example.com",
					},
				},
			},
			contains: []string{
				"meta {",
				"name: create-user",
				"seq: 6",
				"post {",
				"url: {{proto}}://{{host}}/api/users",
				"body: json",
				"json {",
				`{"name": "John"}`,
			},
		},
		{
			name: "request with query string",
			rd: RequestData{
				Name:       "search",
				FilesCount: 0,
				Method:     "GET",
				Path:       "/api/search",
				RawQuery:   "q=test&page=1",
				BodyType:   "none",
				Env: &BrunoEnv{
					Vars: map[string]string{
						"proto": "https",
						"host":  "example.com",
					},
				},
			},
			contains: []string{
				"url: {{proto}}://{{host}}/api/search?q=test&page=1",
			},
		},
		{
			name: "request without env",
			rd: RequestData{
				Name:       "no-env",
				FilesCount: 0,
				Method:     "GET",
				Path:       "/api/test",
				BodyType:   "none",
				Env:        nil,
			},
			contains: []string{
				"url: ://",
			},
		},
		{
			name: "request with path variable substitution",
			rd: RequestData{
				Name:       "get-user",
				FilesCount: 0,
				Method:     "GET",
				Path:       "/users/123",
				BodyType:   "none",
				Env: &BrunoEnv{
					Vars:        map[string]string{"proto": "https", "host": "example.com", "user_id": "123"},
					ReverseVars: map[string]string{"123": "user_id"},
				},
			},
			contains: []string{
				"url: {{proto}}://{{host}}/users/{{user_id}}",
			},
		},
		{
			name: "request with query string variable substitution",
			rd: RequestData{
				Name:       "search-user",
				FilesCount: 0,
				Method:     "GET",
				Path:       "/api/search",
				RawQuery:   "user_id=123&account=456",
				BodyType:   "none",
				Env: &BrunoEnv{
					Vars:        map[string]string{"proto": "https", "host": "example.com", "user_id": "123", "account_id": "456"},
					ReverseVars: map[string]string{"123": "user_id", "456": "account_id"},
				},
			},
			contains: []string{
				"url: {{proto}}://{{host}}/api/search?user_id={{user_id}}&account={{account_id}}",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := requestContent(tt.rd)
			for _, want := range tt.contains {
				if !containsString(got, want) {
					t.Errorf("requestContent() should contain %q\ngot:\n%s", want, got)
				}
			}
		})
	}
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestCreateRequestFileWithBasedir(t *testing.T) {
	tests := []struct {
		name         string
		setupDirs    []string // directories to create relative to tmpDir
		basedir      string   // basedir relative to tmpDir (empty means use tmpDir directly)
		rd           RequestData
		expectedFile string // expected file path relative to tmpDir
		wantErr      bool
	}{
		{
			name:      "file created in basedir when basedir is set",
			setupDirs: []string{"myproject"},
			basedir:   "myproject",
			rd: RequestData{
				Method:   "GET",
				Path:     "/api/users",
				BodyType: "none",
				Env: &BrunoEnv{
					Vars: map[string]string{"proto": "https", "host": "example.com"},
				},
			},
			expectedFile: "myproject/api-users-GET.bru",
			wantErr:      false,
		},
		{
			name:      "file created in nested basedir",
			setupDirs: []string{"projects", "projects/myapi"},
			basedir:   "projects/myapi",
			rd: RequestData{
				Method:   "POST",
				Path:     "/users/create",
				BodyType: "json",
				Body:     `{"name": "test"}`,
				Env: &BrunoEnv{
					Vars: map[string]string{"proto": "https", "host": "api.example.com"},
				},
			},
			expectedFile: "projects/myapi/users-create-POST.bru",
			wantErr:      false,
		},
		{
			name:      "file created in matching subfolder within basedir",
			setupDirs: []string{"collection", "collection/api"},
			basedir:   "collection",
			rd: RequestData{
				Method:   "GET",
				Path:     "/api/users/123",
				BodyType: "none",
				Env: &BrunoEnv{
					Vars: map[string]string{"proto": "https", "host": "example.com"},
				},
			},
			expectedFile: "collection/api/users-123-GET.bru",
			wantErr:      false,
		},
		{
			name:      "file created in deeply nested matching subfolder",
			setupDirs: []string{"collection", "collection/api", "collection/api/v1", "collection/api/v1/users"},
			basedir:   "collection",
			rd: RequestData{
				Method:   "DELETE",
				Path:     "/api/v1/users/456",
				BodyType: "none",
				Env: &BrunoEnv{
					Vars: map[string]string{"proto": "https", "host": "example.com"},
				},
			},
			expectedFile: "collection/api/v1/users/456-DELETE.bru",
			wantErr:      false,
		},
		{
			name:      "basedir with path containing env variables",
			setupDirs: []string{"myproject", "myproject/api"},
			basedir:   "myproject",
			rd: RequestData{
				Method:   "GET",
				Path:     "/api/users/12345",
				BodyType: "none",
				Env: &BrunoEnv{
					Vars:        map[string]string{"proto": "https", "host": "example.com", "user_id": "12345"},
					ReverseVars: map[string]string{"12345": "user_id"},
				},
			},
			expectedFile: "myproject/api/users-USER_ID-GET.bru",
			wantErr:      false,
		},
		{
			name:      "absolute path as basedir",
			setupDirs: []string{}, // will use tmpDir directly as absolute path
			basedir:   "",         // special case: test will set absolute path
			rd: RequestData{
				Method:   "GET",
				Path:     "/health",
				BodyType: "none",
				Env: &BrunoEnv{
					Vars: map[string]string{"proto": "https", "host": "example.com"},
				},
			},
			expectedFile: "health-GET.bru",
			wantErr:      false,
		},

		// Tests for exact folder match - filename should NOT have leading dash
		{
			name:      "path matches exact folder - should not have leading dash",
			setupDirs: []string{"collection", "collection/api", "collection/api/app"},
			basedir:   "collection",
			rd: RequestData{
				Method:   "GET",
				Path:     "/api/app",
				BodyType: "none",
				Env: &BrunoEnv{
					Vars: map[string]string{"proto": "https", "host": "example.com"},
				},
			},
			expectedFile: "collection/api/app/GET.bru",
			wantErr:      false,
		},
		{
			name:      "path matches exact folder - POST method",
			setupDirs: []string{"collection", "collection/api", "collection/api/users"},
			basedir:   "collection",
			rd: RequestData{
				Method:   "POST",
				Path:     "/api/users",
				BodyType: "json",
				Body:     `{"name": "test"}`,
				Env: &BrunoEnv{
					Vars: map[string]string{"proto": "https", "host": "example.com"},
				},
			},
			expectedFile: "collection/api/users/POST.bru",
			wantErr:      false,
		},
		{
			name:      "single segment path matches folder",
			setupDirs: []string{"collection", "collection/users"},
			basedir:   "collection",
			rd: RequestData{
				Method:   "GET",
				Path:     "/users",
				BodyType: "none",
				Env: &BrunoEnv{
					Vars: map[string]string{"proto": "https", "host": "example.com"},
				},
			},
			expectedFile: "collection/users/GET.bru",
			wantErr:      false,
		},
		{
			name:      "deeply nested path matches exact folder",
			setupDirs: []string{"col", "col/api", "col/api/v1", "col/api/v1/users", "col/api/v1/users/profile"},
			basedir:   "col",
			rd: RequestData{
				Method:   "DELETE",
				Path:     "/api/v1/users/profile",
				BodyType: "none",
				Env: &BrunoEnv{
					Vars: map[string]string{"proto": "https", "host": "example.com"},
				},
			},
			expectedFile: "col/api/v1/users/profile/DELETE.bru",
			wantErr:      false,
		},
		{
			name:      "root path with basedir as exact match",
			setupDirs: []string{"api"},
			basedir:   "api",
			rd: RequestData{
				Method:   "GET",
				Path:     "/",
				BodyType: "none",
				Env: &BrunoEnv{
					Vars: map[string]string{"proto": "https", "host": "example.com"},
				},
			},
			expectedFile: "api/GET.bru",
			wantErr:      false,
		},
		{
			name:      "path with trailing slash matches folder",
			setupDirs: []string{"collection", "collection/api"},
			basedir:   "collection",
			rd: RequestData{
				Method:   "GET",
				Path:     "/api/",
				BodyType: "none",
				Env: &BrunoEnv{
					Vars: map[string]string{"proto": "https", "host": "example.com"},
				},
			},
			expectedFile: "collection/api/GET.bru",
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Setup directories
			for _, dir := range tt.setupDirs {
				err := os.MkdirAll(filepath.Join(tmpDir, dir), 0o755)
				if err != nil {
					t.Fatalf("failed to create directory %q: %v", dir, err)
				}
			}

			// Set basedir
			var basedir string
			if tt.basedir == "" {
				basedir = tmpDir // absolute path case
			} else {
				basedir = filepath.Join(tmpDir, tt.basedir)
			}

			// Create RequestData with Basedir set
			rd := tt.rd
			rd.Basedir = basedir

			// Run createRequestFile
			err := createRequestFile(rd)

			if (err != nil) != tt.wantErr {
				t.Errorf("createRequestFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// Verify file was created at expected location
			expectedPath := filepath.Join(tmpDir, tt.expectedFile)
			if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
				// List what files were actually created to help debug
				var createdFiles []string
				filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
					if !info.IsDir() {
						rel, _ := filepath.Rel(tmpDir, path)
						createdFiles = append(createdFiles, rel)
					}
					return nil
				})
				t.Errorf("Expected file %q was not created.\nFiles created: %v", tt.expectedFile, createdFiles)
			}
		})
	}
}

func TestCreateRequestFileBasedirNotSet(t *testing.T) {
	// This test documents the current buggy behavior where if Basedir is not set,
	// the file is created in current directory instead of the intended basedir.

	tmpDir := t.TempDir()

	// Create a subdirectory that should be the basedir
	basedir := filepath.Join(tmpDir, "intended-basedir")
	os.MkdirAll(basedir, 0o755)

	// Create RequestData WITHOUT setting Basedir (simulating the bug in DoRequest)
	rd := RequestData{
		// Basedir is intentionally NOT set - this is the bug!
		Method:   "GET",
		Path:     "/api/test",
		BodyType: "none",
		Env: &BrunoEnv{
			Vars: map[string]string{"proto": "https", "host": "example.com"},
		},
	}

	// Change to tmpDir so we can observe where file is created
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	err := createRequestFile(rd)
	if err != nil {
		t.Fatalf("createRequestFile() error = %v", err)
	}

	// BUG: File is created in current directory (tmpDir) instead of basedir
	// because rd.Basedir is empty, findRequestFolder defaults to "."
	buggyPath := filepath.Join(tmpDir, "api-test-GET.bru")
	intendedPath := filepath.Join(basedir, "api-test-GET.bru")

	buggyExists := false
	if _, err := os.Stat(buggyPath); err == nil {
		buggyExists = true
	}

	intendedExists := false
	if _, err := os.Stat(intendedPath); err == nil {
		intendedExists = true
	}

	// Document the bug: file is in current dir, not in intended basedir
	if buggyExists && !intendedExists {
		t.Logf("BUG CONFIRMED: File created at %q (current dir) instead of intended basedir", buggyPath)
		// This is the expected buggy behavior - test passes to document it
	}

	// Clean up the file we created
	os.Remove(buggyPath)
}

// TestDoRequestPassesBasedirToRequestData tests that DoRequest properly passes
// the basedir parameter to RequestData. This test will FAIL until the bug is fixed.
//
// The bug: In DoRequest(), the RequestData struct is created without setting Basedir:
//
//	rd := RequestData{
//	    Method:   req.Method,
//	    Path:     req.URL.Path,
//	    RawQuery: req.URL.RawQuery,
//	    Env:      envs,
//	    // Basedir is MISSING!
//	}
//
// It should be:
//
//	rd := RequestData{
//	    Basedir:  basedir,  // <-- ADD THIS
//	    Method:   req.Method,
//	    ...
//	}
func TestDoRequestPassesBasedirToRequestData(t *testing.T) {
	tmpDir := t.TempDir()

	// Create basedir structure
	basedir := filepath.Join(tmpDir, "my-collection")
	envDir := filepath.Join(basedir, "environments")
	os.MkdirAll(envDir, 0o755)

	// Create environment file
	envContent := `vars {
  proto: https
  host: example.com
}`
	os.WriteFile(filepath.Join(envDir, "base.bru"), []byte(envContent), 0o644)

	// Create a raw HTTP request
	rawRequest := "GET /api/users HTTP/1.1\r\nHost: example.com\r\n\r\n"

	// Save original stdin and restore after test
	origStdin := os.Stdin
	defer func() { os.Stdin = origStdin }()

	// Create a pipe to simulate stdin
	r, w, _ := os.Pipe()
	os.Stdin = r

	// Write request to pipe in a goroutine
	go func() {
		w.Write([]byte(rawRequest))
		w.Close()
	}()

	// Change to a different directory (not basedir) to verify file is NOT created here
	otherDir := filepath.Join(tmpDir, "other-dir")
	os.MkdirAll(otherDir, 0o755)
	origDir, _ := os.Getwd()
	os.Chdir(otherDir)
	defer os.Chdir(origDir)

	// Call DoRequest with basedir
	err := DoRequest(basedir, "environments/base.bru")
	if err != nil {
		t.Fatalf("DoRequest() error = %v", err)
	}

	// The file SHOULD be created in basedir, NOT in current directory (otherDir)
	expectedInBasedir := filepath.Join(basedir, "api-users-GET.bru")
	wrongInCurrentDir := filepath.Join(otherDir, "api-users-GET.bru")

	basedirExists := false
	if _, err := os.Stat(expectedInBasedir); err == nil {
		basedirExists = true
	}

	currentDirExists := false
	if _, err := os.Stat(wrongInCurrentDir); err == nil {
		currentDirExists = true
	}

	// This assertion will FAIL until the bug is fixed
	if !basedirExists {
		t.Errorf("File should be created in basedir %q but wasn't", expectedInBasedir)
	}

	if currentDirExists {
		t.Errorf("BUG: File was created in current directory %q instead of basedir", wrongInCurrentDir)
	}

	// List all files for debugging
	var allFiles []string
	filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && filepath.Ext(path) == ".bru" {
			rel, _ := filepath.Rel(tmpDir, path)
			allFiles = append(allFiles, rel)
		}
		return nil
	})
	t.Logf("All .bru files created: %v", allFiles)
}
