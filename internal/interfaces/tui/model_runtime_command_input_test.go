package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/dto"
)

func TestHandleKey_CommandInputAcceptsLiteralSpaceForSetLimitCommand(t *testing.T) {
	// Arrange
	runtimeSession := &RuntimeSessionState{}
	model := &Model{
		runtimeSession: runtimeSession,
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
	typeCommandInputText(model, "set")
	model.handleKey(tea.KeyMsg{Type: tea.KeySpace})
	typeCommandInputText(model, "limit=10")
	_, cmd := model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})

	// Assert
	if cmd != nil {
		t.Fatal("expected no immediate records reload outside records view")
	}
	if runtimeSession.RecordsPageLimit != 10 {
		t.Fatalf("expected record limit 10 after typing :set limit=10, got %d", runtimeSession.RecordsPageLimit)
	}
	if model.ui.statusMessage != "Record limit set to 10" {
		t.Fatalf("expected success status, got %q", model.ui.statusMessage)
	}
}

func TestHandleKey_StartingNewCommandPreservesExistingStatusWhileOpeningSpotlight(t *testing.T) {
	// Arrange
	model := &Model{
		ui: runtimeUIState{
			width:  80,
			height: 24,
		},
		read: runtimeReadState{
			focus:    FocusContent,
			viewMode: ViewRecords,
		},
	}
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
	typeCommandInputText(model, "setlimit")
	model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
	status := stripANSI(model.renderStatus(200))
	view := stripANSI(model.View())

	// Assert
	if !model.overlay.commandInput.active {
		t.Fatal("expected command input to be active after pressing :")
	}
	if model.ui.statusMessage != "Unknown command: :setlimit" {
		t.Fatalf("expected existing status message to remain while spotlight opens, got %q", model.ui.statusMessage)
	}
	if !strings.Contains(status, "Unknown command: :setlimit") {
		t.Fatalf("expected rendered status to preserve existing message behind spotlight, got %q", status)
	}
	if strings.Contains(status, "Command:") {
		t.Fatalf("expected fresh command entry not to render in status, got %q", status)
	}
	if !strings.Contains(view, ":|") {
		t.Fatalf("expected fresh command prompt in spotlight overlay, got %q", view)
	}
}

func TestHandleKey_CommandInputSecondColonPreservesTypedCommand(t *testing.T) {
	// Arrange
	model := &Model{}
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
	typeCommandInputText(model, "set")
	model.handleKey(tea.KeyMsg{Type: tea.KeySpace})
	typeCommandInputText(model, "limit=10")

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})

	// Assert
	if !model.overlay.commandInput.active {
		t.Fatal("expected command input to remain active")
	}
	if model.overlay.commandInput.value != "set limit=10:" {
		t.Fatalf("expected repeated : to append to command input, got %q", model.overlay.commandInput.value)
	}
	if model.overlay.commandInput.cursor != len("set limit=10:") {
		t.Fatalf("expected cursor at end of appended command, got %d", model.overlay.commandInput.cursor)
	}
}

func TestHandleKey_InvalidCommandClosesSpotlightAndShowsUnknownCommandStatus(t *testing.T) {
	// Arrange
	model := &Model{}
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
	typeCommandInputText(model, "setlimit")

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})

	// Assert
	if model.overlay.commandInput.active {
		t.Fatal("expected invalid command submission to close spotlight")
	}
	if model.ui.statusMessage != "Unknown command: :setlimit" {
		t.Fatalf("expected unknown-command status after invalid submission, got %q", model.ui.statusMessage)
	}
}

func TestHandleKey_CommandInputEscRestoresPreviousRuntimeView(t *testing.T) {
	// Arrange
	model := &Model{
		ui: runtimeUIState{
			width:  80,
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
	}
	baseView := stripANSI(model.View())
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyEsc})
	restoredView := stripANSI(model.View())

	// Assert
	if model.overlay.commandInput.active {
		t.Fatal("expected Esc to close spotlight")
	}
	if restoredView != baseView {
		t.Fatalf("expected Esc to restore previous runtime view, before=%q after=%q", baseView, restoredView)
	}
}

func TestVisibleCommandPrompt_PreservesCurrentHorizontalScrollBehaviorForLongUnicodeInput(t *testing.T) {
	// Arrange
	model := &Model{
		overlay: runtimeOverlayState{
			commandInput: commandInput{
				active: true,
				value:  "alphażółćbeta",
				cursor: len("alphażółćbeta"),
			},
		},
	}

	// Act
	prompt := model.visibleCommandPrompt(8)

	// Assert
	if prompt != ":łćbeta|" {
		t.Fatalf("expected spotlight to keep current byte-index-based horizontal scroll behavior, got %q", prompt)
	}
}

func typeCommandInputText(model *Model, value string) {
	for _, r := range value {
		model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
}
