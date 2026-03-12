package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/dto"
	"github.com/mgierok/dbc/internal/interfaces/tui/internal/primitives"
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
	if !strings.Contains(popup, "w save") {
		t.Fatalf("expected records save shortcut in help popup, got %q", popup)
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

func TestView_HelpPopupSuppressesBackgroundPanelsAndShowsHelpContent(t *testing.T) {
	// Arrange
	model := &Model{
		ui: runtimeUIState{
			width:  80,
			height: 24,
		},
		overlay: runtimeOverlayState{
			helpPopup: helpPopup{active: true, context: helpPopupContextTables},
		},
	}

	// Act
	view := stripANSI(model.View())

	// Assert
	if strings.Contains(view, primitives.IconSelection+" Tables") || strings.Contains(view, primitives.IconSelection+" Schema") || strings.Contains(view, primitives.IconSelection+" Records") {
		t.Fatalf("expected help modal view without background panels, got %q", view)
	}
	if !strings.Contains(view, "Context Help: Tables") {
		t.Fatalf("expected help popup title in view, got %q", view)
	}
	if !strings.Contains(view, "Tables: Enter records") {
		t.Fatalf("expected tables-specific help content in view, got %q", view)
	}
}

func TestView_DirtyConfigCommandOpensConfirmPopupAndSuppressesBackgroundPanels(t *testing.T) {
	// Arrange
	model := &Model{
		read: runtimeReadState{viewMode: ViewRecords},
		ui: runtimeUIState{
			width:  80,
			height: 24,
		},
		staging: testActiveDatabaseStaging(stagingState{pendingInserts: []pendingInsertRow{{}}}),
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
	for _, r := range "config" {
		model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})
	view := stripANSI(model.View())

	// Assert
	if !strings.Contains(view, "Config") {
		t.Fatalf("expected dirty :config modal title, got %q", view)
	}
	if strings.Contains(view, "Tables") || strings.Contains(view, "Schema") || strings.Contains(view, "Records") {
		t.Fatalf("expected centered modal without background panels, got %q", view)
	}
	if !strings.Contains(view, "Unsaved changes detected.") {
		t.Fatalf("expected dirty :config confirm message, got %q", view)
	}
	if !strings.Contains(view, "Save and open config") || !strings.Contains(view, "Discard and open config") {
		t.Fatalf("expected dirty :config options in popup, got %q", view)
	}
}

func TestRenderConfirmPopup_DirtyConfigShowsMessageAndOptionLabels(t *testing.T) {
	// Arrange
	model := &Model{
		overlay: runtimeOverlayState{
			confirmPopup: confirmPopup{
				active:  true,
				title:   "Config",
				message: "Unsaved changes detected. Save changes before you open config, discard them, or cancel.",
				options: []confirmOption{
					{label: "Save and open config", action: confirmLeaveSave},
					{label: "Discard and open config", action: confirmLeaveDiscard},
					{label: "Cancel", action: confirmLeaveCancel},
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
	if !strings.Contains(popup, "Unsaved changes detected.") {
		t.Fatalf("expected decision summary in popup, got %q", popup)
	}
	if !strings.Contains(popup, "Save and open config") {
		t.Fatalf("expected save option label in popup, got %q", popup)
	}
	if !strings.Contains(popup, "Discard and open config") {
		t.Fatalf("expected discard option label in popup, got %q", popup)
	}
	if !strings.Contains(popup, "Cancel") {
		t.Fatalf("expected cancel option label in popup, got %q", popup)
	}
	if !strings.Contains(popup, primitives.IconSelection+" Save and open config") {
		t.Fatalf("expected selected config option in popup, got %q", popup)
	}
}

func TestRenderConfirmPopup_DirtyQuitShowsMessageAndExplicitActions(t *testing.T) {
	// Arrange
	model := &Model{
		overlay: runtimeOverlayState{
			confirmPopup: confirmPopup{
				active:  true,
				title:   "Quit",
				message: "Unsaved changes detected. Save changes before you quit, discard them, or cancel.",
				options: []confirmOption{
					{label: "Save and quit", action: confirmLeaveSave},
					{label: "Discard and quit", action: confirmLeaveDiscard},
					{label: "Cancel", action: confirmLeaveCancel},
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
	if !strings.Contains(popup, "Save changes before you qui") {
		t.Fatalf("expected quit summary in popup, got %q", popup)
	}
	if !strings.Contains(popup, "Save and quit") {
		t.Fatalf("expected save action label in popup, got %q", popup)
	}
	if !strings.Contains(popup, "Discard and quit") {
		t.Fatalf("expected discard action label in popup, got %q", popup)
	}
	if !strings.Contains(popup, primitives.IconSelection+" Save and quit") {
		t.Fatalf("expected selected save action in popup, got %q", popup)
	}
}
