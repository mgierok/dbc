package tui

import (
	"strings"
	"testing"

	"github.com/mgierok/dbc/internal/application/dto"
	"github.com/mgierok/dbc/internal/interfaces/tui/internal/primitives"
)

func TestView_RuntimeRendersIndependentSectionBoxes(t *testing.T) {
	// Arrange
	model := &Model{
		ui: runtimeUIState{
			width:  80,
			height: 24,
		},
		read: runtimeReadState{
			focus:    FocusTables,
			viewMode: ViewSchema,
			tables: []dto.Table{
				{Name: "users"},
			},
			schema: dto.Schema{
				Columns: []dto.SchemaColumn{
					{Name: "id", Type: "INTEGER"},
				},
			},
		},
	}

	// Act
	view := model.View()
	lines := strings.Split(view, "\n")

	// Assert
	if len(lines) != model.ui.height {
		t.Fatalf("expected runtime view height %d, got %d", model.ui.height, len(lines))
	}
	if len(lines) < 4 {
		t.Fatalf("expected at least 4 lines in runtime view, got %d", len(lines))
	}
	topLine := stripANSI(lines[0])
	if !strings.Contains(topLine, primitives.FrameTopLeft+"Tables") || !strings.Contains(topLine, primitives.FrameTopLeft+"Schema") {
		t.Fatalf("expected titled top borders for both panels, got %q", topLine)
	}
	if strings.Contains(topLine, primitives.FrameJoinCenter) || strings.Contains(topLine, primitives.FrameJoinTop) || strings.Contains(topLine, primitives.FrameJoinBottom) {
		t.Fatalf("expected independent panel boxes without shared joins, got %q", topLine)
	}

	statusTop := stripANSI(lines[len(lines)-3])
	statusContent := stripANSI(lines[len(lines)-2])
	statusBottom := stripANSI(lines[len(lines)-1])

	if !strings.HasPrefix(statusTop, primitives.FrameTopLeft) || !strings.HasSuffix(statusTop, primitives.FrameTopRight) {
		t.Fatalf("expected framed status top border, got %q", statusTop)
	}
	if !strings.HasPrefix(statusContent, primitives.FrameVertical+" ") || !strings.HasSuffix(statusContent, " "+primitives.FrameVertical) {
		t.Fatalf("expected status content with one-space side padding, got %q", statusContent)
	}
	if !strings.HasPrefix(statusBottom, primitives.FrameBottomLeft) || !strings.HasSuffix(statusBottom, primitives.FrameBottomRight) {
		t.Fatalf("expected framed status bottom border, got %q", statusBottom)
	}
	for _, line := range []string{statusTop, statusContent, statusBottom} {
		if primitives.TextWidth(line) != model.ui.width {
			t.Fatalf("expected status row width %d, got %d for %q", model.ui.width, primitives.TextWidth(line), line)
		}
	}
}

func TestView_RuntimeRightPanelTopBorderUsesDynamicTitle(t *testing.T) {
	testCases := []struct {
		name     string
		model    Model
		expected string
	}{
		{
			name: "schema view",
			model: Model{
				ui: runtimeUIState{
					width:  80,
					height: 24,
				},
				read: runtimeReadState{
					viewMode: ViewSchema,
					tables:   []dto.Table{{Name: "users"}},
					schema: dto.Schema{
						Columns: []dto.SchemaColumn{{Name: "id", Type: "INTEGER"}},
					},
				},
			},
			expected: primitives.FrameTopLeft + "Schema",
		},
		{
			name: "records view",
			model: Model{
				ui: runtimeUIState{
					width:  80,
					height: 24,
				},
				read: runtimeReadState{
					viewMode: ViewRecords,
					tables:   []dto.Table{{Name: "users"}},
					schema: dto.Schema{
						Columns: []dto.SchemaColumn{{Name: "id", Type: "INTEGER"}},
					},
					records: []dto.RecordRow{{Values: []string{"1"}}},
				},
			},
			expected: primitives.FrameTopLeft + "Records",
		},
		{
			name: "dirty records view",
			model: Model{
				ui: runtimeUIState{
					width:  80,
					height: 24,
				},
				read: runtimeReadState{
					viewMode: ViewRecords,
					tables:   []dto.Table{{Name: "users"}},
					schema: dto.Schema{
						Columns: []dto.SchemaColumn{{Name: "id", Type: "INTEGER"}},
					},
					records: []dto.RecordRow{{Values: []string{"1"}}},
				},
				staging: stagingState{
					pendingUpdates: map[string]recordEdits{
						"id=1": {
							changes: map[int]stagedEdit{
								0: {Value: dto.StagedValue{Text: "2", Raw: "2"}},
							},
						},
					},
				},
			},
			expected: primitives.FrameTopLeft + "Records [staged rows: 1]",
		},
		{
			name: "record detail view",
			model: Model{
				ui: runtimeUIState{
					width:  80,
					height: 24,
				},
				read: runtimeReadState{
					viewMode: ViewRecords,
					tables:   []dto.Table{{Name: "users"}},
					schema: dto.Schema{
						Columns: []dto.SchemaColumn{{Name: "id", Type: "INTEGER"}},
					},
					records: []dto.RecordRow{{Values: []string{"1"}}},
				},
				overlay: runtimeOverlayState{
					recordDetail: recordDetailState{
						active: true,
					},
				},
			},
			expected: primitives.FrameTopLeft + "Record Detail",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			view := tc.model.View()
			topLine := stripANSI(strings.Split(view, "\n")[0])
			if !strings.Contains(topLine, tc.expected) {
				t.Fatalf("expected top line to contain %q, got %q", tc.expected, topLine)
			}
		})
	}
}

func TestView_CommandSpotlightOverlaysRuntimePanelsAndStatusBar(t *testing.T) {
	// Arrange
	model := &Model{
		styles: primitives.NewRenderStyles(true),
		ui: runtimeUIState{
			width:         140,
			height:        24,
			statusMessage: "Affected rows: 1",
		},
		read: runtimeReadState{
			focus:         FocusContent,
			viewMode:      ViewRecords,
			selectedTable: 0,
			tables:        []dto.Table{{Name: "users"}},
			schema: dto.Schema{
				Columns: []dto.SchemaColumn{{Name: "id", Type: "INTEGER"}},
			},
			records:          []dto.RecordRow{{Values: []string{"1"}}},
			recordTotalCount: 1,
			recordTotalPages: 1,
		},
		overlay: runtimeOverlayState{
			commandInput: commandInput{
				active: true,
				value:  "set limit=10",
				cursor: len("set limit=10"),
			},
		},
	}

	// Act
	rawView := model.View()
	view := stripANSI(rawView)
	lines := strings.Split(view, "\n")

	// Assert
	if !strings.Contains(lines[0], primitives.FrameTopLeft+"Tables") || !strings.Contains(lines[0], primitives.FrameTopLeft+"Records") {
		t.Fatalf("expected runtime panels to remain visible behind spotlight, got %q", lines[0])
	}
	if !strings.Contains(rawView, "\x1b[2mTables\x1b[0m") || !strings.Contains(rawView, "\x1b[2mRecords\x1b[0m") {
		t.Fatalf("expected spotlight backdrop to subdue panel titles, got %q", rawView)
	}
	if !strings.Contains(rawView, "\x1b[2mRecords:\x1b[0m\x1b[2m 1/1\x1b[0m") || !strings.Contains(rawView, "\x1b[2mPage:\x1b[0m\x1b[2m 1/1\x1b[0m") {
		t.Fatalf("expected spotlight backdrop to keep subdued status bar summaries, got %q", rawView)
	}
	if !strings.Contains(view, "Affected rows: 1") {
		t.Fatalf("expected status bar to remain visible behind spotlight, got %q", view)
	}
	if strings.Contains(view, "Command:") {
		t.Fatalf("expected no inline command segment in status while spotlight is active, got %q", view)
	}
	if !strings.Contains(view, "│:set limit=10|") {
		t.Fatalf("expected spotlight input row with prompt and caret, got %q", view)
	}
}

func TestView_CommandSpotlightBackdropSubduesDirtyRecordsTitle(t *testing.T) {
	// Arrange
	model := &Model{
		styles: primitives.NewRenderStyles(true),
		ui: runtimeUIState{
			width:  140,
			height: 24,
		},
		read: runtimeReadState{
			focus:         FocusContent,
			viewMode:      ViewRecords,
			selectedTable: 0,
			tables:        []dto.Table{{Name: "users"}},
			schema: dto.Schema{
				Columns: []dto.SchemaColumn{{Name: "id", Type: "INTEGER"}},
			},
			records: []dto.RecordRow{{Values: []string{"1"}}},
		},
		staging: stagingState{
			pendingInserts: []pendingInsertRow{{}},
		},
		overlay: runtimeOverlayState{
			commandInput: commandInput{
				active: true,
			},
		},
	}

	// Act
	view := model.View()

	// Assert
	if !strings.Contains(view, "\x1b[2mRecords [staged rows: 1]\x1b[0m") {
		t.Fatalf("expected spotlight backdrop to subdue dirty records title, got %q", view)
	}
}

func TestView_CommandSpotlightCentersAndClampsWidthInNarrowWindow(t *testing.T) {
	// Arrange
	model := &Model{
		ui: runtimeUIState{
			width:  13,
			height: 9,
		},
		read: runtimeReadState{
			focus:    FocusContent,
			viewMode: ViewRecords,
		},
		overlay: runtimeOverlayState{
			commandInput: commandInput{
				active: true,
			},
		},
	}

	// Act
	view := stripANSI(model.View())
	lines := strings.Split(view, "\n")

	// Assert
	if len(lines) != 9 {
		t.Fatalf("expected runtime view height 9, got %d", len(lines))
	}
	if lines[3] != "┌───────────┐" {
		t.Fatalf("expected centered spotlight top border with terminal-clamped width, got %q", lines[3])
	}
	if lines[4] != "│:|         │" {
		t.Fatalf("expected spotlight input row to preserve 10-character field, got %q", lines[4])
	}
	if lines[5] != "└───────────┘" {
		t.Fatalf("expected centered spotlight bottom border, got %q", lines[5])
	}
}

func TestView_CommandSpotlightDefaultsToHalfTerminalWidth(t *testing.T) {
	// Arrange
	model := &Model{
		ui: runtimeUIState{
			width:  80,
			height: 24,
		},
		read: runtimeReadState{
			focus:    FocusContent,
			viewMode: ViewRecords,
		},
		overlay: runtimeOverlayState{
			commandInput: commandInput{
				active: true,
			},
		},
	}

	// Act
	view := stripANSI(model.View())
	lines := strings.Split(view, "\n")

	// Assert
	topBorder := "┌──────────────────────────────────────┐"
	contentRow := "│:|                                    │"
	bottomBorder := "└──────────────────────────────────────┘"

	topIndex := strings.Index(lines[10], topBorder)
	if topIndex < 0 {
		t.Fatalf("expected spotlight top border to be rendered, got %q", lines[10])
	}
	if primitives.TextWidth(lines[10][:topIndex]) != 20 {
		t.Fatalf("expected spotlight top border to start at centered column 21, got prefix width %d in %q", primitives.TextWidth(lines[10][:topIndex]), lines[10])
	}
	if !strings.Contains(lines[10], topBorder) {
		t.Fatalf("expected spotlight top border to default to half terminal width, got %q", lines[10])
	}
	if !strings.Contains(lines[11], contentRow) {
		t.Fatalf("expected spotlight content row to use half-width default, got %q", lines[11])
	}
	if !strings.Contains(lines[12], bottomBorder) {
		t.Fatalf("expected spotlight bottom border to use half-width default, got %q", lines[12])
	}
}

func TestView_PendingCommandSpotlightShowsStatusInsteadOfEditablePrompt(t *testing.T) {
	// Arrange
	model := &Model{
		ui: runtimeUIState{
			width:  80,
			height: 24,
		},
		read: runtimeReadState{
			focus:    FocusContent,
			viewMode: ViewRecords,
		},
		overlay: runtimeOverlayState{
			commandInput: commandInput{
				active:        true,
				mode:          commandInputModePending,
				value:         "edit /tmp/analytics.sqlite",
				pendingStatus: "Opening \"/tmp/analytics.sqlite\"...",
			},
		},
	}

	// Act
	view := stripANSI(model.View())

	// Assert
	if !strings.Contains(view, "Opening \"/tmp/analytics.sqlite\"...") {
		t.Fatalf("expected pending spotlight status, got %q", view)
	}
	if strings.Contains(view, ":|") {
		t.Fatalf("expected pending spotlight to hide editable prompt, got %q", view)
	}
}
