package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/dto"
)

func TestHandleKey_UndoRedo_RevertsAndReappliesInsertAdd(t *testing.T) {
	// Arrange
	model := &Model{
		read: runtimeReadState{
			viewMode: ViewRecords,
			focus:    FocusContent,
			schema: dto.Schema{
				Columns: []dto.SchemaColumn{
					{Name: "name", Type: "TEXT", Nullable: false},
				},
			},
		},
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})

	// Assert
	if len(model.currentStagingSnapshot().PendingInserts) != 1 {
		t.Fatalf("expected one pending insert after add, got %d", len(model.currentStagingSnapshot().PendingInserts))
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})

	// Assert
	if len(model.currentStagingSnapshot().PendingInserts) != 0 {
		t.Fatalf("expected insert to be undone")
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyCtrlR})

	// Assert
	if len(model.currentStagingSnapshot().PendingInserts) != 1 {
		t.Fatalf("expected insert to be redone")
	}
}

func TestHandleKey_NewActionClearsRedoStack(t *testing.T) {
	// Arrange
	model := &Model{
		read: runtimeReadState{
			viewMode: ViewRecords,
			focus:    FocusContent,
			schema: dto.Schema{
				Columns: []dto.SchemaColumn{
					{Name: "name", Type: "TEXT", Nullable: false},
				},
			},
		},
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})

	// Assert
	model.handleKey(tea.KeyMsg{Type: tea.KeyCtrlR})

	// Assert
	if len(model.currentStagingSnapshot().PendingInserts) != 1 {
		t.Fatalf("expected redo to have no effect after redo stack clear, got %d inserts", len(model.currentStagingSnapshot().PendingInserts))
	}
}

func TestHandleKey_UndoRedo_RevertsAndReappliesPersistedCellEdit(t *testing.T) {
	// Arrange
	model := &Model{
		read: runtimeReadState{
			viewMode: ViewRecords,
			focus:    FocusContent,
			records:  []dto.RecordRow{{Values: []string{"1", "alice"}}},
			schema: dto.Schema{
				Columns: []dto.SchemaColumn{
					{Name: "id", Type: "INTEGER", PrimaryKey: true},
					{Name: "name", Type: "TEXT"},
				},
			},
		},
	}
	if err := model.stageEdit(0, 1, dto.StagedValue{Text: "bob", Raw: "bob"}); err != nil {
		t.Fatalf("expected staged edit, got error %v", err)
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})

	// Assert
	if len(model.currentStagingSnapshot().PendingUpdates) != 0 {
		t.Fatalf("expected persisted edit to be undone")
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyCtrlR})

	// Assert
	edits := model.currentStagingSnapshot().PendingUpdates["id=1"]
	change, ok := edits.Changes[1]
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
		read: runtimeReadState{
			viewMode: ViewRecords,
			focus:    FocusContent,
			records:  []dto.RecordRow{{Values: []string{"1", "alice"}}},
			schema: dto.Schema{
				Columns: []dto.SchemaColumn{
					{Name: "id", Type: "INTEGER", PrimaryKey: true},
					{Name: "name", Type: "TEXT"},
				},
			},
		},
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})

	// Assert
	if len(model.currentStagingSnapshot().PendingDeletes) != 1 {
		t.Fatalf("expected delete toggle to be staged")
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})

	// Assert
	if len(model.currentStagingSnapshot().PendingDeletes) != 0 {
		t.Fatalf("expected delete toggle to be undone")
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyCtrlR})

	// Assert
	if len(model.currentStagingSnapshot().PendingDeletes) != 1 {
		t.Fatalf("expected delete toggle to be redone")
	}
}

func TestHandleKey_UndoRestorePendingInsert_PreservesAutoFieldVisibility(t *testing.T) {
	// Arrange
	model := withTestStaging(&Model{
		read: runtimeReadState{
			viewMode:         ViewRecords,
			focus:            FocusContent,
			recordSelection:  0,
			recordColumn:     1,
			recordFieldFocus: true,
			schema: dto.Schema{
				Columns: []dto.SchemaColumn{
					{Name: "id", Type: "INTEGER", PrimaryKey: true, AutoIncrement: true},
					{Name: "name", Type: "TEXT", Nullable: false},
				},
			},
		},
	}, stagingState{
		pendingInserts: []pendingInsertRow{
			{
				values: map[int]stagedEdit{
					0: {Value: dto.StagedValue{Text: "42", Raw: "42"}},
					1: {Value: dto.StagedValue{Text: "alice", Raw: "alice"}},
				},
				explicitAuto: map[int]bool{0: true},
				showAuto:     true,
			},
		},
	})

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})

	// Assert
	snapshot := model.currentStagingSnapshot()
	if len(snapshot.PendingInserts) != 1 {
		t.Fatalf("expected pending insert to be restored, got %d inserts", len(snapshot.PendingInserts))
	}
	if !snapshot.PendingInserts[0].ExplicitAuto[0] {
		t.Fatalf("expected restored insert to keep explicit auto flag for id column")
	}
	if got := model.visibleRowValue(0, 0); got != "42" {
		t.Fatalf("expected restored auto value 42 to stay visible, got %q", got)
	}
	if !containsInt(model.visibleColumnIndicesForSelection(), 0) {
		t.Fatalf("expected restored insert to show auto-increment column")
	}
}

func TestHandleKey_RedoRestoresAutoFieldVisibilityForPendingInsert(t *testing.T) {
	// Arrange
	model := &Model{
		read: runtimeReadState{
			viewMode: ViewRecords,
			focus:    FocusContent,
			schema: dto.Schema{
				Columns: []dto.SchemaColumn{
					{Name: "id", Type: "INTEGER", PrimaryKey: true, AutoIncrement: true},
					{Name: "name", Type: "TEXT", Nullable: false},
				},
			},
		},
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	model.handleKey(tea.KeyMsg{Type: tea.KeyCtrlA})
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})
	model.handleKey(tea.KeyMsg{Type: tea.KeyCtrlR})

	// Assert
	snapshot := model.currentStagingSnapshot()
	if len(snapshot.PendingInserts) != 1 {
		t.Fatalf("expected pending insert to be restored by redo, got %d inserts", len(snapshot.PendingInserts))
	}
	if !containsInt(model.visibleColumnIndicesForSelection(), 0) {
		t.Fatalf("expected redone insert to show auto-increment column")
	}
}

func TestClearStagedState_DropsAutoFieldVisibilityAcrossSessions(t *testing.T) {
	// Arrange
	model := withTestStaging(&Model{
		read: runtimeReadState{
			viewMode: ViewRecords,
			focus:    FocusContent,
			schema: dto.Schema{
				Columns: []dto.SchemaColumn{
					{Name: "id", Type: "INTEGER", PrimaryKey: true, AutoIncrement: true},
					{Name: "name", Type: "TEXT", Nullable: false},
				},
			},
		},
	}, stagingState{
		pendingInserts: []pendingInsertRow{
			{
				showAuto: true,
			},
		},
	})

	// Act
	model.clearStagedState()
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})

	// Assert
	if containsInt(model.visibleColumnIndicesForSelection(), 0) {
		t.Fatalf("expected new staging session to hide auto-increment column by default")
	}
}
