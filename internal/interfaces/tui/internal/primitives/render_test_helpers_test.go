package primitives

import "regexp"

var ansiSGRPattern = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func stripANSI(text string) string {
	return ansiSGRPattern.ReplaceAllString(text, "")
}
