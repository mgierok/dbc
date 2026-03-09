package primitives

import "strings"

func renderStatusWithRightHint(left, right string, width int) string {
	if width <= 0 {
		width = 80
	}

	left = strings.TrimSpace(left)
	right = strings.TrimSpace(right)
	if right == "" {
		return padRight(left, width)
	}

	right = truncate(right, width)
	rightWidth := textWidth(right)
	if rightWidth >= width {
		return padRight(right, width)
	}

	leftWidth := width - rightWidth - 1
	if leftWidth <= 0 {
		return padRight(right, width)
	}

	if left == "" {
		return padRight(strings.Repeat(" ", leftWidth+1)+right, width)
	}

	return padRight(truncate(left, leftWidth), leftWidth) + " " + right
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
