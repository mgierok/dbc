package tui

import "github.com/mgierok/dbc/internal/interfaces/tui/internal/primitives"

type RuntimeSessionState struct {
	RecordsPageLimit int
}

func (s *RuntimeSessionState) effectiveRecordsPageLimit() int {
	if s == nil || s.RecordsPageLimit <= 0 {
		return defaultRecordPageLimit
	}
	if s.RecordsPageLimit > primitives.RuntimeMaxRecordPageLimit {
		return primitives.RuntimeMaxRecordPageLimit
	}
	return s.RecordsPageLimit
}
