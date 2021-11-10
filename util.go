package main

import (
	"strings"
)

// result shares slice with lines parameter
func removeCmd(lines []string, prefix botCmd) (result []string) {
	if len(lines) < 1 {
		return
	}
	prefixlen := len(prefix)
	result = lines
	if strings.ToLower(result[0][:prefixlen]) != string(prefix) {
		return
	}
	result[0] = strings.TrimSpace(result[0][prefixlen:])
	return
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
