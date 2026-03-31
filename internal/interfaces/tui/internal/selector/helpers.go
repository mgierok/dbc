package selector

func clamp(value, max int) int {
	if value < 0 {
		return 0
	}
	if value > max {
		return max
	}
	return value
}
