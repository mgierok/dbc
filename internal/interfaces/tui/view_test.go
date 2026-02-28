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

func TestView_RuntimeRendersIndependentSectionBoxes(t *testing.T) {
	// Arrange
	model := &Model{
		width:    80,
		height:   24,
		focus:    FocusTables,
		viewMode: ViewSchema,
		tables: []dto.Table{
			{Name: "users"},
		},
		schema: dto.Schema{
			Columns: []dto.SchemaColumn{
				{Name: "id", Type: "INTEGER"},
			},
		},
	}

	// Act
	view := model.View()
	lines := strings.Split(view, "\n")

	// Assert
	if len(lines) != model.height {
		t.Fatalf("expected runtime view height %d, got %d", model.height, len(lines))
	}
	if len(lines) < 4 {
		t.Fatalf("expected at least 4 lines in runtime view, got %d", len(lines))
	}
	topLine := lines[0]
	if !strings.Contains(topLine, frameTopLeft+"Tables") || !strings.Contains(topLine, frameTopLeft+"Schema") {
		t.Fatalf("expected titled top borders for both panels, got %q", topLine)
	}
	if strings.Contains(topLine, frameJoinCenter) || strings.Contains(topLine, frameJoinTop) || strings.Contains(topLine, frameJoinBottom) {
		t.Fatalf("expected independent panel boxes without shared joins, got %q", topLine)
	}

	statusTop := lines[len(lines)-3]
	statusContent := lines[len(lines)-2]
	statusBottom := lines[len(lines)-1]

	if !strings.HasPrefix(statusTop, frameTopLeft) || !strings.HasSuffix(statusTop, frameTopRight) {
		t.Fatalf("expected framed status top border, got %q", statusTop)
	}
	if !strings.HasPrefix(statusContent, frameVertical+" ") || !strings.HasSuffix(statusContent, " "+frameVertical) {
		t.Fatalf("expected status content with one-space side padding, got %q", statusContent)
	}
	if !strings.HasPrefix(statusBottom, frameBottomLeft) || !strings.HasSuffix(statusBottom, frameBottomRight) {
		t.Fatalf("expected framed status bottom border, got %q", statusBottom)
	}
	for _, line := range []string{statusTop, statusContent, statusBottom} {
		if textWidth(line) != model.width {
			t.Fatalf("expected status row width %d, got %d for %q", model.width, textWidth(line), line)
		}
	}
}

func TestView_RuntimeRightPanelTopBorderUsesDynamicTitle(t *testing.T) {
	testCases := []struct {
		name     string
		model    Model
		expected string
	}{
		{
			name: "schema view",
			model: Model{
				width:    80,
				height:   24,
				viewMode: ViewSchema,
				tables:   []dto.Table{{Name: "users"}},
				schema: dto.Schema{
					Columns: []dto.SchemaColumn{{Name: "id", Type: "INTEGER"}},
				},
			},
			expected: frameTopLeft + "Schema",
		},
		{
			name: "records view",
			model: Model{
				width:    80,
				height:   24,
				viewMode: ViewRecords,
				tables:   []dto.Table{{Name: "users"}},
				schema: dto.Schema{
					Columns: []dto.SchemaColumn{{Name: "id", Type: "INTEGER"}},
				},
				records: []dto.RecordRow{{Values: []string{"1"}}},
			},
			expected: frameTopLeft + "Records",
		},
		{
			name: "record detail view",
			model: Model{
				width:    80,
				height:   24,
				viewMode: ViewRecords,
				tables:   []dto.Table{{Name: "users"}},
				recordDetail: recordDetailState{
					active: true,
				},
				schema: dto.Schema{
					Columns: []dto.SchemaColumn{{Name: "id", Type: "INTEGER"}},
				},
				records: []dto.RecordRow{{Values: []string{"1"}}},
			},
			expected: frameTopLeft + "Record Detail",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			view := tc.model.View()
			topLine := strings.Split(view, "\n")[0]
			if !strings.Contains(topLine, tc.expected) {
				t.Fatalf("expected top line to contain %q, got %q", tc.expected, topLine)
			}
		})
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
	if len(lines) < 3 {
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
	if len(lines) < 3 {
		t.Fatalf("expected header row, got %v", lines)
	}
	header := lines[1]
	if !strings.Contains(header, "name "+iconSortDesc) {
		t.Fatalf("expected desc sort indicator in header, got %q", header)
	}
}

func TestBoxWidthForRecordHeaderColumn_UsesFullColumnWidth(t *testing.T) {
	// Arrange
	testCases := []struct {
		name      string
		width     int
		wantWidth int
	}{
		{name: "odd width", width: 25, wantWidth: 25},
		{name: "even width", width: 24, wantWidth: 24},
		{name: "small width", width: 5, wantWidth: 5},
		{name: "minimum width", width: 1, wantWidth: 1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			gotWidth := boxWidthForRecordHeaderColumn(tc.width)

			// Assert
			if gotWidth != tc.wantWidth {
				t.Fatalf("expected box width %d, got %d", tc.wantWidth, gotWidth)
			}
		})
	}
}

func TestFormatRecordsHeaderRows_UsesFullWidthBoxWithCenteredLabel(t *testing.T) {
	// Arrange
	label := "name " + iconSortAsc

	// Act
	rows := formatRecordsHeaderRows([]string{label}, []int{20})

	// Assert
	if len(rows) != 3 {
		t.Fatalf("expected 3 header rows, got %d", len(rows))
	}
	for _, row := range rows {
		if textWidth(row) != 20 {
			t.Fatalf("expected header row width 20, got %d for row %q", textWidth(row), row)
		}
	}

	middleRow := rows[1]
	leftBorder := strings.Index(middleRow, frameVertical)
	rightBorder := strings.LastIndex(middleRow, frameVertical)
	if leftBorder < 0 || rightBorder <= leftBorder {
		t.Fatalf("expected framed middle row, got %q", middleRow)
	}
	if leftBorder != 0 {
		t.Fatalf("expected left border at column start, got index %d in %q", leftBorder, middleRow)
	}
	if textWidth(middleRow[:rightBorder+len(frameVertical)]) != 20 {
		t.Fatalf("expected right border at column end, got %q", middleRow)
	}
	inside := middleRow[leftBorder+len(frameVertical) : rightBorder]
	if strings.TrimSpace(inside) != label {
		t.Fatalf("expected centered label %q, got %q", label, inside)
	}

	leftPadding := len(inside) - len(strings.TrimLeft(inside, " "))
	rightPadding := len(inside) - len(strings.TrimRight(inside, " "))
	paddingDiff := leftPadding - rightPadding
	if paddingDiff < -1 || paddingDiff > 1 {
		t.Fatalf("expected balanced centering paddings, got left=%d right=%d in %q", leftPadding, rightPadding, inside)
	}
}

func TestRenderRecords_RendersThreeLineFrameHeader(t *testing.T) {
	// Arrange
	model := &Model{
		viewMode: ViewRecords,
		schema: dto.Schema{
			Columns: []dto.SchemaColumn{
				{Name: "id", Type: "INTEGER"},
				{Name: "name", Type: "TEXT"},
			},
		},
		records: []dto.RecordRow{
			{Values: []string{"1", "alice"}},
		},
	}

	// Act
	lines := model.renderRecords(80, 8)

	// Assert
	if len(lines) < 3 {
		t.Fatalf("expected 3-line header, got %v", lines)
	}
	if !strings.Contains(lines[0], frameTopLeft) || !strings.Contains(lines[0], frameTopRight) {
		t.Fatalf("expected top header frame row, got %q", lines[0])
	}
	if !strings.Contains(lines[1], frameVertical) {
		t.Fatalf("expected middle header frame row, got %q", lines[1])
	}
	if !strings.Contains(lines[2], frameBottomLeft) || !strings.Contains(lines[2], frameBottomRight) {
		t.Fatalf("expected bottom header frame row, got %q", lines[2])
	}
}

func TestFormatRecordRow_UsesDoubleSpaceSeparatorBetweenColumns(t *testing.T) {
	// Arrange
	values := []string{"A", "B"}
	widths := []int{1, 1}

	// Act
	row := formatRecordRow(values, widths, -1)

	// Assert
	if row != "A  B" {
		t.Fatalf("expected two-space separator between columns, got %q", row)
	}
}

func TestFormatRecordsHeaderRows_UsesDoubleSpaceSeparatorBetweenColumns(t *testing.T) {
	// Arrange
	values := []string{"id", "name"}
	widths := []int{6, 6}

	// Act
	rows := formatRecordsHeaderRows(values, widths)

	// Assert
	if len(rows) != 3 {
		t.Fatalf("expected 3 header rows, got %d", len(rows))
	}
	for _, row := range rows {
		if !strings.Contains(row, "  ") {
			t.Fatalf("expected two-space separator in header row, got %q", row)
		}
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
	if !strings.Contains(content, iconSelection+" "+iconEdit+" ") {
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
	if len(lines) < 6 {
		t.Fatalf("expected records output with header and rows, got %v", lines)
	}
	secondColumnStarts := make([]int, 0, 3)
	rowLines := lines[3:6]
	secondColumnTokens := []string{"inserted", "alice2", "bob"}
	for i, line := range rowLines {
		colStartByteIndex := strings.Index(line, secondColumnTokens[i])
		if colStartByteIndex < 0 {
			t.Fatalf("expected second column token %q in row line, got %q", secondColumnTokens[i], line)
		}
		colStartDisplayWidth := textWidth(line[:colStartByteIndex])
		secondColumnStarts = append(secondColumnStarts, colStartDisplayWidth)
	}
	for i := 1; i < len(secondColumnStarts); i++ {
		if secondColumnStarts[i] != secondColumnStarts[0] {
			t.Fatalf(
				"expected aligned second-column start positions, got %v in lines %q",
				secondColumnStarts,
				rowLines,
			)
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
	if !strings.Contains(popup, "name (TEXT)"+frameSegmentSeparator+"NULLABLE") {
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

	if !strings.HasPrefix(lines[0], frameTopLeft+"Context Help: Tables") {
		t.Fatalf("expected context-specific help title in top border, got %q", lines[0])
	}

	summary := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(lines[1], frameVertical), frameVertical))
	if summary != "Use j/k, Ctrl+f/Ctrl+b to scroll. Esc closes." {
		t.Fatalf("expected config-style summary row below title, got %q", summary)
	}

	separator := strings.TrimSpace(lines[2])
	if !strings.HasPrefix(separator, frameJoinLeft) || !strings.HasSuffix(separator, frameJoinRight) {
		t.Fatalf("expected separator row with border joins, got %q", separator)
	}
	separatorInner := strings.TrimSuffix(strings.TrimPrefix(separator, frameJoinLeft), frameJoinRight)
	if separatorInner == "" || strings.Trim(separatorInner, frameHorizontal) != "" {
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
	if strings.Contains(view, iconSelection+" Tables") || strings.Contains(view, iconSelection+" Schema") || strings.Contains(view, iconSelection+" Records") {
		t.Fatalf("expected help modal view without background panels, got %q", view)
	}
	if !strings.Contains(view, frameTopLeft+"Context Help: Tables") {
		t.Fatalf("expected help modal frame in view, got %q", view)
	}

	lines := strings.Split(view, "\n")
	helpLine := -1
	for i, line := range lines {
		if strings.Contains(line, frameTopLeft+"Context Help: Tables") {
			helpLine = i
			if strings.Index(line, frameTopLeft+"Context Help: Tables") == 0 {
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
	if !strings.Contains(view, frameTopLeft+"Filter") {
		t.Fatalf("expected filter popup frame in view, got %q", view)
	}
	lines := strings.Split(view, "\n")
	filterLine := -1
	for i, line := range lines {
		if strings.Contains(line, frameTopLeft+"Filter") {
			filterLine = i
			if strings.Index(line, frameTopLeft+"Filter") == 0 {
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
	if !strings.Contains(view, frameTopLeft+"Edit Cell") {
		t.Fatalf("expected edit popup frame in view, got %q", view)
	}
	lines := strings.Split(view, "\n")
	editLine := -1
	for i, line := range lines {
		if strings.Contains(line, frameTopLeft+"Edit Cell") {
			editLine = i
			if strings.Index(line, frameTopLeft+"Edit Cell") == 0 {
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
	if !strings.Contains(view, frameTopLeft+"Config") {
		t.Fatalf("expected dirty :config modal title, got %q", view)
	}
	if strings.Contains(view, "Tables") || strings.Contains(view, "Schema") || strings.Contains(view, "Records") {
		t.Fatalf("expected centered modal without background panels, got %q", view)
	}

	lines := strings.Split(view, "\n")
	configLine := -1
	for i, line := range lines {
		if strings.Contains(line, frameTopLeft+"Config") {
			configLine = i
			if strings.Index(line, frameTopLeft+"Config") == 0 {
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
	if !strings.HasPrefix(lines[0], frameTopLeft+"Config") {
		t.Fatalf("expected config title in top border, got %q", lines[0])
	}
	if !strings.Contains(popup, "Unsaved changes detected.") {
		t.Fatalf("expected decision summary in popup, got %q", popup)
	}

	separator := strings.TrimSpace(lines[2])
	if !strings.HasPrefix(separator, frameJoinLeft) || !strings.HasSuffix(separator, frameJoinRight) {
		t.Fatalf("expected separator row with border joins, got %q", separator)
	}
	separatorInner := strings.TrimSuffix(strings.TrimPrefix(separator, frameJoinLeft), frameJoinRight)
	if separatorInner == "" || strings.Trim(separatorInner, frameHorizontal) != "" {
		t.Fatalf("expected separator row after summary, got %q", separator)
	}
	if !strings.Contains(popup, iconSelection+" Save and open config") {
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
	if !strings.Contains(popup, frameTopLeft+"Switch Table") {
		t.Fatalf("expected switch-table title in popup, got %q", popup)
	}
	if !strings.Contains(popup, "Switching tables will cause loss of unsaved data") {
		t.Fatalf("expected informational switch-table summary in popup, got %q", popup)
	}
	if !strings.Contains(popup, iconSelection+" (y) Yes, discard changes and switch table") {
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
	if !strings.Contains(view, frameTopLeft+"Confirm") {
		t.Fatalf("expected confirm popup frame, got %q", view)
	}
	lines := strings.Split(view, "\n")
	confirmLine := -1
	for i, line := range lines {
		if strings.Contains(line, frameTopLeft+"Confirm") {
			confirmLine = i
			if strings.Index(line, frameTopLeft+"Confirm") == 0 {
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

	separator := strings.TrimSpace(lines[2])
	if !strings.HasPrefix(separator, frameJoinLeft) || !strings.HasSuffix(separator, frameJoinRight) {
		t.Fatalf("expected separator row with border joins, got %q", separator)
	}
	separatorInner := strings.TrimSuffix(strings.TrimPrefix(separator, frameJoinLeft), frameJoinRight)
	if separatorInner == "" || strings.Trim(separatorInner, frameHorizontal) != "" {
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
	nonContentWidth := (panelBoxBorderWidth * 2) + panelBoxGapWidth
	expectedLeftWidth := tablePrefixWidth + textWidth(longName) + nameMargin
	if leftWidth != expectedLeftWidth {
		t.Fatalf("expected left panel width %d, got %d", expectedLeftWidth, leftWidth)
	}
	if rightWidth != model.width-leftWidth-nonContentWidth {
		t.Fatalf("expected right panel width %d, got %d", model.width-leftWidth-nonContentWidth, rightWidth)
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
	if !strings.Contains(lines[0], longName) {
		t.Fatalf("expected full table name in rendered line, got %q", lines[0])
	}
	if strings.Contains(lines[0], "...") {
		t.Fatalf("expected no truncation in rendered line, got %q", lines[0])
	}
}

func TestRenderTables_ShowsSelectionMarkerWithoutBoldWhenTablePanelIsNotFocused(t *testing.T) {
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
	if !strings.Contains(selectedLine, iconSelection+" ") {
		t.Fatalf("expected selection marker for selected table, got %q", selectedLine)
	}
	if strings.Contains(selectedLine, "\x1b[1m") || strings.Contains(selectedLine, "\x1b[0m") {
		t.Fatalf("expected selected table without bold formatting, got %q", selectedLine)
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
	if leftWidth != 16 {
		t.Fatalf("expected adjusted left panel width 16, got %d", leftWidth)
	}
	if leftWidth+rightWidth+(panelBoxBorderWidth*2)+panelBoxGapWidth != model.width {
		t.Fatalf("expected panel widths to match available width, got left=%d right=%d total=%d", leftWidth, rightWidth, model.width)
	}
}
