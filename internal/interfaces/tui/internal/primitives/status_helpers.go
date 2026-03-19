package primitives

import "strings"

func RenderSemanticStatusWithRightHint(left, right SemanticLine, styles RenderStyles, width int) string {
	return renderStatusWithRightHint(styles.RenderLine(left), styles.RenderLine(right), width)
}

func RenderStatusWithRightHint(left, right string, width int) string {
	return renderStatusWithRightHint(left, right, width)
}

func renderStatusWithRightHint(left, right string, width int) string {
	if width <= 0 {
		width = 80
	}

	left = strings.TrimSpace(left)
	right = strings.TrimSpace(right)
	if right == "" {
		return PadRight(left, width)
	}

	right = Truncate(right, width)
	rightWidth := TextWidth(right)
	if rightWidth >= width {
		return PadRight(right, width)
	}

	leftWidth := width - rightWidth - 1
	if leftWidth <= 0 {
		return PadRight(right, width)
	}

	if left == "" {
		return PadRight(strings.Repeat(" ", leftWidth+1)+right, width)
	}

	return PadRight(Truncate(left, leftWidth), leftWidth) + " " + right
}

func clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
