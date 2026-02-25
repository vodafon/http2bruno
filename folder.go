package main

import "strings"

// default folder.bru
//
//	meta {}
//	headers {}
func DefaultFolderBru(name string) string {
	meta := make(map[string]string)
	meta["name"] = name

	var sb strings.Builder
	sb.WriteString(MetaGenerate(meta))
	return sb.String()
}
