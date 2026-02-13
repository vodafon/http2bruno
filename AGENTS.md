# AGENTS.md - http2bruno

AI coding agent instructions for the http2bruno repository.

## Project Overview

A Go CLI tool that converts raw HTTP requests into [Bruno](https://www.usebruno.com/) API client format files (`.bru` files). Single-package application with no external dependencies beyond Go stdlib.

## Build & Development Commands

```bash
# Build
go build

# Run tests (all)
go test ./...

# Run tests with verbose output
go test -v ./...

# Run a single test by name
go test -v -run TestFunctionName
go test -v -run TestFunctionName/subtest_name

# Run tests with coverage
go test -cover ./...

# Run tests with coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Format code
go fmt ./...

# Vet code (static analysis)
go vet ./...

# Install locally
go install
```

## Project Structure

```
http2bruno/
  main.go           # CLI entry point, flag parsing, operation dispatch
  request.go        # HTTP request parsing and .bru file generation
  request_test.go   # Tests for request handling
  env.go            # Environment file parsing and variable substitution
  env_test.go       # Tests for environment handling
  structure.go      # Collection and folder directory creation
  structure_test.go # Tests for structure creation
  collection.go     # Bruno collection configuration (bruno.json)
  folder.go         # Bruno folder configuration (folder.bru)
  folder_test.go    # Tests for folder handling
  helpers.go        # Block formatting utilities
  helpers_test.go   # Tests for helpers
  headers.go        # Headers block generation
  meta.go           # Meta block generation
```

## Code Style Guidelines

### Imports

- Group imports in parentheses, sorted alphabetically
- Standard library only (no third-party dependencies)
- No blank lines between imports

```go
import (
    "bufio"
    "fmt"
    "os"
    "strings"
)
```

### Naming Conventions

| Entity | Convention | Example |
|--------|------------|---------|
| Exported functions | PascalCase | `DoRequest`, `ParseBrunoEnv` |
| Unexported functions | camelCase | `raiseError`, `createRequestFile` |
| Types/Structs | PascalCase | `RequestData`, `BrunoEnv` |
| Variables | camelCase | `flagOp`, `bodyBytes` |
| Package-level vars | camelCase | `flagCollection` |
| Constants | camelCase or PascalCase | `flagEnvFile` |

### Error Handling

- Return errors as last return value
- Check errors immediately after function calls
- Wrap errors with context using `fmt.Errorf` and `%w`
- Use simple error returns for validation failures

```go
// Good: Wrapped error with context
if err != nil {
    return fmt.Errorf("parse raw request error %w", err)
}

// Good: Simple validation error
if collection == "" {
    return fmt.Errorf("-c collection name is required")
}

// Good: Deferred cleanup
req, err := ParseRawRequest(rawReq)
if err != nil {
    return fmt.Errorf("parse raw request error %w", err)
}
defer req.Body.Close()
```

### Function Documentation

- Use `// FunctionName description` format for exported functions
- Include example output in multi-line comments when helpful
- Keep comments concise and focused on "what" not "how"

```go
// DefaultBrunoJSON bruno.json file in collection directory
//
//  {
//    "version": "1",
//    "name": "example.com",
//    ...
//  }
func DefaultBrunoJSON(name string) string {
```

### File Operations

- Use `0o755` for directories, `0o644` for files
- Check file existence before writing: `os.Stat(fp)`
- Use `os.WriteFile` for simple writes
- Use `strings.Builder` for building multi-part strings

### Test Patterns

- Use table-driven tests with `[]struct` and `t.Run`
- Name test cases descriptively: `"single key-value pair"`, `"empty body returns empty map"`
- Use `t.TempDir()` for filesystem tests
- Test both success and error cases
- Use `reflect.DeepEqual` for map/slice comparisons

```go
func TestFunctionName(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
        wantErr  bool
    }{
        {"descriptive case name", "input", "expected", false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := FunctionName(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("FunctionName() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.expected {
                t.Errorf("FunctionName() = %q, want %q", got, tt.expected)
            }
        })
    }
}
```

### String Building

Use `strings.Builder` for concatenation:

```go
var sb strings.Builder
sb.WriteString(name)
sb.WriteString(" {\n")
sb.WriteString(content)
sb.WriteString("}\n")
return sb.String()
```

### Maps for Configuration

Use `map[string]string` for key-value configurations:

```go
meta := make(map[string]string)
meta["name"] = rd.Name
meta["seq"] = strconv.Itoa(rd.FilesCount + 1)
meta["type"] = "http"
```

## Key Patterns in This Codebase

### Bruno File Format

Bruno uses a block-based format:

```
meta {
  name: request-name
  seq: 1
  type: http
}

get {
  url: https://example.com/api
  body: none
  auth: none
}
```

### Environment Variable Substitution

Values are replaced with `{{variable_name}}` placeholders using `BrunoEnv.ReverseVars` mapping.

### Request Folder Detection

`findRequestFolder` matches URL paths to existing directory structure, returning the deepest matching folder.

## Things to Avoid

- No `interface{}` or type assertions without necessity
- No empty error handling (`if err != nil { }`)
- No global state mutations
- No external dependencies - stdlib only
