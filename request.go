package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type RequestData struct {
	FilesCount int
	Name       string
	Basedir    string
	Method     string
	Path       string
	RawQuery   string
	BodyType   string
	Body       string
	Env        *BrunoEnv
	HTTPReq    *http.Request
}

// BodyTypeName returns the value for the `body:value block`.
// xml, json, and text type names are the same as the type,
// but for forms we need to do a conversion.
func BodyTypeName(bt string) string {
	btm := make(map[string]string)
	btm["formUrlEncoded"] = "form-urlencoded"
	btm["multipartForm"] = "multipart-form"

	name := btm[bt]
	if name == "" {
		return bt
	}

	return name
}

// BodyTypeFromContentType returns the Bruno body type based on
// the request's Content-Type header.
//
// Bruno body types:
// none, json, xml, text, multipartForm, formUrlEncoded
func BodyTypeFromContentType(ct string) (string, error) {
	if ct == "" {
		return "none", nil
	}

	mediaType, _, err := mime.ParseMediaType(ct)
	if err != nil {
		return "", fmt.Errorf("failed to parse content type %q: %w", ct, err)
	}

	switch mediaType {
	case "application/json":
		return "json", nil
	case "application/xml", "text/xml":
		return "xml", nil
	case "text/plain":
		return "text", nil
	case "multipart/form-data":
		return "multipartForm", nil
	case "application/x-www-form-urlencoded":
		return "formUrlEncoded", nil
	default:
		return "", fmt.Errorf("unsupported content type: %s", mediaType)
	}
}

func DoRequest(basedir, envfile string) error {
	rawReq, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("read from stdin error %w", err)
	}

	req, err := ParseRawRequest(rawReq)
	if err != nil {
		return fmt.Errorf("parse raw request error %w", err)
	}
	defer req.Body.Close()

	envs, err := EnvFromFile(filepath.Join(basedir, envfile))
	if err != nil {
		fmt.Fprintf(os.Stderr, "[W] read env file: %s", err)
	}

	rd := RequestData{
		Basedir:  basedir,
		Method:   req.Method,
		Path:     req.URL.Path,
		RawQuery: req.URL.RawQuery,
		Env:      envs,
		HTTPReq:  req,
	}

	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		return fmt.Errorf("failed to read request body: %w", err)
	}
	rd.Body = string(bodyBytes)

	if rd.Body != "" {
		rd.BodyType, err = BodyTypeFromContentType(req.Header.Get("Content-Type"))
		if err != nil {
			return fmt.Errorf("detect bodytype error %w", err)
		}
	} else {
		rd.BodyType = "none"
	}

	err = createRequestFile(rd)
	if err != nil {
		return fmt.Errorf("create request file error %w", err)
	}

	// print request back for next processors
	fmt.Print(string(rawReq))

	return nil
}

func createRequestFile(rd RequestData) error {
	dir, tail := findRequestFolder(rd.Basedir, rd.Path)
	tail = EnvToPath(tail, rd.Env)
	name := pathToName(tail)
	if name == "" {
		rd.Name = rd.Method
	} else {
		rd.Name = name + "-" + rd.Method
	}
	rd.FilesCount = DirFilesCount(dir)
	rd.Body = EnvToBody(rd.Body, rd.Env)

	content := requestContent(rd)
	fp := filepath.Join(dir, rd.Name+".bru")

	if _, err := os.Stat(fp); err == nil {
		return fmt.Errorf("file %q already exists", fp)
	}

	if err := os.WriteFile(fp, []byte(content), 0644); err != nil {
		return fmt.Errorf("write request to file %q error %w", fp, err)
	}

	return nil
}

func requestContent(rd RequestData) string {
	var sb strings.Builder

	meta := make(map[string]string)
	meta["name"] = rd.Name
	meta["seq"] = strconv.Itoa(rd.FilesCount + 1)
	meta["type"] = "http"

	sb.WriteString(MetaGenerate(meta))
	sb.WriteString("\n")

	rvars := make(map[string]string)
	path := rd.Path
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	path = EnvToPath(strings.TrimPrefix(path, "/"), rd.Env)
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if rd.RawQuery != "" {
		query := EnvToBody(rd.RawQuery, rd.Env)
		path += "?" + query
	}
	proto, host := "", ""
	if rd.Env != nil {
		proto = "{{proto}}"
		host = "{{host}}"
	}
	if rd.HTTPReq != nil && host != rd.HTTPReq.Host {
		fmt.Fprintf(os.Stderr, "[W] host mismatched. in envs - %s, in request - %s", host, rd.HTTPReq.Host)
	}
	rvars["url"] = fmt.Sprintf("%s://%s%s", proto, host, path)
	rvars["body"] = rd.BodyType
	rvars["auth"] = "none"
	sb.WriteString(NameBlockMap(strings.ToLower(rd.Method), rvars))
	sb.WriteString("\n")

	setts := make(map[string]string)
	setts["encodeUrl"] = "false"
	sb.WriteString(NameBlockMap("settings", setts))
	sb.WriteString("\n")

	docs := []string{"- [ ] methods", "- [ ] params", "- [ ] headers"}

	if rd.BodyType != "none" {
		sb.WriteString(RequestBodyBlock(rd))
		sb.WriteString("\n")
		docs = append(docs, "- [ ] body params")
	}

	sb.WriteString(NameBlockStrings("docs", docs))

	return sb.String()
}

func RequestBodyBlock(rd RequestData) string {
	if rd.BodyType == "formUrlEncoded" {
		return RequestBodyUrlEncoded(rd)
	}
	if rd.BodyType == "multipartForm" {
		return RequestBodyMultipartForm(rd)
	}
	return NameBlockStrings(rd.BodyType, []string{rd.Body})
}

func RequestBodyUrlEncoded(rd RequestData) string {
	name := "body:" + BodyTypeName(rd.BodyType)
	vars := ParseBodyUrlEncoded(rd.Body)
	return NameBlockMap(name, vars)
}

func RequestBodyMultipartForm(rd RequestData) string {
	name := "body:" + BodyTypeName(rd.BodyType)
	vars := ParseBodyMultipartForm(rd.Body)
	return NameBlockMap(name, vars)
}

// ParseBodyMultipartForm parse http multipart/form-data body to map
func ParseBodyMultipartForm(body string) map[string]string {
	result := make(map[string]string)
	if body == "" {
		return result
	}

	reader := multipart.NewReader(strings.NewReader(body), extractBoundary(body))
	form, err := reader.ReadForm(10 << 20) // 10 MB max memory
	if err != nil {
		return result
	}
	defer form.RemoveAll()

	for key, values := range form.Value {
		if len(values) > 0 {
			result[key] = values[0]
		}
	}

	return result
}

// extractBoundary extracts the boundary string from a multipart body
func extractBoundary(body string) string {
	// The first line of a multipart body is typically "--<boundary>"
	firstLine := strings.SplitN(body, "\r\n", 2)[0]
	if strings.HasPrefix(firstLine, "--") {
		return strings.TrimPrefix(firstLine, "--")
	}
	// Try with just \n as line ending
	firstLine = strings.SplitN(body, "\n", 2)[0]
	if strings.HasPrefix(firstLine, "--") {
		return strings.TrimPrefix(firstLine, "--")
	}
	return ""
}

// ParseBodyUrlEncoded parse http x-www-form-urlencoded body to map
func ParseBodyUrlEncoded(body string) map[string]string {
	result := make(map[string]string)
	if body == "" {
		return result
	}
	pairs := strings.Split(body, "&")
	for _, pair := range pairs {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			key, _ := url.QueryUnescape(parts[0])
			value, _ := url.QueryUnescape(parts[1])
			result[key] = value
		} else if len(parts) == 1 {
			key, _ := url.QueryUnescape(parts[0])
			result[key] = ""
		}
	}
	return result
}

// pathToName convert path in format
// api/users/{{user_id}} to filename format api-users-USER_ID
// variables if present converts to uppercase keys
// {{id}} to ID
// {{account_name}} to ACCOUNT_NAME
func pathToName(path string) string {
	// Replace path separators with dashes
	result := strings.ReplaceAll(path, "/", "-")

	// Find and replace {{variable}} patterns with uppercase equivalents
	re := regexp.MustCompile(`\{\{(\w+)\}\}`)
	result = re.ReplaceAllStringFunc(result, func(match string) string {
		// Extract the variable name between {{ and }}
		varName := re.FindStringSubmatch(match)[1]
		return strings.ToUpper(varName)
	})

	return result
}

// findRequestFolder detects Bruno subfolders for a request.
// For example, if the path is /api/users/search/id/123
// and the subfolder api/users exists, it should return
// api/users, search/id/123.
// If the api folder is not found, it returns
// ., api/users/search/id/123.
// It uses the current folder as a starting point too,
// so if api is not found but the current dir is api, it returns
// ., users/search/id/123.
// basedir specifies the base directory to search in (default ".").
func findRequestFolder(basedir string, path string) (string, string) {
	if basedir == "" {
		basedir = "."
	}

	// Clean and split the path into segments
	path = strings.Trim(path, "/")
	if path == "" {
		return basedir, ""
	}
	segments := strings.Split(path, "/")

	// Try to find the longest matching subfolder path
	// Start from the longest possible path and work backwards
	for i := len(segments); i > 0; i-- {
		candidate := filepath.Join(basedir, filepath.Join(segments[:i]...))
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			remaining := strings.Join(segments[i:], "/")
			return candidate, remaining
		}
	}

	// If no subfolder matched, check if we're already inside one of the path segments
	// (current directory matches the first segment)
	resolvedBase, err := filepath.Abs(basedir)
	if err == nil {
		currentDir := filepath.Base(resolvedBase)
		if len(segments) > 0 && segments[0] == currentDir {
			// Current dir matches first segment, strip it and try again
			remainingSegments := segments[1:]
			for i := len(remainingSegments); i > 0; i-- {
				candidate := filepath.Join(basedir, filepath.Join(remainingSegments[:i]...))
				if info, err := os.Stat(candidate); err == nil && info.IsDir() {
					remaining := strings.Join(remainingSegments[i:], "/")
					return candidate, remaining
				}
			}
			// No subfolder found, return basedir with remaining path
			return basedir, strings.Join(remainingSegments, "/")
		}
	}

	// Nothing matched at all
	return basedir, path
}

// ParseRawRequest takes a raw HTTP request as []byte and returns a parsed *http.Request and error.
func ParseRawRequest(rawRequest []byte) (*http.Request, error) {
	reader := bufio.NewReader(bytes.NewReader(rawRequest))

	// Skip lines starting with '#' (meta lines)
	for {
		line, err := reader.Peek(1)
		if err != nil {
			return nil, err
		}
		if line[0] != '#' {
			break
		}
		// Read and discard the entire line
		_, _, err = reader.ReadLine()
		if err != nil {
			return nil, err
		}
	}

	req, err := http.ReadRequest(reader)
	if err != nil {
		return nil, err
	}
	return req, nil
}
