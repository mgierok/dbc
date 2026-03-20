package tui

import (
	"strings"
	"testing"

	"github.com/mgierok/dbc/internal/application/dto"
	"github.com/mgierok/dbc/internal/interfaces/tui/internal/primitives"
)

func TestRenderRecords_ShowsAscSortIndicatorInHeader(t *testing.T) {
	// Arrange
	model := &Model{
		styles: primitives.NewRenderStyles(true),
		read: runtimeReadState{
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
		},
	}

	// Act
	lines := model.renderRecords(80, 6)

	// Assert
	if len(lines) < 3 {
		t.Fatalf("expected header row, got %v", lines)
	}
	header := lines[1]
	if !strings.Contains(header, "name "+primitives.IconSortAsc) {
		t.Fatalf("expected asc sort indicator in header, got %q", header)
	}
	if strings.Contains(header, "id "+primitives.IconSortAsc) || strings.Contains(header, "id "+primitives.IconSortDesc) {
		t.Fatalf("expected indicator only on sorted column, got %q", header)
	}
}

func TestRenderRecords_ShowsDescSortIndicatorInHeader(t *testing.T) {
	// Arrange
	model := &Model{
		read: runtimeReadState{
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
		},
	}

	// Act
	lines := model.renderRecords(80, 6)

	// Assert
	if len(lines) < 3 {
		t.Fatalf("expected header row, got %v", lines)
	}
	header := lines[1]
	if !strings.Contains(header, "name "+primitives.IconSortDesc) {
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
	label := "name " + primitives.IconSortAsc

	// Act
	rows := formatRecordsHeaderRows([]string{label}, []int{20}, primitives.NewRenderStyles(false))

	// Assert
	if len(rows) != 3 {
		t.Fatalf("expected 3 header rows, got %d", len(rows))
	}
	for _, row := range rows {
		if primitives.TextWidth(row) != 20 {
			t.Fatalf("expected header row width 20, got %d for row %q", primitives.TextWidth(row), row)
		}
	}

	if !strings.Contains(rows[0], primitives.FrameTopLeft) || !strings.Contains(rows[0], primitives.FrameTopRight) {
		t.Fatalf("expected framed top row, got %q", rows[0])
	}
	if !strings.Contains(rows[1], primitives.FrameVertical) || !strings.Contains(rows[1], label) {
		t.Fatalf("expected framed middle row with label %q, got %q", label, rows[1])
	}
	if !strings.Contains(rows[2], primitives.FrameBottomLeft) || !strings.Contains(rows[2], primitives.FrameBottomRight) {
		t.Fatalf("expected framed bottom row, got %q", rows[2])
	}
}

func TestRenderRecords_RendersThreeLineFrameHeader(t *testing.T) {
	// Arrange
	model := &Model{
		read: runtimeReadState{
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
		},
	}

	// Act
	lines := model.renderRecords(80, 8)

	// Assert
	if len(lines) < 3 {
		t.Fatalf("expected 3-line header, got %v", lines)
	}
	if !strings.Contains(lines[0], primitives.FrameTopLeft) || !strings.Contains(lines[0], primitives.FrameTopRight) {
		t.Fatalf("expected top header frame row, got %q", lines[0])
	}
	if !strings.Contains(lines[1], primitives.FrameVertical) {
		t.Fatalf("expected middle header frame row, got %q", lines[1])
	}
	if !strings.Contains(lines[2], primitives.FrameBottomLeft) || !strings.Contains(lines[2], primitives.FrameBottomRight) {
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
	tokens := strings.Fields(row)
	if len(tokens) != 2 || tokens[0] != "A" || tokens[1] != "B" {
		t.Fatalf("expected row to preserve column tokens, got %q", row)
	}
	if !strings.Contains(row, "  ") {
		t.Fatalf("expected row to keep visible spacing between columns, got %q", row)
	}
}

func TestFormatRecordsHeaderRows_UsesDoubleSpaceSeparatorBetweenColumns(t *testing.T) {
	// Arrange
	values := []string{"id", "name"}
	widths := []int{6, 6}

	// Act
	rows := formatRecordsHeaderRows(values, widths, primitives.NewRenderStyles(false))

	// Assert
	if len(rows) != 3 {
		t.Fatalf("expected 3 header rows, got %d", len(rows))
	}
	if !strings.Contains(rows[1], "id") || !strings.Contains(rows[1], "name") {
		t.Fatalf("expected header row to contain column labels, got %q", rows[1])
	}
	if !strings.Contains(rows[1], "  ") {
		t.Fatalf("expected visible spacing between header columns, got %q", rows[1])
	}
}

func TestRenderRecords_UsesInsertAndDeleteIconsInRowPrefix(t *testing.T) {
	// Arrange
	model := &Model{
		read: runtimeReadState{
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
		},
		staging: stagingState{
			pendingInserts: []pendingInsertRow{
				{
					values: map[int]stagedEdit{
						0: {Value: dto.StagedValue{Text: "2", Raw: "2"}},
						1: {Value: dto.StagedValue{Text: "new", Raw: "new"}},
					},
				},
			},
		},
	}
	key, ok := model.recordKeyForPersistedRow(0)
	if !ok {
		t.Fatal("expected persisted row key")
	}
	model.staging.pendingDeletes = map[string]recordDelete{
		key: {},
	}

	// Act
	content := strings.Join(model.renderRecords(80, 8), "\n")

	// Assert
	if !strings.Contains(content, primitives.IconInsert+" ") {
		t.Fatalf("expected insert icon row prefix, got %q", content)
	}
	if !strings.Contains(content, primitives.IconDelete+" ") {
		t.Fatalf("expected delete icon row prefix, got %q", content)
	}
}

func TestRenderRecords_UsesEditIconForEditedRows(t *testing.T) {
	// Arrange
	model := &Model{
		read: runtimeReadState{
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
		},
	}
	key, ok := model.recordKeyForPersistedRow(0)
	if !ok {
		t.Fatal("expected persisted row key")
	}
	model.staging.pendingUpdates = map[string]recordEdits{
		key: {
			changes: map[int]stagedEdit{
				1: {Value: dto.StagedValue{Text: "alice2", Raw: "alice2"}},
			},
		},
	}

	// Act
	content := strings.Join(model.renderRecords(80, 6), "\n")

	// Assert
	if !strings.Contains(content, primitives.IconSelection+" "+primitives.IconEdit+" ") {
		t.Fatalf("expected edited row icon marker in row prefix, got %q", content)
	}
	if strings.Contains(content, "alice2"+primitives.IconEdit) {
		t.Fatalf("expected no edited icon marker inside cell value, got %q", content)
	}
}

func TestRenderRecords_PreservesColumnAlignmentWithMixedRowMarkers(t *testing.T) {
	// Arrange
	model := &Model{
		read: runtimeReadState{
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
				{Values: []string{"2", "bob"}},
			},
		},
		staging: stagingState{
			pendingInserts: []pendingInsertRow{
				{
					values: map[int]stagedEdit{
						0: {Value: dto.StagedValue{Text: "10", Raw: "10"}},
						1: {Value: dto.StagedValue{Text: "inserted", Raw: "inserted"}},
					},
				},
			},
		},
	}

	editedKey, ok := model.recordKeyForPersistedRow(0)
	if !ok {
		t.Fatal("expected edited row key")
	}
	model.staging.pendingUpdates = map[string]recordEdits{
		editedKey: {
			changes: map[int]stagedEdit{
				1: {Value: dto.StagedValue{Text: "alice2", Raw: "alice2"}},
			},
		},
	}

	deleteKey, ok := model.recordKeyForPersistedRow(1)
	if !ok {
		t.Fatal("expected delete row key")
	}
	model.staging.pendingDeletes = map[string]recordDelete{
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
		colStartDisplayWidth := primitives.TextWidth(line[:colStartByteIndex])
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

func TestRenderRecords_StrikesThroughDeleteMarkedRowCellContent(t *testing.T) {
	// Arrange
	model := &Model{
		styles: primitives.NewRenderStyles(true),
		read: runtimeReadState{
			viewMode: ViewRecords,
			schema: dto.Schema{
				Columns: []dto.SchemaColumn{
					{Name: "id", Type: "INTEGER", PrimaryKey: true},
					{Name: "name", Type: "TEXT"},
				},
			},
			records: []dto.RecordRow{
				{Values: []string{"1", "alice"}},
			},
		},
	}
	key, ok := model.recordKeyForPersistedRow(0)
	if !ok {
		t.Fatal("expected persisted row key")
	}
	model.staging.pendingUpdates = map[string]recordEdits{
		key: {
			changes: map[int]stagedEdit{
				1: {Value: dto.StagedValue{Text: "alice2", Raw: "alice2"}},
			},
		},
	}
	model.staging.pendingDeletes = map[string]recordDelete{
		key: {},
	}

	// Act
	lines := model.renderRecords(80, 6)

	// Assert
	if len(lines) < 4 {
		t.Fatalf("expected rendered row, got %v", lines)
	}
	rowLine := lines[3]
	if !strings.Contains(rowLine, "\x1b[9m") {
		t.Fatalf("expected delete-marked row values to use strikethrough, got %q", rowLine)
	}
	if !strings.Contains(rowLine, "alice2") {
		t.Fatalf("expected effective staged value in delete-marked row, got %q", rowLine)
	}
}

func TestRenderRecords_CombinesSelectedAndDeletedStylesWithoutStrikingPrefixOrMarker(t *testing.T) {
	// Arrange
	model := &Model{
		styles: primitives.NewRenderStyles(true),
		read: runtimeReadState{
			viewMode:        ViewRecords,
			focus:           FocusContent,
			recordSelection: 0,
			schema: dto.Schema{
				Columns: []dto.SchemaColumn{
					{Name: "id", Type: "INTEGER", PrimaryKey: true},
					{Name: "name", Type: "TEXT"},
				},
			},
			records: []dto.RecordRow{
				{Values: []string{"1", "alice"}},
			},
		},
	}
	key, ok := model.recordKeyForPersistedRow(0)
	if !ok {
		t.Fatal("expected persisted row key")
	}
	model.staging.pendingDeletes = map[string]recordDelete{
		key: {},
	}

	// Act
	lines := model.renderRecords(80, 6)

	// Assert
	if len(lines) < 4 {
		t.Fatalf("expected rendered row, got %v", lines)
	}
	rowLine := lines[3]
	if !strings.Contains(rowLine, "\x1b[7;9m") {
		t.Fatalf("expected selected delete-marked row to preserve reverse video with strikethrough, got %q", rowLine)
	}
	expectedPrefix := primitives.SelectionSelectedPrefix() + primitives.IconDelete + " "
	if !strings.Contains(rowLine, expectedPrefix) {
		t.Fatalf("expected selection prefix and delete marker outside strikethrough, got %q", rowLine)
	}
	if strings.Contains(rowLine, expectedPrefix+"\x1b[9m") || strings.Contains(rowLine, expectedPrefix+"\x1b[7;9m") {
		t.Fatalf("expected delete strikethrough to start after selection prefix and marker, got %q", rowLine)
	}
}

func TestRenderRecordDetail_UsesVerticalLayoutWithoutTruncation(t *testing.T) {
	// Arrange
	longValue := "abcdefghijklmnopqrstuvwxyz0123456789"
	model := &Model{
		styles: primitives.NewRenderStyles(true),
		read: runtimeReadState{
			viewMode: ViewRecords,
			focus:    FocusContent,
			schema: dto.Schema{
				Columns: []dto.SchemaColumn{
					{Name: "id", Type: "INTEGER"},
					{Name: "payload", Type: "TEXT"},
				},
			},
			records: []dto.RecordRow{
				{Values: []string{"1", longValue}},
			},
		},
		overlay: runtimeOverlayState{
			recordDetail: recordDetailState{
				active: true,
			},
		},
	}

	// Act
	lines := model.renderContent(20, 8)
	content := strings.Join(lines, "\n")

	// Assert
	if !strings.Contains(content, primitives.IconInfo+" Persisted record") {
		t.Fatalf("expected information marker in detail layout, got %q", content)
	}
	if strings.Contains(content, "[ROW]") {
		t.Fatalf("expected detail layout to render user-facing labels only, got %q", content)
	}
	if !strings.Contains(content, "\x1b[1mid\x1b[0m (INTEGER)") {
		t.Fatalf("expected id header in detail layout, got %q", content)
	}
	if !strings.Contains(content, "\x1b[1mpayload\x1b[0m (TEXT)") {
		t.Fatalf("expected payload header in detail layout, got %q", content)
	}
	if strings.Contains(content, "[COL]") {
		t.Fatalf("expected detail layout without literal column marker tokens, got %q", content)
	}
	if strings.Contains(content, "...") {
		t.Fatalf("expected no truncation marker in detail layout, got %q", content)
	}
}

func TestRecordDetailContentLines_UsesInformationMarkerForRowStates(t *testing.T) {
	t.Run("persisted row", func(t *testing.T) {
		// Arrange
		model := &Model{
			styles: primitives.NewRenderStyles(true),
			read: runtimeReadState{
				schema: dto.Schema{
					Columns: []dto.SchemaColumn{
						{Name: "id", Type: "INTEGER", PrimaryKey: true},
					},
				},
				records: []dto.RecordRow{
					{Values: []string{"1"}},
				},
			},
		}

		// Act
		lines := model.recordDetailContentLines(40)

		// Assert
		if !strings.Contains(lines[0], primitives.IconInfo+" Persisted record") {
			t.Fatalf("expected persisted information marker, got %q", lines[0])
		}
	})

	t.Run("pending insert row", func(t *testing.T) {
		// Arrange
		model := &Model{
			read: runtimeReadState{
				schema: dto.Schema{
					Columns: []dto.SchemaColumn{
						{Name: "id", Type: "INTEGER"},
					},
				},
			},
			staging: stagingState{
				pendingInserts: []pendingInsertRow{
					{},
				},
			},
		}

		// Act
		lines := model.recordDetailContentLines(40)

		// Assert
		if !strings.Contains(lines[0], primitives.IconInfo+" Pending insert") {
			t.Fatalf("expected pending insert information marker, got %q", lines[0])
		}
	})

	t.Run("delete-marked persisted row", func(t *testing.T) {
		// Arrange
		model := &Model{
			read: runtimeReadState{
				schema: dto.Schema{
					Columns: []dto.SchemaColumn{
						{Name: "id", Type: "INTEGER", PrimaryKey: true},
					},
				},
				records: []dto.RecordRow{
					{Values: []string{"1"}},
				},
			},
		}
		key, ok := model.recordKeyForPersistedRow(0)
		if !ok {
			t.Fatal("expected persisted row key")
		}
		model.staging.pendingDeletes = map[string]recordDelete{
			key: {},
		}

		// Act
		lines := model.recordDetailContentLines(40)

		// Assert
		if !strings.Contains(lines[0], primitives.IconInfo+" Marked for delete") {
			t.Fatalf("expected delete information marker, got %q", lines[0])
		}
	})

	t.Run("edited persisted row", func(t *testing.T) {
		// Arrange
		model := &Model{
			styles: primitives.NewRenderStyles(true),
			read: runtimeReadState{
				schema: dto.Schema{
					Columns: []dto.SchemaColumn{
						{Name: "id", Type: "INTEGER", PrimaryKey: true},
						{Name: "name", Type: "TEXT"},
					},
				},
				records: []dto.RecordRow{
					{Values: []string{"1", "alice"}},
				},
			},
		}
		key, ok := model.recordKeyForPersistedRow(0)
		if !ok {
			t.Fatal("expected persisted row key")
		}
		model.staging.pendingUpdates = map[string]recordEdits{
			key: {
				changes: map[int]stagedEdit{
					1: {Value: dto.StagedValue{Text: "alice2", Raw: "alice2"}},
				},
			},
		}

		// Act
		lines := model.recordDetailContentLines(40)

		// Assert
		if !strings.Contains(lines[0], primitives.IconInfo+" Edited record") {
			t.Fatalf("expected edited information marker, got %q", lines[0])
		}
		content := strings.Join(lines, "\n")
		if !strings.Contains(content, "\x1b[1mname\x1b[0m (TEXT) "+primitives.IconEdit) {
			t.Fatalf("expected edit icon on modified field header in detail content, got %q", content)
		}
		if strings.Contains(content, "\x1b[1mid\x1b[0m (INTEGER) "+primitives.IconEdit) {
			t.Fatalf("expected no edit icon on unmodified field header in detail content, got %q", content)
		}
	})
}

func TestRecordDetailContentLines_StrikesDeleteMarkedFieldValuesOnly(t *testing.T) {
	// Arrange
	model := &Model{
		styles: primitives.NewRenderStyles(true),
		read: runtimeReadState{
			schema: dto.Schema{
				Columns: []dto.SchemaColumn{
					{Name: "id", Type: "INTEGER", PrimaryKey: true},
					{Name: "name", Type: "TEXT"},
				},
			},
			records: []dto.RecordRow{
				{Values: []string{"1", "alice"}},
			},
		},
	}
	key, ok := model.recordKeyForPersistedRow(0)
	if !ok {
		t.Fatal("expected persisted row key")
	}
	model.staging.pendingUpdates = map[string]recordEdits{
		key: {
			changes: map[int]stagedEdit{
				1: {Value: dto.StagedValue{Text: "alice2", Raw: "alice2"}},
			},
		},
	}
	model.staging.pendingDeletes = map[string]recordDelete{
		key: {},
	}

	// Act
	lines := model.recordDetailContentLines(40)
	content := strings.Join(lines, "\n")

	// Assert
	if !strings.Contains(content, "  \x1b[9malice2\x1b[0m") {
		t.Fatalf("expected delete-marked detail value lines to use strikethrough, got %q", content)
	}
	if strings.Contains(lines[0], "\x1b[9m") {
		t.Fatalf("expected delete summary line to remain readable without strikethrough, got %q", lines[0])
	}
	if strings.Contains(content, "\x1b[9m\x1b[1mname\x1b[0m (TEXT)") || strings.Contains(content, "\x1b[1mname\x1b[0m (TEXT)\x1b[9m") {
		t.Fatalf("expected detail field header to remain unstruck, got %q", content)
	}
}

func TestRecordDetailContentLines_StrikesEveryWrappedDeleteMarkedValueLine(t *testing.T) {
	// Arrange
	model := &Model{
		styles: primitives.NewRenderStyles(true),
		read: runtimeReadState{
			schema: dto.Schema{
				Columns: []dto.SchemaColumn{
					{Name: "id", Type: "INTEGER", PrimaryKey: true},
					{Name: "notes", Type: "TEXT"},
				},
			},
			records: []dto.RecordRow{
				{Values: []string{"1", "abcdefghijklmnopqrstuvwx"}},
			},
		},
	}
	key, ok := model.recordKeyForPersistedRow(0)
	if !ok {
		t.Fatal("expected persisted row key")
	}
	model.staging.pendingDeletes = map[string]recordDelete{
		key: {},
	}

	// Act
	lines := model.recordDetailContentLines(14)
	content := strings.Join(lines, "\n")

	// Assert
	if !strings.Contains(content, "  \x1b[9mabcdefghijkl\x1b[0m") {
		t.Fatalf("expected first wrapped delete-marked value line to use strikethrough, got %q", content)
	}
	if !strings.Contains(content, "  \x1b[9mmnopqrstuvwx\x1b[0m") {
		t.Fatalf("expected continuation delete-marked value line to keep strikethrough, got %q", content)
	}
	if !strings.Contains(stripANSI(content), "  abcdefghijkl\n  mnopqrstuvwx") {
		t.Fatalf("expected wrapped delete-marked value lines in plain text, got %q", stripANSI(content))
	}
}

func TestRecordDetailContentLinesWithStyles_BackdropKeepsDeletedStyleAcrossWrappedLines(t *testing.T) {
	// Arrange
	model := &Model{
		read: runtimeReadState{
			schema: dto.Schema{
				Columns: []dto.SchemaColumn{
					{Name: "id", Type: "INTEGER", PrimaryKey: true},
					{Name: "notes", Type: "TEXT"},
				},
			},
			records: []dto.RecordRow{
				{Values: []string{"1", "abcdefghijklmnopqrstuvwx"}},
			},
		},
	}
	key, ok := model.recordKeyForPersistedRow(0)
	if !ok {
		t.Fatal("expected persisted row key")
	}
	model.staging.pendingDeletes = map[string]recordDelete{
		key: {},
	}

	// Act
	lines := model.recordDetailContentLinesWithStyles(14, primitives.NewRenderStyles(true).Backdrop())
	content := strings.Join(lines, "\n")

	// Assert
	if !strings.Contains(content, "  \x1b[2;9mabcdefghijkl\x1b[0m") {
		t.Fatalf("expected first backdrop delete-marked value line to use faint strikethrough, got %q", content)
	}
	if !strings.Contains(content, "  \x1b[2;9mmnopqrstuvwx\x1b[0m") {
		t.Fatalf("expected backdrop continuation line to keep faint strikethrough, got %q", content)
	}
	if !strings.Contains(stripANSI(content), "  abcdefghijkl\n  mnopqrstuvwx") {
		t.Fatalf("expected wrapped backdrop delete-marked value lines in plain text, got %q", stripANSI(content))
	}
}

func TestRecordDetailContentLinesWithStyles_BackdropKeepsDeletedStyleAcrossExplicitLineBreaks(t *testing.T) {
	// Arrange
	model := &Model{
		read: runtimeReadState{
			schema: dto.Schema{
				Columns: []dto.SchemaColumn{
					{Name: "id", Type: "INTEGER", PrimaryKey: true},
					{Name: "notes", Type: "TEXT"},
				},
			},
			records: []dto.RecordRow{
				{Values: []string{"1", "alpha\nbeta"}},
			},
		},
	}
	key, ok := model.recordKeyForPersistedRow(0)
	if !ok {
		t.Fatal("expected persisted row key")
	}
	model.staging.pendingDeletes = map[string]recordDelete{
		key: {},
	}

	// Act
	lines := model.recordDetailContentLinesWithStyles(20, primitives.NewRenderStyles(true).Backdrop())
	content := strings.Join(lines, "\n")

	// Assert
	if !strings.Contains(content, "  \x1b[2;9malpha\x1b[0m") {
		t.Fatalf("expected first delete-marked line to use faint strikethrough, got %q", content)
	}
	if !strings.Contains(content, "  \x1b[2;9mbeta\x1b[0m") {
		t.Fatalf("expected line after explicit break to keep faint strikethrough, got %q", content)
	}
}

func TestDeleteMarkedViews_FallBackToPlainTextWhenStylesAreDisabled(t *testing.T) {
	// Arrange
	t.Setenv("NO_COLOR", "1")
	t.Setenv("TERM", "xterm-256color")

	model := &Model{
		styles: primitives.ResolveRenderStylesFromEnv(),
		read: runtimeReadState{
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
		},
	}
	key, ok := model.recordKeyForPersistedRow(0)
	if !ok {
		t.Fatal("expected persisted row key")
	}
	model.staging.pendingDeletes = map[string]recordDelete{
		key: {},
	}

	// Act
	recordLines := model.renderRecords(80, 6)
	detailLines := model.recordDetailContentLines(40)
	recordContent := strings.Join(recordLines, "\n")
	detailContent := strings.Join(detailLines, "\n")

	// Assert
	if strings.Contains(recordContent, "\x1b[") || strings.Contains(detailContent, "\x1b[") {
		t.Fatalf("expected plain-text fallback without ANSI styling, got records=%q detail=%q", recordContent, detailContent)
	}
	if !strings.Contains(recordContent, primitives.IconDelete+" ") {
		t.Fatalf("expected delete marker to remain visible without ANSI styling, got %q", recordContent)
	}
	if !strings.Contains(detailLines[0], primitives.IconInfo+" Marked for delete") {
		t.Fatalf("expected detail delete summary to remain visible without ANSI styling, got %q", detailLines[0])
	}
}
