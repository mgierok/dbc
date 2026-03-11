package tui

import runtimecontract "github.com/mgierok/dbc/internal/interfaces/tui/internal"

type RuntimeSessionState struct {
	RecordsPageLimit int
}

func (s *RuntimeSessionState) effectiveRecordsPageLimit() int {
	if s == nil || s.RecordsPageLimit <= 0 {
		return defaultRecordPageLimit
	}
	if s.RecordsPageLimit > runtimecontract.MaxRecordPageLimit {
		return runtimecontract.MaxRecordPageLimit
	}
	return s.RecordsPageLimit
}
