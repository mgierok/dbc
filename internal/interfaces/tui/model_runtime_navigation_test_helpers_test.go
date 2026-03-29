package tui

import (
	"context"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/dto"
	"github.com/mgierok/dbc/internal/application/usecase"
)

func runtimeTestDatabaseOption() DatabaseOption {
	return DatabaseOption{
		Name:       "primary",
		ConnString: "/tmp/primary.sqlite",
		Source:     DatabaseOptionSourceConfig,
	}
}

func runtimeNavigationTestSchema() dto.Schema {
	return dto.Schema{
		Columns: []dto.SchemaColumn{
			{Name: "id", Type: "INTEGER", PrimaryKey: true, AutoIncrement: true},
			{Name: "name", Type: "TEXT", Nullable: false},
		},
	}
}

func newDirtyRuntimeNavigationModel(saveChanges *spySaveChangesUseCase) *Model {
	current := runtimeTestDatabaseOption()
	return withTestStaging(&Model{
		ctx:         context.Background(),
		saveChanges: saveChanges,
		read: runtimeReadState{
			viewMode: ViewRecords,
			focus:    FocusContent,
			schema:   runtimeNavigationTestSchema(),
			tables:   []dto.Table{{Name: "users"}},
		},
		runtimeDatabaseSelectorDeps: runtimeDatabaseSelectorDepsForTest(current),
	}, stagingState{
		pendingInserts: []pendingInsertRow{
			{
				values: map[int]stagedEdit{
					0: {Value: dto.StagedValue{Text: "", Raw: ""}},
					1: {Value: dto.StagedValue{Text: "new", Raw: "new"}},
				},
				explicitAuto: map[int]bool{},
			},
		},
	})
}

func newPendingDatabaseNavigationConfirmModel(saveChanges *spySaveChangesUseCase) *Model {
	current := runtimeTestDatabaseOption()
	return withTestStaging(&Model{
		saveChanges: saveChanges,
		read: runtimeReadState{
			viewMode:      ViewRecords,
			focus:         FocusContent,
			tables:        []dto.Table{{Name: "users"}},
			selectedTable: 0,
			schema: dto.Schema{
				Columns: []dto.SchemaColumn{
					{Name: "id", Type: "INTEGER", PrimaryKey: true},
					{Name: "name", Type: "TEXT", Nullable: false},
				},
			},
		},
		overlay: runtimeOverlayState{
			confirmPopup: confirmPopup{
				active: true,
				options: []confirmOption{
					{label: "Save and reload database", decisionID: usecase.DirtyDecisionSave},
				},
			},
		},
		ui: runtimeUIState{
			pendingNavigation: &usecase.PendingRuntimeNavigation{
				Action: usecase.RuntimeNavigationAction{
					Kind: usecase.RuntimeNavigationActionOpenDatabase,
					DatabaseTarget: usecase.RuntimeDatabaseTarget{
						Option: usecase.RuntimeDatabaseOption{
							Name:       current.Name,
							ConnString: current.ConnString,
							Source:     usecase.RuntimeDatabaseOptionSourceConfig,
						},
						TransitionKind: usecase.RuntimeDatabaseTransitionReloadCurrent,
					},
				},
			},
		},
	}, stagingState{
		pendingInserts: []pendingInsertRow{
			{
				values: map[int]stagedEdit{
					0: {Value: dto.StagedValue{Text: "1", Raw: "1"}},
					1: {Value: dto.StagedValue{Text: "bob", Raw: "bob"}},
				},
				explicitAuto: map[int]bool{},
			},
		},
	})
}

func assertQuitCommand(t *testing.T, cmd tea.Cmd, context string) {
	t.Helper()
	if cmd == nil {
		t.Fatalf("expected quit command for %s", context)
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatalf("expected tea.QuitMsg for %s, got %T", context, cmd())
	}
}
