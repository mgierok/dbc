package usecase

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/mgierok/dbc/internal/application/dto"
)

type stagingOperationKind int

const (
	opInsertAdded stagingOperationKind = iota
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
	id    dto.InsertDraftID
	row   dto.PendingInsertRow
}

type cellEditOperation struct {
	target             cellEditTarget
	insertID           dto.InsertDraftID
	recordKey          string
	identity           dto.RecordIdentity
	columnIndex        int
	before             dto.StagedEdit
	beforeExists       bool
	after              dto.StagedEdit
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
	kind   stagingOperationKind
	insert insertOperation
	cell   cellEditOperation
	del    deleteToggleOperation
}

type StagingSession struct {
	policy      *StagingPolicy
	translator  *StagedChangesTranslator
	nextInsert  int
	insertOrder []dto.InsertDraftID
	inserts     map[dto.InsertDraftID]dto.PendingInsertRow
	updates     map[string]dto.PendingRecordEdits
	deletes     map[string]dto.PendingRecordDelete
	history     []stagedOperation
	future      []stagedOperation
}

func NewStagingSession(policy *StagingPolicy, translator *StagedChangesTranslator) *StagingSession {
	if policy == nil {
		policy = NewStagingPolicy()
	}
	if translator == nil {
		translator = NewStagedChangesTranslator()
	}
	return &StagingSession{
		policy:     policy,
		translator: translator,
	}
}

func (s *StagingSession) Reset() {
	if s == nil {
		return
	}
	s.nextInsert = 0
	s.insertOrder = nil
	s.inserts = nil
	s.updates = nil
	s.deletes = nil
	s.history = nil
	s.future = nil
}

func (s *StagingSession) Snapshot() dto.StagingSnapshot {
	if s == nil {
		return dto.StagingSnapshot{}
	}
	snapshot := dto.StagingSnapshot{
		PendingInserts: make([]dto.InsertDraftSnapshot, 0, len(s.insertOrder)),
		PendingUpdates: clonePendingRecordEdits(s.updates),
		PendingDeletes: clonePendingRecordDeletes(s.deletes),
	}
	for _, id := range s.insertOrder {
		row, ok := s.inserts[id]
		if !ok {
			continue
		}
		snapshot.PendingInserts = append(snapshot.PendingInserts, dto.InsertDraftSnapshot{
			ID:           id,
			Values:       cloneStagedEdits(row.Values),
			ExplicitAuto: cloneExplicitAuto(row.ExplicitAuto),
		})
	}
	return snapshot
}

func (s *StagingSession) AddInsert(schema dto.Schema) (dto.InsertDraftID, error) {
	if s == nil {
		return "", fmt.Errorf("staging session unavailable")
	}
	id := dto.InsertDraftID(fmt.Sprintf("insert-%d", s.nextInsert))
	s.nextInsert++
	row := dto.PendingInsertRow{
		Values:       make(map[int]dto.StagedEdit, len(schema.Columns)),
		ExplicitAuto: make(map[int]bool),
	}
	for index, column := range schema.Columns {
		row.Values[index] = dto.StagedEdit{Value: s.policy.InitialInsertValue(column)}
	}
	if err := s.insertPendingRowAt(0, id, row); err != nil {
		return "", err
	}
	s.recordOperation(stagedOperation{
		kind: opInsertAdded,
		insert: insertOperation{
			index: 0,
			id:    id,
			row:   clonePendingInsertRow(row),
		},
	})
	return id, nil
}

func (s *StagingSession) RemoveInsert(insertID dto.InsertDraftID) error {
	if s == nil {
		return fmt.Errorf("staging session unavailable")
	}
	index := s.indexOfInsert(insertID)
	if index < 0 {
		return fmt.Errorf("insert draft not found")
	}
	removedID, removed, err := s.removePendingInsertAt(index)
	if err != nil {
		return err
	}
	s.recordOperation(stagedOperation{
		kind: opInsertRemoved,
		insert: insertOperation{
			index: index,
			id:    removedID,
			row:   removed,
		},
	})
	return nil
}

func (s *StagingSession) StageInsertEdit(insertID dto.InsertDraftID, columnIndex int, value dto.StagedValue) error {
	if s == nil {
		return fmt.Errorf("staging session unavailable")
	}
	if columnIndex < 0 {
		return fmt.Errorf("column index out of range")
	}
	row, ok := s.inserts[insertID]
	if !ok {
		return fmt.Errorf("insert draft not found")
	}
	if row.Values == nil {
		row.Values = make(map[int]dto.StagedEdit)
	}
	if row.ExplicitAuto == nil {
		row.ExplicitAuto = make(map[int]bool)
	}
	before, beforeExists := row.Values[columnIndex]
	beforeExplicitAuto := row.ExplicitAuto[columnIndex]
	after := dto.StagedEdit{Value: value}
	row.Values[columnIndex] = after
	row.ExplicitAuto[columnIndex] = true
	afterExplicitAuto := true
	s.inserts[insertID] = row

	changed := !beforeExists || !stagedEditEqual(before, after) || beforeExplicitAuto != afterExplicitAuto
	if !changed {
		return nil
	}
	s.recordOperation(stagedOperation{
		kind: opCellEdited,
		cell: cellEditOperation{
			target:             cellEditInsert,
			insertID:           insertID,
			columnIndex:        columnIndex,
			before:             before,
			beforeExists:       beforeExists,
			after:              after,
			afterExists:        true,
			beforeExplicitAuto: beforeExplicitAuto,
			afterExplicitAuto:  afterExplicitAuto,
		},
	})
	return nil
}

func (s *StagingSession) StagePersistedEdit(
	recordKey string,
	identity dto.RecordIdentity,
	columnIndex int,
	originalDisplayValue string,
	value dto.StagedValue,
) error {
	if s == nil {
		return fmt.Errorf("staging session unavailable")
	}
	if strings.TrimSpace(recordKey) == "" {
		return fmt.Errorf("record key missing")
	}
	if columnIndex < 0 {
		return fmt.Errorf("column index out of range")
	}
	if s.updates == nil {
		s.updates = make(map[string]dto.PendingRecordEdits)
	}
	edits := s.updates[recordKey]
	if edits.Changes == nil {
		edits.Changes = make(map[int]dto.StagedEdit)
	}
	edits.Identity = identity
	before, beforeExists := edits.Changes[columnIndex]
	after, afterExists := dto.StagedEdit{}, false

	if stagedDisplayValue(value) == originalDisplayValue {
		delete(edits.Changes, columnIndex)
		if len(edits.Changes) == 0 {
			delete(s.updates, recordKey)
		} else {
			s.updates[recordKey] = edits
		}
	} else {
		after = dto.StagedEdit{Value: value}
		afterExists = true
		edits.Changes[columnIndex] = after
		s.updates[recordKey] = edits
	}

	changed := beforeExists != afterExists || (beforeExists && afterExists && !stagedEditEqual(before, after))
	if !changed {
		return nil
	}
	s.recordOperation(stagedOperation{
		kind: opCellEdited,
		cell: cellEditOperation{
			target:       cellEditPersisted,
			recordKey:    recordKey,
			identity:     identity,
			columnIndex:  columnIndex,
			before:       before,
			beforeExists: beforeExists,
			after:        after,
			afterExists:  afterExists,
		},
	})
	return nil
}

func (s *StagingSession) SetDeleteMark(recordKey string, identity dto.RecordIdentity, marked bool) error {
	if s == nil {
		return fmt.Errorf("staging session unavailable")
	}
	if strings.TrimSpace(recordKey) == "" {
		return fmt.Errorf("record key missing")
	}
	_, exists := s.deletes[recordKey]
	if err := s.setDeleteMark(recordKey, identity, marked); err != nil {
		return err
	}
	s.recordOperation(stagedOperation{
		kind: opDeleteToggled,
		del: deleteToggleOperation{
			key:          recordKey,
			identity:     identity,
			beforeMarked: exists,
			afterMarked:  marked,
		},
	})
	return nil
}

func (s *StagingSession) Undo() error {
	if s == nil || len(s.history) == 0 {
		return nil
	}
	lastIndex := len(s.history) - 1
	op := s.history[lastIndex]
	s.history = s.history[:lastIndex]
	if err := s.applyInverseOperation(op); err != nil {
		s.history = append(s.history, op)
		return err
	}
	s.future = append(s.future, op)
	return nil
}

func (s *StagingSession) Redo() error {
	if s == nil || len(s.future) == 0 {
		return nil
	}
	lastIndex := len(s.future) - 1
	op := s.future[lastIndex]
	s.future = s.future[:lastIndex]
	if err := s.applyOperation(op); err != nil {
		s.future = append(s.future, op)
		return err
	}
	s.history = append(s.history, op)
	return nil
}

func (s *StagingSession) BuildTableChanges(schema dto.Schema) (dto.TableChanges, error) {
	if s == nil {
		return dto.TableChanges{}, fmt.Errorf("staging session unavailable")
	}
	return s.translator.BuildTableChanges(
		schema,
		s.pendingInsertRows(),
		s.updates,
		s.deletes,
	)
}

func (s *StagingSession) DirtyEditCount() int {
	if s == nil {
		return 0
	}
	return s.policy.DirtyEditCount(
		s.pendingInsertRows(),
		s.updates,
		s.deletes,
	)
}

func (s *StagingSession) HasDirtyEdits() bool {
	return s.DirtyEditCount() > 0
}

func (s *StagingSession) applyOperation(op stagedOperation) error {
	switch op.kind {
	case opInsertAdded:
		return s.insertPendingRowAt(op.insert.index, op.insert.id, op.insert.row)
	case opInsertRemoved:
		_, _, err := s.removePendingInsertAt(op.insert.index)
		return err
	case opCellEdited:
		return s.applyCellEditState(op.cell, false)
	case opDeleteToggled:
		return s.setDeleteMark(op.del.key, op.del.identity, op.del.afterMarked)
	default:
		return fmt.Errorf("unsupported staged operation")
	}
}

func (s *StagingSession) applyInverseOperation(op stagedOperation) error {
	switch op.kind {
	case opInsertAdded:
		_, _, err := s.removePendingInsertAt(op.insert.index)
		return err
	case opInsertRemoved:
		return s.insertPendingRowAt(op.insert.index, op.insert.id, op.insert.row)
	case opCellEdited:
		return s.applyCellEditState(op.cell, true)
	case opDeleteToggled:
		return s.setDeleteMark(op.del.key, op.del.identity, op.del.beforeMarked)
	default:
		return fmt.Errorf("unsupported staged operation")
	}
}

func (s *StagingSession) applyCellEditState(op cellEditOperation, useBefore bool) error {
	edit := op.after
	exists := op.afterExists
	explicitAuto := op.afterExplicitAuto
	if useBefore {
		edit = op.before
		exists = op.beforeExists
		explicitAuto = op.beforeExplicitAuto
	}
	switch op.target {
	case cellEditInsert:
		row, ok := s.inserts[op.insertID]
		if !ok {
			return fmt.Errorf("insert draft not found")
		}
		if row.Values == nil {
			row.Values = make(map[int]dto.StagedEdit)
		}
		if row.ExplicitAuto == nil {
			row.ExplicitAuto = make(map[int]bool)
		}
		if exists {
			row.Values[op.columnIndex] = edit
		} else {
			delete(row.Values, op.columnIndex)
		}
		if explicitAuto {
			row.ExplicitAuto[op.columnIndex] = true
		} else {
			delete(row.ExplicitAuto, op.columnIndex)
		}
		s.inserts[op.insertID] = row
		return nil
	case cellEditPersisted:
		if strings.TrimSpace(op.recordKey) == "" {
			return fmt.Errorf("record key missing")
		}
		if s.updates == nil {
			s.updates = make(map[string]dto.PendingRecordEdits)
		}
		edits := s.updates[op.recordKey]
		if edits.Changes == nil {
			edits.Changes = make(map[int]dto.StagedEdit)
		}
		edits.Identity = op.identity
		if exists {
			edits.Changes[op.columnIndex] = edit
			s.updates[op.recordKey] = edits
			return nil
		}
		delete(edits.Changes, op.columnIndex)
		if len(edits.Changes) == 0 {
			delete(s.updates, op.recordKey)
			return nil
		}
		s.updates[op.recordKey] = edits
		return nil
	default:
		return fmt.Errorf("unsupported cell edit target")
	}
}

func (s *StagingSession) setDeleteMark(recordKey string, identity dto.RecordIdentity, marked bool) error {
	if strings.TrimSpace(recordKey) == "" {
		return fmt.Errorf("record key missing")
	}
	if s.deletes == nil {
		s.deletes = make(map[string]dto.PendingRecordDelete)
	}
	if marked {
		s.deletes[recordKey] = dto.PendingRecordDelete{Identity: identity}
		return nil
	}
	delete(s.deletes, recordKey)
	return nil
}

func (s *StagingSession) recordOperation(op stagedOperation) {
	s.history = append(s.history, op)
	s.future = nil
}

func (s *StagingSession) pendingInsertRows() []dto.PendingInsertRow {
	rows := make([]dto.PendingInsertRow, 0, len(s.insertOrder))
	for _, id := range s.insertOrder {
		row, ok := s.inserts[id]
		if !ok {
			continue
		}
		rows = append(rows, clonePendingInsertRow(row))
	}
	return rows
}

func (s *StagingSession) insertPendingRowAt(index int, id dto.InsertDraftID, row dto.PendingInsertRow) error {
	if index < 0 || index > len(s.insertOrder) {
		return fmt.Errorf("insert index out of range")
	}
	if s.inserts == nil {
		s.inserts = make(map[dto.InsertDraftID]dto.PendingInsertRow)
	}
	cloned := clonePendingInsertRow(row)
	s.insertOrder = append(s.insertOrder, "")
	copy(s.insertOrder[index+1:], s.insertOrder[index:])
	s.insertOrder[index] = id
	s.inserts[id] = cloned
	return nil
}

func (s *StagingSession) removePendingInsertAt(index int) (dto.InsertDraftID, dto.PendingInsertRow, error) {
	if index < 0 || index >= len(s.insertOrder) {
		return "", dto.PendingInsertRow{}, fmt.Errorf("insert index out of range")
	}
	id := s.insertOrder[index]
	row, ok := s.inserts[id]
	if !ok {
		return "", dto.PendingInsertRow{}, fmt.Errorf("insert draft not found")
	}
	removed := clonePendingInsertRow(row)
	s.insertOrder = append(s.insertOrder[:index], s.insertOrder[index+1:]...)
	delete(s.inserts, id)
	return id, removed, nil
}

func (s *StagingSession) indexOfInsert(insertID dto.InsertDraftID) int {
	for index, currentID := range s.insertOrder {
		if currentID == insertID {
			return index
		}
	}
	return -1
}

func clonePendingInsertRow(row dto.PendingInsertRow) dto.PendingInsertRow {
	return dto.PendingInsertRow{
		Values:       cloneStagedEdits(row.Values),
		ExplicitAuto: cloneExplicitAuto(row.ExplicitAuto),
	}
}

func clonePendingRecordEdits(source map[string]dto.PendingRecordEdits) map[string]dto.PendingRecordEdits {
	if len(source) == 0 {
		return nil
	}
	cloned := make(map[string]dto.PendingRecordEdits, len(source))
	for key, edits := range source {
		cloned[key] = dto.PendingRecordEdits{
			Identity: edits.Identity,
			Changes:  cloneStagedEdits(edits.Changes),
		}
	}
	return cloned
}

func clonePendingRecordDeletes(source map[string]dto.PendingRecordDelete) map[string]dto.PendingRecordDelete {
	if len(source) == 0 {
		return nil
	}
	cloned := make(map[string]dto.PendingRecordDelete, len(source))
	for key, deleteChange := range source {
		cloned[key] = deleteChange
	}
	return cloned
}

func cloneStagedEdits(source map[int]dto.StagedEdit) map[int]dto.StagedEdit {
	if len(source) == 0 {
		return nil
	}
	cloned := make(map[int]dto.StagedEdit, len(source))
	for key, value := range source {
		cloned[key] = value
	}
	return cloned
}

func cloneExplicitAuto(source map[int]bool) map[int]bool {
	if len(source) == 0 {
		return nil
	}
	cloned := make(map[int]bool, len(source))
	for key, value := range source {
		cloned[key] = value
	}
	return cloned
}

func stagedEditEqual(left, right dto.StagedEdit) bool {
	if left.Value.IsNull != right.Value.IsNull || left.Value.Text != right.Value.Text {
		return false
	}
	return reflect.DeepEqual(left.Value.Raw, right.Value.Raw)
}
