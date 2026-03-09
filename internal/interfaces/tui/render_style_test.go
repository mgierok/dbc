package tui

import (
	"context"
	"strings"
	"testing"

	"github.com/mgierok/dbc/internal/application/dto"
)

func TestResolveRenderStylesFromEnv_DisablesStylingWhenNoColorIsSet(t *testing.T) {
	// Arrange
	t.Setenv("NO_COLOR", "1")
	t.Setenv("TERM", "xterm-256color")

	// Act
	styles := resolveRenderStylesFromEnv()

	// Assert
	if styles.enabled {
		t.Fatal("expected render styles to be disabled when NO_COLOR is set")
	}
}

func TestResolveRenderStylesFromEnv_DisablesStylingWhenTermIsDumb(t *testing.T) {
	// Arrange
	t.Setenv("NO_COLOR", "")
	t.Setenv("TERM", "dumb")

	// Act
	styles := resolveRenderStylesFromEnv()

	// Assert
	if styles.enabled {
		t.Fatal("expected render styles to be disabled when TERM is dumb")
	}
}

func TestNewModel_UsesDetectedRenderStyles(t *testing.T) {
	// Arrange
	originalDetector := detectRenderStyles
	t.Cleanup(func() {
		detectRenderStyles = originalDetector
	})
	detectRenderStyles = func() renderStyles {
		return renderStyles{enabled: true}
	}

	// Act
	model := NewModel(context.Background(), nil, nil, nil, nil, nil, nil, nil)

	// Assert
	if !model.styles.enabled {
		t.Fatal("expected model to keep detected render styles")
	}
}

func TestRenderStatus_StylesDirtyModeAndContextHelpWhenEnabled(t *testing.T) {
	// Arrange
	model := &Model{
		styles: renderStyles{enabled: true},
		staging: stagingState{
			pendingUpdates: map[string]recordEdits{
				"id=1": {
					changes: map[int]stagedEdit{
						0: {Value: dto.StagedValue{Text: "bob", Raw: "bob"}},
					},
				},
			},
		},
	}

	// Act
	status := model.renderStatus(120)

	// Assert
	if !strings.Contains(status, "\x1b[1mWRITE (dirty: 1)\x1b[0m") {
		t.Fatalf("expected bold dirty-mode token, got %q", status)
	}
	if !strings.Contains(status, "\x1b[2mContext help: ?\x1b[0m") {
		t.Fatalf("expected faint context-help hint, got %q", status)
	}
}

func TestRenderRecords_StylesSelectedRowAndHeaderWhenEnabled(t *testing.T) {
	// Arrange
	model := &Model{
		styles:          renderStyles{enabled: true},
		viewMode:        ViewRecords,
		focus:           FocusContent,
		recordSelection: 0,
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
	if !strings.Contains(lines[1], "\x1b[1mid\x1b[0m") || !strings.Contains(lines[1], "\x1b[1mname\x1b[0m") {
		t.Fatalf("expected bold record header labels, got %q", lines[1])
	}
	if !strings.Contains(lines[3], "\x1b[7m") {
		t.Fatalf("expected reverse-video selected record row, got %q", lines[3])
	}
}

func TestRenderStandardizedPopup_StylesTitleSelectionAndScrollIndicatorWhenEnabled(t *testing.T) {
	// Arrange
	spec := standardizedPopupSpec{
		title:               "Config",
		summary:             "Choose action.",
		rows:                popupSelectableRows([]string{"Save", "Discard", "Cancel"}, 1),
		scrollOffset:        1,
		visibleRows:         2,
		showScrollIndicator: true,
		defaultWidth:        50,
		minWidth:            20,
		maxWidth:            60,
		styles:              renderStyles{enabled: true},
	}

	// Act
	lines := renderStandardizedPopup(60, 24, spec)
	popup := strings.Join(lines, "\n")

	// Assert
	if !strings.Contains(lines[0], "\x1b[1mConfig\x1b[0m") {
		t.Fatalf("expected bold popup title, got %q", lines[0])
	}
	if !strings.Contains(popup, frameVertical+"\x1b[7m "+selectionSelectedPrefix()+"Discard") {
		t.Fatalf("expected reverse-video selected popup content without styling borders, got %q", popup)
	}
	if !strings.Contains(popup, "\x1b[2mScroll: 2/2\x1b[0m") {
		t.Fatalf("expected faint scroll indicator, got %q", popup)
	}
}

func TestRenderStandardizedPopup_StylesSummaryWhenEnabled(t *testing.T) {
	// Arrange
	spec := standardizedPopupSpec{
		title:        "Filter",
		summary:      "Select column",
		rows:         popupSelectableRows([]string{"id (INTEGER)"}, 0),
		defaultWidth: 50,
		minWidth:     20,
		maxWidth:     60,
		styles:       renderStyles{enabled: true},
	}

	// Act
	lines := renderStandardizedPopup(60, 24, spec)

	// Assert
	if !strings.Contains(lines[1], "\x1b[1mSelect column\x1b[0m") {
		t.Fatalf("expected bold popup summary, got %q", lines[1])
	}
}

func TestRenderEditPopup_StylesExplicitErrorRowsWhenEnabled(t *testing.T) {
	// Arrange
	model := &Model{
		styles: renderStyles{enabled: true},
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
			active:       true,
			columnIndex:  0,
			errorMessage: "invalid value",
		},
	}

	// Act
	popup := strings.Join(model.renderEditPopup(60), "\n")

	// Assert
	if !strings.Contains(popup, "\x1b[1;4mError: invalid value\x1b[0m") {
		t.Fatalf("expected emphasized popup error row, got %q", popup)
	}
}
