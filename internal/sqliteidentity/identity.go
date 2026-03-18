package sqliteidentity

import (
	"path/filepath"
	"strings"
)

func Normalize(connString string) string {
	normalized := strings.TrimSpace(connString)
	if normalized == "" {
		return ""
	}

	normalized = filepath.Clean(normalized)
	if !filepath.IsAbs(normalized) {
		absPath, err := filepath.Abs(normalized)
		if err == nil {
			normalized = absPath
		}
	}

	return normalized
}

func Equivalent(left, right string) bool {
	normalizedLeft := Normalize(left)
	normalizedRight := Normalize(right)
	if normalizedLeft == "" || normalizedRight == "" {
		return false
	}

	return normalizedLeft == normalizedRight
}
