package tui

import (
	"github.com/mgierok/dbc/internal/application/dto"
	"github.com/mgierok/dbc/internal/application/usecase"
)

func (m *Model) buildTableChanges() (dto.TableChanges, error) {
	return m.staging.buildTableChanges(m.translatorUseCase(), m.read.schema)
}

func (m *Model) dirtyEditCount() int {
	return m.staging.dirtyEditCount(m.stagingPolicyUseCase())
}

func (m *Model) hasDirtyEdits() bool {
	return m.dirtyEditCount() > 0
}

func (m *Model) clearStagedState() {
	m.staging.clear()
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
