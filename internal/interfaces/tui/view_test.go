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
