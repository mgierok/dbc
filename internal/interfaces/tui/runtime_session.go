package tui

type RuntimeSessionState struct {
	RecordsPageLimit int
}

func (s *RuntimeSessionState) effectiveRecordsPageLimit() int {
	if s == nil || s.RecordsPageLimit <= 0 {
		return defaultRecordPageLimit
	}
	if s.RecordsPageLimit > maxRuntimeRecordLimit {
		return maxRuntimeRecordLimit
	}
	return s.RecordsPageLimit
}
