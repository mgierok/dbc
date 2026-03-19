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
	sgrStrike    = "9"
	sgrReset     = "\x1b[0m"
)

type RenderStyles struct {
	enabled bool
	variant renderVariant
}

type renderVariant int

const (
	renderVariantNormal renderVariant = iota
	renderVariantBackdrop
)

func NewRenderStyles(enabled bool) RenderStyles {
	return RenderStyles{enabled: enabled, variant: renderVariantNormal}
}

func ResolveRenderStylesFromEnv() RenderStyles {
	if _, disabled := os.LookupEnv("NO_COLOR"); disabled {
		return RenderStyles{}
	}
	if strings.EqualFold(strings.TrimSpace(os.Getenv("TERM")), "dumb") {
		return RenderStyles{}
	}
	return RenderStyles{enabled: true}
}

func (s RenderStyles) Enabled() bool {
	return s.enabled
}

func (s RenderStyles) Backdrop() RenderStyles {
	return RenderStyles{enabled: s.enabled, variant: renderVariantBackdrop}
}

func (s RenderStyles) Title(text string) string {
	if s.variant == renderVariantBackdrop {
		return s.wrap(text, sgrFaint)
	}
	return s.wrap(text, sgrBold)
}

func (s RenderStyles) Selected(text string) string {
	if s.variant == renderVariantBackdrop {
		return s.wrap(text, sgrFaint)
	}
	return s.wrap(text, sgrReverse)
}

func (s RenderStyles) Deleted(text string) string {
	if s.variant == renderVariantBackdrop {
		return s.wrap(text, sgrFaint, sgrStrike)
	}
	return s.wrap(text, sgrStrike)
}

func (s RenderStyles) SelectedDeleted(text string) string {
	if s.variant == renderVariantBackdrop {
		return s.wrap(text, sgrFaint, sgrStrike)
	}
	return s.wrap(text, sgrReverse, sgrStrike)
}

func (s RenderStyles) Muted(text string) string {
	return s.wrap(text, sgrFaint)
}

func (s RenderStyles) Error(text string) string {
	if s.variant == renderVariantBackdrop {
		return s.wrap(text, sgrFaint, sgrUnderline)
	}
	return s.wrap(text, sgrBold, sgrUnderline)
}

func (s RenderStyles) Dirty(text string) string {
	if s.variant == renderVariantBackdrop {
		return s.wrap(text, sgrFaint)
	}
	return s.wrap(text, sgrBold)
}

func (s RenderStyles) Label(text string) string {
	if s.variant == renderVariantBackdrop {
		return s.wrap(text, sgrFaint)
	}
	return s.wrap(text, sgrBold)
}

func (s RenderStyles) Summary(text string) string {
	if s.variant == renderVariantBackdrop {
		return s.wrap(text, sgrFaint)
	}
	return s.wrap(text, sgrBold)
}

func (s RenderStyles) wrap(text string, attrs ...string) string {
	if !s.enabled || text == "" || len(attrs) == 0 {
		return text
	}
	return "\x1b[" + strings.Join(attrs, ";") + "m" + text + sgrReset
}

func IsErrorLikeMessage(message string) bool {
	trimmed := strings.TrimSpace(strings.ToLower(message))
	return strings.HasPrefix(trimmed, "error:") || strings.Contains(trimmed, "failed") || strings.Contains(trimmed, "invalid ")
}
