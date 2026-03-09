package tui

import (
	"strings"
	"testing"
)

func TestRenderStandardizedPopup_RendersSelectableRows(t *testing.T) {
	// Arrange
	spec := standardizedPopupSpec{
		title:        "Config",
		summary:      "Choose action.",
		rows:         popupSelectableRows([]string{"Save", "Discard", "Cancel"}, 1),
		defaultWidth: 50,
		minWidth:     20,
		maxWidth:     60,
	}

	// Act
	lines := renderStandardizedPopup(60, 24, spec)
	popup := stripANSI(strings.Join(lines, "\n"))

	// Assert
	if !strings.Contains(stripANSI(lines[0]), frameTopLeft+"Config") {
		t.Fatalf("expected title in top border, got %q", stripANSI(lines[0]))
	}
	if !strings.Contains(popup, frameVertical+" "+selectionUnselectedPrefix()+"Save") {
		t.Fatalf("expected one-space left padding for rows, got %q", popup)
	}
	if !strings.Contains(popup, selectionSelectedPrefix()+"Discard") {
		t.Fatalf("expected selected row prefix, got %q", popup)
	}
	if !strings.Contains(popup, " "+frameVertical) {
		t.Fatalf("expected one-space right padding before right border, got %q", popup)
	}
}

func TestRenderStandardizedPopup_ShowsScrollIndicatorForOverflow(t *testing.T) {
	// Arrange
	spec := standardizedPopupSpec{
		title:               "Help",
		summary:             "Summary",
		rows:                popupTextRows([]string{"one", "two", "three", "four"}),
		scrollOffset:        1,
		visibleRows:         2,
		showScrollIndicator: true,
		defaultWidth:        50,
		minWidth:            20,
		maxWidth:            60,
	}

	// Act
	popup := stripANSI(strings.Join(renderStandardizedPopup(60, 24, spec), "\n"))

	// Assert
	if !strings.Contains(popup, "two") || !strings.Contains(popup, "three") {
		t.Fatalf("expected scrolled window rows, got %q", popup)
	}
	if strings.Contains(popup, "one") || strings.Contains(popup, "four") {
		t.Fatalf("expected rows outside viewport to be hidden, got %q", popup)
	}
	if !strings.Contains(popup, "Scroll: 2/3") {
		t.Fatalf("expected scroll indicator, got %q", popup)
	}
}

func TestRenderStandardizedPopup_EnforcesMinimumHeight40Percent(t *testing.T) {
	// Arrange
	spec := standardizedPopupSpec{
		title:        "Confirm",
		summary:      "Save changes?",
		rows:         popupSelectableRows([]string{"Yes", "No"}, 0),
		defaultWidth: 50,
		minWidth:     20,
		maxWidth:     60,
	}

	// Act
	lines := renderStandardizedPopup(60, 24, spec)

	// Assert
	minExpectedHeight := (24*40 + 99) / 100
	if len(lines) < minExpectedHeight {
		t.Fatalf("expected min popup height %d, got %d", minExpectedHeight, len(lines))
	}
}

func TestRenderStandardizedPopup_DoesNotEmphasizeOrdinaryRowsThatContainErrorLikeWords(t *testing.T) {
	// Arrange
	spec := standardizedPopupSpec{
		title:        "Filter",
		summary:      "Select column",
		rows:         popupSelectableRows([]string{"failed_login (TEXT)", "invalid status (TEXT)"}, 0),
		defaultWidth: 60,
		minWidth:     20,
		maxWidth:     60,
		styles:       renderStyles{enabled: true},
	}

	// Act
	popup := strings.Join(renderStandardizedPopup(60, 24, spec), "\n")

	// Assert
	if strings.Contains(popup, "\x1b[1;4mfailed_login (TEXT)\x1b[0m") {
		t.Fatalf("expected schema-like row label to remain unstyled, got %q", popup)
	}
	if strings.Contains(popup, "\x1b[1;4minvalid status (TEXT)\x1b[0m") {
		t.Fatalf("expected ordinary row label to remain unstyled, got %q", popup)
	}
}

func TestRenderStandardizedPopup_SelectedRowDoesNotStyleBorders(t *testing.T) {
	// Arrange
	spec := standardizedPopupSpec{
		title:        "Sort",
		summary:      "Select column",
		rows:         popupSelectableRows([]string{"id (TEXT)", "kind (TEXT)"}, 1),
		defaultWidth: 60,
		minWidth:     20,
		maxWidth:     60,
		styles:       renderStyles{enabled: true},
	}

	// Act
	lines := renderStandardizedPopup(60, 24, spec)

	// Assert
	selectedLine := lines[4]
	if strings.HasPrefix(selectedLine, "\x1b[7m"+frameVertical) {
		t.Fatalf("expected left border to remain unstyled, got %q", selectedLine)
	}
	if strings.Contains(selectedLine, frameVertical+"\x1b[0m") {
		t.Fatalf("expected right border to remain outside selected styling, got %q", selectedLine)
	}
	if !strings.Contains(selectedLine, frameVertical+"\x1b[7m "+selectionSelectedPrefix()+"kind (TEXT)") {
		t.Fatalf("expected only popup content to be reverse-video, got %q", selectedLine)
	}
}

func TestRenderStandardizedPopup_ContentWidthModeUsesSelectableRowsAndFooterWidth(t *testing.T) {
	// Arrange
	spec := standardizedPopupSpec{
		title:     "DB",
		summary:   "Cfg",
		rows:      popupSelectableRows([]string{"local"}, 0),
		footer:    standardizedPopupFooter{right: "Context help: ?"},
		widthMode: popupWidthContent,
	}

	// Act
	lines := renderStandardizedPopup(80, 24, spec)

	// Assert
	actualWidth := textWidth(stripANSI(lines[0]))
	minExpectedWidth := textWidth("Context help: ?") + (popupContentSidePadding * 2) + 2
	if actualWidth != minExpectedWidth {
		t.Fatalf("expected content-width popup width %d, got %d for %q", minExpectedWidth, actualWidth, stripANSI(lines[0]))
	}
	if !strings.Contains(stripANSI(lines[len(lines)-2]), "Context help: ?") {
		t.Fatalf("expected footer row before bottom border, got %q", stripANSI(lines[len(lines)-2]))
	}
}
