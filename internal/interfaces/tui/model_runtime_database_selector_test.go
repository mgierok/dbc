package tui

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/dto"
	"github.com/mgierok/dbc/internal/application/port"
	"github.com/mgierok/dbc/internal/application/usecase"
)

type stubRuntimeDatabaseSwitcher struct {
	calls        int
	lastSelected DatabaseOption
	deps         RuntimeRunDeps
	err          error
}

func (s *stubRuntimeDatabaseSwitcher) Switch(ctx context.Context, selected DatabaseOption) (RuntimeRunDeps, error) {
	s.calls++
	s.lastSelected = selected
	return s.deps, s.err
}

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

func runtimeDatabaseSelectorDepsForTest(current DatabaseOption, switcher RuntimeDatabaseSwitcher, additionalOptions ...DatabaseOption) *RuntimeDatabaseSelectorDeps {
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
		SwitchDatabase:           switcher,
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
		runtimeDatabaseSelectorDeps: runtimeDatabaseSelectorDepsForTest(current, nil),
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
		runtimeDatabaseSelectorDeps: runtimeDatabaseSelectorDepsForTest(current, nil),
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

func TestHandleKey_RuntimeDatabaseSelectorSelectionOfCurrentDatabaseIsNoOp(t *testing.T) {
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
		runtimeDatabaseSelectorDeps: runtimeDatabaseSelectorDepsForTest(current, nil),
	}
	model.openRuntimeDatabaseSelectorPopup()

	// Act
	updated, cmd := model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})

	// Assert
	assertRuntimeSessionActive(t, cmd, "Enter on current runtime database")
	runtimeModel := updated.(*Model)
	if runtimeModel.overlay.databaseSelector.active {
		t.Fatal("expected runtime database selector popup to close after selecting current database")
	}
}

func TestHandleRuntimeDatabaseSelection_EquivalentCurrentDatabaseIsNoOp(t *testing.T) {
	// Arrange
	switcher := &stubRuntimeDatabaseSwitcher{}
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
		runtimeDatabaseSelectorDeps: runtimeDatabaseSelectorDepsForTest(current, switcher),
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
	assertRuntimeSessionActive(t, cmd, "equivalent current runtime database")
	runtimeModel := updated.(*Model)
	if runtimeModel.overlay.databaseSelector.active {
		t.Fatal("expected runtime database selector popup to close after equivalent current database selection")
	}
	if switcher.calls != 0 {
		t.Fatalf("expected no runtime switch for equivalent current database, got %d calls", switcher.calls)
	}
}

func TestHandleKey_RuntimeDatabaseSelectorFailedSwitchKeepsPopupOpen(t *testing.T) {
	// Arrange
	switcher := &stubRuntimeDatabaseSwitcher{err: errors.New("ping failed")}
	current := DatabaseOption{
		Name:       "primary",
		ConnString: "/tmp/primary.sqlite",
		Source:     DatabaseOptionSourceConfig,
	}
	model := &Model{
		ctx: context.Background(),
		runtimeDatabaseSelectorDeps: runtimeDatabaseSelectorDepsForTest(current, switcher, DatabaseOption{
			Name:       "analytics",
			ConnString: "/tmp/analytics.sqlite",
			Source:     DatabaseOptionSourceCLI,
		}),
		runtimeClose: func() {
			t.Fatal("expected current runtime close to stay unused after failed switch")
		},
	}
	model.openRuntimeDatabaseSelectorPopup()
	model.overlay.databaseSelector.controller.SetSelectedIndex(1)

	// Act
	updated, cmd := model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected async runtime switch command")
	}
	msg := cmd()
	runtimeModel := updated.(*Model)
	updated, followup := runtimeModel.Update(msg)

	// Assert
	if followup != nil {
		if _, ok := followup().(tea.QuitMsg); ok {
			t.Fatal("expected failed runtime switch to keep session active")
		}
	}
	runtimeModel = updated.(*Model)
	if !runtimeModel.overlay.databaseSelector.active {
		t.Fatal("expected runtime database selector popup to stay open after failed switch")
	}
	if runtimeModel.ui.runtimeSwitchInFlight {
		t.Fatal("expected runtime switch to leave in-flight state after failure")
	}
	if switcher.lastSelected.ConnString != "/tmp/analytics.sqlite" {
		t.Fatalf("expected switch attempt for /tmp/analytics.sqlite, got %q", switcher.lastSelected.ConnString)
	}
	view := runtimeModel.overlay.databaseSelector.controller.View()
	if !strings.Contains(view, "ping failed") {
		t.Fatalf("expected failed switch message in selector status, got %q", view)
	}
	if runtimeModel.currentRuntimeDatabaseConnString() != current.ConnString {
		t.Fatalf("expected current runtime database to stay %q, got %q", current.ConnString, runtimeModel.currentRuntimeDatabaseConnString())
	}
}

func TestHandleKey_RuntimeDatabaseSelectorSuccessfulSwitchReplacesRuntimeDeps(t *testing.T) {
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
	previousClosed := false
	nextClosed := false
	switcher := &stubRuntimeDatabaseSwitcher{
		deps: RuntimeRunDeps{
			DatabaseSelector: runtimeDatabaseSelectorDepsForTest(switched, nil),
			Close: func() {
				nextClosed = true
			},
		},
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
		runtimeDatabaseSelectorDeps: runtimeDatabaseSelectorDepsForTest(current, switcher, switched),
		runtimeClose: func() {
			previousClosed = true
		},
	}
	model.overlay.recordDetail = recordDetailState{active: true, scrollOffset: 4}
	model.openRuntimeDatabaseSelectorPopup()
	model.overlay.databaseSelector.controller.SetSelectedIndex(1)

	// Act
	updated, cmd := model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected async runtime switch command")
	}
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); ok {
		t.Fatal("expected successful runtime switch to stay in the same process")
	}
	runtimeModel := updated.(*Model)
	updated, followup := runtimeModel.Update(msg)

	// Assert
	if followup == nil {
		t.Fatal("expected successful runtime switch to reinitialize runtime data")
	}
	runtimeModel = updated.(*Model)
	if !previousClosed {
		t.Fatal("expected previous runtime resources to close after successful switch")
	}
	if nextClosed {
		t.Fatal("expected replacement runtime resources to remain active after switch")
	}
	if runtimeModel.currentRuntimeDatabaseConnString() != switched.ConnString {
		t.Fatalf("expected switched runtime database %q, got %q", switched.ConnString, runtimeModel.currentRuntimeDatabaseConnString())
	}
	if runtimeModel.overlay.databaseSelector.active {
		t.Fatal("expected runtime database selector popup to close after successful switch")
	}
	if runtimeModel.read.focus != FocusTables {
		t.Fatalf("expected focus reset to tables after switch, got %v", runtimeModel.read.focus)
	}
	if runtimeModel.read.viewMode != ViewSchema {
		t.Fatalf("expected view mode reset to schema after switch, got %v", runtimeModel.read.viewMode)
	}
	if runtimeModel.read.currentFilter != nil {
		t.Fatalf("expected filter reset after switch, got %+v", runtimeModel.read.currentFilter)
	}
	if runtimeModel.read.currentSort != nil {
		t.Fatalf("expected sort reset after switch, got %+v", runtimeModel.read.currentSort)
	}
	if runtimeModel.read.recordSelection != 0 {
		t.Fatalf("expected record selection reset after switch, got %d", runtimeModel.read.recordSelection)
	}
	if runtimeModel.read.recordColumn != 0 {
		t.Fatalf("expected record column reset after switch, got %d", runtimeModel.read.recordColumn)
	}
	if runtimeModel.read.recordFieldFocus {
		t.Fatal("expected record field focus reset after switch")
	}
	if runtimeModel.overlay.recordDetail.active {
		t.Fatal("expected record detail popup to reset after switch")
	}
	if runtimeModel.runtimeSession != runtimeSession {
		t.Fatal("expected runtime session pointer to survive switch")
	}
	if runtimeModel.runtimeSession.RecordsPageLimit != 55 {
		t.Fatalf("expected runtime session state to survive switch, got %d", runtimeModel.runtimeSession.RecordsPageLimit)
	}
}
