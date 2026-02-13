package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
)

// BrunoEnv represents the parsed environment variables
type BrunoEnv struct {
	Vars        map[string]string
	ReverseVars map[string]string
}

// EnvToBody replaces values in the body with {{key}} placeholders using env values.
// For example, if body is x=123&a=1234 and env.ReverseVars["123"] = "user_id",
// it should return x={{user_id}}&a=1234. Matches must not have alphanumeric characters
// on either side: 123 matches, a123 does not match, =123 matches.
func EnvToBody(body string, env *BrunoEnv) string {
	if env == nil || body == "" {
		return body
	}

	// Sort keys by length descending to match longer values first
	keys := make([]string, 0, len(env.ReverseVars))
	for k := range env.ReverseVars {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return len(keys[i]) > len(keys[j])
	})

	for _, value := range keys {
		varName := env.ReverseVars[value]
		// Build a regex that matches the value only when not surrounded by alphanumeric characters
		pattern := `(?:^|(?<=[^a-zA-Z0-9]))` + regexp.QuoteMeta(value) + `(?:$|(?=[^a-zA-Z0-9]))`
		re := regexp.MustCompile(pattern)
		body = re.ReplaceAllString(body, "{{"+varName+"}}")
	}

	return body
}

// EnvToPath replaces parts of the path with {{key}} using env values.
// The path is in the format api/users/123.
// If it finds a part in env.ReverseVars[part], it replaces it with the corresponding value.
// For example, if `123` maps to `user_id`, it returns api/users/{{user_id}}.
func EnvToPath(path string, env *BrunoEnv) string {
	if env == nil || len(env.ReverseVars) == 0 {
		return path
	}

	parts := strings.Split(path, "/")
	for i, part := range parts {
		if key, ok := env.ReverseVars[part]; ok {
			parts[i] = "{{" + key + "}}"
		}
	}
	return strings.Join(parts, "/")
}

func EnvFromFile(path string) (*BrunoEnv, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %q file error %w", path, err)
	}

	return ParseBrunoEnv(string(data))
}

// ParseBrunoEnv parses the content of a Bruno environment file
func ParseBrunoEnv(content string) (*BrunoEnv, error) {
	env := &BrunoEnv{
		Vars:        make(map[string]string),
		ReverseVars: make(map[string]string),
	}

	scanner := bufio.NewScanner(strings.NewReader(content))
	inVarsBlock := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines
		if line == "" {
			continue
		}

		// Detect start of vars block
		if line == "vars {" {
			inVarsBlock = true
			continue
		}

		// Detect end of vars block
		if line == "}" && inVarsBlock {
			inVarsBlock = false
			continue
		}

		// Parse key-value pairs inside the block
		if inVarsBlock {
			// Split by the first colon only
			// This ensures values containing colons (like User-Agents) are safe
			key, value, found := strings.Cut(line, ":")
			if !found {
				continue // Skip malformed lines
			}

			k := strings.TrimSpace(key)
			v := strings.TrimSpace(value)

			env.Vars[k] = v
			env.ReverseVars[v] = k
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return env, nil
}

// EnvGenerate generate file content from map
// env block in format
//
//	vars {
//	  User-Agent: go1.1
//	}
func EnvGenerate(vars map[string]string) string {
	return NameBlockMap("vars", vars)
}

func DefaultEnvBru(name string) string {
	vars := make(map[string]string)
	vars["host"] = name
	vars["proto"] = "https"
	vars["ua"] = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36"

	return EnvGenerate(vars)
}
