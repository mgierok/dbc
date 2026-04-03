package tui

import (
	"context"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/dto"
)

func newRuntimeCommandModel() *Model {
	return &Model{
		read: runtimeReadState{
			viewMode: ViewRecords,
			focus:    FocusContent,
		},
		ui: runtimeUIState{
			height: 40,
		},
	}
}

func newRuntimeCommandModelWithCurrentDatabase(current DatabaseOption) *Model {
	return &Model{
		read: runtimeReadState{
			viewMode: ViewRecords,
			focus:    FocusContent,
		},
		runtimeDatabaseSelectorDeps: runtimeDatabaseSelectorDepsForTest(current),
	}
}

func runtimeSaveTestSchema() dto.Schema {
	return dto.Schema{
		Columns: []dto.SchemaColumn{
			{Name: "id", Type: "INTEGER", PrimaryKey: true},
			{Name: "name", Type: "TEXT", Nullable: false},
		},
	}
}

func runtimeDirtyInsertSeed() stagingState {
	return stagingState{
		pendingInserts: []pendingInsertRow{
			{
				values: map[int]stagedEdit{
					0: {Value: dto.StagedValue{Text: "1", Raw: "1"}},
					1: {Value: dto.StagedValue{Text: "alice", Raw: "alice"}},
				},
				explicitAuto: map[int]bool{},
			},
		},
	}
}

func newRuntimeSaveModel(viewMode ViewMode, focus PanelFocus) *Model {
	return &Model{
		ctx:         context.Background(),
		saveChanges: &spySaveChangesUseCase{},
		read: runtimeReadState{
			viewMode:      viewMode,
			focus:         focus,
			tables:        []dto.Table{{Name: "users"}},
			selectedTable: 0,
			schema:        runtimeSaveTestSchema(),
		},
	}
}

func newDirtyRuntimeSaveModel(viewMode ViewMode, focus PanelFocus) *Model {
	return withTestStaging(newRuntimeSaveModel(viewMode, focus), runtimeDirtyInsertSeed())
}

func submitTypedRuntimeCommand(model *Model, value string) (tea.Model, tea.Cmd) {
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
	for _, r := range value {
		model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	return model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})
}

func assertRuntimeSessionActive(t *testing.T, cmd tea.Cmd, context string) {
	t.Helper()
	if cmd != nil {
		if _, ok := cmd().(tea.QuitMsg); ok {
			t.Fatalf("expected %s to keep session active", context)
		}
	}
}

func assertRuntimeSaveStarted(t *testing.T, model *Model, context string) {
	t.Helper()
	if model.overlay.confirmPopup.active {
		t.Fatalf("expected %s not to open a save popup", context)
	}
	if !model.ui.saveInFlight {
		t.Fatalf("expected %s to start save immediately", context)
	}
	if model.ui.statusMessage != "Saving changes..." {
		t.Fatalf("expected %s to show saving status, got %q", context, model.ui.statusMessage)
	}
}
