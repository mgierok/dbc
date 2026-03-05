package tui

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/mgierok/dbc/internal/application/dto"
	"github.com/mgierok/dbc/internal/application/usecase"
)

func (m *Model) buildTableChanges() (dto.TableChanges, error) {
	return m.translatorUseCase().BuildTableChanges(
		m.schema,
		m.toPendingInsertRowsDTO(),
		m.toPendingRecordEditsDTO(),
		m.toPendingRecordDeletesDTO(),
	)
}

func (m *Model) dirtyEditCount() int {
	return m.stagingPolicyUseCase().DirtyEditCount(
		m.toPendingInsertRowsDTO(),
		m.toPendingRecordEditsDTO(),
		m.toPendingRecordDeletesDTO(),
	)
}

func (m *Model) hasDirtyEdits() bool {
	return m.dirtyEditCount() > 0
}

func (m *Model) clearStagedState() {
	m.pendingInserts = nil
	m.pendingUpdates = nil
	m.pendingDeletes = nil
	m.history = nil
	m.future = nil
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

func (m *Model) translatorUseCase() *usecase.StagedChangesTranslator {
	if m.translator != nil {
		return m.translator
	}
	return usecase.NewStagedChangesTranslator()
}

func (m *Model) stagingPolicyUseCase() *usecase.StagingPolicy {
	if m.stagingPolicy != nil {
		return m.stagingPolicy
	}
	return usecase.NewStagingPolicy()
}

func (m *Model) toPendingInsertRowsDTO() []dto.PendingInsertRow {
	rows := make([]dto.PendingInsertRow, 0, len(m.pendingInserts))
	for _, row := range m.pendingInserts {
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

func (m *Model) toPendingRecordEditsDTO() map[string]dto.PendingRecordEdits {
	edits := make(map[string]dto.PendingRecordEdits, len(m.pendingUpdates))
	for key, update := range m.pendingUpdates {
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

func (m *Model) toPendingRecordDeletesDTO() map[string]dto.PendingRecordDelete {
	deletes := make(map[string]dto.PendingRecordDelete, len(m.pendingDeletes))
	for key, deleteChange := range m.pendingDeletes {
		deletes[key] = dto.PendingRecordDelete{
			Identity: deleteChange.identity,
		}
	}
	return deletes
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
