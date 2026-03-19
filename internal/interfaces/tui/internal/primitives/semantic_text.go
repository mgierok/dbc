package primitives

import "strings"

type SemanticRole int

const (
	SemanticRoleBody SemanticRole = iota
	SemanticRoleMuted
	SemanticRoleTitle
	SemanticRoleHeader
	SemanticRoleSummary
	SemanticRoleLabel
	SemanticRoleDirty
	SemanticRoleError
	SemanticRoleSelected
	SemanticRoleDeleted
	SemanticRoleSelectedDeleted
)

type SemanticSpan struct {
	Text string
	Role SemanticRole
}

type SemanticLine []SemanticSpan

func Span(role SemanticRole, text string) SemanticSpan {
	return SemanticSpan{Text: text, Role: role}
}

func SemanticText(role SemanticRole, text string) SemanticLine {
	return SemanticLine{Span(role, text)}
}

func SemanticTexts(role SemanticRole, texts []string) []SemanticLine {
	lines := make([]SemanticLine, len(texts))
	for i, text := range texts {
		lines[i] = SemanticText(role, text)
	}
	return lines
}

func JoinSemanticLines(lines []SemanticLine, separator string, separatorRole SemanticRole) SemanticLine {
	var joined SemanticLine
	for i, line := range lines {
		if i > 0 && separator != "" {
			joined = append(joined, Span(separatorRole, separator))
		}
		joined = append(joined, line...)
	}
	return joined
}

func PrefixSemanticLine(prefix string, line SemanticLine) SemanticLine {
	if prefix == "" {
		return append(SemanticLine(nil), line...)
	}

	role := SemanticRoleBody
	if len(line) > 0 {
		role = line[0].Role
	}

	prefixed := SemanticLine{Span(role, prefix)}
	prefixed = append(prefixed, line...)
	return prefixed
}

func (l SemanticLine) PlainText() string {
	var builder strings.Builder
	for _, span := range l {
		builder.WriteString(span.Text)
	}
	return builder.String()
}
