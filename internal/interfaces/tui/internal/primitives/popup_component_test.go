package primitives

import (
	"strings"
	"testing"
)

func TestRenderStandardizedPopup_RendersSelectableRows(t *testing.T) {
	// Arrange
	spec := StandardizedPopupSpec{
		Title:        SemanticText(SemanticRoleTitle, "Config"),
		Summary:      SemanticText(SemanticRoleSummary, "Choose action."),
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

func TestRenderStandardizedPopup_RendersSummaryDividerRow(t *testing.T) {
	// Arrange
	spec := StandardizedPopupSpec{
		Title:        SemanticText(SemanticRoleTitle, "Confirm"),
		Summary:      SemanticText(SemanticRoleSummary, "Save staged changes?"),
		Rows:         PopupSelectableRows([]string{"Save", "Discard"}, 0),
		DefaultWidth: 50,
		MinWidth:     20,
		MaxWidth:     60,
	}

	// Act
	lines := RenderStandardizedPopup(60, 24, spec)
	separator := stripANSI(lines[2])

	// Assert
	if !strings.HasPrefix(separator, FrameJoinLeft) || !strings.HasSuffix(separator, FrameJoinRight) {
		t.Fatalf("expected separator row with border joins, got %q", separator)
	}
	separatorInner := strings.TrimSuffix(strings.TrimPrefix(separator, FrameJoinLeft), FrameJoinRight)
	if separatorInner == "" || strings.Trim(separatorInner, FrameHorizontal) != "" {
		t.Fatalf("expected separator row to contain only frame horizontals, got %q", separator)
	}
}

func TestRenderStandardizedPopup_ShowsScrollIndicatorForOverflow(t *testing.T) {
	// Arrange
	spec := StandardizedPopupSpec{
		Title:               SemanticText(SemanticRoleTitle, "Help"),
		Summary:             SemanticText(SemanticRoleSummary, "Summary"),
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
		Title:        SemanticText(SemanticRoleTitle, "Confirm"),
		Summary:      SemanticText(SemanticRoleSummary, "Save changes?"),
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
		Title:        SemanticText(SemanticRoleTitle, "Filter"),
		Summary:      SemanticText(SemanticRoleSummary, "Select column"),
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
		Title:        SemanticText(SemanticRoleTitle, "Sort"),
		Summary:      SemanticText(SemanticRoleSummary, "Select column"),
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
		Title:     SemanticText(SemanticRoleTitle, "DB"),
		Summary:   SemanticText(SemanticRoleSummary, "Cfg"),
		Rows:      PopupSelectableRows([]string{"local"}, 0),
		Footer:    StandardizedPopupFooter{Right: SemanticText(SemanticRoleMuted, "?")},
		WidthMode: PopupWidthContent,
	}

	// Act
	lines := RenderStandardizedPopup(80, 24, spec)

	// Assert
	actualWidth := TextWidth(stripANSI(lines[0]))
	expectedWidth := TextWidth("local") + TextWidth(SelectionSelectedPrefix()) + (popupContentSidePadding * 2) + 2
	if actualWidth != expectedWidth {
		t.Fatalf("expected content-width popup width %d, got %d for %q", expectedWidth, actualWidth, stripANSI(lines[0]))
	}
	if !strings.Contains(stripANSI(lines[len(lines)-2]), "?") {
		t.Fatalf("expected footer row before bottom border, got %q", stripANSI(lines[len(lines)-2]))
	}
}

func TestRenderStandardizedPopup_KeepsFrameWidthAndRightBorderAlignedWithUnicodeSourceMarkers(t *testing.T) {
	// Arrange
	spec := StandardizedPopupSpec{
		Title:   SemanticText(SemanticRoleTitle, "Select database"),
		Summary: SemanticText(SemanticRoleSummary, "Config: /tmp/config.json"),
		Rows: PopupSelectableRows([]string{
			IconConfigSource + " local" + FrameSegmentSeparator + "/tmp/local.sqlite",
			IconCLISource + " /tmp/direct.sqlite" + FrameSegmentSeparator + "/tmp/direct.sqlite",
		}, 0),
		Footer:    StandardizedPopupFooter{Right: SemanticText(SemanticRoleMuted, "?")},
		WidthMode: PopupWidthContent,
		Styles:    NewRenderStyles(true),
	}

	// Act
	lines := RenderStandardizedPopup(120, 24, spec)

	// Assert
	expectedWidth := TextWidth(stripANSI(lines[0]))
	for _, line := range lines {
		stripped := stripANSI(line)
		if TextWidth(stripped) != expectedWidth {
			t.Fatalf("expected consistent popup width %d, got %d for line %q", expectedWidth, TextWidth(stripped), stripped)
		}
		if !strings.HasSuffix(stripped, FrameVertical) &&
			!strings.HasSuffix(stripped, FrameTopRight) &&
			!strings.HasSuffix(stripped, FrameJoinRight) &&
			!strings.HasSuffix(stripped, FrameBottomRight) {
			t.Fatalf("expected popup line to end with right border marker, got %q", stripped)
		}
	}
	if !strings.Contains(stripANSI(strings.Join(lines, "\n")), IconConfigSource+" local"+FrameSegmentSeparator+"/tmp/local.sqlite") {
		t.Fatalf("expected config source marker in popup, got %q", stripANSI(strings.Join(lines, "\n")))
	}
	if !strings.Contains(stripANSI(strings.Join(lines, "\n")), IconCLISource+" /tmp/direct.sqlite"+FrameSegmentSeparator+"/tmp/direct.sqlite") {
		t.Fatalf("expected CLI source marker in popup, got %q", stripANSI(strings.Join(lines, "\n")))
	}
}

func TestCenterBoxLines_AddsHorizontalAndVerticalPadding(t *testing.T) {
	// Arrange
	lines := []string{
		FrameTopLeft + "Box" + FrameTopRight,
		FrameBottomLeft + strings.Repeat(FrameHorizontal, 3) + FrameBottomRight,
	}

	// Act
	centered := CenterBoxLines(lines, 13, 6)
	rows := strings.Split(centered, "\n")

	// Assert
	if len(rows) != 6 {
		t.Fatalf("expected centered output height 6, got %d", len(rows))
	}
	if rows[0] != strings.Repeat(" ", 13) || rows[1] != strings.Repeat(" ", 13) {
		t.Fatalf("expected two blank top padding rows, got %q", rows[:2])
	}
	if rows[2] != "    "+lines[0]+strings.Repeat(" ", 4) {
		t.Fatalf("expected horizontally centered first line, got %q", rows[2])
	}
	if rows[3] != "    "+lines[1]+strings.Repeat(" ", 4) {
		t.Fatalf("expected horizontally centered second line, got %q", rows[3])
	}
	if rows[4] != strings.Repeat(" ", 13) || rows[5] != strings.Repeat(" ", 13) {
		t.Fatalf("expected bottom padding rows, got %q", rows[4:])
	}
}
