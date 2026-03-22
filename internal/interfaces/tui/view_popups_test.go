package tui

import (
	"context"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/dto"
	"github.com/mgierok/dbc/internal/interfaces/tui/internal/primitives"
	selectorpkg "github.com/mgierok/dbc/internal/interfaces/tui/internal/selector"
)

func TestRenderFilterPopup_ValueInputShowsCaretAtCursor(t *testing.T) {
	// Arrange
	model := &Model{
		overlay: runtimeOverlayState{
			filterPopup: filterPopup{
				active: true,
				step:   filterInputValue,
				input:  "abc",
				cursor: 1,
			},
		},
	}

	// Act
	popup := stripANSI(strings.Join(model.renderFilterPopup(60), "\n"))

	// Assert
	if !strings.Contains(popup, "Value: a|bc") {
		t.Fatalf("expected caret in filter value input, got %q", popup)
	}
}

func TestRenderEditPopup_TextInputShowsCaretAtCursor(t *testing.T) {
	// Arrange
	model := &Model{
		read: runtimeReadState{
			schema: dto.Schema{
				Columns: []dto.SchemaColumn{
					{
						Name:  "name",
						Type:  "TEXT",
						Input: dto.ColumnInput{Kind: dto.ColumnInputText},
					},
				},
			},
		},
		overlay: runtimeOverlayState{
			editPopup: editPopup{
				active:      true,
				columnIndex: 0,
				input:       "john",
				cursor:      2,
			},
		},
	}

	// Act
	popup := stripANSI(strings.Join(model.renderEditPopup(60), "\n"))

	// Assert
	if !strings.Contains(popup, "Value: jo|hn") {
		t.Fatalf("expected caret in edit text input, got %q", popup)
	}
}

func TestRenderEditPopup_UsesCombinedSummaryRow(t *testing.T) {
	// Arrange
	model := &Model{
		read: runtimeReadState{
			schema: dto.Schema{
				Columns: []dto.SchemaColumn{
					{
						Name:     "name",
						Type:     "TEXT",
						Nullable: true,
						Input:    dto.ColumnInput{Kind: dto.ColumnInputText},
					},
				},
			},
		},
		overlay: runtimeOverlayState{
			editPopup: editPopup{
				active:      true,
				columnIndex: 0,
			},
		},
	}

	// Act
	popup := stripANSI(strings.Join(model.renderEditPopup(60), "\n"))

	// Assert
	if !strings.Contains(popup, "name (TEXT)"+primitives.FrameSegmentSeparator+"NULLABLE") {
		t.Fatalf("expected combined summary row for column metadata, got %q", popup)
	}
}

func TestRenderHelpPopup_ShowsOnlyCurrentContextBindings(t *testing.T) {
	// Arrange
	model := &Model{
		ui: runtimeUIState{height: 40},
		overlay: runtimeOverlayState{
			helpPopup: helpPopup{
				active:  true,
				context: helpPopupContextRecords,
			},
		},
	}

	// Act
	popup := stripANSI(strings.Join(model.renderHelpPopup(60), "\n"))

	// Assert
	if !strings.Contains(popup, "Records: Esc tables") {
		t.Fatalf("expected records context shortcut row in help popup, got %q", popup)
	}
	if !strings.Contains(popup, ":w / :write save") {
		t.Fatalf("expected records save command hint in help popup, got %q", popup)
	}
	if strings.Contains(popup, "Supported Commands") || strings.Contains(popup, "Supported Keywords") {
		t.Fatalf("expected context-help popup without global help sections, got %q", popup)
	}
}

func TestRenderHelpPopup_ScrollCanReachFinalItemWhenOverflowing(t *testing.T) {
	// Arrange
	model := &Model{
		ui: runtimeUIState{
			height:        12,
			statusMessage: "",
		},
		overlay: runtimeOverlayState{
			helpPopup: helpPopup{active: true, context: helpPopupContextRecords},
		},
	}
	initial := stripANSI(strings.Join(model.renderHelpPopup(60), "\n"))

	// Act
	for range 30 {
		model.handleHelpPopupKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	}
	scrolled := stripANSI(strings.Join(model.renderHelpPopup(60), "\n"))

	// Assert
	if strings.Contains(initial, "Shift+S sort") {
		t.Fatalf("expected final help item to be hidden before scrolling, got %q", initial)
	}
	if !strings.Contains(scrolled, "Shift+S sort") {
		t.Fatalf("expected final help item to be reachable after scrolling, got %q", scrolled)
	}
}

func TestHandleHelpPopupKey_NonScrollKeyDoesNotChangeRenderedWindow(t *testing.T) {
	// Arrange
	model := &Model{
		ui: runtimeUIState{height: 12},
		overlay: runtimeOverlayState{
			helpPopup: helpPopup{active: true, context: helpPopupContextRecords},
		},
	}
	for range 5 {
		model.handleHelpPopupKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	}
	before := strings.Join(model.renderHelpPopup(60), "\n")

	// Act
	model.handleHelpPopupKey(tea.KeyMsg{Type: tea.KeyEnter})
	after := strings.Join(model.renderHelpPopup(60), "\n")

	// Assert
	if stripANSI(before) != stripANSI(after) {
		t.Fatalf("expected non-scroll key to keep help window stable, before=%q after=%q", stripANSI(before), stripANSI(after))
	}
}

type runtimeSelectorManagerStub struct{}

func (runtimeSelectorManagerStub) List(_ context.Context) ([]dto.ConfigDatabase, error) {
	return []dto.ConfigDatabase{{Name: "local", Path: "/tmp/local.sqlite"}}, nil
}

func (runtimeSelectorManagerStub) Create(_ context.Context, _ dto.ConfigDatabase) error {
	return nil
}

func (runtimeSelectorManagerStub) Update(_ context.Context, _ int, _ dto.ConfigDatabase) error {
	return nil
}

func (runtimeSelectorManagerStub) Delete(_ context.Context, _ int) error {
	return nil
}

func (runtimeSelectorManagerStub) ActivePath(_ context.Context) (string, error) {
	return "/tmp/config.json", nil
}

func TestView_HelpPopupShowsBackdropRuntimePanelsAndHelpContent(t *testing.T) {
	// Arrange
	model := &Model{
		styles: primitives.NewRenderStyles(true),
		ui: runtimeUIState{
			width:  80,
			height: 24,
		},
		read: runtimeReadState{
			focus:    FocusTables,
			viewMode: ViewSchema,
			tables:   []dto.Table{{Name: "users"}},
			schema: dto.Schema{
				Columns: []dto.SchemaColumn{{Name: "id", Type: "INTEGER"}},
			},
		},
		overlay: runtimeOverlayState{
			helpPopup: helpPopup{active: true, context: helpPopupContextTables},
		},
	}

	// Act
	view := model.View()
	plainView := stripANSI(view)

	// Assert
	if !strings.Contains(plainView, primitives.FrameTopLeft+"Tables") || !strings.Contains(plainView, primitives.FrameTopLeft+"Schema") {
		t.Fatalf("expected runtime panels to remain visible behind help popup, got %q", plainView)
	}
	if !strings.Contains(view, "\x1b[2mTables\x1b[0m") || !strings.Contains(view, "\x1b[2mSchema\x1b[0m") {
		t.Fatalf("expected help backdrop to subdue panel titles, got %q", view)
	}
	if !strings.Contains(plainView, "Context Help: Tables") {
		t.Fatalf("expected help popup title in view, got %q", plainView)
	}
	if !strings.Contains(plainView, "Tables: Enter records") {
		t.Fatalf("expected tables-specific help content in view, got %q", plainView)
	}
}

func TestView_DirtyConfigCommandOpensRuntimeSelectorWithBackdropRuntimeLayout(t *testing.T) {
	// Arrange
	model := &Model{
		styles: primitives.NewRenderStyles(true),
		read:   runtimeReadState{viewMode: ViewRecords},
		ui: runtimeUIState{
			width:  80,
			height: 24,
		},
		staging: stagingState{pendingInserts: []pendingInsertRow{{}}},
		runtimeDatabaseSelectorDeps: runtimeDatabaseSelectorDepsForTest(DatabaseOption{
			Name:       "primary",
			ConnString: "/tmp/primary.sqlite",
			Source:     DatabaseOptionSourceConfig,
		}),
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
	for _, r := range "config" {
		model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})
	view := model.View()
	plainView := stripANSI(view)

	// Assert
	if !strings.Contains(plainView, "Select database") {
		t.Fatalf("expected dirty :config selector popup, got %q", plainView)
	}
	if !strings.Contains(plainView, primitives.FrameTopLeft+"Tables") || !strings.Contains(plainView, primitives.FrameTopLeft+"Records") {
		t.Fatalf("expected dirty :config selector to keep runtime panels visible, got %q", plainView)
	}
	if !strings.Contains(view, "\x1b[2m✱\x1b[0m") {
		t.Fatalf("expected dirty backdrop status to use subdued styling, got %q", view)
	}
	if !strings.Contains(view, "\x1b[2mRecords [staged rows: 1]\x1b[0m") {
		t.Fatalf("expected dirty backdrop to subdue staged-record title, got %q", view)
	}
	if !strings.Contains(plainView, "primary") {
		t.Fatalf("expected selector content in popup, got %q", plainView)
	}
}

func TestView_RuntimeDatabaseSelectorUsesSharedBackdropPresenter(t *testing.T) {
	// Arrange
	controller, err := selectorpkg.NewRuntimeController(context.Background(), runtimeSelectorManagerStub{}, selectorpkg.SelectorLaunchState{})
	if err != nil {
		t.Fatalf("expected runtime selector controller, got error %v", err)
	}

	model := &Model{
		styles: primitives.NewRenderStyles(true),
		ui: runtimeUIState{
			width:  100,
			height: 24,
		},
		read: runtimeReadState{
			focus:         FocusContent,
			viewMode:      ViewRecords,
			selectedTable: 0,
			tables:        []dto.Table{{Name: "users"}},
			schema: dto.Schema{
				Columns: []dto.SchemaColumn{{Name: "id", Type: "INTEGER"}},
			},
			records:          []dto.RecordRow{{Values: []string{"1"}}},
			recordTotalCount: 1,
			recordTotalPages: 1,
		},
		overlay: runtimeOverlayState{
			databaseSelector: runtimeDatabaseSelectorPopup{
				active:     true,
				controller: controller,
			},
		},
	}

	// Act
	view := model.View()
	plainView := stripANSI(view)

	// Assert
	if !strings.Contains(plainView, primitives.FrameTopLeft+"Tables") || !strings.Contains(plainView, primitives.FrameTopLeft+"Records") {
		t.Fatalf("expected runtime database selector to keep runtime layout visible, got %q", plainView)
	}
	if !strings.Contains(view, "\x1b[2mTables\x1b[0m") || !strings.Contains(view, "\x1b[2mRecords\x1b[0m") {
		t.Fatalf("expected runtime database selector backdrop to subdue panel titles, got %q", view)
	}
	if !strings.Contains(plainView, "Select database") {
		t.Fatalf("expected runtime database selector popup content, got %q", plainView)
	}
}

func TestRenderConfirmPopup_DirtyDatabaseReloadShowsMessageAndOptionLabels(t *testing.T) {
	// Arrange
	model := &Model{
		overlay: runtimeOverlayState{
			confirmPopup: confirmPopup{
				active:  true,
				title:   "Reload Database",
				message: "Reloading the current database will cause loss of unsaved data (3 rows) unless you save first. Choose save, discard, or cancel.",
				options: []confirmOption{
					{label: "Save and reload database", action: confirmDatabaseTransitionSave},
					{label: "Discard changes and reload database", action: confirmDatabaseTransitionDiscard},
					{label: "Cancel", action: confirmDatabaseTransitionCancel},
				},
				selected: 0,
				modal:    true,
			},
		},
	}

	// Act
	lines := model.renderConfirmPopup(60)
	popup := stripANSI(strings.Join(lines, "\n"))

	// Assert
	if !strings.Contains(popup, "Reloading the current database") {
		t.Fatalf("expected decision summary in popup, got %q", popup)
	}
	if !strings.Contains(popup, "Save and reload database") {
		t.Fatalf("expected save option label in popup, got %q", popup)
	}
	if !strings.Contains(popup, "Discard changes and reload database") {
		t.Fatalf("expected discard option label in popup, got %q", popup)
	}
	if !strings.Contains(popup, "Cancel") {
		t.Fatalf("expected cancel option label in popup, got %q", popup)
	}
	if !strings.Contains(popup, primitives.IconSelection+" Save and reload database") {
		t.Fatalf("expected selected config option in popup, got %q", popup)
	}
}

func TestRenderConfirmPopup_DirtyTableSwitchShowsMessageAndExplicitActions(t *testing.T) {
	// Arrange
	model := &Model{
		overlay: runtimeOverlayState{
			confirmPopup: confirmPopup{
				active:  true,
				title:   "Switch Table",
				message: "Switching tables will cause loss of unsaved data (3 rows). Are you sure you want to discard unsaved data?",
				options: []confirmOption{
					{label: "Discard changes and switch table", action: confirmDiscardTable},
					{label: "Continue editing", action: confirmCancelTableSwitch},
				},
				selected: 0,
				modal:    true,
			},
		},
	}

	// Act
	lines := model.renderConfirmPopup(120)
	popup := stripANSI(strings.Join(lines, "\n"))

	// Assert
	if !strings.Contains(popup, "Switching tables will cause loss of unsaved data") {
		t.Fatalf("expected informational switch-table summary in popup, got %q", popup)
	}
	if !strings.Contains(popup, "Discard changes and switch table") {
		t.Fatalf("expected discard action label in popup, got %q", popup)
	}
	if !strings.Contains(popup, "Continue editing") {
		t.Fatalf("expected continue-editing action label in popup, got %q", popup)
	}
	if !strings.Contains(popup, primitives.IconSelection+" Discard changes and switch table") {
		t.Fatalf("expected selected discard action in popup, got %q", popup)
	}
}

func TestRenderConfirmPopup_DirtyQuitShowsMessageAndExplicitActions(t *testing.T) {
	// Arrange
	model := &Model{
		overlay: runtimeOverlayState{
			confirmPopup: confirmPopup{
				active:  true,
				title:   "Quit",
				message: "Quitting will cause loss of unsaved data (3 rows). Are you sure you want to discard unsaved data and quit?",
				options: []confirmOption{
					{label: "Discard changes and quit"},
					{label: "Continue editing"},
				},
				selected: 0,
				modal:    true,
			},
		},
	}

	// Act
	lines := model.renderConfirmPopup(120)
	popup := stripANSI(strings.Join(lines, "\n"))

	// Assert
	if !strings.Contains(popup, "Quit") {
		t.Fatalf("expected quit title in popup, got %q", popup)
	}
	if !strings.Contains(popup, "Quitting will cause loss of unsaved data (3 rows).") {
		t.Fatalf("expected quit message in popup, got %q", popup)
	}
	if !strings.Contains(popup, "Discard changes and quit") {
		t.Fatalf("expected quit discard action label in popup, got %q", popup)
	}
	if !strings.Contains(popup, "Continue editing") {
		t.Fatalf("expected quit continue-editing action label in popup, got %q", popup)
	}
	if !strings.Contains(popup, primitives.IconSelection+" Discard changes and quit") {
		t.Fatalf("expected selected quit discard action in popup, got %q", popup)
	}
}

func TestRenderConfirmPopup_SanitizesMessageAndOptionLabels(t *testing.T) {
	// Arrange
	model := &Model{
		overlay: runtimeOverlayState{
			confirmPopup: confirmPopup{
				active:  true,
				title:   "Delete\x1b[31m",
				message: "Delete selected value?\r\n\x1b]2;ignored\ayep",
				options: []confirmOption{
					{label: "Delete\tentry\x1b[0m"},
					{label: "Cancel\r\nlater"},
				},
				selected: 0,
			},
		},
	}

	// Act
	popup := stripANSI(strings.Join(model.renderConfirmPopup(80), "\n"))

	// Assert
	if strings.Contains(popup, "\x1b") || strings.Contains(popup, "\r") || strings.Contains(popup, "\t") {
		t.Fatalf("expected confirm popup without injected escape/control characters, got %q", popup)
	}
	if !strings.Contains(popup, "Delete") {
		t.Fatalf("expected sanitized popup title, got %q", popup)
	}
	if !strings.Contains(popup, "Delete selected value? yep") {
		t.Fatalf("expected sanitized popup message, got %q", popup)
	}
	if !strings.Contains(popup, "Delete entry") {
		t.Fatalf("expected sanitized selected option label, got %q", popup)
	}
	if !strings.Contains(popup, "Cancel later") {
		t.Fatalf("expected sanitized secondary option label, got %q", popup)
	}
}
