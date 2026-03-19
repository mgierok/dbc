package tui

import (
	"strings"

	"github.com/mgierok/dbc/internal/interfaces/tui/internal/primitives"
)

const (
	commandSpotlightMinFieldWidth = 10
	commandSpotlightMinWidth      = commandSpotlightMinFieldWidth + 1 + 2
	commandSpotlightHeight        = 3
)

func (m *Model) renderCommandSpotlight(totalWidth int) []string {
	boxWidth := m.commandSpotlightWidth(totalWidth)
	innerWidth := boxWidth - 2
	if innerWidth < 1 {
		innerWidth = 1
	}

	return []string{
		primitives.FrameTopLeft + strings.Repeat(primitives.FrameHorizontal, innerWidth) + primitives.FrameTopRight,
		primitives.FrameVertical + primitives.PadRight(m.styles.Render(primitives.SemanticRoleBody, m.visibleCommandPrompt(innerWidth)), innerWidth) + primitives.FrameVertical,
		primitives.FrameBottomLeft + strings.Repeat(primitives.FrameHorizontal, innerWidth) + primitives.FrameBottomRight,
	}
}

func (m *Model) commandSpotlightWidth(totalWidth int) int {
	if totalWidth <= 0 {
		totalWidth = 80
	}

	width := totalWidth / 2
	if width < commandSpotlightMinWidth {
		width = commandSpotlightMinWidth
	}
	if width > totalWidth {
		width = totalWidth
	}
	if width < commandSpotlightHeight {
		width = commandSpotlightHeight
	}
	return width
}

func (m *Model) visibleCommandPrompt(width int) string {
	if !m.overlay.commandInput.active || width <= 0 {
		return ""
	}
	if width == 1 {
		return ":"
	}

	valueWithCaret := []rune(m.commandInputValueWithCaret())
	visibleValueWidth := width - 1
	if len(valueWithCaret) <= visibleValueWidth {
		return ":" + string(valueWithCaret)
	}

	// Keep the current cursor/scroll behavior aligned with commandInput.cursor,
	// which is still a byte offset. This intentionally preserves today's mixed
	// byte/rune clipping semantics for long non-ASCII input until the input
	// model itself is made rune-aware in a dedicated change.
	caretIndex := clamp(m.overlay.commandInput.cursor, 0, len(m.overlay.commandInput.value))
	start := 0
	if caretIndex >= visibleValueWidth {
		start = caretIndex - visibleValueWidth + 1
	}
	maxStart := len(valueWithCaret) - visibleValueWidth
	if maxStart < 0 {
		maxStart = 0
	}
	start = clamp(start, 0, maxStart)
	end := start + visibleValueWidth
	if end > len(valueWithCaret) {
		end = len(valueWithCaret)
	}

	return ":" + string(valueWithCaret[start:end])
}
