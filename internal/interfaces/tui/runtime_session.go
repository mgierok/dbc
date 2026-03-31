package tui

type RuntimeSessionState struct {
	RecordsPageLimit       int
	nextRuntimeBundleToken int
}

func (s *RuntimeSessionState) allocateRuntimeBundleToken() int {
	if s == nil {
		return 0
	}
	s.nextRuntimeBundleToken++
	if s.nextRuntimeBundleToken <= 0 {
		s.nextRuntimeBundleToken = 1
	}
	return s.nextRuntimeBundleToken
}
