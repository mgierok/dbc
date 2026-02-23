package tui

import (
	"context"
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/dto"
	"github.com/mgierok/dbc/internal/application/usecase"
	domainmodel "github.com/mgierok/dbc/internal/domain/model"
)

func TestHandleKey_EnterFromTablesSwitchesToRecordsAndContentFocus(t *testing.T) {
	// Arrange
	model := &Model{
		viewMode: ViewSchema,
		focus:    FocusTables,
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})

	// Assert
	if model.viewMode != ViewRecords {
		t.Fatalf("expected view mode to switch to records, got %v", model.viewMode)
	}
	if model.focus != FocusContent {
		t.Fatalf("expected focus to switch to content, got %v", model.focus)
	}
}

func TestHandleKey_EnterInRecordsEnablesFieldFocus(t *testing.T) {
	// Arrange
	model := &Model{
		viewMode: ViewRecords,
		focus:    FocusContent,
		records:  []dto.RecordRow{{Values: []string{"1"}}},
		schema: dto.Schema{
			Columns: []dto.SchemaColumn{
				{Name: "id", Type: "INTEGER", PrimaryKey: true},
			},
		},
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})

	// Assert
	if !model.recordFieldFocus {
		t.Fatalf("expected record field focus to be enabled")
	}
}

func TestHandleKey_EnterInFieldFocusOpensEditPopup(t *testing.T) {
	// Arrange
	model := &Model{
		viewMode:         ViewRecords,
		focus:            FocusContent,
		recordFieldFocus: true,
		records:          []dto.RecordRow{{Values: []string{"1"}}},
		schema: dto.Schema{
			Columns: []dto.SchemaColumn{
				{Name: "id", Type: "INTEGER", PrimaryKey: true},
			},
		},
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})

	// Assert
	if !model.editPopup.active {
		t.Fatalf("expected edit popup to be active")
	}
}

func TestHandleKey_EscClearsFieldFocus(t *testing.T) {
	// Arrange
	model := &Model{
		viewMode:         ViewRecords,
		focus:            FocusContent,
		recordFieldFocus: true,
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyEsc})

	// Assert
	if model.recordFieldFocus {
		t.Fatalf("expected record field focus to be disabled")
	}
	if model.focus != FocusContent {
		t.Fatalf("expected focus to remain on content in nested context, got %v", model.focus)
	}
}

func TestHandleKey_EscInRightPanelNeutralReturnsToTables(t *testing.T) {
	// Arrange
	model := &Model{
		viewMode: ViewRecords,
		focus:    FocusContent,
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyEsc})

	// Assert
	if model.focus != FocusTables {
		t.Fatalf("expected focus to return to tables, got %v", model.focus)
	}
	if model.viewMode != ViewSchema {
		t.Fatalf("expected schema view to be active, got %v", model.viewMode)
	}
}

func TestHandleKey_EscFromFieldFocusThenNeutralSwitchesToTablesAndSchema(t *testing.T) {
	// Arrange
	model := &Model{
		viewMode:         ViewRecords,
		focus:            FocusContent,
		recordFieldFocus: true,
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyEsc})
	model.handleKey(tea.KeyMsg{Type: tea.KeyEsc})

	// Assert
	if model.recordFieldFocus {
		t.Fatalf("expected record field focus to be disabled")
	}
	if model.focus != FocusTables {
		t.Fatalf("expected focus to return to tables, got %v", model.focus)
	}
	if model.viewMode != ViewSchema {
		t.Fatalf("expected schema view to be active, got %v", model.viewMode)
	}
}

func TestHandleKey_CtrlWPanelShortcutsAreUnsupported(t *testing.T) {
	tests := []struct {
		name       string
		startFocus PanelFocus
		nextKey    tea.KeyMsg
	}{
		{
			name:       "ctrl+w h does not switch to tables",
			startFocus: FocusContent,
			nextKey:    tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}},
		},
		{
			name:       "ctrl+w l does not switch to content",
			startFocus: FocusTables,
			nextKey:    tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}},
		},
		{
			name:       "ctrl+w w does not toggle panel",
			startFocus: FocusTables,
			nextKey:    tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'w'}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			model := &Model{
				viewMode: ViewRecords,
				focus:    tc.startFocus,
			}

			// Act
			model.handleKey(tea.KeyMsg{Type: tea.KeyCtrlW})
			model.handleKey(tc.nextKey)

			// Assert
			if model.focus != tc.startFocus {
				t.Fatalf("expected focus to stay %v, got %v", tc.startFocus, model.focus)
			}
		})
	}
}

func TestHandleKey_CommandConfigQuitsToOpenSelector(t *testing.T) {
	for _, tc := range []struct {
		name    string
		command string
	}{
		{name: "full command", command: "config"},
		{name: "alias", command: "c"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			model := &Model{viewMode: ViewRecords}

			// Act
			model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
			for _, r := range tc.command {
				model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
			}
			_, cmd := model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})

			// Assert
			if cmd == nil {
				t.Fatalf("expected quit command for :%s", tc.command)
			}
			if _, ok := cmd().(tea.QuitMsg); !ok {
				t.Fatalf("expected tea.QuitMsg for :%s, got %T", tc.command, cmd())
			}
		})
	}
}

func TestHandleKey_CommandQuitQuitsRuntime(t *testing.T) {
	for _, tc := range []struct {
		name    string
		command string
	}{
		{name: "short command", command: "q"},
		{name: "full command", command: "quit"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			model := &Model{viewMode: ViewRecords}

			// Act
			model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
			for _, r := range tc.command {
				model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
			}
			_, cmd := model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})

			// Assert
			if cmd == nil {
				t.Fatalf("expected quit command for :%s", tc.command)
			}
			if _, ok := cmd().(tea.QuitMsg); !ok {
				t.Fatalf("expected tea.QuitMsg for :%s, got %T", tc.command, cmd())
			}
		})
	}
}

func TestHandleKey_CommandHelpOpensPopupWithoutUnknownStatus(t *testing.T) {
	newRuntimeModel := func() *Model {
		return &Model{
			viewMode: ViewRecords,
			height:   40,
		}
	}
	submitCommand := func(model *Model, value string) tea.Cmd {
		model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
		for _, r := range value {
			model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		}
		_, cmd := model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})
		return cmd
	}
	assertSessionActive := func(t *testing.T, cmd tea.Cmd, context string) {
		t.Helper()
		if cmd != nil {
			if _, ok := cmd().(tea.QuitMsg); ok {
				t.Fatalf("expected %s to keep session active", context)
			}
		}
	}
	parseHelpContent := func(rendered string) []string {
		lines := strings.Split(rendered, "\n")
		content := make([]string, 0, len(lines))
		for _, line := range lines {
			if !strings.HasPrefix(line, "|") || !strings.HasSuffix(line, "|") {
				continue
			}
			value := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(line, "|"), "|"))
			if value == "Help" || strings.HasPrefix(value, "Scroll: ") {
				continue
			}
			content = append(content, value)
		}
		return content
	}
	sectionRows := func(content []string, section string) []string {
		start := -1
		for i, line := range content {
			if line == section {
				start = i + 1
				break
			}
		}
		if start == -1 {
			return nil
		}

		rows := make([]string, 0)
		for i := start; i < len(content); i++ {
			if content[i] == "" {
				break
			}
			rows = append(rows, content[i])
		}
		return rows
	}

	t.Run("FR-001 happy path opens help popup", func(t *testing.T) {
		// Arrange
		model := newRuntimeModel()

		// Act
		cmd := submitCommand(model, "help")

		// Assert
		assertSessionActive(t, cmd, ":help")
		if !model.helpPopup.active {
			t.Fatal("expected :help to open help popup")
		}
		if strings.Contains(strings.ToLower(model.statusMessage), "unknown command") {
			t.Fatalf("expected no unknown-command status for :help, got %q", model.statusMessage)
		}
	})

	t.Run("FR-001 alias path opens help popup", func(t *testing.T) {
		// Arrange
		model := newRuntimeModel()

		// Act
		cmd := submitCommand(model, "h")

		// Assert
		assertSessionActive(t, cmd, ":h")
		if !model.helpPopup.active {
			t.Fatal("expected :h to open help popup")
		}
		if strings.Contains(strings.ToLower(model.statusMessage), "unknown command") {
			t.Fatalf("expected no unknown-command status for :h, got %q", model.statusMessage)
		}
	})

	t.Run("FR-001 negative path requires explicit command prefix", func(t *testing.T) {
		// Arrange
		model := newRuntimeModel()

		// Act
		for _, r := range "help" {
			model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		}
		_, cmd := model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})

		// Assert
		assertSessionActive(t, cmd, "help without ':'")
		if model.helpPopup.active {
			t.Fatal("expected help popup to stay closed without ':' prefix")
		}
	})

	t.Run("FR-002 happy path renders Supported Commands section", func(t *testing.T) {
		// Arrange
		model := newRuntimeModel()
		submitCommand(model, "help")

		// Act
		popup := strings.Join(model.renderHelpPopup(60), "\n")
		content := parseHelpContent(popup)
		commands := sectionRows(content, "Supported Commands")

		// Assert
		if len(commands) == 0 {
			t.Fatalf("expected Supported Commands section rows, got %q", popup)
		}
		if !strings.Contains(popup, "Supported Commands") {
			t.Fatalf("expected help popup to include Supported Commands section, got %q", popup)
		}
	})

	t.Run("FR-002 negative path rejects missing or mislabelled command section shape", func(t *testing.T) {
		// Arrange
		model := newRuntimeModel()
		submitCommand(model, "help")

		// Act
		popup := strings.Join(model.renderHelpPopup(60), "\n")
		content := parseHelpContent(popup)
		commands := sectionRows(content, "Supported Commands")

		// Assert
		if len(commands) < 2 {
			t.Fatalf("expected multiple command entries under Supported Commands, got %q", popup)
		}
		for _, row := range commands {
			if !strings.Contains(row, " - ") {
				t.Fatalf("expected command row to include one-line description, row=%q popup=%q", row, popup)
			}
		}
		if strings.Contains(popup, "Supported Command") && !strings.Contains(popup, "Supported Commands") {
			t.Fatalf("expected canonical section label 'Supported Commands', got %q", popup)
		}
	})

	t.Run("FR-003 happy path renders Supported Keywords section", func(t *testing.T) {
		// Arrange
		model := newRuntimeModel()
		submitCommand(model, "help")

		// Act
		popup := strings.Join(model.renderHelpPopup(60), "\n")
		content := parseHelpContent(popup)
		keywords := sectionRows(content, "Supported Keywords")

		// Assert
		if len(keywords) == 0 {
			t.Fatalf("expected Supported Keywords section rows, got %q", popup)
		}
		if !strings.Contains(popup, "Supported Keywords") {
			t.Fatalf("expected help popup to include Supported Keywords section, got %q", popup)
		}
	})

	t.Run("FR-003 negative path detects missing keyword descriptions", func(t *testing.T) {
		// Arrange
		model := newRuntimeModel()
		submitCommand(model, "help")

		// Act
		popup := strings.Join(model.renderHelpPopup(60), "\n")
		content := parseHelpContent(popup)
		keywords := sectionRows(content, "Supported Keywords")

		// Assert
		if len(keywords) == 0 {
			t.Fatalf("expected keyword rows to be present, got %q", popup)
		}
		for _, row := range keywords {
			if !strings.Contains(row, " - ") {
				t.Fatalf("expected keyword row to include one-line description, row=%q popup=%q", row, popup)
			}
		}
	})

	t.Run("FR-004 happy path reaches final help item by scrolling", func(t *testing.T) {
		// Arrange
		model := newRuntimeModel()
		model.height = 12
		submitCommand(model, "help")

		// Act
		initial := strings.Join(model.renderHelpPopup(60), "\n")
		for range 30 {
			model.handleHelpPopupKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		}
		scrolled := strings.Join(model.renderHelpPopup(60), "\n")

		// Assert
		if strings.Contains(initial, "Ctrl+a - Toggle auto field visibility for inserts.") {
			t.Fatalf("expected final help item to be hidden before scrolling, got %q", initial)
		}
		if !strings.Contains(scrolled, "Ctrl+a - Toggle auto field visibility for inserts.") {
			t.Fatalf("expected final help item to be reachable after scrolling, got %q", scrolled)
		}
	})

	t.Run("FR-004 negative path keeps help scroll stable on non-scroll key", func(t *testing.T) {
		// Arrange
		model := newRuntimeModel()
		model.height = 12
		submitCommand(model, "help")
		for range 5 {
			model.handleHelpPopupKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		}
		beforeOffset := model.helpPopup.scrollOffset
		before := strings.Join(model.renderHelpPopup(60), "\n")

		// Act
		model.handleHelpPopupKey(tea.KeyMsg{Type: tea.KeyEnter})
		afterOffset := model.helpPopup.scrollOffset
		after := strings.Join(model.renderHelpPopup(60), "\n")

		// Assert
		if beforeOffset != afterOffset {
			t.Fatalf("expected non-scroll key to keep offset stable, before=%d after=%d", beforeOffset, afterOffset)
		}
		if before != after {
			t.Fatalf("expected non-scroll key to keep help window stable, before=%q after=%q", before, after)
		}
	})

	t.Run("FR-005 happy path keeps popup open on repeated :help", func(t *testing.T) {
		// Arrange
		model := newRuntimeModel()
		submitCommand(model, "help")

		// Act
		cmd := submitCommand(model, "help")

		// Assert
		assertSessionActive(t, cmd, "repeated :help")
		if !model.helpPopup.active {
			t.Fatal("expected help popup to remain open when :help is re-entered")
		}
	})

	t.Run("FR-005 negative path does not dismiss popup on repeated :help", func(t *testing.T) {
		// Arrange
		model := newRuntimeModel()
		submitCommand(model, "help")

		// Act
		submitCommand(model, "help")

		// Assert
		if !model.helpPopup.active {
			t.Fatal("expected repeated :help to avoid dismissing help popup")
		}
	})

	t.Run("FR-006 happy path closes help popup on Esc", func(t *testing.T) {
		// Arrange
		model := newRuntimeModel()
		submitCommand(model, "help")

		// Act
		model.handleKey(tea.KeyMsg{Type: tea.KeyEsc})

		// Assert
		if model.helpPopup.active {
			t.Fatal("expected Esc to close help popup")
		}
	})

	t.Run("FR-006 negative path keeps help popup open on unrelated key", func(t *testing.T) {
		// Arrange
		model := newRuntimeModel()
		submitCommand(model, "help")

		// Act
		model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})

		// Assert
		if !model.helpPopup.active {
			t.Fatal("expected unrelated key to keep help popup open")
		}
	})

	t.Run("FR-007 happy path keeps unsupported command fallback", func(t *testing.T) {
		// Arrange
		model := newRuntimeModel()

		// Act
		cmd := submitCommand(model, "unknown")

		// Assert
		assertSessionActive(t, cmd, "unsupported command")
		if !strings.Contains(strings.ToLower(model.statusMessage), "unknown command") {
			t.Fatalf("expected unknown command status message, got %q", model.statusMessage)
		}
	})

	t.Run("FR-007 negative path blocks misspelled :help regression", func(t *testing.T) {
		// Arrange
		model := newRuntimeModel()

		// Act
		cmd := submitCommand(model, "helpp")

		// Assert
		assertSessionActive(t, cmd, "misspelled :help")
		if model.helpPopup.active {
			t.Fatal("expected misspelled :help to keep help popup closed")
		}
		if !strings.Contains(strings.ToLower(model.statusMessage), "unknown command") {
			t.Fatalf("expected unknown-command status for misspelled :help, got %q", model.statusMessage)
		}
	})
}

func TestHandleKey_InvalidCommandShowsErrorAndKeepsSessionActive(t *testing.T) {
	// Arrange
	model := &Model{viewMode: ViewRecords}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
	for _, r := range "unknown" {
		model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	_, cmd := model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})

	// Assert
	if cmd != nil {
		if _, ok := cmd().(tea.QuitMsg); ok {
			t.Fatal("expected invalid command to keep session active")
		}
	}
	if !strings.Contains(strings.ToLower(model.statusMessage), "unknown command") {
		t.Fatalf("expected unknown command status message, got %q", model.statusMessage)
	}
}

func TestHandleKey_HelpCommandRequiresExplicitPrefix(t *testing.T) {
	// Arrange
	model := &Model{viewMode: ViewRecords}

	// Act
	for _, r := range "help" {
		model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	_, cmd := model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})

	// Assert
	if cmd != nil {
		if _, ok := cmd().(tea.QuitMsg); ok {
			t.Fatal("expected no quit command without explicit ':' prefix")
		}
	}
	if model.helpPopup.active {
		t.Fatal("expected help popup to stay closed without ':' prefix")
	}
}

func TestHandleKey_CommandRequiresExplicitPrefix(t *testing.T) {
	// Arrange
	model := &Model{viewMode: ViewRecords}

	// Act
	for _, r := range "config" {
		model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	_, cmd := model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})

	// Assert
	if cmd != nil {
		if _, ok := cmd().(tea.QuitMsg); ok {
			t.Fatal("expected no quit command without explicit ':' prefix")
		}
	}
}

func TestHandleKey_QKeyWithoutCommandPrefixKeepsSessionActive(t *testing.T) {
	// Arrange
	model := &Model{viewMode: ViewRecords}

	// Act
	_, cmd := model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	// Assert
	if cmd != nil {
		if _, ok := cmd().(tea.QuitMsg); ok {
			t.Fatal("expected q without ':' prefix to keep runtime active")
		}
	}
}

func TestHandleKey_CtrlCWithoutCommandPrefixKeepsSessionActive(t *testing.T) {
	// Arrange
	model := &Model{viewMode: ViewRecords}

	// Act
	_, cmd := model.handleKey(tea.KeyMsg{Type: tea.KeyCtrlC})

	// Assert
	if cmd != nil {
		if _, ok := cmd().(tea.QuitMsg); ok {
			t.Fatal("expected Ctrl+c without ':' prefix to keep runtime active")
		}
	}
}

func TestHandleKey_CommandHelpReenterKeepsPopupOpen(t *testing.T) {
	// Arrange
	model := &Model{viewMode: ViewRecords}
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
	for _, r := range "help" {
		model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})
	if !model.helpPopup.active {
		t.Fatal("expected help popup to open before idempotence check")
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
	for _, r := range "help" {
		model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})

	// Assert
	if !model.helpPopup.active {
		t.Fatal("expected help popup to remain open when :help is re-entered")
	}
	if strings.Contains(strings.ToLower(model.statusMessage), "unknown command") {
		t.Fatalf("expected no unknown-command status for repeated :help, got %q", model.statusMessage)
	}
}

func TestHandleKey_CommandHelpReenterDoesNotResetStatusUnexpectedly(t *testing.T) {
	// Arrange
	model := &Model{viewMode: ViewRecords, statusMessage: "existing status"}
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
	for _, r := range "help" {
		model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})
	if !model.helpPopup.active {
		t.Fatal("expected help popup to open before re-entering :help")
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
	for _, r := range "help" {
		model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})

	// Assert
	if model.statusMessage != "existing status" {
		t.Fatalf("expected status message to remain unchanged, got %q", model.statusMessage)
	}
	if !model.helpPopup.active {
		t.Fatal("expected help popup to remain open")
	}
}

func TestHandleKey_HelpPopupEscClosesPopup(t *testing.T) {
	// Arrange
	model := &Model{viewMode: ViewRecords}
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
	for _, r := range "help" {
		model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})
	if !model.helpPopup.active {
		t.Fatal("expected help popup to open before close check")
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyEsc})

	// Assert
	if model.helpPopup.active {
		t.Fatal("expected Esc to close help popup")
	}
}

func TestHandleKey_HelpPopupUnrelatedKeysDoNotClosePopup(t *testing.T) {
	// Arrange
	model := &Model{viewMode: ViewRecords}
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
	for _, r := range "help" {
		model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})
	if !model.helpPopup.active {
		t.Fatal("expected help popup to open before unrelated-key check")
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})

	// Assert
	if !model.helpPopup.active {
		t.Fatal("expected unrelated keys to keep help popup open")
	}
}

func TestHandleKey_MisspelledHelpCommandUsesUnknownCommandFallback(t *testing.T) {
	// Arrange
	model := &Model{viewMode: ViewRecords}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
	for _, r := range "helpp" {
		model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	_, cmd := model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})

	// Assert
	if cmd != nil {
		if _, ok := cmd().(tea.QuitMsg); ok {
			t.Fatal("expected misspelled :help to keep session active")
		}
	}
	if model.helpPopup.active {
		t.Fatal("expected misspelled :help to keep help popup closed")
	}
	if !strings.Contains(strings.ToLower(model.statusMessage), "unknown command") {
		t.Fatalf("expected unknown-command status for misspelled :help, got %q", model.statusMessage)
	}
}

func TestHandleKey_DirtyConfigCommandOpensDecisionPrompt(t *testing.T) {
	for _, tc := range []struct {
		name    string
		command string
	}{
		{name: "full command", command: "config"},
		{name: "alias", command: "c"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			model := &Model{
				viewMode:       ViewRecords,
				pendingInserts: []pendingInsertRow{{}},
			}

			// Act
			model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
			for _, r := range tc.command {
				model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
			}
			_, cmd := model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})

			// Assert
			if cmd != nil {
				if _, ok := cmd().(tea.QuitMsg); ok {
					t.Fatalf("expected dirty :%s to wait for explicit decision", tc.command)
				}
			}
			if !model.confirmPopup.active {
				t.Fatalf("expected dirty :%s decision popup to open", tc.command)
			}
			if model.openConfigSelector {
				t.Fatalf("expected :%s navigation to remain blocked until explicit decision", tc.command)
			}
		})
	}
}

func TestHandleConfirmPopupKey_DirtyConfigCancelKeepsStagedState(t *testing.T) {
	// Arrange
	model := &Model{
		viewMode:       ViewRecords,
		pendingInserts: []pendingInsertRow{{}},
	}
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
	for _, r := range "config" {
		model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})

	// Act
	model.handleConfirmPopupKey(tea.KeyMsg{Type: tea.KeyEsc})

	// Assert
	if model.confirmPopup.active {
		t.Fatal("expected decision popup to close on cancel")
	}
	if !model.hasDirtyEdits() {
		t.Fatal("expected staged changes to stay untouched on cancel")
	}
	if model.openConfigSelector {
		t.Fatal("expected no navigation on cancel")
	}
}

func TestHandleConfirmPopupKey_DirtyConfigDiscardClearsStateAndNavigates(t *testing.T) {
	// Arrange
	model := &Model{
		viewMode:       ViewRecords,
		pendingInserts: []pendingInsertRow{{}},
	}
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
	for _, r := range "config" {
		model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})
	model.handleConfirmPopupKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})

	// Act
	_, cmd := model.handleConfirmPopupKey(tea.KeyMsg{Type: tea.KeyEnter})

	// Assert
	if cmd == nil {
		t.Fatal("expected quit command after discard decision")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatalf("expected tea.QuitMsg after discard decision, got %T", cmd())
	}
	if model.hasDirtyEdits() {
		t.Fatal("expected staged changes to be cleared on discard")
	}
	if !model.openConfigSelector {
		t.Fatal("expected selector navigation after discard")
	}
}

func TestUpdate_DirtyConfigSaveSuccessNavigatesAfterSave(t *testing.T) {
	// Arrange
	engine := &tuiSpyEngine{}
	saveChanges := usecase.NewSaveTableChanges(engine)
	model := &Model{
		ctx:         context.Background(),
		viewMode:    ViewRecords,
		saveChanges: saveChanges,
		schema: dto.Schema{
			Columns: []dto.SchemaColumn{
				{Name: "id", Type: "INTEGER", PrimaryKey: true, AutoIncrement: true},
				{Name: "name", Type: "TEXT", Nullable: false},
			},
		},
		pendingInserts: []pendingInsertRow{
			{
				values: map[int]stagedEdit{
					0: {Value: domainmodel.Value{Text: "", Raw: ""}},
					1: {Value: domainmodel.Value{Text: "new", Raw: "new"}},
				},
				explicitAuto: map[int]bool{},
			},
		},
		tables: []dto.Table{{Name: "users"}},
	}
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
	for _, r := range "config" {
		model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})

	// Act
	_, saveCmd := model.handleConfirmPopupKey(tea.KeyMsg{Type: tea.KeyEnter})
	if saveCmd == nil {
		t.Fatal("expected save command after selecting save decision")
	}
	msg := saveCmd()
	_, quitCmd := model.Update(msg)

	// Assert
	if quitCmd == nil {
		t.Fatal("expected quit command after successful save decision")
	}
	if _, ok := quitCmd().(tea.QuitMsg); !ok {
		t.Fatalf("expected tea.QuitMsg after successful save decision, got %T", quitCmd())
	}
	if model.hasDirtyEdits() {
		t.Fatal("expected staged changes to be cleared after successful save")
	}
	if !model.openConfigSelector {
		t.Fatal("expected selector navigation after successful save")
	}
}

func TestUpdate_DirtyConfigSaveFailureKeepsStateAndBlocksNavigation(t *testing.T) {
	// Arrange
	engine := &tuiSpyEngine{saveErr: errors.New("boom")}
	saveChanges := usecase.NewSaveTableChanges(engine)
	model := &Model{
		ctx:         context.Background(),
		viewMode:    ViewRecords,
		saveChanges: saveChanges,
		schema: dto.Schema{
			Columns: []dto.SchemaColumn{
				{Name: "id", Type: "INTEGER", PrimaryKey: true, AutoIncrement: true},
				{Name: "name", Type: "TEXT", Nullable: false},
			},
		},
		pendingInserts: []pendingInsertRow{
			{
				values: map[int]stagedEdit{
					0: {Value: domainmodel.Value{Text: "", Raw: ""}},
					1: {Value: domainmodel.Value{Text: "new", Raw: "new"}},
				},
				explicitAuto: map[int]bool{},
			},
		},
		tables: []dto.Table{{Name: "users"}},
	}
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
	for _, r := range "config" {
		model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})

	// Act
	_, saveCmd := model.handleConfirmPopupKey(tea.KeyMsg{Type: tea.KeyEnter})
	if saveCmd == nil {
		t.Fatal("expected save command after selecting save decision")
	}
	msg := saveCmd()
	_, quitCmd := model.Update(msg)

	// Assert
	if quitCmd != nil {
		if _, ok := quitCmd().(tea.QuitMsg); ok {
			t.Fatal("expected no navigation when save fails")
		}
	}
	if !model.hasDirtyEdits() {
		t.Fatal("expected staged changes to be preserved on save error")
	}
	if model.openConfigSelector {
		t.Fatal("expected selector navigation to remain blocked on save error")
	}
	if !strings.Contains(model.statusMessage, "boom") {
		t.Fatalf("expected save error status to be surfaced, got %q", model.statusMessage)
	}
}

func TestHandleKey_InsertCreatesPendingRowAtTop(t *testing.T) {
	// Arrange
	defaultName := "guest"
	model := &Model{
		viewMode: ViewRecords,
		focus:    FocusContent,
		schema: dto.Schema{
			Columns: []dto.SchemaColumn{
				{Name: "id", Type: "INTEGER", PrimaryKey: true, AutoIncrement: true},
				{Name: "name", Type: "TEXT", Nullable: false, DefaultValue: &defaultName},
				{Name: "note", Type: "TEXT", Nullable: true},
				{Name: "age", Type: "INTEGER", Nullable: false},
			},
		},
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})

	// Assert
	if len(model.pendingInserts) != 1 {
		t.Fatalf("expected one pending insert, got %d", len(model.pendingInserts))
	}
	if model.recordSelection != 0 {
		t.Fatalf("expected selection at top pending row, got %d", model.recordSelection)
	}
	if model.recordColumn != 1 {
		t.Fatalf("expected first editable column to skip auto field, got %d", model.recordColumn)
	}
	row := model.pendingInserts[0]
	if got := displayValue(row.values[1].Value); got != "guest" {
		t.Fatalf("expected default value guest, got %q", got)
	}
	if !row.values[2].Value.IsNull {
		t.Fatalf("expected nullable column to default to NULL")
	}
	if got := displayValue(row.values[3].Value); got != "" {
		t.Fatalf("expected required no-default column to be empty, got %q", got)
	}
}

func TestHandleKey_DeleteTogglesPersistedRow(t *testing.T) {
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
		t.Fatalf("expected one pending delete, got %d", len(model.pendingDeletes))
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})

	// Assert
	if len(model.pendingDeletes) != 0 {
		t.Fatalf("expected pending delete to toggle off")
	}
}

func TestHandleKey_DeleteRemovesPendingInsert(t *testing.T) {
	// Arrange
	model := &Model{
		viewMode:       ViewRecords,
		focus:          FocusContent,
		pendingInserts: []pendingInsertRow{{values: map[int]stagedEdit{}}},
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})

	// Assert
	if len(model.pendingInserts) != 0 {
		t.Fatalf("expected pending insert to be removed")
	}
}

func TestBuildTableChanges_IgnoresUpdatesForDeletedRows(t *testing.T) {
	// Arrange
	model := &Model{
		schema: dto.Schema{
			Columns: []dto.SchemaColumn{
				{Name: "id", Type: "INTEGER", PrimaryKey: true, AutoIncrement: true},
				{Name: "name", Type: "TEXT", Nullable: false},
			},
		},
		pendingInserts: []pendingInsertRow{
			{
				values: map[int]stagedEdit{
					0: {Value: domainmodel.Value{Text: "", Raw: ""}},
					1: {Value: domainmodel.Value{Text: "new", Raw: "new"}},
				},
				explicitAuto: map[int]bool{},
			},
		},
		pendingUpdates: map[string]recordEdits{
			"id=1": {
				identity: domainmodel.RecordIdentity{
					Keys: []domainmodel.ColumnValue{{Column: "id", Value: domainmodel.Value{Text: "1", Raw: int64(1)}}},
				},
				changes: map[int]stagedEdit{
					1: {Value: domainmodel.Value{Text: "bob", Raw: "bob"}},
				},
			},
		},
		pendingDeletes: map[string]recordDelete{
			"id=1": {
				identity: domainmodel.RecordIdentity{
					Keys: []domainmodel.ColumnValue{{Column: "id", Value: domainmodel.Value{Text: "1", Raw: int64(1)}}},
				},
			},
		},
	}

	// Act
	changes, err := model.buildTableChanges()

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(changes.Inserts) != 1 {
		t.Fatalf("expected one insert, got %d", len(changes.Inserts))
	}
	if len(changes.Updates) != 0 {
		t.Fatalf("expected updates for deleted row to be ignored")
	}
	if len(changes.Deletes) != 1 {
		t.Fatalf("expected one delete, got %d", len(changes.Deletes))
	}
}

func TestDirtyEditCount_IncludesInsertsDeletesAndUpdates(t *testing.T) {
	// Arrange
	model := &Model{
		pendingInserts: []pendingInsertRow{{}},
		pendingDeletes: map[string]recordDelete{"id=1": {}},
		pendingUpdates: map[string]recordEdits{
			"id=2": {
				changes: map[int]stagedEdit{
					0: {Value: domainmodel.Value{Text: "x", Raw: "x"}},
					1: {Value: domainmodel.Value{Text: "y", Raw: "y"}},
				},
			},
		},
	}

	// Act
	dirty := model.dirtyEditCount()

	// Assert
	if dirty != 4 {
		t.Fatalf("expected dirty count 4, got %d", dirty)
	}
}

func TestConfirmSaveChanges_SubmitsBuiltTableChanges(t *testing.T) {
	// Arrange
	engine := &tuiSpyEngine{}
	saveChanges := usecase.NewSaveTableChanges(engine)
	model := &Model{
		ctx:         context.Background(),
		saveChanges: saveChanges,
		schema: dto.Schema{
			Columns: []dto.SchemaColumn{
				{Name: "id", Type: "INTEGER", PrimaryKey: true, AutoIncrement: true},
				{Name: "name", Type: "TEXT", Nullable: false},
			},
		},
		pendingInserts: []pendingInsertRow{
			{
				values: map[int]stagedEdit{
					0: {Value: domainmodel.Value{Text: "", Raw: ""}},
					1: {Value: domainmodel.Value{Text: "new", Raw: "new"}},
				},
				explicitAuto: map[int]bool{},
			},
		},
		tables: []dto.Table{{Name: "users"}},
	}

	// Act
	_, cmd := model.confirmSaveChanges()
	msg := cmd()

	// Assert
	result, ok := msg.(saveChangesMsg)
	if !ok {
		t.Fatalf("expected saveChangesMsg, got %T", msg)
	}
	if result.err != nil {
		t.Fatalf("expected no error, got %v", result.err)
	}
	if len(engine.lastChanges.Inserts) != 1 {
		t.Fatalf("expected one insert payload, got %d", len(engine.lastChanges.Inserts))
	}
}

func TestSetTableSelection_WithDirtyStatePromptsAndDiscardClearsStaging(t *testing.T) {
	// Arrange
	model := &Model{
		tables:            []dto.Table{{Name: "users"}, {Name: "orders"}},
		selectedTable:     0,
		pendingInserts:    []pendingInsertRow{{}},
		pendingUpdates:    map[string]recordEdits{"id=1": {changes: map[int]stagedEdit{0: {Value: domainmodel.Value{Text: "x", Raw: "x"}}}}},
		pendingDeletes:    map[string]recordDelete{"id=2": {}},
		pendingTableIndex: -1,
	}

	// Act
	model.setTableSelection(1)

	// Assert
	if !model.confirmPopup.active || model.confirmPopup.action != confirmDiscardTable {
		t.Fatalf("expected discard confirmation popup")
	}
	if model.selectedTable != 0 {
		t.Fatalf("expected table selection not to change before confirmation")
	}

	// Act
	model.handleConfirmPopupKey(tea.KeyMsg{Type: tea.KeyEnter})

	// Assert
	if model.selectedTable != 1 {
		t.Fatalf("expected table switch after confirmation")
	}
	if model.hasDirtyEdits() {
		t.Fatalf("expected staged state to be cleared after discard")
	}
}

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
	if err := model.stageEdit(0, 1, domainmodel.Value{Text: "bob", Raw: "bob"}); err != nil {
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

func TestHandleKey_FieldFocusNavigationAdjustsColumnForPendingInsertRows(t *testing.T) {
	// Arrange
	model := &Model{
		viewMode:         ViewRecords,
		focus:            FocusContent,
		recordFieldFocus: true,
		recordSelection:  1,
		recordColumn:     0,
		pendingInserts: []pendingInsertRow{
			{
				values: map[int]stagedEdit{
					0: {Value: domainmodel.Value{Text: "", Raw: ""}},
					1: {Value: domainmodel.Value{Text: "new", Raw: "new"}},
				},
				explicitAuto: map[int]bool{},
			},
		},
		records: []dto.RecordRow{
			{Values: []string{"1", "alice"}},
		},
		schema: dto.Schema{
			Columns: []dto.SchemaColumn{
				{Name: "id", Type: "INTEGER", PrimaryKey: true, AutoIncrement: true},
				{Name: "name", Type: "TEXT", Nullable: false},
			},
		},
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})

	// Assert
	if model.recordSelection != 0 {
		t.Fatalf("expected selection to move to pending insert row, got %d", model.recordSelection)
	}
	if model.recordColumn != 1 {
		t.Fatalf("expected focused column to move off hidden auto-increment field, got %d", model.recordColumn)
	}
}

type tuiSpyEngine struct {
	lastChanges domainmodel.TableChanges
	saveErr     error
}

func (s *tuiSpyEngine) ListTables(ctx context.Context) ([]domainmodel.Table, error) {
	return nil, nil
}

func (s *tuiSpyEngine) GetSchema(ctx context.Context, tableName string) (domainmodel.Schema, error) {
	return domainmodel.Schema{}, nil
}

func (s *tuiSpyEngine) ListRecords(ctx context.Context, tableName string, offset, limit int, filter *domainmodel.Filter) (domainmodel.RecordPage, error) {
	return domainmodel.RecordPage{}, nil
}

func (s *tuiSpyEngine) ListOperators(ctx context.Context, columnType string) ([]domainmodel.Operator, error) {
	return nil, nil
}

func (s *tuiSpyEngine) ApplyRecordChanges(ctx context.Context, tableName string, changes domainmodel.TableChanges) error {
	s.lastChanges = changes
	return s.saveErr
}
