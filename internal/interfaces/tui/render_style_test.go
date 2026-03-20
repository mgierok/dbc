package tui

import (
	"context"
	"strings"
	"testing"

	"github.com/mgierok/dbc/internal/application/dto"
	"github.com/mgierok/dbc/internal/interfaces/tui/internal/primitives"
)

func TestResolveRenderStylesFromEnv_DisablesStylingWhenNoColorIsSet(t *testing.T) {
	// Arrange
	t.Setenv("NO_COLOR", "1")
	t.Setenv("TERM", "xterm-256color")

	// Act
	styles := primitives.ResolveRenderStylesFromEnv()

	// Assert
	if styles.Enabled() {
		t.Fatal("expected render styles to be disabled when NO_COLOR is set")
	}
}

func TestResolveRenderStylesFromEnv_DisablesStylingWhenTermIsDumb(t *testing.T) {
	// Arrange
	t.Setenv("NO_COLOR", "")
	t.Setenv("TERM", "dumb")

	// Act
	styles := primitives.ResolveRenderStylesFromEnv()

	// Assert
	if styles.Enabled() {
		t.Fatal("expected render styles to be disabled when TERM is dumb")
	}
}

func TestNewModel_UsesDetectedRenderStyles(t *testing.T) {
	// Arrange
	originalDetector := detectRenderStyles
	t.Cleanup(func() {
		detectRenderStyles = originalDetector
	})
	detectRenderStyles = func() primitives.RenderStyles {
		return primitives.NewRenderStyles(true)
	}

	// Act
	model := NewModel(context.Background(), RuntimeRunDeps{}, nil)

	// Assert
	if !model.styles.Enabled() {
		t.Fatal("expected model to keep detected render styles")
	}
}

func TestRenderStatus_StylesDirtyModeAndContextHelpWhenEnabled(t *testing.T) {
	// Arrange
	model := &Model{
		styles: primitives.NewRenderStyles(true),
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
	if !strings.Contains(status, "\x1b[1m✱\x1b[0m") {
		t.Fatalf("expected bold dirty-mode icon, got %q", status)
	}
	if !strings.Contains(status, "\x1b[2mContext help: ?\x1b[0m") {
		t.Fatalf("expected faint context-help hint, got %q", status)
	}
}

func TestRenderRecords_StylesSelectedRowAndHeaderWhenEnabled(t *testing.T) {
	// Arrange
	model := &Model{
		styles: primitives.NewRenderStyles(true),
		read: runtimeReadState{
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

func TestRenderRecordsWithStyles_BackdropSubduesOrdinaryBodyRows(t *testing.T) {
	// Arrange
	model := &Model{
		read: runtimeReadState{
			viewMode: ViewRecords,
			focus:    FocusTables,
			schema: dto.Schema{
				Columns: []dto.SchemaColumn{
					{Name: "id", Type: "INTEGER"},
				},
			},
			records: []dto.RecordRow{
				{Values: []string{"1"}},
			},
		},
	}

	// Act
	lines := model.renderRecordsWithStyles(40, 8, primitives.NewRenderStyles(true).Backdrop())

	// Assert
	if !strings.Contains(lines[3], "\x1b[2m") {
		t.Fatalf("expected backdrop body row to be subdued, got %q", lines[3])
	}
	if strings.Contains(lines[3], "[7m") {
		t.Fatalf("expected backdrop body row to avoid selected reverse-video styling, got %q", lines[3])
	}
}

func TestRenderStandardizedPopup_StylesTitleSelectionAndScrollIndicatorWhenEnabled(t *testing.T) {
	// Arrange
	spec := primitives.StandardizedPopupSpec{
		Title:               primitives.SemanticText(primitives.SemanticRoleTitle, "Config"),
		Summary:             primitives.SemanticText(primitives.SemanticRoleSummary, "Choose action."),
		Rows:                primitives.PopupSelectableRows([]string{"Save", "Discard", "Cancel"}, 1),
		ScrollOffset:        1,
		VisibleRows:         2,
		ShowScrollIndicator: true,
		DefaultWidth:        50,
		MinWidth:            20,
		MaxWidth:            60,
		Styles:              primitives.NewRenderStyles(true),
	}

	// Act
	lines := primitives.RenderStandardizedPopup(60, 24, spec)
	popup := strings.Join(lines, "\n")

	// Assert
	if !strings.Contains(lines[0], "\x1b[1mConfig\x1b[0m") {
		t.Fatalf("expected bold popup title, got %q", lines[0])
	}
	if !strings.Contains(popup, primitives.FrameVertical+"\x1b[7m "+primitives.SelectionSelectedPrefix()+"Discard") {
		t.Fatalf("expected reverse-video selected popup content without styling borders, got %q", popup)
	}
	if !strings.Contains(popup, "\x1b[2mScroll: 2/2\x1b[0m") {
		t.Fatalf("expected faint scroll indicator, got %q", popup)
	}
}

func TestRenderStandardizedPopup_StylesSummaryWhenEnabled(t *testing.T) {
	// Arrange
	spec := primitives.StandardizedPopupSpec{
		Title:        primitives.SemanticText(primitives.SemanticRoleTitle, "Filter"),
		Summary:      primitives.SemanticText(primitives.SemanticRoleSummary, "Select column"),
		Rows:         primitives.PopupSelectableRows([]string{"id (INTEGER)"}, 0),
		DefaultWidth: 50,
		MinWidth:     20,
		MaxWidth:     60,
		Styles:       primitives.NewRenderStyles(true),
	}

	// Act
	lines := primitives.RenderStandardizedPopup(60, 24, spec)

	// Assert
	if !strings.Contains(lines[1], "\x1b[1mSelect column\x1b[0m") {
		t.Fatalf("expected bold popup summary, got %q", lines[1])
	}
}

func TestRenderEditPopup_StylesExplicitErrorRowsWhenEnabled(t *testing.T) {
	// Arrange
	model := &Model{
		styles: primitives.NewRenderStyles(true),
		read: runtimeReadState{
			schema: dto.Schema{
				Columns: []dto.SchemaColumn{
					{
						Name:  "name",
						Type:  "TEXT",
						Input: dto.ColumnInput{Kind: dto.ColumnInputText},
					},
				},
			},
		},
		overlay: runtimeOverlayState{
			editPopup: editPopup{
				active:       true,
				columnIndex:  0,
				errorMessage: "invalid value",
			},
		},
	}

	// Act
	popup := strings.Join(model.renderEditPopup(60), "\n")

	// Assert
	if !strings.Contains(popup, "\x1b[1;4mError: invalid value\x1b[0m") {
		t.Fatalf("expected emphasized popup error row, got %q", popup)
	}
}

func TestRenderStyles_BackdropSubduesSelectedAndDeleteStates(t *testing.T) {
	// Arrange
	backdrop := primitives.NewRenderStyles(true).Backdrop()

	// Act
	selected := backdrop.Selected("active")
	deleted := backdrop.Deleted("removed")
	selectedDeleted := backdrop.SelectedDeleted("removed active")

	// Assert
	if selected != "\x1b[2mactive\x1b[0m" {
		t.Fatalf("expected backdrop selected content to use faint styling only, got %q", selected)
	}
	if strings.Contains(selected, "[7m") {
		t.Fatalf("expected backdrop selected content to avoid reverse video, got %q", selected)
	}
	if deleted != "\x1b[2;9mremoved\x1b[0m" {
		t.Fatalf("expected backdrop deleted content to keep strike and add faint, got %q", deleted)
	}
	if selectedDeleted != "\x1b[2;9mremoved active\x1b[0m" {
		t.Fatalf("expected backdrop selected-deleted content to keep strike without reverse video, got %q", selectedDeleted)
	}
}

func TestResolveRenderStylesFromEnv_BackdropStaysUnstyledWhenNoColorIsSet(t *testing.T) {
	// Arrange
	t.Setenv("NO_COLOR", "1")
	t.Setenv("TERM", "xterm-256color")

	// Act
	backdrop := primitives.ResolveRenderStylesFromEnv().Backdrop()

	// Assert
	if got := backdrop.Selected("active"); got != "active" {
		t.Fatalf("expected NO_COLOR backdrop selected content to stay plain, got %q", got)
	}
	if got := backdrop.Error("Error: invalid value"); got != "Error: invalid value" {
		t.Fatalf("expected NO_COLOR backdrop errors to stay plain, got %q", got)
	}
}

func TestResolveRenderStylesFromEnv_BackdropStaysUnstyledWhenTermIsDumb(t *testing.T) {
	// Arrange
	t.Setenv("NO_COLOR", "")
	t.Setenv("TERM", "dumb")

	// Act
	backdrop := primitives.ResolveRenderStylesFromEnv().Backdrop()

	// Assert
	if got := backdrop.Title("Tables"); got != "Tables" {
		t.Fatalf("expected TERM=dumb backdrop title to stay plain, got %q", got)
	}
	if got := backdrop.Deleted("removed"); got != "removed" {
		t.Fatalf("expected TERM=dumb backdrop deleted content to stay plain, got %q", got)
	}
}
