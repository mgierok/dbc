package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/dto"
	domainmodel "github.com/mgierok/dbc/internal/domain/model"
)

func TestRenderStatus_ShowsContextHelpHintOnRight(t *testing.T) {
	// Arrange
	model := &Model{
		focus:    FocusContent,
		viewMode: ViewRecords,
	}

	// Act
	status := model.renderStatus(120)

	// Assert
	if !strings.HasSuffix(status, "Context help: ?") {
		t.Fatalf("expected right-aligned context-help hint suffix, got %q", status)
	}
}

func TestRenderStatus_DoesNotIncludeContextShortcutList(t *testing.T) {
	// Arrange
	model := &Model{
		focus:    FocusContent,
		viewMode: ViewRecords,
	}

	// Act
	status := model.renderStatus(200)

	// Assert
	if strings.Contains(status, "Records: Esc tables | e edit | Enter detail") {
		t.Fatalf("expected context shortcut list to be removed from status line, got %q", status)
	}
}

func TestRenderStatus_RightHintPriorityOnNarrowWidth(t *testing.T) {
	// Arrange
	model := &Model{
		focus:    FocusContent,
		viewMode: ViewRecords,
	}

	// Act
	status := model.renderStatus(20)

	// Assert
	if !strings.HasSuffix(status, "Context help: ?") {
		t.Fatalf("expected narrow status line to keep full right hint, got %q", status)
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

func TestRenderStatus_RecordsViewShowsTotalAndPagination(t *testing.T) {
	// Arrange
	model := &Model{
		viewMode:         ViewRecords,
		recordPageIndex:  0,
		recordTotalPages: 7,
		recordTotalCount: 137,
		records:          make([]dto.RecordRow, 20),
	}

	// Act
	status := model.renderStatus(220)

	// Assert
	if !strings.Contains(status, "Records: 20/137") {
		t.Fatalf("expected records total summary in status, got %q", status)
	}
	if !strings.Contains(status, "Page: 1/7") {
		t.Fatalf("expected pagination summary in status, got %q", status)
	}
}

func TestRenderStatus_RecordsViewShowsSinglePageSummaryForEmptyResult(t *testing.T) {
	// Arrange
	model := &Model{
		viewMode:         ViewRecords,
		recordPageIndex:  0,
		recordTotalPages: 1,
		recordTotalCount: 0,
	}

	// Act
	status := model.renderStatus(220)

	// Assert
	if !strings.Contains(status, "Records: 0/0") {
		t.Fatalf("expected empty records summary in status, got %q", status)
	}
	if !strings.Contains(status, "Page: 1/1") {
		t.Fatalf("expected single-page summary in status, got %q", status)
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
	if !strings.Contains(header, "name "+iconSortAsc) {
		t.Fatalf("expected asc sort indicator in header, got %q", header)
	}
	if strings.Contains(header, "id "+iconSortAsc) || strings.Contains(header, "id "+iconSortDesc) {
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
	if !strings.Contains(header, "name "+iconSortDesc) {
		t.Fatalf("expected desc sort indicator in header, got %q", header)
	}
}

func TestRenderRecords_UsesInsertAndDeleteIconsInRowPrefix(t *testing.T) {
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
		pendingInserts: []pendingInsertRow{
			{
				values: map[int]stagedEdit{
					0: {Value: domainmodel.Value{Text: "2", Raw: "2"}},
					1: {Value: domainmodel.Value{Text: "new", Raw: "new"}},
				},
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
	model.pendingDeletes = map[string]recordDelete{
		key: {},
	}

	// Act
	content := strings.Join(model.renderRecords(80, 8), "\n")

	// Assert
	if !strings.Contains(content, iconInsert+" ") {
		t.Fatalf("expected insert icon row prefix, got %q", content)
	}
	if !strings.Contains(content, iconDelete+" ") {
		t.Fatalf("expected delete icon row prefix, got %q", content)
	}
}

func TestRenderRecords_UsesEditIconForEditedRows(t *testing.T) {
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
	content := strings.Join(model.renderRecords(80, 6), "\n")

	// Assert
	if !strings.Contains(content, "> "+iconEdit+" ") {
		t.Fatalf("expected edited row icon marker in row prefix, got %q", content)
	}
	if strings.Contains(content, "alice2"+iconEdit) {
		t.Fatalf("expected no edited icon marker inside cell value, got %q", content)
	}
}

func TestRenderRecords_PreservesColumnAlignmentWithMixedRowMarkers(t *testing.T) {
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
		pendingInserts: []pendingInsertRow{
			{
				values: map[int]stagedEdit{
					0: {Value: domainmodel.Value{Text: "10", Raw: "10"}},
					1: {Value: domainmodel.Value{Text: "inserted", Raw: "inserted"}},
				},
			},
		},
		records: []dto.RecordRow{
			{Values: []string{"1", "alice"}},
			{Values: []string{"2", "bob"}},
		},
	}

	editedKey, ok := model.recordKeyForPersistedRow(0)
	if !ok {
		t.Fatal("expected edited row key")
	}
	model.pendingUpdates = map[string]recordEdits{
		editedKey: {
			changes: map[int]stagedEdit{
				1: {Value: domainmodel.Value{Text: "alice2", Raw: "alice2"}},
			},
		},
	}

	deleteKey, ok := model.recordKeyForPersistedRow(1)
	if !ok {
		t.Fatal("expected delete row key")
	}
	model.pendingDeletes = map[string]recordDelete{
		deleteKey: {},
	}

	// Act
	lines := model.renderRecords(90, 8)

	// Assert
	if len(lines) < 5 {
		t.Fatalf("expected records output with header and rows, got %v", lines)
	}
	separatorColumns := make([]int, 0, 3)
	for _, line := range lines[2:5] {
		sep := strings.Index(line, dividerColumn)
		if sep < 0 {
			t.Fatalf("expected column separator in row line, got %q", line)
		}
		separatorColumns = append(separatorColumns, sep)
	}
	for i := 1; i < len(separatorColumns); i++ {
		if separatorColumns[i] != separatorColumns[0] {
			t.Fatalf("expected aligned column separators, got %v in lines %q", separatorColumns, lines[2:5])
		}
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
	if !strings.Contains(content, iconInfo+" Persisted record") {
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
		if !strings.Contains(lines[0], iconInfo+" Persisted record") {
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
		if !strings.Contains(lines[0], iconInfo+" Pending insert") {
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
		if !strings.Contains(lines[0], iconInfo+" Marked for delete") {
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
		if !strings.Contains(lines[0], iconInfo+" Edited record") {
			t.Fatalf("expected edited information marker, got %q", lines[0])
		}
		content := strings.Join(lines, "\n")
		if !strings.Contains(content, "\x1b[1mname\x1b[0m (TEXT) "+iconEdit) {
			t.Fatalf("expected edit icon on modified field header in detail content, got %q", content)
		}
		if strings.Contains(content, "\x1b[1mid\x1b[0m (INTEGER) "+iconEdit) {
			t.Fatalf("expected no edit icon on unmodified field header in detail content, got %q", content)
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
	if !strings.Contains(popup, "name (TEXT)"+segmentSeparator+"NULLABLE") {
		t.Fatalf("expected combined summary row for column metadata, got %q", popup)
	}
}

func TestRenderHelpPopup_ShowsOnlyCurrentContextBindings(t *testing.T) {
	// Arrange
	model := &Model{
		height: 40,
		helpPopup: helpPopup{
			active:  true,
			context: helpPopupContextRecords,
		},
	}

	// Act
	popup := strings.Join(model.renderHelpPopup(60), "\n")

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

func TestRenderHelpPopup_UsesConfigPopupHeaderLayout(t *testing.T) {
	// Arrange
	model := &Model{
		height: 40,
		helpPopup: helpPopup{
			active:  true,
			context: helpPopupContextTables,
		},
	}

	// Act
	lines := model.renderHelpPopup(60)

	// Assert
	if len(lines) < 5 {
		t.Fatalf("expected help popup to include framed header and content, got %q", strings.Join(lines, "\n"))
	}

	title := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(lines[1], "|"), "|"))
	if title != "Context Help: Tables" {
		t.Fatalf("expected context-specific help title, got %q", title)
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
		helpPopup:     helpPopup{active: true, context: helpPopupContextRecords},
		statusMessage: "",
	}
	initial := strings.Join(model.renderHelpPopup(60), "\n")

	// Act
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
}

func TestHandleHelpPopupKey_NonScrollKeyDoesNotChangeRenderedWindow(t *testing.T) {
	// Arrange
	model := &Model{
		height:    12,
		helpPopup: helpPopup{active: true, context: helpPopupContextRecords},
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
		helpPopup: helpPopup{active: true, context: helpPopupContextTables},
	}

	// Act
	view := model.View()

	// Assert
	if strings.Contains(view, activePanelTitle("Tables")) || strings.Contains(view, activePanelTitle("Schema")) || strings.Contains(view, activePanelTitle("Records")) {
		t.Fatalf("expected help modal view without background panels, got %q", view)
	}
	if !strings.Contains(view, "|Context Help: Tables") {
		t.Fatalf("expected help modal frame in view, got %q", view)
	}

	lines := strings.Split(view, "\n")
	helpLine := -1
	for i, line := range lines {
		if strings.Contains(line, "|Context Help: Tables") {
			helpLine = i
			if strings.Index(line, "|Context Help: Tables") == 0 {
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
	)
	separatorWidth := textWidth(dividerColumn)
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
	if rightWidth != 11 {
		t.Fatalf("expected minimum right panel width 11, got %d", rightWidth)
	}
	if leftWidth != 18 {
		t.Fatalf("expected adjusted left panel width 18, got %d", leftWidth)
	}
	if leftWidth+rightWidth+textWidth(dividerColumn) != model.width {
		t.Fatalf("expected panel widths to match available width, got left=%d right=%d total=%d", leftWidth, rightWidth, model.width)
	}
}
