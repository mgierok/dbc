package primitives

import (
	"strings"
	"testing"
)

func TestRenderStandardizedPopup_RendersSelectableRows(t *testing.T) {
	// Arrange
	spec := StandardizedPopupSpec{
		Title:        "Config",
		Summary:      "Choose action.",
		Rows:         PopupSelectableRows([]string{"Save", "Discard", "Cancel"}, 1),
		DefaultWidth: 50,
		MinWidth:     20,
		MaxWidth:     60,
	}

	// Act
	lines := RenderStandardizedPopup(60, 24, spec)
	popup := stripANSI(strings.Join(lines, "\n"))

	// Assert
	if !strings.Contains(stripANSI(lines[0]), FrameTopLeft+"Config") {
		t.Fatalf("expected title in top border, got %q", stripANSI(lines[0]))
	}
	if !strings.Contains(popup, FrameVertical+" "+SelectionUnselectedPrefix()+"Save") {
		t.Fatalf("expected one-space left padding for rows, got %q", popup)
	}
	if !strings.Contains(popup, SelectionSelectedPrefix()+"Discard") {
		t.Fatalf("expected selected row prefix, got %q", popup)
	}
	if !strings.Contains(popup, " "+FrameVertical) {
		t.Fatalf("expected one-space right padding before right border, got %q", popup)
	}
}

func TestRenderStandardizedPopup_ShowsScrollIndicatorForOverflow(t *testing.T) {
	// Arrange
	spec := StandardizedPopupSpec{
		Title:               "Help",
		Summary:             "Summary",
		Rows:                PopupTextRows([]string{"one", "two", "three", "four"}),
		ScrollOffset:        1,
		VisibleRows:         2,
		ShowScrollIndicator: true,
		DefaultWidth:        50,
		MinWidth:            20,
		MaxWidth:            60,
	}

	// Act
	popup := stripANSI(strings.Join(RenderStandardizedPopup(60, 24, spec), "\n"))

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
	spec := StandardizedPopupSpec{
		Title:        "Confirm",
		Summary:      "Save changes?",
		Rows:         PopupSelectableRows([]string{"Yes", "No"}, 0),
		DefaultWidth: 50,
		MinWidth:     20,
		MaxWidth:     60,
	}

	// Act
	lines := RenderStandardizedPopup(60, 24, spec)

	// Assert
	minExpectedHeight := (24*40 + 99) / 100
	if len(lines) < minExpectedHeight {
		t.Fatalf("expected min popup height %d, got %d", minExpectedHeight, len(lines))
	}
}

func TestRenderStandardizedPopup_DoesNotEmphasizeOrdinaryRowsThatContainErrorLikeWords(t *testing.T) {
	// Arrange
	spec := StandardizedPopupSpec{
		Title:        "Filter",
		Summary:      "Select column",
		Rows:         PopupSelectableRows([]string{"failed_login (TEXT)", "invalid status (TEXT)"}, 0),
		DefaultWidth: 60,
		MinWidth:     20,
		MaxWidth:     60,
		Styles:       NewRenderStyles(true),
	}

	// Act
	popup := strings.Join(RenderStandardizedPopup(60, 24, spec), "\n")

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
	spec := StandardizedPopupSpec{
		Title:        "Sort",
		Summary:      "Select column",
		Rows:         PopupSelectableRows([]string{"id (TEXT)", "kind (TEXT)"}, 1),
		DefaultWidth: 60,
		MinWidth:     20,
		MaxWidth:     60,
		Styles:       NewRenderStyles(true),
	}

	// Act
	lines := RenderStandardizedPopup(60, 24, spec)

	// Assert
	selectedLine := lines[4]
	if strings.HasPrefix(selectedLine, "\x1b[7m"+FrameVertical) {
		t.Fatalf("expected left border to remain unstyled, got %q", selectedLine)
	}
	if strings.Contains(selectedLine, FrameVertical+"\x1b[0m") {
		t.Fatalf("expected right border to remain outside selected styling, got %q", selectedLine)
	}
	if !strings.Contains(selectedLine, FrameVertical+"\x1b[7m "+SelectionSelectedPrefix()+"kind (TEXT)") {
		t.Fatalf("expected only popup content to be reverse-video, got %q", selectedLine)
	}
}

func TestRenderStandardizedPopup_ContentWidthModeUsesSelectableRowsAndFooterWidth(t *testing.T) {
	// Arrange
	spec := StandardizedPopupSpec{
		Title:     "DB",
		Summary:   "Cfg",
		Rows:      PopupSelectableRows([]string{"local"}, 0),
		Footer:    StandardizedPopupFooter{Right: "Context help: ?"},
		WidthMode: PopupWidthContent,
	}

	// Act
	lines := RenderStandardizedPopup(80, 24, spec)

	// Assert
	actualWidth := TextWidth(stripANSI(lines[0]))
	minExpectedWidth := TextWidth("Context help: ?") + (popupContentSidePadding * 2) + 2
	if actualWidth != minExpectedWidth {
		t.Fatalf("expected content-width popup width %d, got %d for %q", minExpectedWidth, actualWidth, stripANSI(lines[0]))
	}
	if !strings.Contains(stripANSI(lines[len(lines)-2]), "Context help: ?") {
		t.Fatalf("expected footer row before bottom border, got %q", stripANSI(lines[len(lines)-2]))
	}
}
