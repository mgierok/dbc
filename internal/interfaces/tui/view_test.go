package tui

import (
	"strings"
	"testing"

	"github.com/mgierok/dbc/internal/application/dto"
)

func TestView_RuntimeRendersIndependentSectionBoxes(t *testing.T) {
	// Arrange
	model := &Model{
		width:    80,
		height:   24,
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
	}

	// Act
	view := model.View()
	lines := strings.Split(view, "\n")

	// Assert
	if len(lines) != model.height {
		t.Fatalf("expected runtime view height %d, got %d", model.height, len(lines))
	}
	if len(lines) < 4 {
		t.Fatalf("expected at least 4 lines in runtime view, got %d", len(lines))
	}
	topLine := lines[0]
	if !strings.Contains(topLine, frameTopLeft+"Tables") || !strings.Contains(topLine, frameTopLeft+"Schema") {
		t.Fatalf("expected titled top borders for both panels, got %q", topLine)
	}
	if strings.Contains(topLine, frameJoinCenter) || strings.Contains(topLine, frameJoinTop) || strings.Contains(topLine, frameJoinBottom) {
		t.Fatalf("expected independent panel boxes without shared joins, got %q", topLine)
	}

	statusTop := lines[len(lines)-3]
	statusContent := lines[len(lines)-2]
	statusBottom := lines[len(lines)-1]

	if !strings.HasPrefix(statusTop, frameTopLeft) || !strings.HasSuffix(statusTop, frameTopRight) {
		t.Fatalf("expected framed status top border, got %q", statusTop)
	}
	if !strings.HasPrefix(statusContent, frameVertical+" ") || !strings.HasSuffix(statusContent, " "+frameVertical) {
		t.Fatalf("expected status content with one-space side padding, got %q", statusContent)
	}
	if !strings.HasPrefix(statusBottom, frameBottomLeft) || !strings.HasSuffix(statusBottom, frameBottomRight) {
		t.Fatalf("expected framed status bottom border, got %q", statusBottom)
	}
	for _, line := range []string{statusTop, statusContent, statusBottom} {
		if textWidth(line) != model.width {
			t.Fatalf("expected status row width %d, got %d for %q", model.width, textWidth(line), line)
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
				width:    80,
				height:   24,
				viewMode: ViewSchema,
				tables:   []dto.Table{{Name: "users"}},
				schema: dto.Schema{
					Columns: []dto.SchemaColumn{{Name: "id", Type: "INTEGER"}},
				},
			},
			expected: frameTopLeft + "Schema",
		},
		{
			name: "records view",
			model: Model{
				width:    80,
				height:   24,
				viewMode: ViewRecords,
				tables:   []dto.Table{{Name: "users"}},
				schema: dto.Schema{
					Columns: []dto.SchemaColumn{{Name: "id", Type: "INTEGER"}},
				},
				records: []dto.RecordRow{{Values: []string{"1"}}},
			},
			expected: frameTopLeft + "Records",
		},
		{
			name: "record detail view",
			model: Model{
				width:    80,
				height:   24,
				viewMode: ViewRecords,
				tables:   []dto.Table{{Name: "users"}},
				recordDetail: recordDetailState{
					active: true,
				},
				schema: dto.Schema{
					Columns: []dto.SchemaColumn{{Name: "id", Type: "INTEGER"}},
				},
				records: []dto.RecordRow{{Values: []string{"1"}}},
			},
			expected: frameTopLeft + "Record Detail",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			view := tc.model.View()
			topLine := strings.Split(view, "\n")[0]
			if !strings.Contains(topLine, tc.expected) {
				t.Fatalf("expected top line to contain %q, got %q", tc.expected, topLine)
			}
		})
	}
}
