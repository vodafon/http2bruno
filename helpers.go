package main

import (
	"os"
	"strings"
)

func NameBlockMap(name string, heads map[string]string) string {
	if len(heads) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(name)
	sb.WriteString(" {\n")
	sb.WriteString(BlockMap(heads))
	sb.WriteString("}\n")
	return sb.String()
}

func NameBlockStrings(name string, list []string) string {
	if len(list) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(name)
	sb.WriteString(" {\n")
	sb.WriteString(BlockStrings(list))
	sb.WriteString("}\n")
	return sb.String()
}

func BlockMap(m map[string]string) string {
	if len(m) == 0 {
		return ""
	}

	var sb strings.Builder
	for key, value := range m {
		sb.WriteString("  ")
		sb.WriteString(key)
		sb.WriteString(": ")
		sb.WriteString(value)
		sb.WriteString("\n")
	}
	return sb.String()
}

func BlockStrings(m []string) string {
	if len(m) == 0 {
		return ""
	}

	var sb strings.Builder
	for _, el := range m {
		sb.WriteString("  ")
		sb.WriteString(el)
		sb.WriteString("\n")
	}
	return sb.String()
}

// DirFilesCount returns count file in folder,
// non recursive, without subfolders
func DirFilesCount(dir string) int {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	count := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			count++
		}
	}
	return count
}
