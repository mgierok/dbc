package tui

import (
	"context"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/dto"
	"github.com/mgierok/dbc/internal/interfaces/tui/internal/primitives"
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

func TestSubmitCommandInput_SetLimitOverwritesPreviousSessionValue(t *testing.T) {
	// Arrange
	runtimeSession := &RuntimeSessionState{RecordsPageLimit: 10}
	model := &Model{
		read:           runtimeReadState{viewMode: ViewSchema},
		runtimeSession: runtimeSession,
	}

	// Act
	cmd := submitRuntimeCommand(t, model, "set limit=25")

	// Assert
	if cmd != nil {
		t.Fatal("expected no immediate records load outside records view")
	}
	if runtimeSession.RecordsPageLimit != 25 {
		t.Fatalf("expected record limit overwrite to 25, got %d", runtimeSession.RecordsPageLimit)
	}
	if model.ui.statusMessage != "Record limit set to 25" {
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

func TestSubmitCommandInput_SetLimitIgnoresStaleRecordsResponse(t *testing.T) {
	// Arrange
	recordsSpy := &spyListRecordsUseCase{
		page: dto.RecordPage{
			Rows:       makeRecordRows(10),
			TotalCount: 30,
		},
	}
	model := &Model{
		ctx:            context.Background(),
		listRecords:    recordsSpy,
		runtimeSession: &RuntimeSessionState{},
		read: runtimeReadState{
			viewMode:        ViewRecords,
			focus:           FocusContent,
			tables:          []dto.Table{{Name: "users"}},
			recordRequestID: 1,
			recordLoading:   true,
			records: []dto.RecordRow{
				{Values: []string{"old"}},
			},
		},
	}
	staleMsg := recordsMsg{
		tableName: "users",
		requestID: 1,
		page: dto.RecordPage{
			Rows:       []dto.RecordRow{{Values: []string{"stale"}}},
			TotalCount: 99,
		},
	}

	// Act
	cmd := submitRuntimeCommand(t, model, "set limit=10")
	if cmd == nil {
		t.Fatal("expected records reload command after changing record limit")
	}
	model.Update(staleMsg)
	model.Update(cmd())

	// Assert
	if len(model.read.records) != 10 {
		t.Fatalf("expected fresh records page to be applied, got %d rows", len(model.read.records))
	}
	if len(model.read.records) > 0 && model.read.records[0].Values[0] == "stale" {
		t.Fatalf("expected stale response to be ignored, got %+v", model.read.records[0])
	}
	if recordsSpy.lastRecordsLimit != 10 {
		t.Fatalf("expected fresh reload to use limit 10, got %d", recordsSpy.lastRecordsLimit)
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

func TestLoadRecordsCmd_ClampsOversizedSessionLimitBeforeUse(t *testing.T) {
	// Arrange
	recordsSpy := &spyListRecordsUseCase{
		page: dto.RecordPage{
			Rows:       makeRecordRows(1),
			TotalCount: 1,
		},
	}
	model := &Model{
		ctx:         context.Background(),
		listRecords: recordsSpy,
		read:        runtimeReadState{tables: []dto.Table{{Name: "users"}}},
		runtimeSession: &RuntimeSessionState{
			RecordsPageLimit: primitives.RuntimeMaxRecordPageLimit + 1,
		},
	}

	// Act
	cmd := model.loadRecordsCmd(true)
	if cmd == nil {
		t.Fatal("expected records load command")
	}
	model.Update(cmd())

	// Assert
	if recordsSpy.lastRecordsLimit != primitives.RuntimeMaxRecordPageLimit {
		t.Fatalf("expected oversized session value to clamp to %d, got %d", primitives.RuntimeMaxRecordPageLimit, recordsSpy.lastRecordsLimit)
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
