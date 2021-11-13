package main

import (
	"strings"
)

// result shares slice with lines parameter
func removeCmd(lines []string, prefix botCmd) []string {
	if len(lines) < 1 {
		return lines
	}
	prefixlen := len(prefix)
	if strings.ToLower(lines[0][:prefixlen]) != string(prefix) {
		return lines
	}
	lines[0] = strings.TrimSpace(lines[0][prefixlen:])
	return lines
}

func dropRunes(line string, runes string) string {
	filter := func(r rune) rune {
		for _, drop := range runes {
			if r == drop {
				return -1
			}
		}
		return r
	}
	return strings.Map(filter, line)
}

func endsWith(line string, suffixes []string) bool {
	for _, suffix := range suffixes {
		if strings.HasSuffix(line, suffix) {
			return true
		}
	}
	return false
}

func splitTrim(line string, separator string) (result []string) {
	result = strings.Split(line, separator)
	for i, line := range result {
		result[i] = strings.TrimSpace(line)
	}
	return
}
