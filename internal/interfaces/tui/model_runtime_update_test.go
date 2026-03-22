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

type stubListRecordsUseCase struct {
	lastTableName string
	page          dto.RecordPage
	err           error
}

func (s *stubGetSchemaUseCase) Execute(ctx context.Context, tableName string) (dto.Schema, error) {
	s.lastTableName = tableName
	if s.err != nil {
		return dto.Schema{}, s.err
	}
	return s.schema, nil
}

func (s *stubListRecordsUseCase) Execute(ctx context.Context, tableName string, offset, limit int, filter *dto.Filter, sort *dto.Sort) (dto.RecordPage, error) {
	s.lastTableName = tableName
	if s.err != nil {
		return dto.RecordPage{}, s.err
	}
	return s.page, nil
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

func TestUpdate_SaveChangesMsgPendingDatabaseTransitionClearsStateAndStartsTransition(t *testing.T) {
	// Arrange
	current := DatabaseOption{
		Name:       "primary",
		ConnString: "/tmp/primary.sqlite",
		Source:     DatabaseOptionSourceConfig,
	}
	model := &Model{
		staging: stagingState{
			pendingInserts: []pendingInsertRow{
				{
					values: map[int]stagedEdit{
						0: {Value: dto.StagedValue{Text: "new", Raw: "new"}},
					},
					explicitAuto: map[int]bool{},
				},
			},
		},
		ui: runtimeUIState{
			saveInFlight: true,
			pendingDatabaseTransition: &runtimeDatabaseTransitionRequest{
				Target: runtimeDatabaseTransitionTarget{
					Option: current,
					Kind:   reloadCurrentDatabase,
				},
				Origin: runtimeDatabaseTransitionOriginEditCommand,
			},
		},
		runtimeDatabaseSelectorDeps: runtimeDatabaseSelectorDepsForTest(current),
	}

	// Act
	_, cmd := model.Update(saveChangesMsg{count: 1})

	// Assert
	if model.hasDirtyEdits() {
		t.Fatal("expected staged state to be cleared after successful save")
	}
	if model.ui.pendingDatabaseTransition != nil {
		t.Fatal("expected pending database transition to be cleared after successful save")
	}
	if model.ui.saveInFlight {
		t.Fatal("expected save-in-flight flag to be cleared after successful save")
	}
	if model.exitResult.Action != RuntimeExitActionOpenDatabaseNext {
		t.Fatalf("expected successful save to request runtime reopen, got %v", model.exitResult.Action)
	}
	if cmd == nil {
		t.Fatal("expected save-and-transition flow to quit runtime")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatalf("expected tea.QuitMsg for save-and-transition flow, got %T", cmd())
	}
}

func TestUpdate_SaveChangesMsgPendingQuitAfterSaveClearsStateAndQuits(t *testing.T) {
	// Arrange
	model := &Model{
		staging: stagingState{
			pendingInserts: []pendingInsertRow{
				{
					values: map[int]stagedEdit{
						0: {Value: dto.StagedValue{Text: "new", Raw: "new"}},
					},
					explicitAuto: map[int]bool{},
				},
			},
		},
		ui: runtimeUIState{
			pendingQuitAfterSave: true,
			saveInFlight:         true,
		},
	}

	// Act
	_, cmd := model.Update(saveChangesMsg{count: 1})

	// Assert
	if model.hasDirtyEdits() {
		t.Fatal("expected staged state to be cleared after successful save")
	}
	if model.ui.pendingQuitAfterSave {
		t.Fatal("expected pending quit flag to be cleared after successful save")
	}
	if model.ui.saveInFlight {
		t.Fatal("expected save-in-flight flag to be cleared after successful save")
	}
	if model.ui.statusMessage != "" {
		t.Fatalf("expected immediate quit flow not to overwrite status, got %q", model.ui.statusMessage)
	}
	if cmd == nil {
		t.Fatal("expected quit command after successful save-and-quit flow")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatalf("expected tea.QuitMsg from quit command, got %T", cmd())
	}
}

func TestUpdate_SaveChangesMsgShowsSavedRowsStatusAndReloadsRecords(t *testing.T) {
	// Arrange
	listRecords := &stubListRecordsUseCase{
		page: dto.RecordPage{
			Rows:       []dto.RecordRow{{Values: []string{"1", "alice"}}},
			TotalCount: 1,
		},
	}
	model := &Model{
		ctx:         context.Background(),
		listRecords: listRecords,
		read: runtimeReadState{
			viewMode: ViewRecords,
			tables:   []dto.Table{{Name: "users"}},
		},
		staging: stagingState{
			pendingUpdates: map[string]recordEdits{
				"id=1": {
					changes: map[int]stagedEdit{
						0: {Value: dto.StagedValue{Text: "alice", Raw: "alice"}},
						1: {Value: dto.StagedValue{Text: "alice@example.com", Raw: "alice@example.com"}},
					},
				},
			},
		},
		ui: runtimeUIState{
			saveInFlight: true,
		},
	}

	// Act
	_, cmd := model.Update(saveChangesMsg{count: 1})

	// Assert
	if model.ui.statusMessage != "Affected rows: 1" {
		t.Fatalf("expected affected-row status message, got %q", model.ui.statusMessage)
	}
	if model.ui.saveInFlight {
		t.Fatal("expected save-in-flight flag to be cleared after successful save")
	}
	if cmd == nil {
		t.Fatal("expected records reload command after successful save")
	}

	msg := cmd()
	recordsMsg, ok := msg.(recordsMsg)
	if !ok {
		t.Fatalf("expected recordsMsg from reload command, got %T", msg)
	}
	if listRecords.lastTableName != "users" {
		t.Fatalf("expected records reload for users table, got %q", listRecords.lastTableName)
	}
	if recordsMsg.tableName != "users" {
		t.Fatalf("expected records message for users table, got %q", recordsMsg.tableName)
	}
}

func TestFormatSavedRowsMessage_AllowsZeroAffectedRows(t *testing.T) {
	// Arrange

	// Act
	message := formatSavedRowsMessage(0)

	// Assert
	if message != "Affected rows: 0" {
		t.Fatalf("expected zero-count affected-row message, got %q", message)
	}
}

func TestUpdate_SaveChangesMsgErrorClearsPendingQuitAfterSaveAndPreservesDirtyState(t *testing.T) {
	// Arrange
	model := &Model{
		staging: stagingState{
			pendingInserts: []pendingInsertRow{
				{
					values: map[int]stagedEdit{
						0: {Value: dto.StagedValue{Text: "new", Raw: "new"}},
					},
					explicitAuto: map[int]bool{},
				},
			},
		},
		ui: runtimeUIState{
			pendingQuitAfterSave: true,
			saveInFlight:         true,
		},
	}

	// Act
	_, cmd := model.Update(saveChangesMsg{err: errors.New("boom")})

	// Assert
	if cmd != nil {
		t.Fatal("expected no follow-up command for failed save-and-quit flow")
	}
	if !model.hasDirtyEdits() {
		t.Fatal("expected staged state to stay dirty after failed save")
	}
	if model.ui.pendingQuitAfterSave {
		t.Fatal("expected pending quit flag to be cleared after failed save")
	}
	if model.ui.saveInFlight {
		t.Fatal("expected save-in-flight flag to be cleared after failed save")
	}
	if model.ui.statusMessage != "Error: boom" {
		t.Fatalf("expected surfaced save error, got %q", model.ui.statusMessage)
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

func TestUpdate_TablesMsgStartsSchemaLoadWithoutAnyReloadRestoreState(t *testing.T) {
	// Arrange
	getSchema := &stubGetSchemaUseCase{
		schema: dto.Schema{
			Columns: []dto.SchemaColumn{
				{Name: "id", Type: "INTEGER"},
				{Name: "name", Type: "TEXT"},
			},
		},
	}
	recordsSpy := &spyListRecordsUseCase{}
	model := &Model{
		ctx:         context.Background(),
		getSchema:   getSchema,
		listRecords: recordsSpy,
		read: runtimeReadState{
			focus:           FocusContent,
			viewMode:        ViewRecords,
			recordPageIndex: 2,
			currentFilter: &dto.Filter{
				Column:   "name",
				Operator: dto.Operator{Name: "Equals", Kind: dto.OperatorKindEq, RequiresValue: true},
				Value:    "alice",
			},
			currentSort: &dto.Sort{
				Column:    "id",
				Direction: dto.SortDirectionDesc,
			},
		},
	}

	// Act
	_, cmd := model.Update(tablesMsg{
		tables: []dto.Table{
			{Name: "users"},
			{Name: "orders"},
		},
	})

	// Assert
	if model.read.selectedTable != 0 {
		t.Fatalf("expected restored table selection index 0, got %d", model.read.selectedTable)
	}
	if model.read.focus != FocusContent {
		t.Fatalf("expected focus to stay unchanged, got %v", model.read.focus)
	}
	if model.read.viewMode != ViewRecords {
		t.Fatalf("expected view mode to stay unchanged, got %v", model.read.viewMode)
	}
	if model.read.recordPageIndex != 2 {
		t.Fatalf("expected page index to stay unchanged before any records reset, got %d", model.read.recordPageIndex)
	}
	assertFilterEqual(t, model.read.currentFilter, &dto.Filter{
		Column:   "name",
		Operator: dto.Operator{Name: "Equals", Kind: dto.OperatorKindEq, RequiresValue: true},
		Value:    "alice",
	})
	assertSortEqual(t, model.read.currentSort, &dto.Sort{
		Column:    "id",
		Direction: dto.SortDirectionDesc,
	})
	if model.read.recordLoading {
		t.Fatal("expected records load to stay idle until explicitly requested")
	}
	if model.read.recordRequestID != 0 {
		t.Fatalf("expected record request id to stay unchanged, got %d", model.read.recordRequestID)
	}
	if cmd == nil {
		t.Fatal("expected schema-load command after restore tables step")
	}

	msg := cmd()
	schemaMessage, ok := msg.(schemaMsg)
	if !ok {
		t.Fatalf("expected schemaMsg after restore tables step, got %T", msg)
	}
	if getSchema.lastTableName != "users" {
		t.Fatalf("expected schema load for users table, got %q", getSchema.lastTableName)
	}
	if schemaMessage.tableName != "users" {
		t.Fatalf("expected schema message for users table, got %q", schemaMessage.tableName)
	}
	if recordsSpy.lastFilter != nil || recordsSpy.lastSort != nil {
		t.Fatal("expected records use case not to run while handling tables only")
	}
}

func TestUpdate_SchemaMsgKeepsCurrentBrowseStateWithoutReloadRestoreLogic(t *testing.T) {
	// Arrange
	recordsSpy := &spyListRecordsUseCase{}
	model := &Model{
		ctx:         context.Background(),
		listRecords: recordsSpy,
		read: runtimeReadState{
			tables:          []dto.Table{{Name: "users"}},
			focus:           FocusContent,
			viewMode:        ViewRecords,
			recordPageIndex: 2,
			currentFilter: &dto.Filter{
				Column:   "name",
				Operator: dto.Operator{Name: "Equals", Kind: dto.OperatorKindEq, RequiresValue: true},
				Value:    "alice",
			},
			currentSort: &dto.Sort{
				Column:    "id",
				Direction: dto.SortDirectionDesc,
			},
		},
	}

	// Act
	_, cmd := model.Update(schemaMsg{
		tableName: "users",
		schema: dto.Schema{
			Columns: []dto.SchemaColumn{
				{Name: "id", Type: "INTEGER"},
				{Name: "email", Type: "TEXT"},
			},
		},
	})

	// Assert
	if cmd != nil {
		t.Fatal("expected schema update not to trigger implicit reload")
	}
	if model.read.viewMode != ViewRecords {
		t.Fatalf("expected records view to stay active, got %v", model.read.viewMode)
	}
	assertFilterEqual(t, model.read.currentFilter, &dto.Filter{
		Column:   "name",
		Operator: dto.Operator{Name: "Equals", Kind: dto.OperatorKindEq, RequiresValue: true},
		Value:    "alice",
	})
	assertSortEqual(t, model.read.currentSort, &dto.Sort{
		Column:    "id",
		Direction: dto.SortDirectionDesc,
	})
	if recordsSpy.lastFilter != nil || recordsSpy.lastSort != nil {
		t.Fatal("expected records use case not to run while handling schema only")
	}
}

func TestUpdate_SaveChangesMsgPendingDatabaseTransitionPreservesCurrentBrowseStateUntilExit(t *testing.T) {
	// Arrange
	current := DatabaseOption{
		Name:       "primary",
		ConnString: "/tmp/primary.sqlite",
		Source:     DatabaseOptionSourceConfig,
	}
	model := &Model{
		staging: stagingState{
			pendingInserts: []pendingInsertRow{
				{
					values: map[int]stagedEdit{
						0: {Value: dto.StagedValue{Text: "new", Raw: "new"}},
					},
					explicitAuto: map[int]bool{},
				},
			},
		},
		read: runtimeReadState{
			focus:            FocusContent,
			viewMode:         ViewRecords,
			recordPageIndex:  2,
			recordSelection:  3,
			recordColumn:     1,
			recordFieldFocus: true,
			currentFilter: &dto.Filter{
				Column:   "name",
				Operator: dto.Operator{Name: "Equals", Kind: dto.OperatorKindEq, RequiresValue: true},
				Value:    "alice",
			},
			currentSort: &dto.Sort{
				Column:    "id",
				Direction: dto.SortDirectionDesc,
			},
		},
		ui: runtimeUIState{
			saveInFlight: true,
			pendingDatabaseTransition: &runtimeDatabaseTransitionRequest{
				Target: runtimeDatabaseTransitionTarget{
					Option: current,
					Kind:   reloadCurrentDatabase,
				},
				Origin: runtimeDatabaseTransitionOriginConfigSelector,
			},
		},
	}

	// Act
	_, cmd := model.Update(saveChangesMsg{count: 1})

	// Assert
	if cmd == nil {
		t.Fatal("expected save-and-reopen flow to quit runtime")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatalf("expected tea.QuitMsg for save-and-reopen flow, got %T", cmd())
	}
	if model.read.focus != FocusContent {
		t.Fatalf("expected focus to stay unchanged until runtime exits, got %v", model.read.focus)
	}
	if model.read.viewMode != ViewRecords {
		t.Fatalf("expected view mode to stay unchanged until runtime exits, got %v", model.read.viewMode)
	}
	assertFilterEqual(t, model.read.currentFilter, &dto.Filter{
		Column:   "name",
		Operator: dto.Operator{Name: "Equals", Kind: dto.OperatorKindEq, RequiresValue: true},
		Value:    "alice",
	})
	assertSortEqual(t, model.read.currentSort, &dto.Sort{
		Column:    "id",
		Direction: dto.SortDirectionDesc,
	})
}

func TestUpdate_ErrMsgIgnoresStaleBundleToken(t *testing.T) {
	// Arrange
	model := &Model{
		runtimeBundleToken: 2,
		read: runtimeReadState{
			recordLoading: true,
		},
		ui: runtimeUIState{
			statusMessage: "fresh status",
		},
	}

	// Act
	_, cmd := model.Update(errMsg{bundleToken: 1, err: errors.New("stale boom")})

	// Assert
	if cmd != nil {
		t.Fatal("expected no follow-up command for stale bundle error")
	}
	if !model.read.recordLoading {
		t.Fatal("expected stale bundle error to leave record loading state unchanged")
	}
	if model.ui.statusMessage != "fresh status" {
		t.Fatalf("expected stale bundle error to be ignored, got %q", model.ui.statusMessage)
	}
}

func assertFilterEqual(t *testing.T, actual, expected *dto.Filter) {
	t.Helper()
	if expected == nil {
		if actual != nil {
			t.Fatalf("expected nil filter, got %+v", actual)
		}
		return
	}
	if actual == nil {
		t.Fatal("expected non-nil filter")
	}
	if actual.Column != expected.Column {
		t.Fatalf("expected filter column %q, got %q", expected.Column, actual.Column)
	}
	if actual.Operator != expected.Operator {
		t.Fatalf("expected filter operator %+v, got %+v", expected.Operator, actual.Operator)
	}
	if actual.Value != expected.Value {
		t.Fatalf("expected filter value %q, got %q", expected.Value, actual.Value)
	}
}

func assertSortEqual(t *testing.T, actual, expected *dto.Sort) {
	t.Helper()
	if expected == nil {
		if actual != nil {
			t.Fatalf("expected nil sort, got %+v", actual)
		}
		return
	}
	if actual == nil {
		t.Fatal("expected non-nil sort")
	}
	if actual.Column != expected.Column || actual.Direction != expected.Direction {
		t.Fatalf("expected sort %+v, got %+v", expected, actual)
	}
}
