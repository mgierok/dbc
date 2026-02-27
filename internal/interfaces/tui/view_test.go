package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/dto"
	domainmodel "github.com/mgierok/dbc/internal/domain/model"
)

func TestStatusShortcuts_TablesPanel(t *testing.T) {
	// Arrange
	model := &Model{focus: FocusTables}

	// Act
	shortcuts := model.statusShortcuts()

	// Assert
	if shortcuts != "Tables: Enter records | F filter" {
		t.Fatalf("expected table shortcuts, got %q", shortcuts)
	}
}

func TestStatusShortcuts_SchemaPanel(t *testing.T) {
	// Arrange
	model := &Model{
		focus:    FocusContent,
		viewMode: ViewSchema,
	}

	// Act
	shortcuts := model.statusShortcuts()

	// Assert
	if shortcuts != "Schema: Esc tables | F filter" {
		t.Fatalf("expected schema shortcuts, got %q", shortcuts)
	}
}

func TestStatusShortcuts_Popup(t *testing.T) {
	// Arrange
	model := &Model{
		filterPopup: filterPopup{active: true},
	}

	// Act
	shortcuts := model.statusShortcuts()

	// Assert
	if shortcuts != "Popup: Enter apply | Esc close" {
		t.Fatalf("expected popup shortcuts, got %q", shortcuts)
	}
}

func TestStatusShortcuts_RecordsPanel(t *testing.T) {
	// Arrange
	model := &Model{
		focus:    FocusContent,
		viewMode: ViewRecords,
	}

	// Act
	shortcuts := model.statusShortcuts()

	// Assert
	if shortcuts != "Records: Esc tables | e edit | Enter detail | i insert | d delete | u undo | Ctrl+r redo | w save | F filter | Shift+S sort" {
		t.Fatalf("expected records shortcuts, got %q", shortcuts)
	}
}

func TestStatusShortcuts_RecordDetailPanel(t *testing.T) {
	// Arrange
	model := &Model{
		focus:    FocusContent,
		viewMode: ViewRecords,
		recordDetail: recordDetailState{
			active: true,
		},
	}

	// Act
	shortcuts := model.statusShortcuts()

	// Assert
	if shortcuts != "Detail: Esc back | j/k scroll | Ctrl+f/Ctrl+b page" {
		t.Fatalf("expected detail shortcuts, got %q", shortcuts)
	}
}

func TestRenderStatus_ShowsDirtyCount(t *testing.T) {
	// Arrange
	model := &Model{
		pendingUpdates: map[string]recordEdits{
			"id=1": {
				changes: map[int]stagedEdit{
					0: {Value: domainmodel.Value{Text: "bob", Raw: "bob"}},
				},
			},
		},
		pendingInserts: []pendingInsertRow{{}},
		pendingDeletes: map[string]recordDelete{"id=2": {}},
	}

	// Act
	status := model.renderStatus(80)

	// Assert
	if !strings.Contains(status, "WRITE (dirty: 3)") {
		t.Fatalf("expected dirty status, got %q", status)
	}
}

func TestRenderStatus_CommandPromptShowsCaretAtCursor(t *testing.T) {
	// Arrange
	model := &Model{
		commandInput: commandInput{
			active: true,
			value:  "config",
			cursor: 3,
		},
	}

	// Act
	status := model.renderStatus(200)

	// Assert
	if !strings.Contains(status, "Command: :con|fig") {
		t.Fatalf("expected command prompt caret in status, got %q", status)
	}
}

func TestRenderFilterPopup_ValueInputShowsCaretAtCursor(t *testing.T) {
	// Arrange
	model := &Model{
		filterPopup: filterPopup{
			active: true,
			step:   filterInputValue,
			input:  "abc",
			cursor: 1,
		},
	}

	// Act
	popup := strings.Join(model.renderFilterPopup(60), "\n")

	// Assert
	if !strings.Contains(popup, "Value: a|bc") {
		t.Fatalf("expected caret in filter value input, got %q", popup)
	}
}

func TestRenderRecords_ShowsAscSortIndicatorInHeader(t *testing.T) {
	// Arrange
	model := &Model{
		viewMode: ViewRecords,
		schema: dto.Schema{
			Columns: []dto.SchemaColumn{
				{Name: "id", Type: "INTEGER"},
				{Name: "name", Type: "TEXT"},
			},
		},
		currentSort: &dto.Sort{
			Column:    "name",
			Direction: dto.SortDirectionAsc,
		},
		records: []dto.RecordRow{
			{Values: []string{"1", "alice"}},
		},
	}

	// Act
	lines := model.renderRecords(80, 6)

	// Assert
	if len(lines) < 2 {
		t.Fatalf("expected header row, got %v", lines)
	}
	header := lines[1]
	if !strings.Contains(header, "name ↑") {
		t.Fatalf("expected asc sort indicator in header, got %q", header)
	}
	if strings.Contains(header, "id ↑") || strings.Contains(header, "id ↓") {
		t.Fatalf("expected indicator only on sorted column, got %q", header)
	}
}

func TestRenderRecords_ShowsDescSortIndicatorInHeader(t *testing.T) {
	// Arrange
	model := &Model{
		viewMode: ViewRecords,
		schema: dto.Schema{
			Columns: []dto.SchemaColumn{
				{Name: "id", Type: "INTEGER"},
				{Name: "name", Type: "TEXT"},
			},
		},
		currentSort: &dto.Sort{
			Column:    "name",
			Direction: dto.SortDirectionDesc,
		},
		records: []dto.RecordRow{
			{Values: []string{"1", "alice"}},
		},
	}

	// Act
	lines := model.renderRecords(80, 6)

	// Assert
	if len(lines) < 2 {
		t.Fatalf("expected header row, got %v", lines)
	}
	header := lines[1]
	if !strings.Contains(header, "name ↓") {
		t.Fatalf("expected desc sort indicator in header, got %q", header)
	}
}

func TestRenderRecordDetail_UsesVerticalLayoutWithoutTruncation(t *testing.T) {
	// Arrange
	longValue := "abcdefghijklmnopqrstuvwxyz0123456789"
	model := &Model{
		viewMode: ViewRecords,
		focus:    FocusContent,
		recordDetail: recordDetailState{
			active: true,
		},
		schema: dto.Schema{
			Columns: []dto.SchemaColumn{
				{Name: "id", Type: "INTEGER"},
				{Name: "payload", Type: "TEXT"},
			},
		},
		records: []dto.RecordRow{
			{Values: []string{"1", longValue}},
		},
	}

	// Act
	lines := model.renderContent(20, 8)
	content := strings.Join(lines, "\n")

	// Assert
	if !strings.Contains(content, "ⓘ  Persisted record") {
		t.Fatalf("expected information marker in detail layout, got %q", content)
	}
	if strings.Contains(content, "[ROW]") {
		t.Fatalf("expected [ROW] marker to be removed in detail layout, got %q", content)
	}
	if !strings.Contains(content, "\x1b[1mid\x1b[0m (INTEGER)") {
		t.Fatalf("expected id header in detail layout, got %q", content)
	}
	if !strings.Contains(content, "\x1b[1mpayload\x1b[0m (TEXT)") {
		t.Fatalf("expected payload header in detail layout, got %q", content)
	}
	if strings.Contains(content, "[COL]") {
		t.Fatalf("expected [COL] marker to be removed in detail layout, got %q", content)
	}
	if strings.Contains(content, "...") {
		t.Fatalf("expected no truncation marker in detail layout, got %q", content)
	}
}

func TestRecordDetailContentLines_UsesInformationMarkerForRowStates(t *testing.T) {
	t.Run("persisted row", func(t *testing.T) {
		// Arrange
		model := &Model{
			schema: dto.Schema{
				Columns: []dto.SchemaColumn{
					{Name: "id", Type: "INTEGER", PrimaryKey: true},
				},
			},
			records: []dto.RecordRow{
				{Values: []string{"1"}},
			},
		}

		// Act
		lines := model.recordDetailContentLines(40)

		// Assert
		if !strings.Contains(lines[0], "ⓘ  Persisted record") {
			t.Fatalf("expected persisted information marker, got %q", lines[0])
		}
	})

	t.Run("pending insert row", func(t *testing.T) {
		// Arrange
		model := &Model{
			schema: dto.Schema{
				Columns: []dto.SchemaColumn{
					{Name: "id", Type: "INTEGER"},
				},
			},
			pendingInserts: []pendingInsertRow{
				{},
			},
		}

		// Act
		lines := model.recordDetailContentLines(40)

		// Assert
		if !strings.Contains(lines[0], "ⓘ  Pending insert") {
			t.Fatalf("expected pending insert information marker, got %q", lines[0])
		}
	})

	t.Run("delete-marked persisted row", func(t *testing.T) {
		// Arrange
		model := &Model{
			schema: dto.Schema{
				Columns: []dto.SchemaColumn{
					{Name: "id", Type: "INTEGER", PrimaryKey: true},
				},
			},
			records: []dto.RecordRow{
				{Values: []string{"1"}},
			},
		}
		key, ok := model.recordKeyForPersistedRow(0)
		if !ok {
			t.Fatal("expected persisted row key")
		}
		model.pendingDeletes = map[string]recordDelete{
			key: {},
		}

		// Act
		lines := model.recordDetailContentLines(40)

		// Assert
		if !strings.Contains(lines[0], "ⓘ  Marked for delete") {
			t.Fatalf("expected delete information marker, got %q", lines[0])
		}
	})

	t.Run("edited persisted row", func(t *testing.T) {
		// Arrange
		model := &Model{
			schema: dto.Schema{
				Columns: []dto.SchemaColumn{
					{Name: "id", Type: "INTEGER", PrimaryKey: true},
					{Name: "name", Type: "TEXT"},
				},
			},
			records: []dto.RecordRow{
				{Values: []string{"1", "alice"}},
			},
		}
		key, ok := model.recordKeyForPersistedRow(0)
		if !ok {
			t.Fatal("expected persisted row key")
		}
		model.pendingUpdates = map[string]recordEdits{
			key: {
				changes: map[int]stagedEdit{
					1: {Value: domainmodel.Value{Text: "alice2", Raw: "alice2"}},
				},
			},
		}

		// Act
		lines := model.recordDetailContentLines(40)

		// Assert
		if !strings.Contains(lines[0], "ⓘ  Edited record") {
			t.Fatalf("expected edited information marker, got %q", lines[0])
		}
	})
}

func TestRenderEditPopup_TextInputShowsCaretAtCursor(t *testing.T) {
	// Arrange
	model := &Model{
		schema: dto.Schema{
			Columns: []dto.SchemaColumn{
				{
					Name:  "name",
					Type:  "TEXT",
					Input: dto.ColumnInput{Kind: dto.ColumnInputText},
				},
			},
		},
		editPopup: editPopup{
			active:      true,
			columnIndex: 0,
			input:       "john",
			cursor:      2,
		},
	}

	// Act
	popup := strings.Join(model.renderEditPopup(60), "\n")

	// Assert
	if !strings.Contains(popup, "Value: jo|hn") {
		t.Fatalf("expected caret in edit text input, got %q", popup)
	}
}

func TestRenderEditPopup_UsesCombinedSummaryRow(t *testing.T) {
	// Arrange
	model := &Model{
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
		editPopup: editPopup{
			active:      true,
			columnIndex: 0,
		},
	}

	// Act
	popup := strings.Join(model.renderEditPopup(60), "\n")

	// Assert
	if !strings.Contains(popup, "name (TEXT) | NULLABLE") {
		t.Fatalf("expected combined summary row for column metadata, got %q", popup)
	}
}

func TestRenderHelpPopup_IncludesRequiredSectionsAndOneLineDescriptions(t *testing.T) {
	// Arrange
	model := &Model{
		height:    40,
		helpPopup: helpPopup{active: true},
	}

	// Act
	popup := strings.Join(model.renderHelpPopup(60), "\n")

	// Assert
	if !strings.Contains(popup, "Supported Commands") {
		t.Fatalf("expected help popup to include Supported Commands section, got %q", popup)
	}
	if !strings.Contains(popup, "Supported Keywords") {
		t.Fatalf("expected help popup to include Supported Keywords section, got %q", popup)
	}
	if !strings.Contains(popup, ":config / :c - Open database selector and config manager.") {
		t.Fatalf("expected help popup to include :config alias one-line description, got %q", popup)
	}
	if !strings.Contains(popup, ":help / :h - Open runtime help popup reference.") {
		t.Fatalf("expected help popup to include :help alias one-line description, got %q", popup)
	}
	if !strings.Contains(popup, ":quit / :q - Quit the application.") {
		t.Fatalf("expected help popup to include :quit alias one-line description, got %q", popup)
	}
	if !strings.Contains(strings.Join(helpPopupContentLines(), "\n"), "Enter - Open selected record detail view.") {
		t.Fatalf("expected help content to include Enter one-line description, got %q", popup)
	}
	if strings.Contains(popup, "q / Ctrl+c - Quit the application.") {
		t.Fatalf("expected help popup to avoid runtime q/Ctrl+c quit shortcut, got %q", popup)
	}
	if !strings.Contains(strings.Join(helpPopupContentLines(), "\n"), "Esc - Close active popup/context.") {
		t.Fatalf("expected help popup to include Esc one-line description, got %q", popup)
	}
}

func TestRenderHelpPopup_UsesConfigPopupHeaderLayout(t *testing.T) {
	// Arrange
	model := &Model{
		height:    40,
		helpPopup: helpPopup{active: true},
	}

	// Act
	lines := model.renderHelpPopup(60)

	// Assert
	if len(lines) < 5 {
		t.Fatalf("expected help popup to include framed header and content, got %q", strings.Join(lines, "\n"))
	}

	summary := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(lines[2], "|"), "|"))
	if summary != "Use j/k, Ctrl+f/Ctrl+b to scroll. Esc closes." {
		t.Fatalf("expected config-style summary row below title, got %q", summary)
	}

	separator := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(lines[3], "|"), "|"))
	if separator == "" || strings.Trim(separator, "-") != "" {
		t.Fatalf("expected separator row after summary, got %q", separator)
	}
}

func TestRenderHelpPopup_ScrollCanReachFinalItemWhenOverflowing(t *testing.T) {
	// Arrange
	model := &Model{
		height:        12,
		helpPopup:     helpPopup{active: true},
		statusMessage: "",
	}
	initial := strings.Join(model.renderHelpPopup(60), "\n")

	// Act
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
}

func TestHandleHelpPopupKey_NonScrollKeyDoesNotChangeRenderedWindow(t *testing.T) {
	// Arrange
	model := &Model{
		height:    12,
		helpPopup: helpPopup{active: true},
	}
	for range 5 {
		model.handleHelpPopupKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	}
	before := strings.Join(model.renderHelpPopup(60), "\n")

	// Act
	model.handleHelpPopupKey(tea.KeyMsg{Type: tea.KeyEnter})
	after := strings.Join(model.renderHelpPopup(60), "\n")

	// Assert
	if before != after {
		t.Fatalf("expected non-scroll key to keep help window stable, before=%q after=%q", before, after)
	}
}

func TestView_HelpPopupRendersModalLikeConfigSelector(t *testing.T) {
	// Arrange
	model := &Model{
		width:     80,
		height:    24,
		helpPopup: helpPopup{active: true},
	}

	// Act
	view := model.View()

	// Assert
	if strings.Contains(view, "Tables") || strings.Contains(view, "Schema") || strings.Contains(view, "Records") {
		t.Fatalf("expected help modal view without background panels, got %q", view)
	}
	if !strings.Contains(view, "|Help") {
		t.Fatalf("expected help modal frame in view, got %q", view)
	}

	lines := strings.Split(view, "\n")
	helpLine := -1
	for i, line := range lines {
		if strings.Contains(line, "|Help") {
			helpLine = i
			if strings.Index(line, "|Help") == 0 {
				t.Fatalf("expected centered modal line with left padding, got %q", line)
			}
			break
		}
	}
	if helpLine <= 0 || helpLine >= len(lines)-1 {
		t.Fatalf("expected help frame to be vertically centered, line=%d total=%d", helpLine, len(lines))
	}
}

func TestView_FilterPopupRendersAsCenteredModal(t *testing.T) {
	// Arrange
	model := &Model{
		width:  80,
		height: 24,
		filterPopup: filterPopup{
			active: true,
			step:   filterInputValue,
			input:  "abc",
			cursor: 1,
		},
	}

	// Act
	view := model.View()

	// Assert
	if strings.Contains(view, "Tables") || strings.Contains(view, "Schema") || strings.Contains(view, "Records") {
		t.Fatalf("expected filter popup modal view without background panels, got %q", view)
	}
	if !strings.Contains(view, "|Filter") {
		t.Fatalf("expected filter popup frame in view, got %q", view)
	}
	lines := strings.Split(view, "\n")
	filterLine := -1
	for i, line := range lines {
		if strings.Contains(line, "|Filter") {
			filterLine = i
			if strings.Index(line, "|Filter") == 0 {
				t.Fatalf("expected centered filter popup line with left padding, got %q", line)
			}
			break
		}
	}
	if filterLine <= 0 || filterLine >= len(lines)-1 {
		t.Fatalf("expected filter popup to be vertically centered, line=%d total=%d", filterLine, len(lines))
	}
}

func TestView_EditPopupRendersAsCenteredModal(t *testing.T) {
	// Arrange
	model := &Model{
		width:  80,
		height: 24,
		schema: dto.Schema{
			Columns: []dto.SchemaColumn{
				{
					Name:  "name",
					Type:  "TEXT",
					Input: dto.ColumnInput{Kind: dto.ColumnInputText},
				},
			},
		},
		editPopup: editPopup{
			active:      true,
			columnIndex: 0,
			input:       "john",
			cursor:      2,
		},
	}

	// Act
	view := model.View()

	// Assert
	if strings.Contains(view, "Tables") || strings.Contains(view, "Schema") || strings.Contains(view, "Records") {
		t.Fatalf("expected edit popup modal view without background panels, got %q", view)
	}
	if !strings.Contains(view, "|Edit Cell") {
		t.Fatalf("expected edit popup frame in view, got %q", view)
	}
	lines := strings.Split(view, "\n")
	editLine := -1
	for i, line := range lines {
		if strings.Contains(line, "|Edit Cell") {
			editLine = i
			if strings.Index(line, "|Edit Cell") == 0 {
				t.Fatalf("expected centered edit popup line with left padding, got %q", line)
			}
			break
		}
	}
	if editLine <= 0 || editLine >= len(lines)-1 {
		t.Fatalf("expected edit popup to be vertically centered, line=%d total=%d", editLine, len(lines))
	}
}

func TestView_DirtyConfigPopupRendersAsCenteredModal(t *testing.T) {
	// Arrange
	model := &Model{
		viewMode:       ViewRecords,
		width:          80,
		height:         24,
		pendingInserts: []pendingInsertRow{{}},
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
	for _, r := range "config" {
		model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})
	view := model.View()

	// Assert
	if !strings.Contains(view, "|Config") {
		t.Fatalf("expected dirty :config modal title, got %q", view)
	}
	if strings.Contains(view, "Tables") || strings.Contains(view, "Schema") || strings.Contains(view, "Records") {
		t.Fatalf("expected centered modal without background panels, got %q", view)
	}

	lines := strings.Split(view, "\n")
	configLine := -1
	for i, line := range lines {
		if strings.Contains(line, "|Config") {
			configLine = i
			if strings.Index(line, "|Config") == 0 {
				t.Fatalf("expected centered config modal line with left padding, got %q", line)
			}
			break
		}
	}
	if configLine <= 0 || configLine >= len(lines)-1 {
		t.Fatalf("expected config modal to be vertically centered, line=%d total=%d", configLine, len(lines))
	}
}

func TestRenderConfirmPopup_DirtyConfigUsesStandardizedHeaderAndOptionsLayout(t *testing.T) {
	// Arrange
	model := &Model{
		confirmPopup: confirmPopup{
			active:  true,
			title:   "Config",
			message: "Unsaved changes detected. Choose save, discard, or cancel.",
			options: []confirmOption{
				{label: "Save and open config", action: confirmConfigSaveAndOpen},
				{label: "Discard and open config", action: confirmConfigDiscardAndOpen},
				{label: "Cancel", action: confirmConfigCancel},
			},
			selected: 0,
			modal:    true,
		},
	}

	// Act
	lines := model.renderConfirmPopup(60)
	popup := strings.Join(lines, "\n")

	// Assert
	if len(lines) < 6 {
		t.Fatalf("expected modal config popup with framed header and options, got %q", popup)
	}
	if !strings.Contains(popup, "|Config") {
		t.Fatalf("expected config title in popup, got %q", popup)
	}
	if !strings.Contains(popup, "Unsaved changes detected. Choose save, discard, or cancel.") {
		t.Fatalf("expected decision summary in popup, got %q", popup)
	}

	separator := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(lines[3], "|"), "|"))
	if separator == "" || strings.Trim(separator, "-") != "" {
		t.Fatalf("expected separator row after summary, got %q", separator)
	}
	if !strings.Contains(popup, "> Save and open config") {
		t.Fatalf("expected selected option marker in popup, got %q", popup)
	}
}

func TestRenderConfirmPopup_DirtyTableSwitchUsesInformationalMessageAndExplicitActions(t *testing.T) {
	// Arrange
	model := &Model{
		confirmPopup: confirmPopup{
			active:  true,
			title:   "Switch Table",
			message: "Switching tables will cause loss of unsaved data (3 changes). Are you sure you want to discard unsaved data?",
			options: []confirmOption{
				{label: "(y) Yes, discard changes and switch table", action: confirmDiscardTable},
				{label: "(n) No, continue editing", action: confirmCancelTableSwitch},
			},
			selected: 0,
			modal:    true,
		},
	}

	// Act
	lines := model.renderConfirmPopup(120)
	popup := strings.Join(lines, "\n")

	// Assert
	if !strings.Contains(popup, "|Switch Table") {
		t.Fatalf("expected switch-table title in popup, got %q", popup)
	}
	if !strings.Contains(popup, "Switching tables will cause loss of unsaved data") {
		t.Fatalf("expected informational switch-table summary in popup, got %q", popup)
	}
	if !strings.Contains(popup, "> (y) Yes, discard changes and switch table") {
		t.Fatalf("expected explicit yes action in popup, got %q", popup)
	}
	if !strings.Contains(popup, "(n) No, continue editing") {
		t.Fatalf("expected explicit no action in popup, got %q", popup)
	}
}

func TestView_RegularConfirmPopupRendersAsCenteredModal(t *testing.T) {
	// Arrange
	model := &Model{
		width:  80,
		height: 24,
		confirmPopup: confirmPopup{
			active:  true,
			title:   "Confirm",
			message: "Save staged changes?",
		},
	}

	// Act
	view := model.View()

	// Assert
	if strings.Contains(view, "Tables") || strings.Contains(view, "Schema") || strings.Contains(view, "Records") {
		t.Fatalf("expected non-modal confirm as centered popup without background panels, got %q", view)
	}
	if !strings.Contains(view, "|Confirm") {
		t.Fatalf("expected confirm popup frame, got %q", view)
	}
	lines := strings.Split(view, "\n")
	confirmLine := -1
	for i, line := range lines {
		if strings.Contains(line, "|Confirm") {
			confirmLine = i
			if strings.Index(line, "|Confirm") == 0 {
				t.Fatalf("expected centered confirm popup line with left padding, got %q", line)
			}
			break
		}
	}
	if confirmLine <= 0 || confirmLine >= len(lines)-1 {
		t.Fatalf("expected confirm popup to be vertically centered, line=%d total=%d", confirmLine, len(lines))
	}
}

func TestRenderConfirmPopup_InlineUsesStandardizedSeparatorRow(t *testing.T) {
	// Arrange
	model := &Model{
		confirmPopup: confirmPopup{
			active:  true,
			title:   "Confirm",
			message: "Save staged changes?",
		},
	}

	// Act
	lines := model.renderConfirmPopup(60)

	// Assert
	if len(lines) < 5 {
		t.Fatalf("expected standardized confirm popup layout, got %q", strings.Join(lines, "\n"))
	}

	separator := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(lines[3], "|"), "|"))
	if separator == "" || strings.Trim(separator, "-") != "" {
		t.Fatalf("expected separator row after summary, got %q", separator)
	}
}

func TestPanelWidths_UsesLongestTableNameAsMaxWidthInWideWindow(t *testing.T) {
	// Arrange
	longName := "table_name_for_dynamic_left_panel_width"
	model := &Model{
		width: 180,
		tables: []dto.Table{
			{Name: "users"},
			{Name: longName},
		},
	}

	// Act
	leftWidth, rightWidth := model.panelWidths()

	// Assert
	const (
		tablePrefixWidth = 2
		nameMargin       = 1
		separatorWidth   = 3
	)
	expectedLeftWidth := tablePrefixWidth + textWidth(longName) + nameMargin
	if leftWidth != expectedLeftWidth {
		t.Fatalf("expected left panel width %d, got %d", expectedLeftWidth, leftWidth)
	}
	if rightWidth != model.width-leftWidth-separatorWidth {
		t.Fatalf("expected right panel width %d, got %d", model.width-leftWidth-separatorWidth, rightWidth)
	}
}

func TestRenderTables_DoesNotTruncateLongestNameAtComputedMaxWidth(t *testing.T) {
	// Arrange
	longName := "table_name_for_dynamic_left_panel_width"
	model := &Model{
		width: 180,
		focus: FocusTables,
		tables: []dto.Table{
			{Name: longName},
		},
	}

	leftWidth, _ := model.panelWidths()

	// Act
	lines := model.renderTables(leftWidth, 4)

	// Assert
	if !strings.Contains(lines[1], longName) {
		t.Fatalf("expected full table name in rendered line, got %q", lines[1])
	}
	if strings.Contains(lines[1], "...") {
		t.Fatalf("expected no truncation in rendered line, got %q", lines[1])
	}
}

func TestRenderTables_BoldsSelectedTableWithoutTableFocus(t *testing.T) {
	// Arrange
	model := &Model{
		width:         80,
		focus:         FocusContent,
		selectedTable: 1,
		tables: []dto.Table{
			{Name: "users"},
			{Name: "orders"},
		},
	}

	// Act
	lines := model.renderTables(20, 4)

	// Assert
	var selectedLine string
	for _, line := range lines {
		if strings.Contains(line, "orders") {
			selectedLine = line
			break
		}
	}
	if selectedLine == "" {
		t.Fatalf("expected selected table line to be rendered, got %q", strings.Join(lines, "\n"))
	}
	if !strings.Contains(selectedLine, "\x1b[1m") || !strings.Contains(selectedLine, "\x1b[0m") {
		t.Fatalf("expected selected table to be bold, got %q", selectedLine)
	}
	if strings.Contains(selectedLine, "> ") {
		t.Fatalf("expected no focus marker when tables panel is not focused, got %q", selectedLine)
	}
}

func TestPanelWidths_PreservesMinimumRightPanelWidthInNarrowWindow(t *testing.T) {
	// Arrange
	model := &Model{
		width: 30,
		tables: []dto.Table{
			{Name: "table_name_for_dynamic_left_panel_width"},
		},
	}

	// Act
	leftWidth, rightWidth := model.panelWidths()

	// Assert
	if rightWidth != 10 {
		t.Fatalf("expected minimum right panel width 10, got %d", rightWidth)
	}
	if leftWidth != 17 {
		t.Fatalf("expected adjusted left panel width 17, got %d", leftWidth)
	}
	if leftWidth+rightWidth+3 != model.width {
		t.Fatalf("expected panel widths to match available width, got left=%d right=%d total=%d", leftWidth, rightWidth, model.width)
	}
}
