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
	popup := strings.Join(renderStandardizedPopup(60, spec), "\n")

	// Assert
	if !strings.Contains(popup, frameVertical+"Config") {
		t.Fatalf("expected title row, got %q", popup)
	}
	if !strings.Contains(popup, frameVertical+"  Save") {
		t.Fatalf("expected unselected row prefix, got %q", popup)
	}
	if !strings.Contains(popup, frameVertical+"> Discard") {
		t.Fatalf("expected selected row prefix, got %q", popup)
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
	popup := strings.Join(renderStandardizedPopup(60, spec), "\n")

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
