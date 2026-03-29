package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/dto"
)

func TestHandleKey_EInRecordsEnablesFieldFocus(t *testing.T) {
	// Arrange
	model := &Model{
		read: runtimeReadState{
			viewMode: ViewRecords,
			focus:    FocusContent,
			records:  []dto.RecordRow{{Values: []string{"1"}}},
			schema: dto.Schema{
				Columns: []dto.SchemaColumn{
					{Name: "id", Type: "INTEGER", PrimaryKey: true},
				},
			},
		},
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})

	// Assert
	if !model.read.recordFieldFocus {
		t.Fatalf("expected record field focus to be enabled")
	}
}

func TestHandleKey_EInFieldFocusOpensEditPopup(t *testing.T) {
	// Arrange
	model := &Model{
		read: runtimeReadState{
			viewMode:         ViewRecords,
			focus:            FocusContent,
			recordFieldFocus: true,
			records:          []dto.RecordRow{{Values: []string{"1"}}},
			schema: dto.Schema{
				Columns: []dto.SchemaColumn{
					{Name: "id", Type: "INTEGER", PrimaryKey: true},
				},
			},
		},
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})

	// Assert
	if !model.overlay.editPopup.active {
		t.Fatalf("expected edit popup to be active")
	}
}

func TestHandleKey_EInFieldFocusBlocksRowsWithUnavailableIdentity(t *testing.T) {
	// Arrange
	model := &Model{
		read: runtimeReadState{
			viewMode:         ViewRecords,
			focus:            FocusContent,
			recordFieldFocus: true,
			records: []dto.RecordRow{
				{
					Values:              []string{"<truncated 262145 bytes>"},
					IdentityUnavailable: true,
				},
			},
			schema: dto.Schema{
				Columns: []dto.SchemaColumn{
					{Name: "id", Type: "TEXT", PrimaryKey: true},
				},
			},
		},
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})

	// Assert
	if model.overlay.editPopup.active {
		t.Fatal("expected edit popup to stay closed")
	}
	if model.ui.statusMessage != "Error: selected record identity exceeds safe browse limit" {
		t.Fatalf("expected explicit oversized-identity status, got %q", model.ui.statusMessage)
	}
}

func TestHandleKey_EInFieldFocusBlocksPlaceholderBackedTextCells(t *testing.T) {
	// Arrange
	model := &Model{
		read: runtimeReadState{
			viewMode:         ViewRecords,
			focus:            FocusContent,
			recordFieldFocus: true,
			recordColumn:     1,
			records: []dto.RecordRow{
				{
					Values:             []string{"1", "<truncated 262145 bytes>"},
					RowKey:             "id=1",
					Identity:           dto.RecordIdentity{Keys: []dto.RecordIdentityKey{{Column: "id", Value: dto.StagedValue{Text: "1", Raw: int64(1)}}}},
					EditableFromBrowse: []bool{true, false},
				},
			},
			schema: dto.Schema{
				Columns: []dto.SchemaColumn{
					{Name: "id", Type: "INTEGER", PrimaryKey: true},
					{Name: "note", Type: "TEXT"},
				},
			},
		},
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})

	// Assert
	if model.overlay.editPopup.active {
		t.Fatal("expected edit popup to stay closed")
	}
	if model.ui.statusMessage != "Error: selected cell has no safe editable source" {
		t.Fatalf("expected placeholder edit block status, got %q", model.ui.statusMessage)
	}
}

func TestHandleKey_EInFieldFocusBlocksPlaceholderBackedBlobCells(t *testing.T) {
	// Arrange
	model := &Model{
		read: runtimeReadState{
			viewMode:         ViewRecords,
			focus:            FocusContent,
			recordFieldFocus: true,
			recordColumn:     1,
			records: []dto.RecordRow{
				{
					Values:             []string{"1", "<blob 2 bytes>"},
					RowKey:             "id=1",
					Identity:           dto.RecordIdentity{Keys: []dto.RecordIdentityKey{{Column: "id", Value: dto.StagedValue{Text: "1", Raw: int64(1)}}}},
					EditableFromBrowse: []bool{true, false},
				},
			},
			schema: dto.Schema{
				Columns: []dto.SchemaColumn{
					{Name: "id", Type: "INTEGER", PrimaryKey: true},
					{Name: "payload", Type: "BLOB"},
				},
			},
		},
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})

	// Assert
	if model.overlay.editPopup.active {
		t.Fatal("expected edit popup to stay closed")
	}
	if model.ui.statusMessage != "Error: selected cell has no safe editable source" {
		t.Fatalf("expected placeholder edit block status, got %q", model.ui.statusMessage)
	}
}

func TestHandleKey_EInFieldFocusAllowsEditingExistingStagedValueOnPlaceholderCell(t *testing.T) {
	// Arrange
	model := withTestStaging(&Model{
		read: runtimeReadState{
			viewMode:         ViewRecords,
			focus:            FocusContent,
			recordFieldFocus: true,
			recordColumn:     1,
			records: []dto.RecordRow{
				{
					Values:             []string{"1", "<truncated 262145 bytes>"},
					RowKey:             "id=1",
					Identity:           dto.RecordIdentity{Keys: []dto.RecordIdentityKey{{Column: "id", Value: dto.StagedValue{Text: "1", Raw: int64(1)}}}},
					EditableFromBrowse: []bool{true, false},
				},
			},
			schema: dto.Schema{
				Columns: []dto.SchemaColumn{
					{Name: "id", Type: "INTEGER", PrimaryKey: true},
					{Name: "note", Type: "TEXT"},
				},
			},
		},
	}, stagingState{
		pendingUpdates: map[string]recordEdits{
			"id=1": {
				identity: dto.RecordIdentity{Keys: []dto.RecordIdentityKey{{Column: "id", Value: dto.StagedValue{Text: "1", Raw: int64(1)}}}},
				changes: map[int]stagedEdit{
					1: {Value: dto.StagedValue{Text: "edited", Raw: "edited"}},
				},
			},
		},
	})

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})

	// Assert
	if !model.overlay.editPopup.active {
		t.Fatal("expected edit popup to open for staged value")
	}
	if model.overlay.editPopup.input != "edited" {
		t.Fatalf("expected staged value to seed popup, got %q", model.overlay.editPopup.input)
	}
}
