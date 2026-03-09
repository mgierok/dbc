package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/dto"
)

func TestHandleKey_UndoRedo_RevertsAndReappliesInsertAdd(t *testing.T) {
	// Arrange
	model := &Model{
		viewMode: ViewRecords,
		focus:    FocusContent,
		schema: dto.Schema{
			Columns: []dto.SchemaColumn{
				{Name: "name", Type: "TEXT", Nullable: false},
			},
		},
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})

	// Assert
	if len(model.pendingInserts) != 1 {
		t.Fatalf("expected one pending insert after add, got %d", len(model.pendingInserts))
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})

	// Assert
	if len(model.pendingInserts) != 0 {
		t.Fatalf("expected insert to be undone")
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyCtrlR})

	// Assert
	if len(model.pendingInserts) != 1 {
		t.Fatalf("expected insert to be redone")
	}
}

func TestHandleKey_NewActionClearsRedoStack(t *testing.T) {
	// Arrange
	model := &Model{
		viewMode: ViewRecords,
		focus:    FocusContent,
		schema: dto.Schema{
			Columns: []dto.SchemaColumn{
				{Name: "name", Type: "TEXT", Nullable: false},
			},
		},
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})

	// Assert
	if len(model.future) != 0 {
		t.Fatalf("expected redo stack to be cleared by new staged action")
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyCtrlR})

	// Assert
	if len(model.pendingInserts) != 1 {
		t.Fatalf("expected redo to have no effect after redo stack clear, got %d inserts", len(model.pendingInserts))
	}
}

func TestHandleKey_UndoRedo_RevertsAndReappliesPersistedCellEdit(t *testing.T) {
	// Arrange
	model := &Model{
		viewMode: ViewRecords,
		focus:    FocusContent,
		records:  []dto.RecordRow{{Values: []string{"1", "alice"}}},
		schema: dto.Schema{
			Columns: []dto.SchemaColumn{
				{Name: "id", Type: "INTEGER", PrimaryKey: true},
				{Name: "name", Type: "TEXT"},
			},
		},
	}
	if err := model.stageEdit(0, 1, dto.StagedValue{Text: "bob", Raw: "bob"}); err != nil {
		t.Fatalf("expected staged edit, got error %v", err)
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})

	// Assert
	if len(model.pendingUpdates) != 0 {
		t.Fatalf("expected persisted edit to be undone")
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyCtrlR})

	// Assert
	edits := model.pendingUpdates["id=1"]
	change, ok := edits.changes[1]
	if !ok {
		t.Fatalf("expected persisted edit to be restored by redo")
	}
	if got := displayValue(change.Value); got != "bob" {
		t.Fatalf("expected restored value bob, got %q", got)
	}
}

func TestHandleKey_UndoRedo_RevertsAndReappliesDeleteToggle(t *testing.T) {
	// Arrange
	model := &Model{
		viewMode: ViewRecords,
		focus:    FocusContent,
		records:  []dto.RecordRow{{Values: []string{"1", "alice"}}},
		schema: dto.Schema{
			Columns: []dto.SchemaColumn{
				{Name: "id", Type: "INTEGER", PrimaryKey: true},
				{Name: "name", Type: "TEXT"},
			},
		},
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})

	// Assert
	if len(model.pendingDeletes) != 1 {
		t.Fatalf("expected delete toggle to be staged")
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})

	// Assert
	if len(model.pendingDeletes) != 0 {
		t.Fatalf("expected delete toggle to be undone")
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyCtrlR})

	// Assert
	if len(model.pendingDeletes) != 1 {
		t.Fatalf("expected delete toggle to be redone")
	}
}
