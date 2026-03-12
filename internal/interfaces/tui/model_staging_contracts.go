package tui

import (
	"github.com/mgierok/dbc/internal/application/dto"
	"github.com/mgierok/dbc/internal/application/usecase"
)

func (m *Model) buildTableChanges() (dto.TableChanges, error) {
	return m.activeTableStaging().buildTableChanges(m.translatorUseCase(), m.read.schema)
}

func (m *Model) buildDatabaseChanges() ([]dto.NamedTableChanges, error) {
	return m.staging.buildDatabaseChanges(m.translatorUseCase())
}

func (m *Model) dirtyEditCount() int {
	return m.staging.dirtyEditCount(m.stagingPolicyUseCase())
}

func (m *Model) dirtyTableCount() int {
	return m.staging.dirtyTableCount(m.stagingPolicyUseCase())
}

func (m *Model) hasDirtyEdits() bool {
	return m.dirtyEditCount() > 0
}

func (m *Model) clearStagedState() {
	m.staging.clear()
}

func (m *Model) activeTableStaging() stagingState {
	return m.staging.table(m.currentTableName())
}

func (m *Model) activeTableStagingPtr() *stagingState {
	return m.staging.ensureTable(m.currentTableName())
}

func (m *Model) syncActiveTableSchema() {
	if len(m.read.schema.Columns) == 0 {
		return
	}
	m.staging.setSchema(m.currentTableName(), m.read.schema)
}

func (m *Model) tableHasDirtyEdits(tableName string) bool {
	return m.staging.hasDirtyTable(tableName, m.stagingPolicyUseCase())
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
