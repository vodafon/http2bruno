package main

import "strings"

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
