package tui

import (
	"github.com/mgierok/dbc/internal/application/dto"
	"github.com/mgierok/dbc/internal/application/usecase"
)

func (m *Model) buildTableChanges() (dto.TableChanges, error) {
	return m.stagingSessionUseCase().BuildTableChanges(m.read.schema)
}

func (m *Model) dirtyEditCount() int {
	return m.stagingSessionUseCase().DirtyEditCount()
}

func (m *Model) hasDirtyEdits() bool {
	return m.stagingSessionUseCase().HasDirtyEdits()
}

func (m *Model) clearStagedState() {
	m.stagingSessionUseCase().Reset()
	m.stagingUI = stagingUIState{}
	m.syncStagingSnapshot()
}

func (m *Model) translatorUseCase() *usecase.StagedChangesTranslator {
	if m.translator != nil {
		return m.translator
	}
	return usecase.NewStagedChangesTranslator()
}

func (m *Model) saveWorkflowUseCase() *usecase.RuntimeSaveWorkflow {
	if m.saveWorkflow != nil {
		return m.saveWorkflow
	}
	return usecase.NewRuntimeSaveWorkflow()
}

func (m *Model) recordAccessResolverUseCase() *usecase.PersistedRecordAccessResolver {
	if m.recordAccessResolver != nil {
		return m.recordAccessResolver
	}
	return usecase.NewPersistedRecordAccessResolver()
}

func (m *Model) stagingPolicyUseCase() *usecase.StagingPolicy {
	if m.stagingPolicy != nil {
		return m.stagingPolicy
	}
	return usecase.NewStagingPolicy()
}

func (m *Model) stagingSessionUseCase() *usecase.StagingSession {
	if m.stagingSession != nil {
		return m.stagingSession
	}
	m.stagingSession = usecase.NewStagingSession(m.stagingPolicyUseCase(), m.translatorUseCase())
	return m.stagingSession
}

func (m *Model) syncStagingSnapshot() {
	m.stagingSnapshot = m.stagingSessionUseCase().Snapshot()
}

func (m *Model) currentStagingSnapshot() dto.StagingSnapshot {
	if m.stagingSession == nil {
		m.syncStagingSnapshot()
	}
	return m.stagingSnapshot
}

func (m *Model) showAutoForInsert(insertID dto.InsertDraftID) bool {
	if len(m.stagingUI.showAuto) == 0 {
		return false
	}
	return m.stagingUI.showAuto[insertID]
}

func (m *Model) setShowAutoForInsert(insertID dto.InsertDraftID, show bool) {
	if !show {
		if len(m.stagingUI.showAuto) == 0 {
			return
		}
		delete(m.stagingUI.showAuto, insertID)
		if len(m.stagingUI.showAuto) == 0 {
			m.stagingUI.showAuto = nil
		}
		return
	}
	if m.stagingUI.showAuto == nil {
		m.stagingUI.showAuto = make(map[dto.InsertDraftID]bool)
	}
	m.stagingUI.showAuto[insertID] = true
}
