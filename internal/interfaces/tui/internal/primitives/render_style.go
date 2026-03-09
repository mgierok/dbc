package primitives

import (
	"os"
	"strings"
)

const (
	sgrBold      = "1"
	sgrFaint     = "2"
	sgrUnderline = "4"
	sgrReverse   = "7"
	sgrReset     = "\x1b[0m"
)

type renderStyles struct {
	enabled bool
}

func resolveRenderStylesFromEnv() renderStyles {
	if _, disabled := os.LookupEnv("NO_COLOR"); disabled {
		return renderStyles{}
	}
	if strings.EqualFold(strings.TrimSpace(os.Getenv("TERM")), "dumb") {
		return renderStyles{}
	}
	return renderStyles{enabled: true}
}

func (s renderStyles) title(text string) string {
	return s.wrap(text, sgrBold)
}

func (s renderStyles) selected(text string) string {
	return s.wrap(text, sgrReverse)
}

func (s renderStyles) muted(text string) string {
	return s.wrap(text, sgrFaint)
}

func (s renderStyles) error(text string) string {
	return s.wrap(text, sgrBold, sgrUnderline)
}

func (s renderStyles) dirty(text string) string {
	return s.wrap(text, sgrBold)
}

func (s renderStyles) label(text string) string {
	return s.wrap(text, sgrBold)
}

func (s renderStyles) summary(text string) string {
	return s.wrap(text, sgrBold)
}

func (s renderStyles) wrap(text string, attrs ...string) string {
	if !s.enabled || text == "" || len(attrs) == 0 {
		return text
	}
	return "\x1b[" + strings.Join(attrs, ";") + "m" + text + sgrReset
}

func isErrorLikeMessage(message string) bool {
	trimmed := strings.TrimSpace(strings.ToLower(message))
	return strings.HasPrefix(trimmed, "error:") || strings.Contains(trimmed, "failed") || strings.Contains(trimmed, "invalid ")
}
