package tui

import (
	"context"
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/dto"
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

func TestHandleKey_ShiftSOpensSortPopupInRecordsView(t *testing.T) {
	// Arrange
	model := &Model{
		viewMode: ViewRecords,
		focus:    FocusContent,
		tables:   []dto.Table{{Name: "users"}},
		schema: dto.Schema{
			Columns: []dto.SchemaColumn{
				{Name: "id", Type: "INTEGER"},
				{Name: "name", Type: "TEXT"},
			},
		},
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'S'}})

	// Assert
	if !model.sortPopup.active {
		t.Fatal("expected sort popup to be active")
	}
	if model.sortPopup.step != sortSelectColumn {
		t.Fatalf("expected sort popup to start at column step, got %v", model.sortPopup.step)
	}
}

func TestHandleKey_ShiftSIgnoredOutsideRecordsView(t *testing.T) {
	// Arrange
	model := &Model{
		viewMode: ViewSchema,
		focus:    FocusContent,
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'S'}})

	// Assert
	if model.sortPopup.active {
		t.Fatal("expected sort popup to stay closed outside records view")
	}
}

func TestHandleKey_ShiftFOpensFilterPopupInRecordsView(t *testing.T) {
	// Arrange
	model := &Model{
		viewMode: ViewRecords,
		focus:    FocusContent,
		tables:   []dto.Table{{Name: "users"}},
		schema: dto.Schema{
			Columns: []dto.SchemaColumn{
				{Name: "id", Type: "INTEGER"},
				{Name: "name", Type: "TEXT"},
			},
		},
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'F'}})

	// Assert
	if !model.filterPopup.active {
		t.Fatal("expected filter popup to be active")
	}
	if model.filterPopup.step != filterSelectColumn {
		t.Fatalf("expected filter popup to start at column step, got %v", model.filterPopup.step)
	}
}

func TestHandleKey_ShiftFIgnoredOutsideRecordsContext(t *testing.T) {
	tests := []struct {
		name     string
		viewMode ViewMode
		focus    PanelFocus
	}{
		{
			name:     "schema content",
			viewMode: ViewSchema,
			focus:    FocusContent,
		},
		{
			name:     "records tables panel",
			viewMode: ViewRecords,
			focus:    FocusTables,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			model := &Model{
				viewMode: tc.viewMode,
				focus:    tc.focus,
				tables:   []dto.Table{{Name: "users"}},
				schema: dto.Schema{
					Columns: []dto.SchemaColumn{
						{Name: "id", Type: "INTEGER"},
						{Name: "name", Type: "TEXT"},
					},
				},
			}

			// Act
			model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'F'}})

			// Assert
			if model.filterPopup.active {
				t.Fatal("expected filter popup to stay closed outside records context")
			}
		})
	}
}

func TestHandleFilterPopupKey_EnterProgressesStepsAndAppliesFilter(t *testing.T) {
	// Arrange
	operatorsSpy := &spyListOperatorsUseCase{
		operators: []dto.Operator{
			{Name: "Equals", Kind: dto.OperatorKindEq, RequiresValue: true},
		},
	}
	recordsSpy := &spyListRecordsUseCase{}
	model := &Model{
		ctx:           context.Background(),
		viewMode:      ViewRecords,
		focus:         FocusContent,
		listOperators: operatorsSpy,
		listRecords:   recordsSpy,
		tables:        []dto.Table{{Name: "users"}},
		schema: dto.Schema{
			Columns: []dto.SchemaColumn{
				{Name: "id", Type: "INTEGER"},
				{Name: "name", Type: "TEXT"},
			},
		},
		filterPopup: filterPopup{
			active:      true,
			step:        filterSelectColumn,
			columnIndex: 1,
		},
		recordPageIndex: 3,
	}

	// Act
	_, cmd := model.handleFilterPopupKey(tea.KeyMsg{Type: tea.KeyEnter})

	// Assert
	if cmd != nil {
		t.Fatal("expected no command when moving from column to operator step")
	}
	if model.filterPopup.step != filterSelectOperator {
		t.Fatalf("expected operator step, got %v", model.filterPopup.step)
	}
	if operatorsSpy.lastColumnType != "TEXT" {
		t.Fatalf("expected operator lookup for TEXT column, got %q", operatorsSpy.lastColumnType)
	}

	// Act
	_, cmd = model.handleFilterPopupKey(tea.KeyMsg{Type: tea.KeyEnter})

	// Assert
	if cmd != nil {
		t.Fatal("expected no command when moving to value-input step")
	}
	if model.filterPopup.step != filterInputValue {
		t.Fatalf("expected value-input step, got %v", model.filterPopup.step)
	}

	// Act
	model.handleFilterPopupKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("alice")})
	_, cmd = model.handleFilterPopupKey(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected records reload command after filter apply")
	}
	msg := cmd()
	model.Update(msg)

	// Assert
	if model.filterPopup.active {
		t.Fatal("expected filter popup to close after apply")
	}
	if model.currentFilter == nil {
		t.Fatal("expected current filter to be set")
	}
	if model.currentFilter.Column != "name" {
		t.Fatalf("expected filter column name, got %q", model.currentFilter.Column)
	}
	if model.currentFilter.Value != "alice" {
		t.Fatalf("expected filter value alice, got %q", model.currentFilter.Value)
	}
	if model.currentFilter.Operator.Kind != dto.OperatorKindEq {
		t.Fatalf("expected operator kind %q, got %q", dto.OperatorKindEq, model.currentFilter.Operator.Kind)
	}
	if model.recordPageIndex != 0 {
		t.Fatalf("expected page index reset to 0 after filter apply, got %d", model.recordPageIndex)
	}
	if recordsSpy.lastFilter == nil {
		t.Fatal("expected filter forwarded to list-records use case")
	}
	if recordsSpy.lastFilter.Column != "name" {
		t.Fatalf("expected forwarded filter column name, got %q", recordsSpy.lastFilter.Column)
	}
	if recordsSpy.lastFilter.Value != "alice" {
		t.Fatalf("expected forwarded filter value alice, got %q", recordsSpy.lastFilter.Value)
	}
}

func TestHandleFilterPopupKey_InputEditingSupportsCursorMovementAndBackspace(t *testing.T) {
	// Arrange
	model := &Model{
		filterPopup: filterPopup{
			active: true,
			step:   filterInputValue,
			input:  "ac",
			cursor: 1,
		},
	}

	// Act
	model.handleFilterPopupKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	model.handleFilterPopupKey(tea.KeyMsg{Type: tea.KeyLeft})
	model.handleFilterPopupKey(tea.KeyMsg{Type: tea.KeyBackspace})
	model.handleFilterPopupKey(tea.KeyMsg{Type: tea.KeyLeft})
	model.handleFilterPopupKey(tea.KeyMsg{Type: tea.KeyRight})
	model.handleFilterPopupKey(tea.KeyMsg{Type: tea.KeyRight})
	model.handleFilterPopupKey(tea.KeyMsg{Type: tea.KeyRight})

	// Assert
	if model.filterPopup.input != "bc" {
		t.Fatalf("expected edited input bc, got %q", model.filterPopup.input)
	}
	if model.filterPopup.cursor != 2 {
		t.Fatalf("expected cursor clamped at input end, got %d", model.filterPopup.cursor)
	}
}

func TestHandleFilterPopupKey_EscClosesPopup(t *testing.T) {
	// Arrange
	model := &Model{
		filterPopup: filterPopup{
			active: true,
			step:   filterSelectOperator,
		},
	}

	// Act
	_, cmd := model.handleFilterPopupKey(tea.KeyMsg{Type: tea.KeyEsc})

	// Assert
	if cmd != nil {
		t.Fatal("expected no command when closing filter popup")
	}
	if model.filterPopup.active {
		t.Fatal("expected filter popup to close on Esc")
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
		pendingUpdates: map[string]recordEdits{
			"id=1": {
				changes: map[int]stagedEdit{
					1: {Value: dto.StagedValue{Text: "bob", Raw: "bob"}},
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

func TestHandleKey_ShiftSApplySortReloadsRecords(t *testing.T) {
	// Arrange
	recordsSpy := &spyListRecordsUseCase{
		page: dto.RecordPage{
			Rows: []dto.RecordRow{
				{Values: []string{"1", "alice"}},
			},
		},
	}
	model := &Model{
		ctx:         context.Background(),
		viewMode:    ViewRecords,
		focus:       FocusContent,
		listRecords: recordsSpy,
		tables:      []dto.Table{{Name: "users"}},
		schema: dto.Schema{
			Columns: []dto.SchemaColumn{
				{Name: "id", Type: "INTEGER"},
				{Name: "name", Type: "TEXT"},
			},
		},
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'S'}})
	model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	_, cmd := model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected records reload command after applying sort")
	}
	msg := cmd()
	model.Update(msg)

	// Assert
	if model.sortPopup.active {
		t.Fatal("expected sort popup to close after apply")
	}
	if model.currentSort == nil {
		t.Fatal("expected current sort to be set")
	}
	if model.currentSort.Column != "id" {
		t.Fatalf("expected sorted column id, got %q", model.currentSort.Column)
	}
	if model.currentSort.Direction != dto.SortDirectionDesc {
		t.Fatalf("expected sort direction DESC, got %s", model.currentSort.Direction)
	}
	if recordsSpy.lastSort == nil {
		t.Fatal("expected engine to receive sort")
	}
	if recordsSpy.lastSort.Column != "id" || recordsSpy.lastSort.Direction != dto.SortDirectionDesc {
		t.Fatalf("expected engine sort id DESC, got %+v", recordsSpy.lastSort)
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
			if !model.confirmPopup.modal {
				t.Fatalf("expected dirty :%s decision popup to be modal", tc.command)
			}
			if model.confirmPopup.title != "Config" {
				t.Fatalf("expected dirty :%s popup title Config, got %q", tc.command, model.confirmPopup.title)
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
	saveChanges := &spySaveChangesUseCase{}
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
					0: {Value: dto.StagedValue{Text: "", Raw: ""}},
					1: {Value: dto.StagedValue{Text: "new", Raw: "new"}},
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
	saveChanges := &spySaveChangesUseCase{err: errors.New("boom")}
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
					0: {Value: dto.StagedValue{Text: "", Raw: ""}},
					1: {Value: dto.StagedValue{Text: "new", Raw: "new"}},
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
