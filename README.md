# http2bruno

A CLI tool that converts raw HTTP requests into [Bruno](https://www.usebruno.com/) API client format files (`.bru` files).

## Installation

```bash
go install github.com/vodafon/http2bruno@latest
```

Or build from source:

```bash
git clone https://github.com/vodafon/http2bruno.git
cd http2bruno
go build
```

## Usage

### Create a new collection

```bash
http2bruno -o collection -c my-api.example.com
```

This creates:
```
my-api/
  bruno.json
  collection.bru
  environments/
    base.bru
```

### Create a new collection with a folder

```bash
http2bruno -o collection -c my-api.example.com -f users
```

This creates:
```
my-api/
  bruno.json
  collection.bru
  environments/
    base.bru
  users/
    folder.bru
```

### Create a folder in an existing collection

```bash
http2bruno -o folder -f api/v1/users -base ./my-api.example.com
```

### Convert HTTP request to Bruno format

Pipe a raw HTTP request to create a `.bru` file:

```bash
cat request.http | http2bruno -base ./my-api.example.com
```

Or capture from clipboard/other tools:

```bash
echo "GET /api/users HTTP/1.1
Host: example.com

" | http2bruno -base ./my-api.example.com
```

The tool will:
1. Parse the raw HTTP request
2. Detect the appropriate subfolder based on the request path
3. Replace values with environment variables (if matching values found in env file)
4. Create a `.bru` file with the request

## Command Line Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-o` | `request` | Operation: `collection`, `folder`, or `request` |
| `-c` | `""` | Collection name (for `-o collection`) |
| `-f` | `""` | Folder name/path (for `-o collection` or `-o folder`) |
| `-base` | `.` | Base collection directory (for `-o request` or `-o folder`) |
| `-e` | `environments/base.bru` | Environment file path relative to base directory |

## Supported Content Types

| Content-Type | Bruno Body Type |
|--------------|-----------------|
| `application/json` | `json` |
| `application/xml`, `text/xml` | `xml` |
| `text/plain` | `text` |
| `multipart/form-data` | `multipartForm` |
| `application/x-www-form-urlencoded` | `formUrlEncoded` |

## Project Structure

```
http2bruno/
  main.go           # CLI entry point
  request.go        # HTTP request parsing and .bru file generation
  env.go            # Environment file parsing and variable substitution
  structure.go      # Collection and folder creation
  collection.go     # Bruno collection configuration
  folder.go         # Bruno folder configuration
  helpers.go        # Block formatting utilities
  headers.go        # Headers block generation
  meta.go           # Meta block generation
```

## Testing

```bash
go test ./...
```

Run with coverage:
```bash
go test -cover ./...
```

## License

MIT
