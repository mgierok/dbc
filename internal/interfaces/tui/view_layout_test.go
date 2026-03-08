package tui

import (
	"strings"
	"testing"

	"github.com/mgierok/dbc/internal/application/dto"
)

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
