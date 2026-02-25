package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

type BrunoJSON struct {
	Version string   `json:"version"`
	Name    string   `json:"name"`
	Type    string   `json:"type"`
	Ignore  []string `json:"ignore"`
}

// DefaultBrunoJSON bruno.json file in collection directory
//
//	{
//	  "version": "1",
//	  "name": "comic.pixiv.net",
//	  "type": "collection",
//	  "ignore": [
//	    "node_modules",
//	    ".git"
//	  ]
//	}
func DefaultBrunoJSON(name string) string {
	brunoJSON := BrunoJSON{
		Version: "1",
		Name:    name,
		Type:    "collection",
		Ignore:  []string{"node_modules", ".git"},
	}
	data, err := json.MarshalIndent(brunoJSON, "", "  ")
	if err != nil {
		return fmt.Sprintf("{\"error\": \"%q\"}", err)
	}
	return string(data)
}

// DefaultCollectionBru collection.bru file in collection directory
//
//	headers {
//	  User-Agent: {{ua}}
//	}
func DefaultCollectionBru() string {
	heads := make(map[string]string)
	heads["User-Agent"] = "{{ua}}"
	heads["Cookie"] = "{{cook}}"
	heads["Authorization"] = "Bearer {{token}}"

	var sb strings.Builder
	sb.WriteString(HeadersGenerate(heads))
	sb.WriteString("\n")
	sb.WriteString(NameBlockStrings("script:pre-request", []string{"req.setMaxRedirects(0);"}))
	return sb.String()
}
