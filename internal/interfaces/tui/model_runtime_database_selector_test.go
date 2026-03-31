package tui

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/dto"
	"github.com/mgierok/dbc/internal/application/port"
	"github.com/mgierok/dbc/internal/application/usecase"
)

type fakeRuntimeSelectorConfigStore struct {
	entries []port.ConfigEntry
}

func (f *fakeRuntimeSelectorConfigStore) List(ctx context.Context) ([]port.ConfigEntry, error) {
	result := make([]port.ConfigEntry, len(f.entries))
	copy(result, f.entries)
	return result, nil
}

func (f *fakeRuntimeSelectorConfigStore) Create(ctx context.Context, entry port.ConfigEntry) error {
	f.entries = append(f.entries, entry)
	return nil
}

func (f *fakeRuntimeSelectorConfigStore) Update(ctx context.Context, index int, entry port.ConfigEntry) error {
	if index < 0 || index >= len(f.entries) {
		return errors.New("index out of range")
	}
	f.entries[index] = entry
	return nil
}

func (f *fakeRuntimeSelectorConfigStore) Delete(ctx context.Context, index int) error {
	if index < 0 || index >= len(f.entries) {
		return errors.New("index out of range")
	}
	f.entries = append(f.entries[:index], f.entries[index+1:]...)
	return nil
}

func (f *fakeRuntimeSelectorConfigStore) ActivePath(ctx context.Context) (string, error) {
	return "/tmp/config.json", nil
}

type fakeRuntimeSelectorConnectionChecker struct{}

func (fakeRuntimeSelectorConnectionChecker) CanConnect(ctx context.Context, dbPath string) error {
	return nil
}

func runtimeDatabaseSelectorDepsForTest(current DatabaseOption, additionalOptions ...DatabaseOption) *RuntimeDatabaseSelectorDeps {
	store := &fakeRuntimeSelectorConfigStore{
		entries: []port.ConfigEntry{
			{Name: current.Name, DBPath: current.ConnString},
		},
	}
	checker := fakeRuntimeSelectorConnectionChecker{}
	return &RuntimeDatabaseSelectorDeps{
		ListConfiguredDatabases:  usecase.NewListConfiguredDatabases(store),
		CreateConfiguredDatabase: usecase.NewCreateConfiguredDatabase(store, checker),
		UpdateConfiguredDatabase: usecase.NewUpdateConfiguredDatabase(store, checker),
		DeleteConfiguredDatabase: usecase.NewDeleteConfiguredDatabase(store),
		GetActiveConfigPath:      usecase.NewGetActiveConfigPath(store),
		CurrentDatabase:          current,
		AdditionalOptions:        additionalOptions,
	}
}

func TestHandleKey_ConfigCommandOpensRuntimeDatabaseSelectorPopupWithoutQuitting(t *testing.T) {
	// Arrange
	current := DatabaseOption{
		Name:       "primary",
		ConnString: "/tmp/primary.sqlite",
		Source:     DatabaseOptionSourceConfig,
	}
	model := &Model{
		ctx: context.Background(),
		read: runtimeReadState{
			viewMode: ViewRecords,
			focus:    FocusContent,
		},
		runtimeDatabaseSelectorDeps: runtimeDatabaseSelectorDepsForTest(current),
	}

	// Act
	updated, cmd := submitTypedRuntimeCommand(model, "config")

	// Assert
	assertRuntimeSessionActive(t, cmd, ":config")
	runtimeModel := updated.(*Model)
	if !runtimeModel.overlay.databaseSelector.active {
		t.Fatal("expected runtime database selector popup to open")
	}
}

func TestHandleKey_RuntimeDatabaseSelectorEscClosesPopupAndPreservesRuntimeState(t *testing.T) {
	// Arrange
	current := DatabaseOption{
		Name:       "primary",
		ConnString: "/tmp/primary.sqlite",
		Source:     DatabaseOptionSourceConfig,
	}
	model := &Model{
		ctx: context.Background(),
		read: runtimeReadState{
			focus:            FocusContent,
			viewMode:         ViewRecords,
			selectedTable:    1,
			recordPageIndex:  2,
			recordSelection:  3,
			recordColumn:     1,
			recordFieldFocus: true,
			currentFilter: &dto.Filter{
				Column:   "name",
				Operator: dto.Operator{Name: "Equals", Kind: dto.OperatorKindEq, RequiresValue: true},
				Value:    "alice",
			},
			currentSort: &dto.Sort{
				Column:    "id",
				Direction: dto.SortDirectionDesc,
			},
		},
		runtimeDatabaseSelectorDeps: runtimeDatabaseSelectorDepsForTest(current),
	}
	model.overlay.recordDetail = recordDetailState{
		active:       true,
		scrollOffset: 4,
	}
	model.openRuntimeDatabaseSelectorPopup()

	// Act
	updated, cmd := model.handleKey(tea.KeyMsg{Type: tea.KeyEsc})

	// Assert
	assertRuntimeSessionActive(t, cmd, "Esc in runtime database selector")
	runtimeModel := updated.(*Model)
	if runtimeModel.overlay.databaseSelector.active {
		t.Fatal("expected runtime database selector popup to close")
	}
	if runtimeModel.read.focus != FocusContent {
		t.Fatalf("expected focus to be preserved, got %v", runtimeModel.read.focus)
	}
	if runtimeModel.read.viewMode != ViewRecords {
		t.Fatalf("expected records view to be preserved, got %v", runtimeModel.read.viewMode)
	}
	if runtimeModel.read.recordPageIndex != 2 {
		t.Fatalf("expected record page index to stay unchanged, got %d", runtimeModel.read.recordPageIndex)
	}
	if runtimeModel.read.recordSelection != 3 {
		t.Fatalf("expected record selection to stay unchanged, got %d", runtimeModel.read.recordSelection)
	}
	if runtimeModel.read.recordColumn != 1 {
		t.Fatalf("expected record column to stay unchanged, got %d", runtimeModel.read.recordColumn)
	}
	if !runtimeModel.read.recordFieldFocus {
		t.Fatal("expected field focus to stay unchanged")
	}
	if !runtimeModel.overlay.recordDetail.active {
		t.Fatal("expected record detail popup to be restored after closing database selector")
	}
}

func TestHandleKey_RuntimeDatabaseSelectorSelectionOfCurrentDatabaseRequestsReopenAndQuits(t *testing.T) {
	// Arrange
	current := DatabaseOption{
		Name:       "primary",
		ConnString: "/tmp/primary.sqlite",
		Source:     DatabaseOptionSourceConfig,
	}
	model := &Model{
		ctx: context.Background(),
		read: runtimeReadState{
			viewMode: ViewRecords,
			focus:    FocusContent,
		},
		runtimeDatabaseSelectorDeps: runtimeDatabaseSelectorDepsForTest(current),
	}
	model.openRuntimeDatabaseSelectorPopup()

	// Act
	updated, cmd := model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})

	// Assert
	if cmd == nil {
		t.Fatal("expected quit command after selecting current runtime database")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatalf("expected tea.QuitMsg after selecting current runtime database, got %T", cmd())
	}
	runtimeModel := updated.(*Model)
	if runtimeModel.exitResult.Action != RuntimeExitActionOpenDatabaseNext {
		t.Fatalf("expected runtime reopen exit action, got %v", runtimeModel.exitResult.Action)
	}
	if runtimeModel.exitResult.NextDatabase.ConnString != current.ConnString {
		t.Fatalf("expected reopen target %q, got %q", current.ConnString, runtimeModel.exitResult.NextDatabase.ConnString)
	}
}

func TestHandleRuntimeDatabaseSelection_EquivalentCurrentDatabaseRequestsReopen(t *testing.T) {
	// Arrange
	basePath := filepath.Join(t.TempDir(), "primary.sqlite")
	current := DatabaseOption{
		Name:       "primary",
		ConnString: basePath,
		Source:     DatabaseOptionSourceConfig,
	}
	model := &Model{
		ctx: context.Background(),
		read: runtimeReadState{
			viewMode: ViewRecords,
			focus:    FocusContent,
		},
		runtimeDatabaseSelectorDeps: runtimeDatabaseSelectorDepsForTest(current),
	}
	model.openRuntimeDatabaseSelectorPopup()

	selected := DatabaseOption{
		Name:       "primary equivalent",
		ConnString: basePath + string(os.PathSeparator) + ".",
		Source:     DatabaseOptionSourceCLI,
	}

	// Act
	updated, cmd := model.handleRuntimeDatabaseSelection(selected)

	// Assert
	if cmd == nil {
		t.Fatal("expected quit command for equivalent current runtime database")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatalf("expected tea.QuitMsg for equivalent current runtime database, got %T", cmd())
	}
	runtimeModel := updated.(*Model)
	if runtimeModel.exitResult.Action != RuntimeExitActionOpenDatabaseNext {
		t.Fatalf("expected runtime reopen exit action, got %v", runtimeModel.exitResult.Action)
	}
	if runtimeModel.exitResult.NextDatabase.ConnString != current.ConnString {
		t.Fatalf("expected reopen target %q, got %q", current.ConnString, runtimeModel.exitResult.NextDatabase.ConnString)
	}
}

func TestHandleKey_RuntimeDatabaseSelectorSelectionUsesLiveConfiguredIdentityWhenCurrentConfigIdentityIsStale(t *testing.T) {
	// Arrange
	basePath := filepath.Join(t.TempDir(), "primary.sqlite")
	current := DatabaseOption{
		Name:       "primary-old",
		ConnString: basePath,
		Source:     DatabaseOptionSourceConfig,
	}
	store := &fakeRuntimeSelectorConfigStore{
		entries: []port.ConfigEntry{
			{Name: "primary-renamed", DBPath: basePath},
		},
	}
	checker := fakeRuntimeSelectorConnectionChecker{}
	model := &Model{
		ctx: context.Background(),
		read: runtimeReadState{
			viewMode: ViewRecords,
			focus:    FocusContent,
		},
		runtimeDatabaseSelectorDeps: &RuntimeDatabaseSelectorDeps{
			ListConfiguredDatabases:  usecase.NewListConfiguredDatabases(store),
			CreateConfiguredDatabase: usecase.NewCreateConfiguredDatabase(store, checker),
			UpdateConfiguredDatabase: usecase.NewUpdateConfiguredDatabase(store, checker),
			DeleteConfiguredDatabase: usecase.NewDeleteConfiguredDatabase(store),
			GetActiveConfigPath:      usecase.NewGetActiveConfigPath(store),
			CurrentDatabase:          current,
		},
	}
	model.openRuntimeDatabaseSelectorPopup()

	// Act
	updated, cmd := model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})

	// Assert
	if cmd == nil {
		t.Fatal("expected quit command after selecting runtime database")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatalf("expected tea.QuitMsg after selecting runtime database, got %T", cmd())
	}
	runtimeModel := updated.(*Model)
	if runtimeModel.exitResult.Action != RuntimeExitActionOpenDatabaseNext {
		t.Fatalf("expected runtime reopen exit action, got %v", runtimeModel.exitResult.Action)
	}
	if runtimeModel.exitResult.NextDatabase.ConnString != basePath {
		t.Fatalf("expected reopen target %q, got %q", basePath, runtimeModel.exitResult.NextDatabase.ConnString)
	}
	if runtimeModel.exitResult.NextDatabase.Name != "primary-renamed" {
		t.Fatalf("expected live configured name %q, got %q", "primary-renamed", runtimeModel.exitResult.NextDatabase.Name)
	}
	if runtimeModel.exitResult.NextDatabase.Source != DatabaseOptionSourceConfig {
		t.Fatalf("expected configured source, got %q", runtimeModel.exitResult.NextDatabase.Source)
	}
}

func TestHandleKey_RuntimeDatabaseSelectorSelectionHonorsExplicitConfiguredAliasChoiceForEquivalentTarget(t *testing.T) {
	// Arrange
	basePath := filepath.Join(t.TempDir(), "primary.sqlite")
	current := DatabaseOption{
		Name:       "primary",
		ConnString: basePath,
		Source:     DatabaseOptionSourceConfig,
	}
	store := &fakeRuntimeSelectorConfigStore{
		entries: []port.ConfigEntry{
			{Name: "primary", DBPath: basePath},
			{Name: "primary-copy", DBPath: basePath},
		},
	}
	checker := fakeRuntimeSelectorConnectionChecker{}
	model := &Model{
		ctx: context.Background(),
		read: runtimeReadState{
			viewMode: ViewRecords,
			focus:    FocusContent,
		},
		runtimeDatabaseSelectorDeps: &RuntimeDatabaseSelectorDeps{
			ListConfiguredDatabases:  usecase.NewListConfiguredDatabases(store),
			CreateConfiguredDatabase: usecase.NewCreateConfiguredDatabase(store, checker),
			UpdateConfiguredDatabase: usecase.NewUpdateConfiguredDatabase(store, checker),
			DeleteConfiguredDatabase: usecase.NewDeleteConfiguredDatabase(store),
			GetActiveConfigPath:      usecase.NewGetActiveConfigPath(store),
			CurrentDatabase:          current,
		},
	}
	model.openRuntimeDatabaseSelectorPopup()
	model.overlay.databaseSelector.controller.SetSelectedIndex(1)

	// Act
	updated, cmd := model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})

	// Assert
	if cmd == nil {
		t.Fatal("expected quit command after selecting runtime database")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatalf("expected tea.QuitMsg after selecting runtime database, got %T", cmd())
	}
	runtimeModel := updated.(*Model)
	if runtimeModel.exitResult.Action != RuntimeExitActionOpenDatabaseNext {
		t.Fatalf("expected runtime reopen exit action, got %v", runtimeModel.exitResult.Action)
	}
	if runtimeModel.exitResult.NextDatabase.ConnString != basePath {
		t.Fatalf("expected reopen target %q, got %q", basePath, runtimeModel.exitResult.NextDatabase.ConnString)
	}
	if runtimeModel.exitResult.NextDatabase.Name != "primary-copy" {
		t.Fatalf("expected selected configured name %q, got %q", "primary-copy", runtimeModel.exitResult.NextDatabase.Name)
	}
	if runtimeModel.exitResult.NextDatabase.Source != DatabaseOptionSourceConfig {
		t.Fatalf("expected configured source, got %q", runtimeModel.exitResult.NextDatabase.Source)
	}
}

func TestHandleKey_RuntimeDatabaseSelectorSelectionRequestsChosenTarget(t *testing.T) {
	// Arrange
	current := DatabaseOption{
		Name:       "primary",
		ConnString: "/tmp/primary.sqlite",
		Source:     DatabaseOptionSourceConfig,
	}
	model := &Model{
		ctx: context.Background(),
		runtimeDatabaseSelectorDeps: runtimeDatabaseSelectorDepsForTest(current, DatabaseOption{
			Name:       "analytics",
			ConnString: "/tmp/analytics.sqlite",
			Source:     DatabaseOptionSourceCLI,
		}),
	}
	model.openRuntimeDatabaseSelectorPopup()
	model.overlay.databaseSelector.controller.SetSelectedIndex(1)

	// Act
	updated, cmd := model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})

	// Assert
	if cmd == nil {
		t.Fatal("expected quit command after selecting runtime database")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatalf("expected tea.QuitMsg after selecting runtime database, got %T", cmd())
	}
	runtimeModel := updated.(*Model)
	if runtimeModel.exitResult.Action != RuntimeExitActionOpenDatabaseNext {
		t.Fatalf("expected runtime reopen exit action, got %v", runtimeModel.exitResult.Action)
	}
	if runtimeModel.exitResult.NextDatabase.ConnString != "/tmp/analytics.sqlite" {
		t.Fatalf("expected selected reopen target /tmp/analytics.sqlite, got %q", runtimeModel.exitResult.NextDatabase.ConnString)
	}
}

func TestHandleKey_RuntimeDatabaseSelectorSelectionDoesNotResetCurrentRuntimeStateBeforeExit(t *testing.T) {
	// Arrange
	current := DatabaseOption{
		Name:       "primary",
		ConnString: "/tmp/primary.sqlite",
		Source:     DatabaseOptionSourceConfig,
	}
	switched := DatabaseOption{
		Name:       "analytics",
		ConnString: "/tmp/analytics.sqlite",
		Source:     DatabaseOptionSourceCLI,
	}
	runtimeSession := &RuntimeSessionState{RecordsPageLimit: 55}
	model := &Model{
		ctx:            context.Background(),
		runtimeSession: runtimeSession,
		read: runtimeReadState{
			focus:            FocusContent,
			viewMode:         ViewRecords,
			selectedTable:    1,
			recordPageIndex:  2,
			recordSelection:  3,
			recordColumn:     1,
			recordFieldFocus: true,
			currentFilter: &dto.Filter{
				Column:   "name",
				Operator: dto.Operator{Name: "Equals", Kind: dto.OperatorKindEq, RequiresValue: true},
				Value:    "alice",
			},
			currentSort: &dto.Sort{
				Column:    "id",
				Direction: dto.SortDirectionDesc,
			},
		},
		runtimeDatabaseSelectorDeps: runtimeDatabaseSelectorDepsForTest(current, switched),
	}
	model.overlay.recordDetail = recordDetailState{active: true, scrollOffset: 4}
	model.openRuntimeDatabaseSelectorPopup()
	model.overlay.databaseSelector.controller.SetSelectedIndex(1)

	// Act
	updated, cmd := model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected quit command after selecting runtime database")
	}

	// Assert
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatalf("expected tea.QuitMsg after selecting runtime database, got %T", cmd())
	}
	runtimeModel := updated.(*Model)
	if runtimeModel.read.focus != FocusContent {
		t.Fatalf("expected focus to stay unchanged before runtime exit, got %v", runtimeModel.read.focus)
	}
	if runtimeModel.read.viewMode != ViewRecords {
		t.Fatalf("expected view mode to stay unchanged before runtime exit, got %v", runtimeModel.read.viewMode)
	}
	if runtimeModel.runtimeSession != runtimeSession {
		t.Fatal("expected runtime session pointer to survive switch")
	}
	if runtimeModel.runtimeSession.RecordsPageLimit != 55 {
		t.Fatalf("expected runtime session state to survive switch, got %d", runtimeModel.runtimeSession.RecordsPageLimit)
	}
	if runtimeModel.exitResult.NextDatabase.ConnString != switched.ConnString {
		t.Fatalf("expected selected reopen target %q, got %q", switched.ConnString, runtimeModel.exitResult.NextDatabase.ConnString)
	}
}
