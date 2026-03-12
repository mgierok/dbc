package tui

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/mgierok/dbc/internal/application/dto"
	"github.com/mgierok/dbc/internal/application/usecase"
)

type stagedEdit struct {
	Value dto.StagedValue
}

type recordEdits struct {
	identity dto.RecordIdentity
	changes  map[int]stagedEdit
}

type pendingInsertRow struct {
	values       map[int]stagedEdit
	explicitAuto map[int]bool
	showAuto     bool
}

type recordDelete struct {
	identity dto.RecordIdentity
}

type stagedOperationKind int

const (
	opInsertAdded stagedOperationKind = iota
	opInsertRemoved
	opCellEdited
	opDeleteToggled
)

type cellEditTarget int

const (
	cellEditPersisted cellEditTarget = iota
	cellEditInsert
)

type insertOperation struct {
	index int
	row   pendingInsertRow
}

type cellEditOperation struct {
	target             cellEditTarget
	insertIndex        int
	recordKey          string
	identity           dto.RecordIdentity
	columnIndex        int
	before             stagedEdit
	beforeExists       bool
	after              stagedEdit
	afterExists        bool
	beforeExplicitAuto bool
	afterExplicitAuto  bool
}

type deleteToggleOperation struct {
	key          string
	identity     dto.RecordIdentity
	beforeMarked bool
	afterMarked  bool
}

type stagedOperation struct {
	kind   stagedOperationKind
	insert insertOperation
	cell   cellEditOperation
	del    deleteToggleOperation
}

type pkColumn struct {
	index  int
	column dto.SchemaColumn
}

// stagingState keeps the mutable write-side runtime session state together so
// edit/save changes stay local to the staging workflow instead of the top-level
// runtime router.
type stagingState struct {
	schema         dto.Schema
	pendingInserts []pendingInsertRow
	pendingUpdates map[string]recordEdits
	pendingDeletes map[string]recordDelete
	history        []stagedOperation
	future         []stagedOperation
}

type databaseStagingState struct {
	tables map[string]*stagingState
}

func (s *databaseStagingState) clear() {
	s.tables = nil
}

func (s *databaseStagingState) table(tableName string) stagingState {
	if s == nil || len(s.tables) == 0 {
		return stagingState{}
	}
	if state, ok := s.tables[tableName]; ok && state != nil {
		return *state
	}
	return stagingState{}
}

func (s *databaseStagingState) ensureTable(tableName string) *stagingState {
	if s.tables == nil {
		s.tables = make(map[string]*stagingState)
	}
	if state, ok := s.tables[tableName]; ok && state != nil {
		return state
	}
	state := &stagingState{}
	s.tables[tableName] = state
	return state
}

func (s *databaseStagingState) setTable(tableName string, state stagingState) {
	if s.tables == nil {
		s.tables = make(map[string]*stagingState)
	}
	cloned := state
	s.tables[tableName] = &cloned
}

func (s *databaseStagingState) setSchema(tableName string, schema dto.Schema) {
	if strings.TrimSpace(tableName) == "" {
		return
	}
	state := s.table(tableName)
	state.schema = schema
	s.setTable(tableName, state)
}

func (s *databaseStagingState) dirtyEditCount(policy *usecase.StagingPolicy) int {
	count := 0
	for _, state := range s.tables {
		if state != nil {
			count += state.dirtyEditCount(policy)
		}
	}
	return count
}

func (s *databaseStagingState) dirtyTableCount(policy *usecase.StagingPolicy) int {
	count := 0
	for _, state := range s.tables {
		if state != nil && state.dirtyEditCount(policy) > 0 {
			count++
		}
	}
	return count
}

func (s *databaseStagingState) hasDirtyTable(tableName string, policy *usecase.StagingPolicy) bool {
	return s.table(tableName).dirtyEditCount(policy) > 0
}

func (s *databaseStagingState) buildDatabaseChanges(translator *usecase.StagedChangesTranslator) ([]dto.NamedTableChanges, error) {
	tableNames := make([]string, 0, len(s.tables))
	for tableName, state := range s.tables {
		if state != nil && state.hasDirtyEdits() {
			tableNames = append(tableNames, tableName)
		}
	}
	sort.Strings(tableNames)

	changes := make([]dto.NamedTableChanges, 0, len(tableNames))
	for _, tableName := range tableNames {
		state := s.tables[tableName]
		if state == nil {
			continue
		}
		if len(state.schema.Columns) == 0 {
			return nil, fmt.Errorf("schema is required for table %q", tableName)
		}
		tableChanges, err := state.buildTableChanges(translator, state.schema)
		if err != nil {
			return nil, err
		}
		changes = append(changes, dto.NamedTableChanges{
			TableName: tableName,
			Changes:   tableChanges,
		})
	}
	return changes, nil
}

func (s stagingState) buildTableChanges(translator *usecase.StagedChangesTranslator, schema dto.Schema) (dto.TableChanges, error) {
	return translator.BuildTableChanges(
		schema,
		s.toPendingInsertRowsDTO(),
		s.toPendingRecordEditsDTO(),
		s.toPendingRecordDeletesDTO(),
	)
}

func (s stagingState) dirtyEditCount(policy *usecase.StagingPolicy) int {
	return policy.DirtyEditCount(
		s.toPendingInsertRowsDTO(),
		s.toPendingRecordEditsDTO(),
		s.toPendingRecordDeletesDTO(),
	)
}

func (s stagingState) hasDirtyEdits() bool {
	return len(s.pendingInserts) > 0 || len(s.pendingUpdates) > 0 || len(s.pendingDeletes) > 0
}

func (s stagingState) toPendingInsertRowsDTO() []dto.PendingInsertRow {
	rows := make([]dto.PendingInsertRow, 0, len(s.pendingInserts))
	for _, row := range s.pendingInserts {
		dtoRow := dto.PendingInsertRow{
			Values:       make(map[int]dto.StagedEdit, len(row.values)),
			ExplicitAuto: make(map[int]bool, len(row.explicitAuto)),
		}
		for index, value := range row.values {
			dtoRow.Values[index] = dto.StagedEdit{Value: value.Value}
		}
		for index, explicit := range row.explicitAuto {
			dtoRow.ExplicitAuto[index] = explicit
		}
		rows = append(rows, dtoRow)
	}
	return rows
}

func (s stagingState) toPendingRecordEditsDTO() map[string]dto.PendingRecordEdits {
	edits := make(map[string]dto.PendingRecordEdits, len(s.pendingUpdates))
	for key, update := range s.pendingUpdates {
		dtoChanges := make(map[int]dto.StagedEdit, len(update.changes))
		for columnIndex, change := range update.changes {
			dtoChanges[columnIndex] = dto.StagedEdit{Value: change.Value}
		}
		edits[key] = dto.PendingRecordEdits{
			Identity: update.identity,
			Changes:  dtoChanges,
		}
	}
	return edits
}

func (s stagingState) toPendingRecordDeletesDTO() map[string]dto.PendingRecordDelete {
	deletes := make(map[string]dto.PendingRecordDelete, len(s.pendingDeletes))
	for key, deleteChange := range s.pendingDeletes {
		deletes[key] = dto.PendingRecordDelete{
			Identity: deleteChange.identity,
		}
	}
	return deletes
}

func displayValue(value dto.StagedValue) string {
	if value.IsNull {
		return "NULL"
	}
	if strings.TrimSpace(value.Text) != "" {
		return value.Text
	}
	if value.Raw != nil {
		return fmt.Sprint(value.Raw)
	}
	return ""
}

func stagedEditEqual(left, right stagedEdit) bool {
	if left.Value.IsNull != right.Value.IsNull || left.Value.Text != right.Value.Text {
		return false
	}
	return reflect.DeepEqual(left.Value.Raw, right.Value.Raw)
}

func clonePendingInsertRow(row pendingInsertRow) pendingInsertRow {
	cloned := pendingInsertRow{
		values:       make(map[int]stagedEdit, len(row.values)),
		explicitAuto: make(map[int]bool, len(row.explicitAuto)),
		showAuto:     row.showAuto,
	}
	for key, value := range row.values {
		cloned.values[key] = value
	}
	for key, value := range row.explicitAuto {
		cloned.explicitAuto[key] = value
	}
	return cloned
}
