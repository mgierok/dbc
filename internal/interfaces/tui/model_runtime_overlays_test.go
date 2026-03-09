package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/dto"
)

func TestHandleKey_EInRecordsEnablesFieldFocus(t *testing.T) {
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
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})

	// Assert
	if !model.recordFieldFocus {
		t.Fatalf("expected record field focus to be enabled")
	}
}

func TestHandleKey_EInFieldFocusOpensEditPopup(t *testing.T) {
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
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})

	// Assert
	if !model.editPopup.active {
		t.Fatalf("expected edit popup to be active")
	}
}

func TestHandleKey_EnterOpensRecordDetailInRecordsView(t *testing.T) {
	// Arrange
	model := &Model{
		viewMode: ViewRecords,
		focus:    FocusContent,
		records:  []dto.RecordRow{{Values: []string{"1", "alice"}}},
		schema: dto.Schema{
			Columns: []dto.SchemaColumn{
				{Name: "id", Type: "INTEGER"},
				{Name: "name", Type: "TEXT"},
			},
		},
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})

	// Assert
	if !model.recordDetail.active {
		t.Fatal("expected record detail to open in records view")
	}
	if model.recordDetail.scrollOffset != 0 {
		t.Fatalf("expected record detail scroll offset reset to 0, got %d", model.recordDetail.scrollOffset)
	}
}

func TestHandleKey_EnterIgnoredOutsideRecordsViewForDetail(t *testing.T) {
	// Arrange
	model := &Model{
		viewMode: ViewSchema,
		focus:    FocusContent,
		records:  []dto.RecordRow{{Values: []string{"1", "alice"}}},
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})

	// Assert
	if model.recordDetail.active {
		t.Fatal("expected record detail to stay closed outside records view")
	}
}

func TestHandleKey_RecordDetailEscClosesDetailBeforeSwitchingPanels(t *testing.T) {
	// Arrange
	model := &Model{
		viewMode: ViewRecords,
		focus:    FocusContent,
		records:  []dto.RecordRow{{Values: []string{"1", "alice"}}},
		schema: dto.Schema{
			Columns: []dto.SchemaColumn{
				{Name: "id", Type: "INTEGER"},
				{Name: "name", Type: "TEXT"},
			},
		},
		recordDetail: recordDetailState{
			active: true,
		},
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyEsc})

	// Assert
	if model.recordDetail.active {
		t.Fatal("expected Esc to close record detail")
	}
	if model.focus != FocusContent {
		t.Fatalf("expected focus to stay in content after closing detail, got %v", model.focus)
	}
	if model.viewMode != ViewRecords {
		t.Fatalf("expected records view to remain active after closing detail, got %v", model.viewMode)
	}
}

func TestHandleKey_RecordDetailScrollMovesOffset(t *testing.T) {
	// Arrange
	model := &Model{
		viewMode: ViewRecords,
		focus:    FocusContent,
		width:    40,
		height:   8,
		recordDetail: recordDetailState{
			active: true,
		},
		schema: dto.Schema{
			Columns: []dto.SchemaColumn{
				{Name: "payload", Type: "TEXT"},
			},
		},
		records: []dto.RecordRow{
			{Values: []string{"abcdefghijklmnopqrstuvwxyz0123456789abcdefghijklmnopqrstuvwxyz0123456789"}},
		},
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})

	// Assert
	if model.recordDetail.scrollOffset <= 0 {
		t.Fatalf("expected detail scroll offset to increase, got %d", model.recordDetail.scrollOffset)
	}
}

func TestRecordDetailContentLines_UsesStagedEffectiveValue(t *testing.T) {
	// Arrange
	model := &Model{
		viewMode: ViewRecords,
		focus:    FocusContent,
		schema: dto.Schema{
			Columns: []dto.SchemaColumn{
				{Name: "id", Type: "INTEGER", PrimaryKey: true},
				{Name: "name", Type: "TEXT"},
			},
		},
		records: []dto.RecordRow{
			{Values: []string{"1", "alice"}},
		},
		staging: stagingState{
			pendingUpdates: map[string]recordEdits{
				"id=1": {
					changes: map[int]stagedEdit{
						1: {Value: dto.StagedValue{Text: "bob", Raw: "bob"}},
					},
				},
			},
		},
	}

	// Act
	content := strings.Join(model.recordDetailContentLines(80), "\n")

	// Assert
	if !strings.Contains(content, "bob") {
		t.Fatalf("expected detail to include staged value, got %q", content)
	}
	if strings.Contains(content, "alice") {
		t.Fatalf("expected original value to be replaced by staged value, got %q", content)
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

func TestHandleKey_ContextHelpPopupBehaviors(t *testing.T) {
	newRuntimeModel := func() *Model {
		return &Model{
			viewMode: ViewRecords,
			focus:    FocusContent,
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

	t.Run("FR-001 happy path opens context help popup with ?", func(t *testing.T) {
		// Arrange
		model := newRuntimeModel()

		// Act
		model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})

		// Assert
		if !model.helpPopup.active {
			t.Fatal("expected ? to open help popup")
		}
		if model.helpPopup.context != helpPopupContextRecords {
			t.Fatalf("expected records context, got %v", model.helpPopup.context)
		}
	})

	t.Run("FR-001 compatibility path keeps :help alias behavior", func(t *testing.T) {
		// Arrange
		model := newRuntimeModel()

		// Act
		cmd := submitCommand(model, "help")

		// Assert
		assertSessionActive(t, cmd, ":help")
		if !model.helpPopup.active {
			t.Fatal("expected :help to open help popup")
		}
		if model.helpPopup.context != helpPopupContextRecords {
			t.Fatalf("expected records context for :help, got %v", model.helpPopup.context)
		}
		if strings.Contains(strings.ToLower(model.statusMessage), "unknown command") {
			t.Fatalf("expected no unknown-command status for :help, got %q", model.statusMessage)
		}
	})

	t.Run("FR-002 popup renders only current context bindings", func(t *testing.T) {
		// Arrange
		model := newRuntimeModel()
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
	})

	t.Run("FR-003 scrolling reaches final context shortcut", func(t *testing.T) {
		// Arrange
		model := newRuntimeModel()
		model.height = 12
		model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})

		// Act
		initial := strings.Join(model.renderHelpPopup(60), "\n")
		for range 30 {
			model.handleHelpPopupKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		}
		scrolled := strings.Join(model.renderHelpPopup(60), "\n")

		// Assert
		if strings.Contains(initial, "Shift+S sort") {
			t.Fatalf("expected final help item to be hidden before scrolling, got %q", initial)
		}
		if !strings.Contains(scrolled, "Shift+S sort") {
			t.Fatalf("expected final help item to be reachable after scrolling, got %q", scrolled)
		}
	})

	t.Run("FR-004 happy path keeps popup open on repeated :help", func(t *testing.T) {
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

	t.Run("FR-005 closes help popup on Esc", func(t *testing.T) {
		// Arrange
		model := newRuntimeModel()
		model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})

		// Act
		model.handleKey(tea.KeyMsg{Type: tea.KeyEsc})

		// Assert
		if model.helpPopup.active {
			t.Fatal("expected Esc to close help popup")
		}
	})

	t.Run("FR-006 keeps unsupported command fallback", func(t *testing.T) {
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

func TestHandleKey_CommandHelpReenterClearsStaleStatusBeforeNewCommand(t *testing.T) {
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
	if model.statusMessage != "" {
		t.Fatalf("expected stale status message to clear before new command, got %q", model.statusMessage)
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

func TestHandleKey_ContextHelpFromFilterPopupUsesFilterContext(t *testing.T) {
	// Arrange
	model := &Model{
		viewMode: ViewRecords,
		focus:    FocusContent,
		filterPopup: filterPopup{
			active: true,
			step:   filterSelectColumn,
		},
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})

	// Assert
	if !model.helpPopup.active {
		t.Fatal("expected ? to open help popup from filter context")
	}
	if model.helpPopup.context != helpPopupContextFilterPopup {
		t.Fatalf("expected filter-popup context help, got %v", model.helpPopup.context)
	}
	if !model.filterPopup.active {
		t.Fatal("expected filter popup state to stay preserved under help overlay")
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

func TestHandleKey_PopupPriority_HelpPopupConsumesEscBeforeOtherPopups(t *testing.T) {
	// Arrange
	model := &Model{
		helpPopup:   helpPopup{active: true},
		filterPopup: filterPopup{active: true},
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyEsc})

	// Assert
	if model.helpPopup.active {
		t.Fatal("expected help popup to close first")
	}
	if !model.filterPopup.active {
		t.Fatal("expected filter popup to remain active when help popup handled Esc")
	}
}
