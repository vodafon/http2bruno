package main

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseBrunoEnv(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		expectedLen int
		checkVars   map[string]string
		wantErr     bool
	}{
		{
			name:        "empty content",
			content:     "",
			expectedLen: 0,
			checkVars:   map[string]string{},
			wantErr:     false,
		},
		{
			name: "single variable",
			content: `vars {
  host: example.com
}`,
			expectedLen: 1,
			checkVars:   map[string]string{"host": "example.com"},
			wantErr:     false,
		},
		{
			name: "multiple variables",
			content: `vars {
  host: example.com
  proto: https
  ua: Mozilla/5.0
}`,
			expectedLen: 3,
			checkVars: map[string]string{
				"host":  "example.com",
				"proto": "https",
				"ua":    "Mozilla/5.0",
			},
			wantErr: false,
		},
		{
			name: "value with colon",
			content: `vars {
  url: https://example.com:8080/api
}`,
			expectedLen: 1,
			checkVars:   map[string]string{"url": "https://example.com:8080/api"},
			wantErr:     false,
		},
		{
			name: "skip malformed lines",
			content: `vars {
  valid: value
  malformed_without_colon
  another: good
}`,
			expectedLen: 2,
			checkVars: map[string]string{
				"valid":   "value",
				"another": "good",
			},
			wantErr: false,
		},
		{
			name: "with empty lines",
			content: `vars {

  host: example.com

  proto: https

}`,
			expectedLen: 2,
			checkVars: map[string]string{
				"host":  "example.com",
				"proto": "https",
			},
			wantErr: false,
		},
		{
			name: "complex user agent",
			content: `vars {
  ua: Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36
}`,
			expectedLen: 1,
			checkVars: map[string]string{
				"ua": "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36",
			},
			wantErr: false,
		},
		{
			name:        "no vars block",
			content:     "some random content without vars block",
			expectedLen: 0,
			checkVars:   map[string]string{},
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env, err := ParseBrunoEnv(tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseBrunoEnv() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(env.Vars) != tt.expectedLen {
				t.Errorf("ParseBrunoEnv() returned %d vars, want %d", len(env.Vars), tt.expectedLen)
			}

			for k, v := range tt.checkVars {
				if env.Vars[k] != v {
					t.Errorf("ParseBrunoEnv() Vars[%q] = %q, want %q", k, env.Vars[k], v)
				}
				// Also check reverse mapping
				if env.ReverseVars[v] != k {
					t.Errorf("ParseBrunoEnv() ReverseVars[%q] = %q, want %q", v, env.ReverseVars[v], k)
				}
			}
		})
	}
}

func TestEnvToPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		env      *BrunoEnv
		expected string
	}{
		{
			name:     "nil env returns path unchanged",
			path:     "api/users/123",
			env:      nil,
			expected: "api/users/123",
		},
		{
			name: "empty reverse vars returns path unchanged",
			path: "api/users/123",
			env: &BrunoEnv{
				Vars:        map[string]string{},
				ReverseVars: map[string]string{},
			},
			expected: "api/users/123",
		},
		{
			name: "single segment replacement",
			path: "api/users/123",
			env: &BrunoEnv{
				Vars:        map[string]string{"user_id": "123"},
				ReverseVars: map[string]string{"123": "user_id"},
			},
			expected: "api/users/{{user_id}}",
		},
		{
			name: "multiple segment replacement",
			path: "api/org1/users/456",
			env: &BrunoEnv{
				Vars:        map[string]string{"org": "org1", "user_id": "456"},
				ReverseVars: map[string]string{"org1": "org", "456": "user_id"},
			},
			expected: "api/{{org}}/users/{{user_id}}",
		},
		{
			name: "no matching segments",
			path: "api/users/list",
			env: &BrunoEnv{
				Vars:        map[string]string{"user_id": "123"},
				ReverseVars: map[string]string{"123": "user_id"},
			},
			expected: "api/users/list",
		},
		{
			name:     "empty path",
			path:     "",
			env:      nil,
			expected: "",
		},
		{
			name: "single segment path",
			path: "123",
			env: &BrunoEnv{
				Vars:        map[string]string{"id": "123"},
				ReverseVars: map[string]string{"123": "id"},
			},
			expected: "{{id}}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EnvToPath(tt.path, tt.env)
			if got != tt.expected {
				t.Errorf("EnvToPath(%q) = %q, want %q", tt.path, got, tt.expected)
			}
		})
	}
}

func TestEnvToBody(t *testing.T) {
	// NOTE: Some of these tests will fail/panic because EnvToBody uses lookbehind
	// assertions (?<=) which are not supported by Go's RE2 regex engine.
	// These tests document the expected behavior once the regex is fixed.
	tests := []struct {
		name     string
		body     string
		env      *BrunoEnv
		expected string
	}{
		// Safe cases (nil env or empty ReverseVars)
		{
			name:     "nil env returns body unchanged",
			body:     "key=123&other=456",
			env:      nil,
			expected: "key=123&other=456",
		},
		{
			name:     "empty body returns empty",
			body:     "",
			env:      nil,
			expected: "",
		},
		{
			name: "empty reverse vars returns body unchanged",
			body: "key=123&other=456",
			env: &BrunoEnv{
				Vars:        map[string]string{"id": "123"},
				ReverseVars: map[string]string{},
			},
			expected: "key=123&other=456",
		},
		{
			name:     "empty body with nil env",
			body:     "",
			env:      &BrunoEnv{Vars: map[string]string{}, ReverseVars: map[string]string{}},
			expected: "",
		},

		// Single replacement cases
		{
			name: "single replacement after equals",
			body: "user_id=123",
			env: &BrunoEnv{
				Vars:        map[string]string{"id": "123"},
				ReverseVars: map[string]string{"123": "id"},
			},
			expected: "user_id={{id}}",
		},
		{
			name: "single replacement with descriptive var name",
			body: "account=ABC123",
			env: &BrunoEnv{
				Vars:        map[string]string{"account_id": "ABC123"},
				ReverseVars: map[string]string{"ABC123": "account_id"},
			},
			expected: "account={{account_id}}",
		},

		// Multiple replacements
		{
			name: "multiple different replacements",
			body: "x=123&y=456",
			env: &BrunoEnv{
				Vars:        map[string]string{"user_id": "123", "account_id": "456"},
				ReverseVars: map[string]string{"123": "user_id", "456": "account_id"},
			},
			expected: "x={{user_id}}&y={{account_id}}",
		},
		{
			name: "multiple occurrences of same value",
			body: "id=123&ref=123&other=123",
			env: &BrunoEnv{
				Vars:        map[string]string{"user_id": "123"},
				ReverseVars: map[string]string{"123": "user_id"},
			},
			expected: "id={{user_id}}&ref={{user_id}}&other={{user_id}}",
		},

		// Boundary conditions - should match
		{
			name: "value at start of body",
			body: "123&other=456",
			env: &BrunoEnv{
				Vars:        map[string]string{"id": "123"},
				ReverseVars: map[string]string{"123": "id"},
			},
			expected: "{{id}}&other=456",
		},
		{
			name: "value at end of body",
			body: "key=123",
			env: &BrunoEnv{
				Vars:        map[string]string{"id": "123"},
				ReverseVars: map[string]string{"123": "id"},
			},
			expected: "key={{id}}",
		},
		{
			name: "value after non-alphanumeric (equals sign)",
			body: "data=123",
			env: &BrunoEnv{
				Vars:        map[string]string{"val": "123"},
				ReverseVars: map[string]string{"123": "val"},
			},
			expected: "data={{val}}",
		},
		{
			name: "value before non-alphanumeric (ampersand)",
			body: "x=123&y=456",
			env: &BrunoEnv{
				Vars:        map[string]string{"id": "123"},
				ReverseVars: map[string]string{"123": "id"},
			},
			expected: "x={{id}}&y=456",
		},

		// Boundary conditions - should NOT match
		{
			name: "value preceded by alphanumeric should not match",
			body: "id=a123&name=test",
			env: &BrunoEnv{
				Vars:        map[string]string{"id": "123"},
				ReverseVars: map[string]string{"123": "id"},
			},
			expected: "id=a123&name=test", // unchanged
		},
		{
			name: "value followed by alphanumeric should not match",
			body: "id=123a&name=test",
			env: &BrunoEnv{
				Vars:        map[string]string{"id": "123"},
				ReverseVars: map[string]string{"123": "id"},
			},
			expected: "id=123a&name=test", // unchanged
		},
		{
			name: "value surrounded by alphanumeric should not match",
			body: "code=a123b",
			env: &BrunoEnv{
				Vars:        map[string]string{"id": "123"},
				ReverseVars: map[string]string{"123": "id"},
			},
			expected: "code=a123b", // unchanged
		},
		{
			name: "partial match should not replace",
			body: "id=12345",
			env: &BrunoEnv{
				Vars:        map[string]string{"id": "123"},
				ReverseVars: map[string]string{"123": "id"},
			},
			expected: "id=12345", // unchanged - 123 is followed by 45
		},

		// Longer values matched first
		{
			name: "longer values matched first",
			body: "val=12345",
			env: &BrunoEnv{
				Vars:        map[string]string{"short": "123", "long": "12345"},
				ReverseVars: map[string]string{"123": "short", "12345": "long"},
			},
			expected: "val={{long}}",
		},
		{
			name: "longer value takes precedence over shorter",
			body: "a=12345&b=123",
			env: &BrunoEnv{
				Vars:        map[string]string{"short_id": "123", "long_id": "12345"},
				ReverseVars: map[string]string{"123": "short_id", "12345": "long_id"},
			},
			expected: "a={{long_id}}&b={{short_id}}",
		},

		// JSON body replacement
		{
			name: "JSON body with numeric value",
			body: `{"user_id": 123, "account": 456}`,
			env: &BrunoEnv{
				Vars:        map[string]string{"user_id": "123", "account_id": "456"},
				ReverseVars: map[string]string{"123": "user_id", "456": "account_id"},
			},
			expected: `{"user_id": {{user_id}}, "account": {{account_id}}}`,
		},
		{
			name: "JSON body with string value",
			body: `{"token": "abc123xyz"}`,
			env: &BrunoEnv{
				Vars:        map[string]string{"auth_token": "abc123xyz"},
				ReverseVars: map[string]string{"abc123xyz": "auth_token"},
			},
			expected: `{"token": "{{auth_token}}"}`,
		},
		{
			name: "JSON nested object",
			body: `{"data": {"id": 999, "ref": 999}}`,
			env: &BrunoEnv{
				Vars:        map[string]string{"item_id": "999"},
				ReverseVars: map[string]string{"999": "item_id"},
			},
			expected: `{"data": {"id": {{item_id}}, "ref": {{item_id}}}}`,
		},

		// URL-encoded form data
		{
			name: "URL encoded form body",
			body: "username=john&password=secret123&remember=true",
			env: &BrunoEnv{
				Vars:        map[string]string{"pwd": "secret123"},
				ReverseVars: map[string]string{"secret123": "pwd"},
			},
			expected: "username=john&password={{pwd}}&remember=true",
		},

		// Special characters in values (regex metacharacters)
		{
			name: "value with dots",
			body: "domain=example.com",
			env: &BrunoEnv{
				Vars:        map[string]string{"host": "example.com"},
				ReverseVars: map[string]string{"example.com": "host"},
			},
			expected: "domain={{host}}",
		},
		{
			name: "value with plus sign",
			body: "email=user+tag@example.com",
			env: &BrunoEnv{
				Vars:        map[string]string{"mail": "user+tag@example.com"},
				ReverseVars: map[string]string{"user+tag@example.com": "mail"},
			},
			expected: "email={{mail}}",
		},
		{
			name: "value with brackets",
			body: "filter=[active]",
			env: &BrunoEnv{
				Vars:        map[string]string{"status": "[active]"},
				ReverseVars: map[string]string{"[active]": "status"},
			},
			expected: "filter={{status}}",
		},

		// Edge cases
		{
			name: "body is exactly the value",
			body: "123",
			env: &BrunoEnv{
				Vars:        map[string]string{"id": "123"},
				ReverseVars: map[string]string{"123": "id"},
			},
			expected: "{{id}}",
		},
		{
			name: "empty value in env should not cause issues",
			body: "key=value",
			env: &BrunoEnv{
				Vars:        map[string]string{"empty": ""},
				ReverseVars: map[string]string{"": "empty"},
			},
			expected: "key=value", // empty string replacement is tricky
		},
		{
			name: "value with only special chars",
			body: "sep=---",
			env: &BrunoEnv{
				Vars:        map[string]string{"separator": "---"},
				ReverseVars: map[string]string{"---": "separator"},
			},
			expected: "sep={{separator}}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EnvToBody(tt.body, tt.env)
			if got != tt.expected {
				t.Errorf("EnvToBody(%q) = %q, want %q", tt.body, got, tt.expected)
			}
		})
	}
}

func TestEnvGenerate(t *testing.T) {
	tests := []struct {
		name     string
		vars     map[string]string
		contains []string
	}{
		{
			name:     "empty map returns empty",
			vars:     map[string]string{},
			contains: []string{},
		},
		{
			name: "single variable",
			vars: map[string]string{"host": "example.com"},
			contains: []string{
				"vars {",
				"host: example.com",
				"}",
			},
		},
		{
			name: "multiple variables",
			vars: map[string]string{
				"host":  "example.com",
				"proto": "https",
			},
			contains: []string{
				"vars {",
				"host: example.com",
				"proto: https",
				"}",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EnvGenerate(tt.vars)

			if len(tt.contains) == 0 && got != "" {
				t.Errorf("EnvGenerate() = %q, want empty string", got)
				return
			}

			for _, want := range tt.contains {
				if !strings.Contains(got, want) {
					t.Errorf("EnvGenerate() = %q, should contain %q", got, want)
				}
			}
		})
	}
}

func TestDefaultEnvBru(t *testing.T) {
	tests := []struct {
		name     string
		hostName string
		contains []string
	}{
		{
			name:     "default env with host",
			hostName: "api.example.com",
			contains: []string{
				"vars {",
				"host: api.example.com",
				"proto: https",
				"ua: Mozilla/5.0",
				"}",
			},
		},
		{
			name:     "empty host",
			hostName: "",
			contains: []string{
				"vars {",
				"host:",
				"proto: https",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DefaultEnvBru(tt.hostName)
			for _, want := range tt.contains {
				if !strings.Contains(got, want) {
					t.Errorf("DefaultEnvBru(%q) = %q, should contain %q", tt.hostName, got, want)
				}
			}
		})
	}
}

func TestBrunoEnvReverseMapping(t *testing.T) {
	content := `vars {
  user_id: 12345
  account: abc123
}`

	env, err := ParseBrunoEnv(content)
	if err != nil {
		t.Fatalf("ParseBrunoEnv() error = %v", err)
	}

	// Test that reverse mapping works correctly
	expectedReverse := map[string]string{
		"12345":  "user_id",
		"abc123": "account",
	}

	if !reflect.DeepEqual(env.ReverseVars, expectedReverse) {
		t.Errorf("ReverseVars = %v, want %v", env.ReverseVars, expectedReverse)
	}
}
