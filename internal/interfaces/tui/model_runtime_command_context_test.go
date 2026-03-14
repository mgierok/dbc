package tui

import "testing"

func TestNonBlockingRuntimeCommandContextActive_ReturnsFalseWhileSaveInFlight(t *testing.T) {
	// Arrange
	model := &Model{
		read: runtimeReadState{
			focus:    FocusContent,
			viewMode: ViewRecords,
		},
		ui: runtimeUIState{
			saveInFlight: true,
		},
	}

	// Act
	active := model.nonBlockingRuntimeCommandContextActive()

	// Assert
	if active {
		t.Fatal("expected non-blocking runtime context to be disabled while save is in flight")
	}
}
