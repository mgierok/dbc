package tui

import (
	"context"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/dto"
)

func TestConfirmSaveChanges_SubmitsAllDirtyTableChanges_FromSaveFlow(t *testing.T) {
	saveChanges := &spySaveChangesUseCase{}
	model := &Model{
		ctx:         context.Background(),
		saveChanges: saveChanges,
		read: runtimeReadState{
			tables:        []dto.Table{{Name: "users"}, {Name: "orders"}},
			selectedTable: 1,
		},
		staging: testDatabaseStaging(map[string]stagingState{
			"users": {
				schema: dto.Schema{
					Columns: []dto.SchemaColumn{
						{Name: "id", Type: "INTEGER", PrimaryKey: true, AutoIncrement: true},
						{Name: "name", Type: "TEXT", Nullable: false},
					},
				},
				pendingInserts: []pendingInsertRow{
					{
						values: map[int]stagedEdit{
							0: {Value: dto.StagedValue{Text: "", Raw: ""}},
							1: {Value: dto.StagedValue{Text: "new", Raw: "new"}},
						},
						explicitAuto: map[int]bool{},
					},
				},
			},
			"orders": {
				schema: dto.Schema{
					Columns: []dto.SchemaColumn{
						{Name: "id", Type: "INTEGER", PrimaryKey: true},
						{Name: "status", Type: "TEXT", Nullable: false},
					},
				},
				pendingUpdates: map[string]recordEdits{
					"id=7": {
						identity: dto.RecordIdentity{
							Keys: []dto.RecordIdentityKey{{Column: "id", Value: dto.StagedValue{Text: "7", Raw: int64(7)}}},
						},
						changes: map[int]stagedEdit{
							1: {Value: dto.StagedValue{Text: "paid", Raw: "paid"}},
						},
					},
				},
			},
		}),
	}

	_, cmd := model.confirmSaveChanges()
	if cmd == nil {
		t.Fatal("expected save command")
	}
	msg := cmd()

	result, ok := msg.(saveChangesMsg)
	if !ok {
		t.Fatalf("expected saveChangesMsg, got %T", msg)
	}
	if result.err != nil {
		t.Fatalf("expected no error, got %v", result.err)
	}
	if len(saveChanges.lastChanges) != 2 {
		t.Fatalf("expected two dirty table batches, got %d", len(saveChanges.lastChanges))
	}
}

func TestSetTableSelection_WithDirtyStateSwitchesWithoutPopupAndPreservesStaging(t *testing.T) {
	model := &Model{
		read: runtimeReadState{
			tables:        []dto.Table{{Name: "users"}, {Name: "orders"}},
			selectedTable: 0,
		},
		staging: testDatabaseStaging(map[string]stagingState{
			"users": {
				pendingInserts: []pendingInsertRow{{}},
			},
			"orders": {
				pendingInserts: []pendingInsertRow{{}},
			},
		}),
	}

	model.setTableSelection(1)

	if model.overlay.confirmPopup.active {
		t.Fatal("expected no switch-table popup")
	}
	if model.read.selectedTable != 1 {
		t.Fatalf("expected selected table 1, got %d", model.read.selectedTable)
	}
	if model.totalRecordRows() != 1 {
		t.Fatalf("expected orders staging to stay visible, got %d rows", model.totalRecordRows())
	}

	model.setTableSelection(0)

	if model.totalRecordRows() != 1 {
		t.Fatalf("expected users staging to be restored, got %d rows", model.totalRecordRows())
	}
}

func TestHandleConfirmPopupKey_DirtyConfigSaveStartsDatabaseSaveFlow(t *testing.T) {
	saveChanges := &spySaveChangesUseCase{}
	model := &Model{
		saveChanges: saveChanges,
		read: runtimeReadState{
			viewMode:      ViewRecords,
			focus:         FocusContent,
			tables:        []dto.Table{{Name: "users"}},
			selectedTable: 0,
		},
		staging: testDatabaseStaging(map[string]stagingState{
			"users": {
				schema: dto.Schema{
					Columns: []dto.SchemaColumn{
						{Name: "id", Type: "INTEGER", PrimaryKey: true},
						{Name: "name", Type: "TEXT", Nullable: false},
					},
				},
				pendingInserts: []pendingInsertRow{
					{
						values: map[int]stagedEdit{
							0: {Value: dto.StagedValue{Text: "1", Raw: "1"}},
							1: {Value: dto.StagedValue{Text: "bob", Raw: "bob"}},
						},
						explicitAuto: map[int]bool{},
					},
				},
			},
		}),
		overlay: runtimeOverlayState{
			confirmPopup: confirmPopup{
				active: true,
				action: confirmLeaveSave,
			},
		},
		ui: runtimeUIState{
			pendingLeaveTarget: leaveRuntimeConfig,
		},
	}

	_, cmd := model.handleConfirmPopupKey(tea.KeyMsg{Type: tea.KeyEnter})

	if cmd == nil {
		t.Fatal("expected save command to be returned")
	}
	if model.ui.pendingLeaveTarget != leaveRuntimeConfig {
		t.Fatalf("expected leave target to stay pending until save result, got %d", model.ui.pendingLeaveTarget)
	}
}

func TestToggleDeleteSelection_AfterSuccessfulSaveBuildsDeleteOnlyChangesForSameTable(t *testing.T) {
	model := &Model{
		read: runtimeReadState{
			viewMode:      ViewRecords,
			focus:         FocusContent,
			tables:        []dto.Table{{Name: "users"}},
			selectedTable: 0,
			records: []dto.RecordRow{
				{Values: []string{"1", "alice"}},
			},
			schema: dto.Schema{
				Columns: []dto.SchemaColumn{
					{Name: "id", Type: "INTEGER", PrimaryKey: true},
					{Name: "name", Type: "TEXT", Nullable: false},
				},
			},
		},
		staging: testDatabaseStaging(map[string]stagingState{
			"users": {
				schema: dto.Schema{
					Columns: []dto.SchemaColumn{
						{Name: "id", Type: "INTEGER", PrimaryKey: true},
						{Name: "name", Type: "TEXT", Nullable: false},
					},
				},
				pendingUpdates: map[string]recordEdits{
					"id=1": {
						identity: dto.RecordIdentity{
							Keys: []dto.RecordIdentityKey{{Column: "id", Value: dto.StagedValue{Text: "1", Raw: int64(1)}}},
						},
						changes: map[int]stagedEdit{
							1: {Value: dto.StagedValue{Text: "ally", Raw: "ally"}},
						},
					},
				},
			},
		}),
	}

	_, cmd := model.Update(saveChangesMsg{count: 1})
	if cmd == nil {
		t.Fatal("expected records reload command after successful save")
	}
	if model.hasDirtyEdits() {
		t.Fatal("expected staged state to be cleared after successful save")
	}

	model.Update(recordsMsg{
		tableName: "users",
		requestID: model.read.recordRequestID,
		page: dto.RecordPage{
			Rows: []dto.RecordRow{
				{Values: []string{"1", "alice"}},
			},
			TotalCount: 1,
		},
	})

	model.toggleDeleteSelection()

	changes, err := model.buildDatabaseChanges()
	if err != nil {
		t.Fatalf("expected delete-only changes to build after save, got %v", err)
	}
	if len(changes) != 1 {
		t.Fatalf("expected one table change batch, got %d", len(changes))
	}
	if changes[0].TableName != "users" {
		t.Fatalf("expected users table batch, got %q", changes[0].TableName)
	}
	if len(changes[0].Changes.Deletes) != 1 {
		t.Fatalf("expected one delete change, got %d", len(changes[0].Changes.Deletes))
	}
}
