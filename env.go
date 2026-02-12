package main

import (
	"bufio"
	"strings"
)

// BrunoEnv represents the parsed environment variables
type BrunoEnv struct {
	Vars        map[string]string
	ReverseVars map[string]string
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
	vars["ua"] = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36"

	return EnvGenerate(vars)
}
