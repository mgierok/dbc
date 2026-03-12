package tui

import (
	"context"
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/dto"
)

type stubGetSchemaUseCase struct {
	lastTableName string
	schema        dto.Schema
	err           error
}

func (s *stubGetSchemaUseCase) Execute(ctx context.Context, tableName string) (dto.Schema, error) {
	s.lastTableName = tableName
	if s.err != nil {
		return dto.Schema{}, s.err
	}
	return s.schema, nil
}

func TestUpdate_TablesMsgStoresTablesSelectsFirstTableAndStartsSchemaLoad(t *testing.T) {
	// Arrange
	getSchema := &stubGetSchemaUseCase{
		schema: dto.Schema{
			Columns: []dto.SchemaColumn{
				{Name: "id", Type: "INTEGER"},
			},
		},
	}
	model := &Model{
		ctx:       context.Background(),
		getSchema: getSchema,
	}
	msg := tablesMsg{
		tables: []dto.Table{
			{Name: "users"},
			{Name: "orders"},
		},
	}

	// Act
	_, cmd := model.Update(msg)

	// Assert
	if len(model.read.tables) != 2 {
		t.Fatalf("expected 2 tables to be stored, got %d", len(model.read.tables))
	}
	if model.read.tables[0].Name != "users" {
		t.Fatalf("expected first stored table to be users, got %q", model.read.tables[0].Name)
	}
	if model.read.selectedTable != 0 {
		t.Fatalf("expected first table to be selected, got index %d", model.read.selectedTable)
	}
	if cmd == nil {
		t.Fatal("expected schema-load command after receiving tables")
	}

	schemaMessage, ok := cmd().(schemaMsg)
	if !ok {
		t.Fatalf("expected schemaMsg from schema-load command, got %T", cmd())
	}
	if getSchema.lastTableName != "users" {
		t.Fatalf("expected schema load for users table, got %q", getSchema.lastTableName)
	}
	if schemaMessage.tableName != "users" {
		t.Fatalf("expected schema message for users table, got %q", schemaMessage.tableName)
	}
}

func TestUpdate_RecordsMsgIgnoresStaleRequestIDAndPreservesCurrentRecords(t *testing.T) {
	// Arrange
	currentRecords := []dto.RecordRow{
		{Values: []string{"current"}},
	}
	model := &Model{
		read: runtimeReadState{
			tables:           []dto.Table{{Name: "users"}},
			records:          append([]dto.RecordRow(nil), currentRecords...),
			recordRequestID:  2,
			recordLoading:    true,
			recordTotalCount: 1,
			recordTotalPages: 1,
		},
	}
	msg := recordsMsg{
		tableName: "users",
		requestID: 1,
		page: dto.RecordPage{
			Rows: []dto.RecordRow{
				{Values: []string{"stale"}},
			},
			TotalCount: 99,
		},
	}

	// Act
	_, cmd := model.Update(msg)

	// Assert
	if cmd != nil {
		t.Fatal("expected no follow-up command for stale records response")
	}
	if len(model.read.records) != 1 {
		t.Fatalf("expected current records to stay unchanged, got %d rows", len(model.read.records))
	}
	if model.read.records[0].Values[0] != "current" {
		t.Fatalf("expected current records to be preserved, got %q", model.read.records[0].Values[0])
	}
	if !model.read.recordLoading {
		t.Fatal("expected stale response to leave record loading state unchanged")
	}
	if model.read.recordTotalCount != 1 {
		t.Fatalf("expected total count to stay unchanged, got %d", model.read.recordTotalCount)
	}
}

func TestUpdate_SaveChangesMsgPendingConfigOpenClearsStateSetsHandoffAndQuits(t *testing.T) {
	// Arrange
	model := &Model{
		staging: testActiveDatabaseStaging(stagingState{
			pendingInserts: []pendingInsertRow{
				{
					values: map[int]stagedEdit{
						0: {Value: dto.StagedValue{Text: "new", Raw: "new"}},
					},
					explicitAuto: map[int]bool{},
				},
			},
		}),
		ui: runtimeUIState{
			pendingLeaveTarget: leaveRuntimeConfig,
		},
	}

	// Act
	_, cmd := model.Update(saveChangesMsg{count: 1})

	// Assert
	if model.hasDirtyEdits() {
		t.Fatal("expected staged state to be cleared after successful save")
	}
	if model.ui.pendingLeaveTarget != leaveRuntimeNone {
		t.Fatalf("expected pending leave target to be cleared after successful save, got %d", model.ui.pendingLeaveTarget)
	}
	if !model.ui.openConfigSelector {
		t.Fatal("expected config-selector handoff to be enabled after successful save")
	}
	if model.ui.statusMessage != "Opening database selector" {
		t.Fatalf("expected config handoff status message, got %q", model.ui.statusMessage)
	}
	if cmd == nil {
		t.Fatal("expected quit command after successful save-and-open flow")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatalf("expected tea.QuitMsg from quit command, got %T", cmd())
	}
}

func TestUpdate_ErrMsgClearsRecordLoadingAndSurfacesErrorStatus(t *testing.T) {
	// Arrange
	model := &Model{
		read: runtimeReadState{
			recordLoading: true,
		},
	}

	// Act
	_, cmd := model.Update(errMsg{err: errors.New("boom")})

	// Assert
	if cmd != nil {
		t.Fatal("expected no follow-up command for runtime error")
	}
	if model.read.recordLoading {
		t.Fatal("expected record loading state to clear after runtime error")
	}
	if model.ui.statusMessage != "Error: boom" {
		t.Fatalf("expected surfaced error status, got %q", model.ui.statusMessage)
	}
}
