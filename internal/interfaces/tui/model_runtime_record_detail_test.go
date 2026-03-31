package tui

import (
	"context"
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

func TestHandleKey_RecordDetailColonOpensCommandInputWithoutClearingStatus(t *testing.T) {
	// Arrange
	model := &Model{
		read: runtimeReadState{
			viewMode: ViewRecords,
			focus:    FocusContent,
		},
		overlay: runtimeOverlayState{
			recordDetail: recordDetailState{active: true},
		},
		ui: runtimeUIState{
			statusMessage: "stale status",
		},
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})

	// Assert
	if !model.overlay.commandInput.active {
		t.Fatal("expected : to open command input from record detail")
	}
	if !model.overlay.recordDetail.active {
		t.Fatal("expected record detail to stay open behind command input")
	}
	if model.ui.statusMessage != "stale status" {
		t.Fatalf("expected existing status to stay visible when command input opens, got %q", model.ui.statusMessage)
	}
}

func TestHandleKey_CommandInputEscClosesBeforeRecordDetail(t *testing.T) {
	// Arrange
	model := &Model{
		read: runtimeReadState{
			viewMode: ViewRecords,
			focus:    FocusContent,
		},
		overlay: runtimeOverlayState{
			commandInput: commandInput{
				active: true,
				value:  "set limit=10",
				cursor: len("set limit=10"),
			},
			recordDetail: recordDetailState{active: true},
		},
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyEsc})

	// Assert
	if model.overlay.commandInput.active {
		t.Fatal("expected Esc to close command input first")
	}
	if !model.overlay.recordDetail.active {
		t.Fatal("expected record detail to remain open after command input closes")
	}
}

func TestHandleKey_RecordDetailSetLimitCommandPreservesReloadBehavior(t *testing.T) {
	// Arrange
	recordsSpy := &spyListRecordsUseCase{
		page: dto.RecordPage{
			Rows:       makeRecordRows(10),
			TotalCount: 45,
		},
	}
	runtimeSession := &RuntimeSessionState{}
	model := &Model{
		ctx:            context.Background(),
		listRecords:    recordsSpy,
		runtimeSession: runtimeSession,
		read: runtimeReadState{
			viewMode:         ViewRecords,
			focus:            FocusContent,
			tables:           []dto.Table{{Name: "users"}},
			recordPageIndex:  2,
			recordTotalPages: 5,
			recordTotalCount: 81,
			recordSelection:  3,
			recordColumn:     1,
			recordFieldFocus: true,
		},
		overlay: runtimeOverlayState{
			recordDetail: recordDetailState{active: true},
		},
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
	typeCommandInputText(model, "set")
	model.handleKey(tea.KeyMsg{Type: tea.KeySpace})
	typeCommandInputText(model, "limit=10")
	_, cmd := model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected records reload command after submitting :set limit from record detail")
	}
	model.Update(cmd())

	// Assert
	if runtimeSession.RecordsPageLimit != 10 {
		t.Fatalf("expected runtime session record limit 10, got %d", runtimeSession.RecordsPageLimit)
	}
	if cmd == nil {
		t.Fatal("expected record-detail submission to delegate to records reload flow")
	}
	if recordsSpy.lastRecordsLimit != 10 {
		t.Fatalf("expected delegated reload to use limit 10, got %d", recordsSpy.lastRecordsLimit)
	}
	if model.overlay.recordDetail.active {
		t.Fatal("expected record detail to close after changing record limit")
	}
	if model.ui.statusMessage != "Record limit set to 10" {
		t.Fatalf("expected success status message, got %q", model.ui.statusMessage)
	}
}

func TestRecordDetailContentLines_RendersSafeMaterializationPlaceholders(t *testing.T) {
	// Arrange
	model := &Model{
		read: runtimeReadState{
			viewMode: ViewRecords,
			focus:    FocusContent,
			records: []dto.RecordRow{
				{
					Values: []string{"<truncated 262145 bytes>", "<blob 2 bytes>"},
				},
			},
			schema: dto.Schema{
				Columns: []dto.SchemaColumn{
					{Name: "note", Type: "TEXT"},
					{Name: "payload", Type: "BLOB"},
				},
			},
		},
	}

	// Act
	lines := model.recordDetailContentLines(80)
	content := strings.Join(lines, "\n")

	// Assert
	if !strings.Contains(content, "<truncated 262145 bytes>") {
		t.Fatalf("expected truncated placeholder in record detail, got %q", content)
	}
	if !strings.Contains(content, "<blob 2 bytes>") {
		t.Fatalf("expected blob placeholder in record detail, got %q", content)
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
	model := withTestStaging(&Model{
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
	}, stagingState{
		pendingUpdates: map[string]recordEdits{
			"id=1": {
				changes: map[int]stagedEdit{
					1: {Value: dto.StagedValue{Text: "bob", Raw: "bob"}},
				},
			},
		},
	})

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
