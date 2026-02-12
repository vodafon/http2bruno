package main

import "strings"

// default folder.bru
//
//	meta {}
//	headers {}
func DefaultFolderBru(name string) string {
	meta := make(map[string]string)
	meta["name"] = name

	heads := make(map[string]string)
	heads["Cookie"] = "{{cook}}"
	heads["Authorization"] = "Bearer {{token}}"

	var sb strings.Builder
	sb.WriteString(MetaGenerate(meta))
	sb.WriteString("\n")
	sb.WriteString(HeadersGenerate(heads))
	return sb.String()
}
