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

func (s RenderStyles) Render(role SemanticRole, text string) string {
	switch role {
	case SemanticRoleBody:
		if s.variant == renderVariantBackdrop {
			return s.wrap(text, sgrFaint)
		}
		return text
	case SemanticRoleMuted:
		return s.wrap(text, sgrFaint)
	case SemanticRoleTitle:
		if s.variant == renderVariantBackdrop {
			return s.wrap(text, sgrFaint)
		}
		return s.wrap(text, sgrBold)
	case SemanticRoleHeader:
		if s.variant == renderVariantBackdrop {
			return s.wrap(text, sgrFaint)
		}
		return s.wrap(text, sgrBold)
	case SemanticRoleSummary:
		if s.variant == renderVariantBackdrop {
			return s.wrap(text, sgrFaint)
		}
		return s.wrap(text, sgrBold)
	case SemanticRoleLabel:
		if s.variant == renderVariantBackdrop {
			return s.wrap(text, sgrFaint)
		}
		return s.wrap(text, sgrBold)
	case SemanticRoleDirty:
		if s.variant == renderVariantBackdrop {
			return s.wrap(text, sgrFaint)
		}
		return s.wrap(text, sgrBold)
	case SemanticRoleError:
		if s.variant == renderVariantBackdrop {
			return s.wrap(text, sgrFaint, sgrUnderline)
		}
		return s.wrap(text, sgrBold, sgrUnderline)
	case SemanticRoleSelected:
		if s.variant == renderVariantBackdrop {
			return s.wrap(text, sgrFaint)
		}
		return s.wrap(text, sgrReverse)
	case SemanticRoleDeleted:
		if s.variant == renderVariantBackdrop {
			return s.wrap(text, sgrFaint, sgrStrike)
		}
		return s.wrap(text, sgrStrike)
	case SemanticRoleSelectedDeleted:
		if s.variant == renderVariantBackdrop {
			return s.wrap(text, sgrFaint, sgrStrike)
		}
		return s.wrap(text, sgrReverse, sgrStrike)
	default:
		if s.variant == renderVariantBackdrop {
			return s.wrap(text, sgrFaint)
		}
		return text
	}
}

func (s RenderStyles) RenderLine(line SemanticLine) string {
	var builder strings.Builder
	for _, span := range line {
		builder.WriteString(s.Render(span.Role, span.Text))
	}
	return builder.String()
}

func (s RenderStyles) Body(text string) string {
	return s.Render(SemanticRoleBody, text)
}

func (s RenderStyles) Title(text string) string {
	return s.Render(SemanticRoleTitle, text)
}

func (s RenderStyles) Header(text string) string {
	return s.Render(SemanticRoleHeader, text)
}

func (s RenderStyles) Selected(text string) string {
	return s.Render(SemanticRoleSelected, text)
}

func (s RenderStyles) Deleted(text string) string {
	return s.Render(SemanticRoleDeleted, text)
}

func (s RenderStyles) SelectedDeleted(text string) string {
	return s.Render(SemanticRoleSelectedDeleted, text)
}

func (s RenderStyles) Muted(text string) string {
	return s.Render(SemanticRoleMuted, text)
}

func (s RenderStyles) Error(text string) string {
	return s.Render(SemanticRoleError, text)
}

func (s RenderStyles) Dirty(text string) string {
	return s.Render(SemanticRoleDirty, text)
}

func (s RenderStyles) Label(text string) string {
	return s.Render(SemanticRoleLabel, text)
}

func (s RenderStyles) Summary(text string) string {
	return s.Render(SemanticRoleSummary, text)
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
