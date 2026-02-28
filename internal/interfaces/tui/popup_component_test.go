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
		rows:         []string{"Save", "Discard", "Cancel"},
		selected:     1,
		defaultWidth: 50,
		minWidth:     20,
		maxWidth:     60,
	}

	// Act
	lines := renderStandardizedPopup(60, 24, spec)
	popup := strings.Join(lines, "\n")

	// Assert
	if !strings.Contains(lines[0], frameTopLeft+"Config") {
		t.Fatalf("expected title in top border, got %q", lines[0])
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
		rows:                []string{"one", "two", "three", "four"},
		selected:            -1,
		scrollOffset:        1,
		visibleRows:         2,
		showScrollIndicator: true,
		defaultWidth:        50,
		minWidth:            20,
		maxWidth:            60,
	}

	// Act
	popup := strings.Join(renderStandardizedPopup(60, 24, spec), "\n")

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
		rows:         []string{"Yes", "No"},
		selected:     0,
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
