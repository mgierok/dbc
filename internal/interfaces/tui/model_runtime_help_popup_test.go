package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestHandleKey_ContextHelpQuestionMarkOpensRecordsHelpPopup(t *testing.T) {
	// Arrange
	model := newRuntimeCommandModel()

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})

	// Assert
	if !model.overlay.helpPopup.active {
		t.Fatal("expected ? to open help popup")
	}
	if model.overlay.helpPopup.context != helpPopupContextRecords {
		t.Fatalf("expected records context, got %v", model.overlay.helpPopup.context)
	}
}

func TestHandleKey_CommandHelpAliasOpensRecordsHelpPopup(t *testing.T) {
	// Arrange
	model := newRuntimeCommandModel()

	// Act
	_, cmd := submitTypedRuntimeCommand(model, "help")

	// Assert
	assertRuntimeSessionActive(t, cmd, ":help")
	if !model.overlay.helpPopup.active {
		t.Fatal("expected :help to open help popup")
	}
	if model.overlay.helpPopup.context != helpPopupContextRecords {
		t.Fatalf("expected records context for :help, got %v", model.overlay.helpPopup.context)
	}
	if strings.Contains(strings.ToLower(model.ui.statusMessage), "unknown command") {
		t.Fatalf("expected no unknown-command status for :help, got %q", model.ui.statusMessage)
	}
}

func TestHandleKey_ContextHelpPopupShowsCurrentContextBindings(t *testing.T) {
	// Arrange
	model := newRuntimeCommandModel()
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})

	// Act
	popup := strings.Join(model.renderHelpPopup(60), "\n")

	// Assert
	if !strings.Contains(popup, "Records: Esc tables") {
		t.Fatalf("expected records shortcuts in context help popup, got %q", popup)
	}
	if strings.Contains(popup, "Supported Commands") || strings.Contains(popup, "Supported Keywords") {
		t.Fatalf("expected context-only help content, got %q", popup)
	}
}

func TestHandleKey_ContextHelpPopupShowsSaveInSchemaTablesAndRecordDetail(t *testing.T) {
	for _, tc := range []struct {
		name           string
		model          *Model
		expectedHeader string
		expectedRow    string
	}{
		{
			name: "schema",
			model: &Model{
				ui: runtimeUIState{height: 40},
				read: runtimeReadState{
					viewMode: ViewSchema,
					focus:    FocusContent,
				},
			},
			expectedHeader: "Context Help: Schema",
			expectedRow:    "Schema: Esc tables",
		},
		{
			name: "tables",
			model: &Model{
				ui: runtimeUIState{height: 40},
				read: runtimeReadState{
					viewMode: ViewSchema,
					focus:    FocusTables,
				},
			},
			expectedHeader: "Context Help: Tables",
			expectedRow:    "Tables: Enter records",
		},
		{
			name: "record detail",
			model: &Model{
				ui: runtimeUIState{height: 40},
				read: runtimeReadState{
					viewMode: ViewRecords,
					focus:    FocusContent,
				},
				overlay: runtimeOverlayState{
					recordDetail: recordDetailState{active: true},
				},
			},
			expectedHeader: "Context Help: Record Detail",
			expectedRow:    "Detail: Esc back",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			tc.model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
			popup := strings.Join(tc.model.renderHelpPopup(60), "\n")

			// Assert
			if !tc.model.overlay.helpPopup.active {
				t.Fatal("expected help popup to open")
			}
			if !strings.Contains(popup, tc.expectedHeader) {
				t.Fatalf("expected help popup header %q, got %q", tc.expectedHeader, popup)
			}
			if !strings.Contains(popup, tc.expectedRow) {
				t.Fatalf("expected context row %q, got %q", tc.expectedRow, popup)
			}
			if !strings.Contains(popup, ":w / :write save") {
				t.Fatalf("expected save shortcut in %s help popup, got %q", tc.name, popup)
			}
		})
	}
}

func TestHandleKey_HelpPopupScrollCanReachFinalContextShortcut(t *testing.T) {
	// Arrange
	model := newRuntimeCommandModel()
	model.ui.height = 12
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	initial := strings.Join(model.renderHelpPopup(60), "\n")

	// Act
	for range 30 {
		model.handleHelpPopupKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	}
	scrolled := strings.Join(model.renderHelpPopup(60), "\n")

	// Assert
	if strings.Contains(initial, "Shift+S sort") {
		t.Fatalf("expected final shortcut to start below initial viewport, got %q", initial)
	}
	if !strings.Contains(scrolled, "Shift+S sort") {
		t.Fatalf("expected help popup scroll to reach final shortcut, got %q", scrolled)
	}
}

func TestHandleKey_CommandHelpReenterKeepsPopupOpen(t *testing.T) {
	// Arrange
	model := newRuntimeCommandModel()
	submitTypedRuntimeCommand(model, "help")
	if !model.overlay.helpPopup.active {
		t.Fatal("expected help popup to open on first :help")
	}

	// Act
	_, cmd := submitTypedRuntimeCommand(model, "help")

	// Assert
	assertRuntimeSessionActive(t, cmd, "repeated :help")
	if !model.overlay.helpPopup.active {
		t.Fatal("expected repeated :help to keep help popup open")
	}
}

func TestHandleKey_CommandHelpReenterPreservesExistingStatusWhileOpeningSpotlight(t *testing.T) {
	// Arrange
	model := newRuntimeCommandModel()
	model.ui.statusMessage = "stale status"
	submitTypedRuntimeCommand(model, "help")
	if !model.overlay.helpPopup.active {
		t.Fatal("expected help popup to open on first :help")
	}

	// Act
	_, cmd := submitTypedRuntimeCommand(model, "help")

	// Assert
	assertRuntimeSessionActive(t, cmd, "repeated :help after stale status")
	if model.ui.statusMessage != "stale status" {
		t.Fatalf("expected repeated :help to preserve status message, got %q", model.ui.statusMessage)
	}
}

func TestHandleKey_HelpPopupEscClosesPopup(t *testing.T) {
	// Arrange
	model := newRuntimeCommandModel()
	submitTypedRuntimeCommand(model, "help")

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyEsc})

	// Assert
	if model.overlay.helpPopup.active {
		t.Fatal("expected Esc to close help popup")
	}
}

func TestHandleKey_HelpPopupColonDoesNotOpenCommandInput(t *testing.T) {
	// Arrange
	model := &Model{
		read: runtimeReadState{
			viewMode: ViewRecords,
			focus:    FocusContent,
		},
		overlay: runtimeOverlayState{
			helpPopup: helpPopup{
				active:  true,
				context: helpPopupContextRecords,
			},
		},
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})

	// Assert
	if model.overlay.commandInput.active {
		t.Fatal("expected help popup to block command input")
	}
	if !model.overlay.helpPopup.active {
		t.Fatal("expected help popup to stay open after :")
	}
}

func TestHandleKey_HelpPopupUnrelatedKeysDoNotClosePopup(t *testing.T) {
	// Arrange
	model := &Model{read: runtimeReadState{viewMode: ViewRecords}}
	submitTypedRuntimeCommand(model, "help")
	if !model.overlay.helpPopup.active {
		t.Fatal("expected help popup to open before unrelated-key check")
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})

	// Assert
	if !model.overlay.helpPopup.active {
		t.Fatal("expected unrelated keys to keep help popup open")
	}
}

func TestHandleKey_ContextHelpFromFilterPopupUsesFilterContext(t *testing.T) {
	// Arrange
	model := &Model{
		read: runtimeReadState{
			viewMode: ViewRecords,
			focus:    FocusContent,
		},
		overlay: runtimeOverlayState{
			filterPopup: filterPopup{
				active: true,
				step:   filterSelectColumn,
			},
		},
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})

	// Assert
	if !model.overlay.helpPopup.active {
		t.Fatal("expected ? to open help popup from filter context")
	}
	if model.overlay.helpPopup.context != helpPopupContextFilterPopup {
		t.Fatalf("expected filter-popup context help, got %v", model.overlay.helpPopup.context)
	}
	if !model.overlay.filterPopup.active {
		t.Fatal("expected filter popup state to stay preserved under help overlay")
	}
}

func TestHandleKey_PopupPriority_HelpPopupConsumesEscBeforeOtherPopups(t *testing.T) {
	// Arrange
	model := &Model{
		overlay: runtimeOverlayState{
			helpPopup:   helpPopup{active: true},
			filterPopup: filterPopup{active: true},
		},
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyEsc})

	// Assert
	if model.overlay.helpPopup.active {
		t.Fatal("expected help popup to close first")
	}
	if !model.overlay.filterPopup.active {
		t.Fatal("expected filter popup to remain active when help popup handled Esc")
	}
}
