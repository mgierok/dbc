package tui

import runtimecontract "github.com/mgierok/dbc/internal/interfaces/tui/internal"

type RuntimeSessionState struct {
	RecordsPageLimit       int
	nextRuntimeBundleToken int
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
