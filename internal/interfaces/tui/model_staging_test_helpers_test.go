package tui

import (
	"fmt"
	"reflect"

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

type stagingState struct {
	pendingInserts []pendingInsertRow
	pendingUpdates map[string]recordEdits
	pendingDeletes map[string]recordDelete
}

func withTestStaging(model *Model, seed stagingState) *Model {
	if model == nil {
		panic("model is nil")
	}
	model.stagingSession = usecase.NewStagingSession(model.stagingPolicyUseCase(), model.translatorUseCase())
	model.stagingUI = stagingUIState{}
	model.stagingSnapshot = dto.StagingSnapshot{}

	for index := len(seed.pendingInserts) - 1; index >= 0; index-- {
		row := seed.pendingInserts[index]
		insertID, err := model.stagingSessionUseCase().AddInsert(model.read.schema)
		if err != nil {
			panic(fmt.Sprintf("seed insert %d failed: %v", index, err))
		}
		currentRow := snapshotInsertByID(model.stagingSessionUseCase().Snapshot(), insertID)
		for columnIndex, value := range row.values {
			explicitAuto := row.explicitAuto[columnIndex]
			if existing, ok := currentRow.Values[columnIndex]; ok && stagedValueEqual(existing.Value, value.Value) && !explicitAuto {
				continue
			}
			if err := model.stagingSessionUseCase().StageInsertEdit(insertID, columnIndex, value.Value); err != nil {
				panic(fmt.Sprintf("seed insert column %d failed: %v", columnIndex, err))
			}
		}
		if row.showAuto {
			model.setShowAutoForInsert(insertID, true)
		}
	}

	for key, update := range seed.pendingUpdates {
		for columnIndex, change := range update.changes {
			original := model.originalDisplayValueForSeed(key, columnIndex)
			if err := model.stagingSessionUseCase().StagePersistedEdit(key, update.identity, columnIndex, original, change.Value); err != nil {
				panic(fmt.Sprintf("seed update %q column %d failed: %v", key, columnIndex, err))
			}
		}
	}

	for key, deleteChange := range seed.pendingDeletes {
		if err := model.stagingSessionUseCase().SetDeleteMark(key, deleteChange.identity, true); err != nil {
			panic(fmt.Sprintf("seed delete %q failed: %v", key, err))
		}
	}

	model.syncStagingSnapshot()
	return model
}

func setTestPendingUpdates(model *Model, updates map[string]recordEdits) {
	if model == nil {
		panic("model is nil")
	}
	seed := stagingState{
		pendingInserts: testPendingInsertsFromSnapshot(model.currentStagingSnapshot()),
		pendingDeletes: testPendingDeletesFromSnapshot(model.currentStagingSnapshot()),
		pendingUpdates: updates,
	}
	withTestStaging(model, seed)
}

func setTestPendingDeletes(model *Model, deletes map[string]recordDelete) {
	if model == nil {
		panic("model is nil")
	}
	seed := stagingState{
		pendingInserts: testPendingInsertsFromSnapshot(model.currentStagingSnapshot()),
		pendingUpdates: testPendingUpdatesFromSnapshot(model.currentStagingSnapshot()),
		pendingDeletes: deletes,
	}
	withTestStaging(model, seed)
}

func snapshotInsertByID(snapshot dto.StagingSnapshot, insertID dto.InsertDraftID) dto.InsertDraftSnapshot {
	for _, row := range snapshot.PendingInserts {
		if row.ID == insertID {
			return row
		}
	}
	panic(fmt.Sprintf("insert draft %q not found in snapshot", insertID))
}

func stagedValueEqual(left, right dto.StagedValue) bool {
	if left.IsNull != right.IsNull || left.Text != right.Text {
		return false
	}
	return reflect.DeepEqual(left.Raw, right.Raw)
}

func (m *Model) originalDisplayValueForSeed(recordKey string, columnIndex int) string {
	for rowIndex := range m.read.records {
		recordRef, err := m.persistedRecordRefForPersistedRow(rowIndex)
		if err != nil || recordRef.RowKey != recordKey {
			continue
		}
		return m.recordValue(rowIndex, columnIndex)
	}
	return ""
}

func testPendingInsertsFromSnapshot(snapshot dto.StagingSnapshot) []pendingInsertRow {
	if len(snapshot.PendingInserts) == 0 {
		return nil
	}
	rows := make([]pendingInsertRow, 0, len(snapshot.PendingInserts))
	for _, insert := range snapshot.PendingInserts {
		row := pendingInsertRow{
			values:       make(map[int]stagedEdit, len(insert.Values)),
			explicitAuto: make(map[int]bool, len(insert.ExplicitAuto)),
		}
		for columnIndex, value := range insert.Values {
			row.values[columnIndex] = stagedEdit{Value: value.Value}
		}
		for columnIndex, explicit := range insert.ExplicitAuto {
			row.explicitAuto[columnIndex] = explicit
		}
		rows = append(rows, row)
	}
	return rows
}

func testPendingUpdatesFromSnapshot(snapshot dto.StagingSnapshot) map[string]recordEdits {
	if len(snapshot.PendingUpdates) == 0 {
		return nil
	}
	updates := make(map[string]recordEdits, len(snapshot.PendingUpdates))
	for key, update := range snapshot.PendingUpdates {
		seed := recordEdits{
			identity: update.Identity,
			changes:  make(map[int]stagedEdit, len(update.Changes)),
		}
		for columnIndex, value := range update.Changes {
			seed.changes[columnIndex] = stagedEdit{Value: value.Value}
		}
		updates[key] = seed
	}
	return updates
}

func testPendingDeletesFromSnapshot(snapshot dto.StagingSnapshot) map[string]recordDelete {
	if len(snapshot.PendingDeletes) == 0 {
		return nil
	}
	deletes := make(map[string]recordDelete, len(snapshot.PendingDeletes))
	for key, deleteChange := range snapshot.PendingDeletes {
		deletes[key] = recordDelete{identity: deleteChange.Identity}
	}
	return deletes
}
