package tui

import (
	"strings"
	"testing"

	"github.com/mgierok/dbc/internal/application/dto"
)

func TestRenderStatus_ShowsContextHelpHintOnRight(t *testing.T) {
	// Arrange
	model := &Model{
		focus:    FocusContent,
		viewMode: ViewRecords,
	}

	// Act
	status := stripANSI(model.renderStatus(120))

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
	status := stripANSI(model.renderStatus(200))

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
	status := stripANSI(model.renderStatus(20))

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
					0: {Value: dto.StagedValue{Text: "bob", Raw: "bob"}},
				},
			},
		},
		pendingInserts: []pendingInsertRow{{}},
		pendingDeletes: map[string]recordDelete{"id=2": {}},
	}

	// Act
	status := stripANSI(model.renderStatus(80))

	// Assert
	if !strings.Contains(status, "WRITE (dirty: 3)") {
		t.Fatalf("expected dirty status, got %q", status)
	}
}

func TestRenderStatus_DoesNotShowViewIndicator(t *testing.T) {
	// Arrange
	model := &Model{
		viewMode: ViewRecords,
	}

	// Act
	status := stripANSI(model.renderStatus(200))

	// Assert
	if strings.Contains(status, "View: ") {
		t.Fatalf("expected status without view indicator, got %q", status)
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
	status := stripANSI(model.renderStatus(200))

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
	status := stripANSI(model.renderStatus(220))

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
	status := stripANSI(model.renderStatus(220))

	// Assert
	if !strings.Contains(status, "Records: 0/0") {
		t.Fatalf("expected empty records summary in status, got %q", status)
	}
	if !strings.Contains(status, "Page: 1/1") {
		t.Fatalf("expected single-page summary in status, got %q", status)
	}
}
