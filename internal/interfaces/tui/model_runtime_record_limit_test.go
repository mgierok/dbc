package tui

import (
	"context"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/dto"
)

func TestSubmitCommandInput_SetLimitInRecordsModeReloadsFromFirstPage(t *testing.T) {
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
	cmd := submitRuntimeCommand(t, model, "set limit=10")
	if cmd == nil {
		t.Fatal("expected records reload command after setting record limit")
	}
	model.Update(cmd())

	// Assert
	if runtimeSession.RecordsPageLimit != 10 {
		t.Fatalf("expected runtime session record limit 10, got %d", runtimeSession.RecordsPageLimit)
	}
	if recordsSpy.lastRecordsOffset != 0 {
		t.Fatalf("expected records reload from first page, got offset %d", recordsSpy.lastRecordsOffset)
	}
	if recordsSpy.lastRecordsLimit != 10 {
		t.Fatalf("expected records reload with limit 10, got %d", recordsSpy.lastRecordsLimit)
	}
	if model.read.recordPageIndex != 0 {
		t.Fatalf("expected page index reset to 0, got %d", model.read.recordPageIndex)
	}
	if model.read.recordSelection != 0 {
		t.Fatalf("expected record selection reset to 0, got %d", model.read.recordSelection)
	}
	if model.read.recordFieldFocus {
		t.Fatal("expected field focus to be cleared after changing record limit")
	}
	if model.overlay.recordDetail.active {
		t.Fatal("expected record detail to close after changing record limit")
	}
	if model.read.recordTotalPages != 5 {
		t.Fatalf("expected total pages recomputed for limit 10, got %d", model.read.recordTotalPages)
	}
	if model.ui.statusMessage != "Record limit set to 10" {
		t.Fatalf("expected success status message, got %q", model.ui.statusMessage)
	}
}

func TestSubmitCommandInput_SetLimitOutsideRecordsForcesNextRecordsEntryToReload(t *testing.T) {
	// Arrange
	recordsSpy := &spyListRecordsUseCase{
		page: dto.RecordPage{
			Rows:       makeRecordRows(15),
			TotalCount: 32,
		},
	}
	model := &Model{
		ctx:         context.Background(),
		listRecords: recordsSpy,
		read: runtimeReadState{
			viewMode: ViewSchema,
			focus:    FocusTables,
			tables:   []dto.Table{{Name: "users"}},
			schema: dto.Schema{
				Columns: []dto.SchemaColumn{
					{Name: "id", Type: "INTEGER"},
				},
			},
			records: []dto.RecordRow{
				{Values: []string{"cached"}},
			},
		},
	}

	// Act
	cmd := submitRuntimeCommand(t, model, "set limit=15")
	if cmd != nil {
		t.Fatal("expected no immediate records load outside records view")
	}
	if recordsSpy.lastRecordsLimit != 0 {
		t.Fatalf("expected no records load before switching to records, got limit %d", recordsSpy.lastRecordsLimit)
	}
	if len(model.read.records) != 0 {
		t.Fatal("expected cached records to be cleared after limit change outside records view")
	}
	_, switchCmd := model.switchToRecords()
	if switchCmd == nil {
		t.Fatal("expected records load on next switch to records")
	}
	model.Update(switchCmd())

	// Assert
	if recordsSpy.lastRecordsLimit != 15 {
		t.Fatalf("expected next records load to use limit 15, got %d", recordsSpy.lastRecordsLimit)
	}
	if recordsSpy.lastRecordsOffset != 0 {
		t.Fatalf("expected next records load to start from first page, got offset %d", recordsSpy.lastRecordsOffset)
	}
	if model.read.recordPageIndex != 0 {
		t.Fatalf("expected page index reset to 0, got %d", model.read.recordPageIndex)
	}
}

func TestSubmitCommandInput_InvalidSetLimitKeepsPreviousValueAndShowsExplicitError(t *testing.T) {
	// Arrange
	runtimeSession := &RuntimeSessionState{RecordsPageLimit: 12}
	model := &Model{
		runtimeSession: runtimeSession,
	}

	// Act
	cmd := submitRuntimeCommand(t, model, "set limit=0")

	// Assert
	if cmd != nil {
		t.Fatal("expected invalid set-limit command to avoid records reload")
	}
	if runtimeSession.RecordsPageLimit != 12 {
		t.Fatalf("expected previous record limit to stay unchanged, got %d", runtimeSession.RecordsPageLimit)
	}
	if !strings.Contains(model.ui.statusMessage, "expected :set limit=<1-1000>") {
		t.Fatalf("expected explicit validation error, got %q", model.ui.statusMessage)
	}
	if strings.Contains(strings.ToLower(model.ui.statusMessage), "unknown command") {
		t.Fatalf("expected validation error instead of unknown command, got %q", model.ui.statusMessage)
	}
}

func submitRuntimeCommand(t *testing.T, model *Model, command string) tea.Cmd {
	t.Helper()

	model.overlay.commandInput = commandInput{
		active: true,
		value:  command,
		cursor: len(command),
	}

	_, cmd := model.submitCommandInput()
	if cmd == nil {
		return nil
	}
	return cmd
}

func makeRecordRows(count int) []dto.RecordRow {
	rows := make([]dto.RecordRow, count)
	for i := 0; i < count; i++ {
		rows[i] = dto.RecordRow{Values: []string{string(rune('a' + i%26))}}
	}
	return rows
}
