package tui

import (
	"context"
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/dto"
	"github.com/mgierok/dbc/internal/application/usecase"
	domainmodel "github.com/mgierok/dbc/internal/domain/model"
)

type stubGetSchemaUseCase struct {
	lastTableName string
	schema        dto.Schema
	err           error
}

type fakeRuntimeSwitchEngine struct {
	tables               []domainmodel.Table
	schema               domainmodel.Schema
	page                 domainmodel.RecordPage
	lastSchemaTableName  string
	lastRecordsTableName string
	listTablesErr        error
	getSchemaErr         error
	listRecordsErr       error
}

type stubListRecordsUseCase struct {
	lastTableName string
	page          dto.RecordPage
	err           error
}

func (f *fakeRuntimeSwitchEngine) ListTables(ctx context.Context) ([]domainmodel.Table, error) {
	if f.listTablesErr != nil {
		return nil, f.listTablesErr
	}
	return append([]domainmodel.Table(nil), f.tables...), nil
}

func (f *fakeRuntimeSwitchEngine) GetSchema(ctx context.Context, tableName string) (domainmodel.Schema, error) {
	f.lastSchemaTableName = tableName
	if f.getSchemaErr != nil {
		return domainmodel.Schema{}, f.getSchemaErr
	}
	return f.schema, nil
}

func (f *fakeRuntimeSwitchEngine) ListRecords(ctx context.Context, tableName string, offset, limit int, filter *domainmodel.Filter, sort *domainmodel.Sort) (domainmodel.RecordPage, error) {
	f.lastRecordsTableName = tableName
	if f.listRecordsErr != nil {
		return domainmodel.RecordPage{}, f.listRecordsErr
	}
	return f.page, nil
}

func (f *fakeRuntimeSwitchEngine) ListOperators(ctx context.Context, columnType string) ([]domainmodel.Operator, error) {
	return nil, nil
}

func (f *fakeRuntimeSwitchEngine) ApplyRecordChanges(ctx context.Context, tableName string, changes domainmodel.TableChanges) (int, error) {
	return 0, nil
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
		runtimeDatabaseSelectorDeps: runtimeDatabaseSelectorDepsForTest(current, &stubRuntimeDatabaseSwitcher{}),
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
	if !model.ui.runtimeSwitchInFlight {
		t.Fatal("expected successful save to continue pending database transition")
	}
	if cmd != nil {
		if _, ok := cmd().(tea.QuitMsg); ok {
			t.Fatalf("expected runtime to stay active after successful save-and-transition flow, got %T", cmd())
		}
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

func TestUpdate_TablesMsgPendingReloadRestoreDefersRecordLoadUntilSchema(t *testing.T) {
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
		ui: runtimeUIState{
			pendingDatabaseReloadRestore: cloneRuntimeDatabaseReloadRestoreState(runtimeDatabaseTransitionSnapshot{
				SelectedTableName: "users",
				Focus:             FocusContent,
				ViewMode:          ViewRecords,
				Filter: &dto.Filter{
					Column:   "name",
					Operator: dto.Operator{Name: "Equals", Kind: dto.OperatorKindEq, RequiresValue: true},
					Value:    "alice",
				},
				Sort: &dto.Sort{
					Column:    "id",
					Direction: dto.SortDirectionDesc,
				},
				PageIndex: 2,
			}),
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
		t.Fatalf("expected restored focus %v, got %v", FocusContent, model.read.focus)
	}
	if model.read.viewMode != ViewRecords {
		t.Fatalf("expected restored view mode %v, got %v", ViewRecords, model.read.viewMode)
	}
	if model.read.recordPageIndex != 2 {
		t.Fatalf("expected restored page index 2, got %d", model.read.recordPageIndex)
	}
	if model.read.currentFilter != nil {
		t.Fatalf("expected filter restore to wait for schema validation, got %+v", model.read.currentFilter)
	}
	if model.read.currentSort != nil {
		t.Fatalf("expected sort restore to wait for schema validation, got %+v", model.read.currentSort)
	}
	if model.read.recordLoading {
		t.Fatal("expected records load to stay idle until schema arrives")
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
		t.Fatal("expected records use case not to run before schema validation")
	}
}

func TestUpdate_SchemaMsgSanitizesPendingReloadRestoreInRecordsView(t *testing.T) {
	tests := []struct {
		name           string
		schemaColumns  []dto.SchemaColumn
		expectedFilter *dto.Filter
		expectedSort   *dto.Sort
		expectedStatus string
	}{
		{
			name: "invalid filter preserves sort",
			schemaColumns: []dto.SchemaColumn{
				{Name: "id", Type: "INTEGER"},
				{Name: "email", Type: "TEXT"},
			},
			expectedFilter: nil,
			expectedSort: &dto.Sort{
				Column:    "id",
				Direction: dto.SortDirectionDesc,
			},
			expectedStatus: "Database reloaded; filter cleared after schema change.",
		},
		{
			name: "invalid sort preserves filter",
			schemaColumns: []dto.SchemaColumn{
				{Name: "name", Type: "TEXT"},
				{Name: "email", Type: "TEXT"},
			},
			expectedFilter: &dto.Filter{
				Column:   "name",
				Operator: dto.Operator{Name: "Equals", Kind: dto.OperatorKindEq, RequiresValue: true},
				Value:    "alice",
			},
			expectedSort:   nil,
			expectedStatus: "Database reloaded; sort cleared after schema change.",
		},
		{
			name: "invalid filter and sort clear both",
			schemaColumns: []dto.SchemaColumn{
				{Name: "email", Type: "TEXT"},
			},
			expectedFilter: nil,
			expectedSort:   nil,
			expectedStatus: "Database reloaded; filter and sort cleared after schema change.",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			recordsSpy := &spyListRecordsUseCase{
				page: dto.RecordPage{
					Rows:       []dto.RecordRow{{Values: []string{"1", "alice@example.com"}}},
					TotalCount: 41,
				},
			}
			model := &Model{
				ctx:         context.Background(),
				listRecords: recordsSpy,
				read: runtimeReadState{
					tables:          []dto.Table{{Name: "users"}},
					focus:           FocusContent,
					viewMode:        ViewRecords,
					recordPageIndex: 2,
				},
				ui: runtimeUIState{
					statusMessage: "Database reloaded.",
					pendingDatabaseReloadRestore: cloneRuntimeDatabaseReloadRestoreState(runtimeDatabaseTransitionSnapshot{
						SelectedTableName: "users",
						Focus:             FocusContent,
						ViewMode:          ViewRecords,
						Filter: &dto.Filter{
							Column:   "name",
							Operator: dto.Operator{Name: "Equals", Kind: dto.OperatorKindEq, RequiresValue: true},
							Value:    "alice",
						},
						Sort: &dto.Sort{
							Column:    "id",
							Direction: dto.SortDirectionDesc,
						},
						PageIndex: 2,
					}),
				},
			}

			// Act
			_, cmd := model.Update(schemaMsg{
				tableName: "users",
				schema: dto.Schema{
					Columns: tc.schemaColumns,
				},
			})

			// Assert
			assertFilterEqual(t, model.read.currentFilter, tc.expectedFilter)
			assertSortEqual(t, model.read.currentSort, tc.expectedSort)
			if model.ui.statusMessage != tc.expectedStatus {
				t.Fatalf("expected reload status %q, got %q", tc.expectedStatus, model.ui.statusMessage)
			}
			if model.ui.pendingDatabaseReloadRestore == nil {
				t.Fatal("expected reload restore state to stay active until records response")
			}
			if model.ui.pendingDatabaseReloadRestore.requestedPageIndex != 2 {
				t.Fatalf("expected requested page index 2 to survive until records response, got %d", model.ui.pendingDatabaseReloadRestore.requestedPageIndex)
			}
			if cmd == nil {
				t.Fatal("expected records-load command after schema sanitization in records view")
			}

			msg := cmd()
			recordsMessage, ok := msg.(recordsMsg)
			if !ok {
				t.Fatalf("expected recordsMsg after schema sanitization, got %T", msg)
			}
			if recordsMessage.tableName != "users" {
				t.Fatalf("expected records reload for users table, got %q", recordsMessage.tableName)
			}
			assertFilterEqual(t, recordsSpy.lastFilter, tc.expectedFilter)
			assertSortEqual(t, recordsSpy.lastSort, tc.expectedSort)
			if recordsSpy.lastRecordsOffset != 40 {
				t.Fatalf("expected records reload offset 40 for restored page 3, got %d", recordsSpy.lastRecordsOffset)
			}
			if recordsSpy.lastRecordsLimit != defaultRecordPageLimit {
				t.Fatalf("expected default records limit %d, got %d", defaultRecordPageLimit, recordsSpy.lastRecordsLimit)
			}
		})
	}
}

func TestUpdate_SchemaMsgSanitizesPendingReloadRestoreOutsideRecordsViewWithoutLoadingRecords(t *testing.T) {
	// Arrange
	recordsSpy := &spyListRecordsUseCase{}
	model := &Model{
		ctx:         context.Background(),
		listRecords: recordsSpy,
		read: runtimeReadState{
			tables:   []dto.Table{{Name: "users"}},
			focus:    FocusContent,
			viewMode: ViewSchema,
		},
		ui: runtimeUIState{
			statusMessage: "Database reloaded.",
			pendingDatabaseReloadRestore: cloneRuntimeDatabaseReloadRestoreState(runtimeDatabaseTransitionSnapshot{
				SelectedTableName: "users",
				Focus:             FocusContent,
				ViewMode:          ViewSchema,
				Filter: &dto.Filter{
					Column:   "name",
					Operator: dto.Operator{Name: "Equals", Kind: dto.OperatorKindEq, RequiresValue: true},
					Value:    "alice",
				},
				Sort: &dto.Sort{
					Column:    "id",
					Direction: dto.SortDirectionDesc,
				},
			}),
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
		t.Fatal("expected no records-load command outside records view")
	}
	if model.read.viewMode != ViewSchema {
		t.Fatalf("expected schema view to stay active, got %v", model.read.viewMode)
	}
	if model.read.currentFilter != nil {
		t.Fatalf("expected invalid filter to be cleared outside records view, got %+v", model.read.currentFilter)
	}
	if model.read.currentSort == nil {
		t.Fatal("expected valid sort to be preserved outside records view")
	}
	if model.read.currentSort.Column != "id" || model.read.currentSort.Direction != dto.SortDirectionDesc {
		t.Fatalf("expected preserved sort id DESC, got %+v", model.read.currentSort)
	}
	if model.ui.pendingDatabaseReloadRestore != nil {
		t.Fatal("expected reload restore state to finish after schema sanitization outside records view")
	}
	if model.ui.statusMessage != "Database reloaded; filter cleared after schema change." {
		t.Fatalf("expected sanitized reload status, got %q", model.ui.statusMessage)
	}
	if recordsSpy.lastFilter != nil || recordsSpy.lastSort != nil {
		t.Fatal("expected records use case not to run outside records view")
	}
}

func TestUpdate_RecordsMsgKeepsReloadPageCorrectionAfterSchemaSanitization(t *testing.T) {
	// Arrange
	recordsSpy := &spyListRecordsUseCase{
		page: dto.RecordPage{
			Rows:       []dto.RecordRow{{Values: []string{"1", "alice"}}},
			TotalCount: 1,
		},
	}
	model := &Model{
		ctx:         context.Background(),
		listRecords: recordsSpy,
		read: runtimeReadState{
			tables:          []dto.Table{{Name: "users"}},
			focus:           FocusContent,
			viewMode:        ViewRecords,
			recordPageIndex: 2,
		},
		ui: runtimeUIState{
			statusMessage: "Database reloaded.",
			pendingDatabaseReloadRestore: cloneRuntimeDatabaseReloadRestoreState(runtimeDatabaseTransitionSnapshot{
				SelectedTableName: "users",
				Focus:             FocusContent,
				ViewMode:          ViewRecords,
				PageIndex:         2,
			}),
		},
	}

	// Act
	_, schemaFollowUpCmd := model.Update(schemaMsg{
		tableName: "users",
		schema: dto.Schema{
			Columns: []dto.SchemaColumn{
				{Name: "id", Type: "INTEGER"},
			},
		},
	})
	if schemaFollowUpCmd == nil {
		t.Fatal("expected initial records load after schema restore")
	}
	firstRecordsMsg, ok := schemaFollowUpCmd().(recordsMsg)
	if !ok {
		t.Fatalf("expected recordsMsg after schema restore, got %T", schemaFollowUpCmd())
	}

	updated, correctionCmd := model.Update(firstRecordsMsg)
	model = updated.(*Model)

	// Assert
	if correctionCmd == nil {
		t.Fatal("expected corrective records reload when restored page is now out of range")
	}
	if model.read.recordPageIndex != 0 {
		t.Fatalf("expected page index to clamp to 0 after first records load, got %d", model.read.recordPageIndex)
	}
	if model.ui.pendingDatabaseReloadRestore == nil {
		t.Fatal("expected reload restore state to stay active for corrective reload")
	}
	if model.ui.pendingDatabaseReloadRestore.requestedPageIndex != -1 {
		t.Fatalf("expected requested page index to switch to corrective mode, got %d", model.ui.pendingDatabaseReloadRestore.requestedPageIndex)
	}

	secondRecordsMsg, ok := correctionCmd().(recordsMsg)
	if !ok {
		t.Fatalf("expected recordsMsg from corrective reload, got %T", correctionCmd())
	}
	updated, _ = model.Update(secondRecordsMsg)
	model = updated.(*Model)

	if model.ui.pendingDatabaseReloadRestore != nil {
		t.Fatal("expected reload restore state to clear after corrective records reload")
	}
	if model.read.recordPageIndex != 0 {
		t.Fatalf("expected corrected page index 0 after reload completion, got %d", model.read.recordPageIndex)
	}
}

func TestUpdate_RecordsMsgBeforeSchemaFinalizeKeepsPendingReloadRestore(t *testing.T) {
	// Arrange
	model := &Model{
		ctx: context.Background(),
		read: runtimeReadState{
			tables:   []dto.Table{{Name: "users"}},
			focus:    FocusContent,
			viewMode: ViewSchema,
		},
		ui: runtimeUIState{
			pendingDatabaseReloadRestore: cloneRuntimeDatabaseReloadRestoreState(runtimeDatabaseTransitionSnapshot{
				SelectedTableName: "users",
				Focus:             FocusContent,
				ViewMode:          ViewSchema,
				Filter: &dto.Filter{
					Column:   "name",
					Operator: dto.Operator{Name: "Equals", Kind: dto.OperatorKindEq, RequiresValue: true},
					Value:    "alice",
				},
				Sort: &dto.Sort{
					Column:    "id",
					Direction: dto.SortDirectionDesc,
				},
			}),
		},
	}
	if _, cmd := model.switchToRecords(); cmd == nil {
		t.Fatal("expected switch to records to start async loads")
	}

	// Act
	_, cmd := model.Update(recordsMsg{
		tableName: "users",
		requestID: model.read.recordRequestID,
		page: dto.RecordPage{
			Rows:       []dto.RecordRow{{Values: []string{"1", "alice"}}},
			TotalCount: 1,
		},
	})

	// Assert
	if cmd != nil {
		t.Fatal("expected no corrective reload before schema finalization")
	}
	if model.ui.pendingDatabaseReloadRestore == nil {
		t.Fatal("expected reload restore to survive early records response")
	}
	if model.read.currentFilter != nil {
		t.Fatalf("expected filter restore to wait for schema finalization, got %+v", model.read.currentFilter)
	}
	if model.read.currentSort != nil {
		t.Fatalf("expected sort restore to wait for schema finalization, got %+v", model.read.currentSort)
	}
	if model.read.recordLoading {
		t.Fatal("expected early records response to clear loading state")
	}
}

func TestUpdate_SchemaMsgAfterSchemaToRecordsTransitionStartsReloadUsingCurrentViewMode(t *testing.T) {
	// Arrange
	recordsSpy := &spyListRecordsUseCase{
		page: dto.RecordPage{
			Rows:       []dto.RecordRow{{Values: []string{"1", "alice"}}},
			TotalCount: 1,
		},
	}
	model := &Model{
		ctx:         context.Background(),
		listRecords: recordsSpy,
		read: runtimeReadState{
			tables:   []dto.Table{{Name: "users"}},
			focus:    FocusContent,
			viewMode: ViewSchema,
		},
		ui: runtimeUIState{
			statusMessage: "Database reloaded.",
			pendingDatabaseReloadRestore: cloneRuntimeDatabaseReloadRestoreState(runtimeDatabaseTransitionSnapshot{
				SelectedTableName: "users",
				Focus:             FocusContent,
				ViewMode:          ViewSchema,
				Filter: &dto.Filter{
					Column:   "name",
					Operator: dto.Operator{Name: "Equals", Kind: dto.OperatorKindEq, RequiresValue: true},
					Value:    "alice",
				},
				Sort: &dto.Sort{
					Column:    "id",
					Direction: dto.SortDirectionDesc,
				},
			}),
		},
	}
	if _, cmd := model.switchToRecords(); cmd == nil {
		t.Fatal("expected switch to records to start async loads")
	}
	if _, cmd := model.Update(recordsMsg{
		tableName: "users",
		requestID: model.read.recordRequestID,
		page: dto.RecordPage{
			Rows:       []dto.RecordRow{{Values: []string{"stale"}}},
			TotalCount: 1,
		},
	}); cmd != nil {
		t.Fatal("expected early records response not to trigger corrective reload")
	}

	// Act
	_, cmd := model.Update(schemaMsg{
		tableName: "users",
		schema: dto.Schema{
			Columns: []dto.SchemaColumn{
				{Name: "id", Type: "INTEGER"},
				{Name: "name", Type: "TEXT"},
			},
		},
	})

	// Assert
	if cmd == nil {
		t.Fatal("expected schema finalization to start records reload in current records view")
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
	if model.ui.pendingDatabaseReloadRestore == nil {
		t.Fatal("expected restore state to stay active until the follow-up records reload completes")
	}

	msg := cmd()
	recordsMessage, ok := msg.(recordsMsg)
	if !ok {
		t.Fatalf("expected recordsMsg after schema finalization, got %T", msg)
	}
	if recordsMessage.requestID != 2 {
		t.Fatalf("expected follow-up records reload request id 2, got %d", recordsMessage.requestID)
	}
	assertFilterEqual(t, recordsSpy.lastFilter, model.read.currentFilter)
	assertSortEqual(t, recordsSpy.lastSort, model.read.currentSort)
	if recordsSpy.lastRecordsOffset != 0 {
		t.Fatalf("expected follow-up reload offset 0, got %d", recordsSpy.lastRecordsOffset)
	}
}

func TestUpdate_SchemaMsgStartsSanitizedReloadWhileEarlierRecordsLoadIsStillInFlight(t *testing.T) {
	// Arrange
	recordsSpy := &spyListRecordsUseCase{
		page: dto.RecordPage{
			Rows:       []dto.RecordRow{{Values: []string{"1", "alice"}}},
			TotalCount: 1,
		},
	}
	model := &Model{
		ctx:         context.Background(),
		listRecords: recordsSpy,
		read: runtimeReadState{
			tables:   []dto.Table{{Name: "users"}},
			focus:    FocusContent,
			viewMode: ViewSchema,
		},
		ui: runtimeUIState{
			statusMessage: "Database reloaded.",
			pendingDatabaseReloadRestore: cloneRuntimeDatabaseReloadRestoreState(runtimeDatabaseTransitionSnapshot{
				SelectedTableName: "users",
				Focus:             FocusContent,
				ViewMode:          ViewSchema,
				Filter: &dto.Filter{
					Column:   "name",
					Operator: dto.Operator{Name: "Equals", Kind: dto.OperatorKindEq, RequiresValue: true},
					Value:    "alice",
				},
			}),
		},
	}
	if _, cmd := model.switchToRecords(); cmd == nil {
		t.Fatal("expected switch to records to start async loads")
	}
	earlierRequestID := model.read.recordRequestID

	// Act
	_, cmd := model.Update(schemaMsg{
		tableName: "users",
		schema: dto.Schema{
			Columns: []dto.SchemaColumn{
				{Name: "id", Type: "INTEGER"},
				{Name: "name", Type: "TEXT"},
			},
		},
	})

	// Assert
	if cmd == nil {
		t.Fatal("expected schema finalization to force a sanitized follow-up records reload")
	}
	if model.read.recordRequestID != earlierRequestID+1 {
		t.Fatalf("expected sanitized reload to bump request id from %d to %d, got %d", earlierRequestID, earlierRequestID+1, model.read.recordRequestID)
	}

	staleRecordsMsg := recordsMsg{
		tableName: "users",
		requestID: earlierRequestID,
		page: dto.RecordPage{
			Rows:       []dto.RecordRow{{Values: []string{"stale"}}},
			TotalCount: 1,
		},
	}
	if _, staleCmd := model.Update(staleRecordsMsg); staleCmd != nil {
		t.Fatal("expected stale earlier records response to be ignored")
	}
	if model.ui.pendingDatabaseReloadRestore == nil {
		t.Fatal("expected restore state to survive until sanitized reload completes")
	}

	msg := cmd()
	recordsMessage, ok := msg.(recordsMsg)
	if !ok {
		t.Fatalf("expected recordsMsg after sanitized reload, got %T", msg)
	}
	if recordsMessage.requestID != model.read.recordRequestID {
		t.Fatalf("expected sanitized reload response to use active request id %d, got %d", model.read.recordRequestID, recordsMessage.requestID)
	}
	assertFilterEqual(t, recordsSpy.lastFilter, model.read.currentFilter)
}

func TestUpdate_RuntimeSwitchIgnoresLateMessagesFromPreviousBundle(t *testing.T) {
	// Arrange
	runtimeSession := &RuntimeSessionState{RecordsPageLimit: 55, nextRuntimeBundleToken: 1}
	previousBundleCanceled := false
	previousCloseCalled := false
	replacementEngine := &fakeRuntimeSwitchEngine{
		tables: []domainmodel.Table{{Name: "fresh_users"}},
	}
	model := &Model{
		ctx:                context.Background(),
		runtimeSession:     runtimeSession,
		runtimeBundleToken: 1,
		runtimeBundleCancel: func() {
			previousBundleCanceled = true
		},
		runtimeClose: func() {
			previousCloseCalled = true
		},
	}
	updated, initCmd := model.Update(runtimeDatabaseSwitchCompletedMsg{
		deps: RuntimeRunDeps{
			ListTables: usecase.NewListTables(replacementEngine),
		},
	})
	if initCmd == nil {
		t.Fatal("expected replacement runtime initialization command after successful switch")
	}
	runtimeModel := updated.(*Model)
	runtimeModel.read.tables = []dto.Table{{Name: "fresh_users"}}
	runtimeModel.read.selectedTable = 0
	runtimeModel.read.schema = dto.Schema{
		Columns: []dto.SchemaColumn{{Name: "fresh_id", Type: "INTEGER"}},
	}
	runtimeModel.read.records = []dto.RecordRow{{Values: []string{"fresh"}}}
	runtimeModel.read.recordRequestID = 4
	runtimeModel.read.recordLoading = true
	runtimeModel.ui.statusMessage = "fresh status"
	staleBundleToken := 1

	// Act
	runtimeModel.Update(tablesMsg{
		bundleToken: staleBundleToken,
		tables:      []dto.Table{{Name: "stale_users"}},
	})
	runtimeModel.Update(schemaMsg{
		bundleToken: staleBundleToken,
		tableName:   "fresh_users",
		schema: dto.Schema{
			Columns: []dto.SchemaColumn{{Name: "stale_id", Type: "TEXT"}},
		},
	})
	runtimeModel.Update(recordsMsg{
		bundleToken: staleBundleToken,
		tableName:   "fresh_users",
		requestID:   4,
		page: dto.RecordPage{
			Rows:       []dto.RecordRow{{Values: []string{"stale"}}},
			TotalCount: 99,
		},
	})
	runtimeModel.Update(errMsg{bundleToken: staleBundleToken, err: errors.New("stale boom")})

	// Assert
	if !previousBundleCanceled {
		t.Fatal("expected previous runtime bundle context to be canceled after switch")
	}
	if !previousCloseCalled {
		t.Fatal("expected previous runtime resources to close after switch")
	}
	if runtimeModel.runtimeBundleToken == staleBundleToken {
		t.Fatalf("expected replacement runtime bundle token to differ from previous token %d", staleBundleToken)
	}
	if len(runtimeModel.read.tables) != 1 || runtimeModel.read.tables[0].Name != "fresh_users" {
		t.Fatalf("expected stale tables message to be ignored, got %+v", runtimeModel.read.tables)
	}
	if len(runtimeModel.read.schema.Columns) != 1 || runtimeModel.read.schema.Columns[0].Name != "fresh_id" {
		t.Fatalf("expected stale schema message to be ignored, got %+v", runtimeModel.read.schema.Columns)
	}
	if len(runtimeModel.read.records) != 1 || runtimeModel.read.records[0].Values[0] != "fresh" {
		t.Fatalf("expected stale records message to be ignored, got %+v", runtimeModel.read.records)
	}
	if !runtimeModel.read.recordLoading {
		t.Fatal("expected stale bundle error to leave active record-loading state unchanged")
	}
	if runtimeModel.ui.statusMessage != "fresh status" {
		t.Fatalf("expected stale bundle error to be ignored, got status %q", runtimeModel.ui.statusMessage)
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

func TestUpdate_RuntimeSwitchInitializesReplacementBundleAndPreservesActiveRecordRequestGuard(t *testing.T) {
	// Arrange
	runtimeSession := &RuntimeSessionState{nextRuntimeBundleToken: 7}
	replacementEngine := &fakeRuntimeSwitchEngine{
		tables: []domainmodel.Table{{Name: "users"}},
		schema: domainmodel.Schema{
			Table: domainmodel.Table{Name: "users"},
			Columns: []domainmodel.Column{
				{Name: "id", Type: "INTEGER"},
			},
		},
		page: domainmodel.RecordPage{
			Records: []domainmodel.Record{
				{Values: []domainmodel.Value{{Text: "fresh"}}},
			},
			TotalCount: 1,
		},
	}
	model := &Model{
		ctx:                 context.Background(),
		runtimeSession:      runtimeSession,
		runtimeBundleToken:  7,
		runtimeBundleCancel: func() {},
	}

	// Act
	updated, initCmd := model.Update(runtimeDatabaseSwitchCompletedMsg{
		deps: RuntimeRunDeps{
			ListTables:  usecase.NewListTables(replacementEngine),
			GetSchema:   usecase.NewGetSchema(replacementEngine),
			ListRecords: usecase.NewListRecords(replacementEngine),
		},
	})
	if initCmd == nil {
		t.Fatal("expected replacement runtime initialization command after successful switch")
	}
	runtimeModel := updated.(*Model)
	initMsg := initCmd()
	tablesMessage, ok := initMsg.(tablesMsg)
	if !ok {
		t.Fatalf("expected replacement init to load tables, got %T", initMsg)
	}
	if tablesMessage.bundleToken != runtimeModel.runtimeBundleToken {
		t.Fatalf("expected tables load to use active bundle token %d, got %d", runtimeModel.runtimeBundleToken, tablesMessage.bundleToken)
	}
	updated, schemaCmd := runtimeModel.Update(initMsg)
	if schemaCmd == nil {
		t.Fatal("expected schema load after replacement tables init")
	}
	runtimeModel = updated.(*Model)
	schemaMsgFromCmd, ok := schemaCmd().(schemaMsg)
	if !ok {
		t.Fatalf("expected schemaMsg from schema load, got %T", schemaCmd())
	}
	if schemaMsgFromCmd.bundleToken != runtimeModel.runtimeBundleToken {
		t.Fatalf("expected schema load to use active bundle token %d, got %d", runtimeModel.runtimeBundleToken, schemaMsgFromCmd.bundleToken)
	}
	updated, _ = runtimeModel.Update(schemaMsgFromCmd)
	runtimeModel = updated.(*Model)
	runtimeModel.read.viewMode = ViewRecords
	runtimeModel.read.focus = FocusContent
	recordReloadCmd := runtimeModel.loadRecordsCmd(true)
	if recordReloadCmd == nil {
		t.Fatal("expected active replacement bundle to load records")
	}
	staleActiveBundleMsg := recordsMsg{
		bundleToken: runtimeModel.runtimeBundleToken,
		tableName:   "users",
		requestID:   runtimeModel.read.recordRequestID - 1,
		page: dto.RecordPage{
			Rows:       []dto.RecordRow{{Values: []string{"stale"}}},
			TotalCount: 99,
		},
	}
	runtimeModel.Update(staleActiveBundleMsg)
	updated, _ = runtimeModel.Update(recordReloadCmd())
	runtimeModel = updated.(*Model)

	// Assert
	if runtimeModel.runtimeBundleToken == 7 {
		t.Fatalf("expected replacement runtime bundle token to differ from previous token %d", 7)
	}
	if len(runtimeModel.read.tables) != 1 || runtimeModel.read.tables[0].Name != "users" {
		t.Fatalf("expected replacement tables init to succeed, got %+v", runtimeModel.read.tables)
	}
	if len(runtimeModel.read.schema.Columns) != 1 || runtimeModel.read.schema.Columns[0].Name != "id" {
		t.Fatalf("expected replacement schema init to succeed, got %+v", runtimeModel.read.schema.Columns)
	}
	if len(runtimeModel.read.records) != 1 || runtimeModel.read.records[0].Values[0] != "fresh" {
		t.Fatalf("expected active records reload to apply fresh page, got %+v", runtimeModel.read.records)
	}
	if replacementEngine.lastSchemaTableName != "users" {
		t.Fatalf("expected schema reload for users table, got %q", replacementEngine.lastSchemaTableName)
	}
	if replacementEngine.lastRecordsTableName != "users" {
		t.Fatalf("expected records reload for users table, got %q", replacementEngine.lastRecordsTableName)
	}
}
