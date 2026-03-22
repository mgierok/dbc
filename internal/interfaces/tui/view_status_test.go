package tui

import (
	"strings"
	"testing"

	"github.com/mgierok/dbc/internal/application/dto"
)

func TestRenderStatus_ShowsContextHelpHintOnRight(t *testing.T) {
	// Arrange
	model := &Model{
		read: runtimeReadState{
			focus:    FocusContent,
			viewMode: ViewRecords,
		},
	}

	// Act
	status := stripANSI(model.renderStatus(120))

	// Assert
	if !strings.HasSuffix(status, "Context help: ?") {
		t.Fatalf("expected right-aligned context-help hint suffix, got %q", status)
	}
}

func TestRenderStatus_UsesSummarySegmentsInsteadOfInlineShortcutRows(t *testing.T) {
	// Arrange
	model := &Model{
		read: runtimeReadState{
			focus:    FocusContent,
			viewMode: ViewRecords,
		},
	}

	// Act
	status := stripANSI(model.renderStatus(200))

	// Assert
	if strings.Contains(status, "Records: Esc tables | e edit | Enter detail") {
		t.Fatalf("expected status line to stay focused on summaries, got %q", status)
	}
}

func TestRenderStatus_RightHintPriorityOnNarrowWidth(t *testing.T) {
	// Arrange
	model := &Model{
		read: runtimeReadState{
			focus:    FocusContent,
			viewMode: ViewRecords,
		},
	}

	// Act
	status := stripANSI(model.renderStatus(20))

	// Assert
	if !strings.HasSuffix(status, "Context help: ?") {
		t.Fatalf("expected narrow status line to keep full right hint, got %q", status)
	}
}

func TestRenderStatus_ShowsCleanIconInsteadOfReadOnlyLabel(t *testing.T) {
	// Arrange
	model := &Model{}

	// Act
	status := stripANSI(model.renderStatus(80))

	// Assert
	if !strings.Contains(status, "○") {
		t.Fatalf("expected clean status icon, got %q", status)
	}
	if strings.Contains(status, "READ-ONLY") {
		t.Fatalf("expected clean status to omit read-only label, got %q", status)
	}
}

func TestRenderStatus_ShowsDirtyIconInsteadOfDirtyCountText(t *testing.T) {
	// Arrange
	model := &Model{
		staging: stagingState{
			pendingUpdates: map[string]recordEdits{
				"id=1": {
					changes: map[int]stagedEdit{
						0: {Value: dto.StagedValue{Text: "bob", Raw: "bob"}},
					},
				},
			},
			pendingInserts: []pendingInsertRow{{}},
			pendingDeletes: map[string]recordDelete{"id=2": {}},
		},
	}

	// Act
	status := stripANSI(model.renderStatus(80))

	// Assert
	if !strings.Contains(status, "✱") {
		t.Fatalf("expected dirty status icon, got %q", status)
	}
	if strings.Contains(status, "WRITE (dirty: 3)") {
		t.Fatalf("expected dirty status to omit dirty-count label, got %q", status)
	}
	if strings.Contains(status, "dirty: 3") {
		t.Fatalf("expected dirty status, got %q", status)
	}
}

func TestRenderStatus_DoesNotRenderViewLabel(t *testing.T) {
	// Arrange
	model := &Model{
		read: runtimeReadState{viewMode: ViewRecords},
	}

	// Act
	status := stripANSI(model.renderStatus(200))

	// Assert
	if strings.Contains(status, "View: ") {
		t.Fatalf("expected status line without a separate view label, got %q", status)
	}
}

func TestRenderStatus_DoesNotRenderInlineCommandPromptWhenSpotlightIsActive(t *testing.T) {
	// Arrange
	model := &Model{
		read: runtimeReadState{
			viewMode: ViewRecords,
		},
		overlay: runtimeOverlayState{
			commandInput: commandInput{
				active: true,
				value:  "config",
				cursor: 3,
			},
		},
	}

	// Act
	status := stripANSI(model.renderStatus(200))

	// Assert
	if strings.Contains(status, "Command:") {
		t.Fatalf("expected status to omit inline command prompt while spotlight is active, got %q", status)
	}
	if !strings.Contains(status, "Records: 0/0") {
		t.Fatalf("expected status summaries to remain visible while spotlight is active, got %q", status)
	}
}

func TestRenderStatus_RecordsViewShowsTotalAndPagination(t *testing.T) {
	// Arrange
	model := &Model{
		read: runtimeReadState{
			viewMode:         ViewRecords,
			recordPageIndex:  0,
			recordTotalPages: 7,
			recordTotalCount: 137,
			records:          make([]dto.RecordRow, 20),
		},
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
		read: runtimeReadState{
			viewMode:         ViewRecords,
			recordPageIndex:  0,
			recordTotalPages: 1,
			recordTotalCount: 0,
		},
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

func TestRenderStatus_SanitizesFilterSortAndErrorSegments(t *testing.T) {
	// Arrange
	model := &Model{
		read: runtimeReadState{
			viewMode: ViewRecords,
			currentFilter: &dto.Filter{
				Column:   "na\x1b[31mme",
				Operator: dto.Operator{Name: "Eq\r\nuals", RequiresValue: true},
				Value:    "ali\tce\x1b]2;ignored\a",
			},
			currentSort: &dto.Sort{
				Column:    "id\x1b[32m",
				Direction: dto.SortDirection("DES\x1b[0mC"),
			},
		},
		ui: runtimeUIState{
			statusMessage: "Error: boom\x1b[31m\r\nnext",
		},
	}

	// Act
	status := model.renderStatus(220)
	plainStatus := stripANSI(status)

	// Assert
	if strings.Contains(status, "\x1b]") || strings.Contains(plainStatus, "\x1b") {
		t.Fatalf("expected status output without injected terminal escape sequences, got %q", status)
	}
	if strings.Contains(plainStatus, "\n") || strings.Contains(plainStatus, "\r") || strings.Contains(plainStatus, "\t") {
		t.Fatalf("expected status output to flatten control characters, got %q", plainStatus)
	}
	if !strings.Contains(plainStatus, "Filter: name Eq uals ali ce") {
		t.Fatalf("expected sanitized filter summary, got %q", plainStatus)
	}
	if !strings.Contains(plainStatus, "Sort: id DESC") {
		t.Fatalf("expected sanitized sort summary, got %q", plainStatus)
	}
	if !strings.Contains(plainStatus, "Error: boom next") {
		t.Fatalf("expected sanitized error status message, got %q", plainStatus)
	}
}
