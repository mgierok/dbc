package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/dto"
)

func TestHandleKey_EnterOpensRecordDetailInRecordsView(t *testing.T) {
	// Arrange
	model := &Model{
		read: runtimeReadState{
			viewMode: ViewRecords,
			focus:    FocusContent,
			records:  []dto.RecordRow{{Values: []string{"1", "alice"}}},
			schema: dto.Schema{
				Columns: []dto.SchemaColumn{
					{Name: "id", Type: "INTEGER"},
					{Name: "name", Type: "TEXT"},
				},
			},
		},
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})

	// Assert
	if !model.overlay.recordDetail.active {
		t.Fatal("expected record detail to open in records view")
	}
	if model.overlay.recordDetail.scrollOffset != 0 {
		t.Fatalf("expected record detail scroll offset reset to 0, got %d", model.overlay.recordDetail.scrollOffset)
	}
}

func TestHandleKey_EnterIgnoredOutsideRecordsViewForDetail(t *testing.T) {
	// Arrange
	model := &Model{
		read: runtimeReadState{
			viewMode: ViewSchema,
			focus:    FocusContent,
			records:  []dto.RecordRow{{Values: []string{"1", "alice"}}},
		},
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})

	// Assert
	if model.overlay.recordDetail.active {
		t.Fatal("expected record detail to stay closed outside records view")
	}
}

func TestHandleKey_RecordDetailEscClosesDetailBeforeSwitchingPanels(t *testing.T) {
	// Arrange
	model := &Model{
		read: runtimeReadState{
			viewMode: ViewRecords,
			focus:    FocusContent,
			records:  []dto.RecordRow{{Values: []string{"1", "alice"}}},
			schema: dto.Schema{
				Columns: []dto.SchemaColumn{
					{Name: "id", Type: "INTEGER"},
					{Name: "name", Type: "TEXT"},
				},
			},
		},
		overlay: runtimeOverlayState{
			recordDetail: recordDetailState{
				active: true,
			},
		},
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyEsc})

	// Assert
	if model.overlay.recordDetail.active {
		t.Fatal("expected Esc to close record detail")
	}
	if model.read.focus != FocusContent {
		t.Fatalf("expected focus to stay in content after closing detail, got %v", model.read.focus)
	}
	if model.read.viewMode != ViewRecords {
		t.Fatalf("expected records view to remain active after closing detail, got %v", model.read.viewMode)
	}
}

func TestHandleKey_RecordDetailScrollMovesOffset(t *testing.T) {
	// Arrange
	model := &Model{
		read: runtimeReadState{
			viewMode: ViewRecords,
			focus:    FocusContent,
			schema: dto.Schema{
				Columns: []dto.SchemaColumn{
					{Name: "payload", Type: "TEXT"},
				},
			},
			records: []dto.RecordRow{
				{Values: []string{"abcdefghijklmnopqrstuvwxyz0123456789abcdefghijklmnopqrstuvwxyz0123456789"}},
			},
		},
		overlay: runtimeOverlayState{
			recordDetail: recordDetailState{
				active: true,
			},
		},
		ui: runtimeUIState{
			width:  40,
			height: 8,
		},
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})

	// Assert
	if model.overlay.recordDetail.scrollOffset <= 0 {
		t.Fatalf("expected detail scroll offset to increase, got %d", model.overlay.recordDetail.scrollOffset)
	}
}

func TestRecordDetailContentLines_UsesStagedEffectiveValue(t *testing.T) {
	// Arrange
	model := &Model{
		read: runtimeReadState{
			viewMode: ViewRecords,
			focus:    FocusContent,
			schema: dto.Schema{
				Columns: []dto.SchemaColumn{
					{Name: "id", Type: "INTEGER", PrimaryKey: true},
					{Name: "name", Type: "TEXT"},
				},
			},
			records: []dto.RecordRow{
				{Values: []string{"1", "alice"}},
			},
		},
		staging: stagingState{
			pendingUpdates: map[string]recordEdits{
				"id=1": {
					changes: map[int]stagedEdit{
						1: {Value: dto.StagedValue{Text: "bob", Raw: "bob"}},
					},
				},
			},
		},
	}

	// Act
	content := strings.Join(model.recordDetailContentLines(80), "\n")

	// Assert
	if !strings.Contains(content, "bob") {
		t.Fatalf("expected detail to include staged value, got %q", content)
	}
	if strings.Contains(content, "alice") {
		t.Fatalf("expected original value to be replaced by staged value, got %q", content)
	}
}
