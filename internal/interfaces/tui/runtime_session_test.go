package tui

import (
	"testing"

	runtimecontract "github.com/mgierok/dbc/internal/interfaces/tui/internal"
)

func TestRuntimeSessionState_EffectiveRecordsPageLimitClampsOversizedStoredValue(t *testing.T) {
	// Arrange
	session := &RuntimeSessionState{RecordsPageLimit: runtimecontract.MaxRecordPageLimit + 1}

	// Act
	limit := session.effectiveRecordsPageLimit()

	// Assert
	if limit != runtimecontract.MaxRecordPageLimit {
		t.Fatalf("expected oversized session limit to clamp to %d, got %d", runtimecontract.MaxRecordPageLimit, limit)
	}
}
