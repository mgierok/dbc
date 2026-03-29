package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/dto"
	"github.com/mgierok/dbc/internal/application/usecase"
)

func (m *Model) requestSaveChanges() (tea.Model, tea.Cmd) {
	if !m.nonBlockingRuntimeCommandContextActive() {
		return m, nil
	}
	m.ui.pendingDatabaseTransition = nil
	return m.requestRuntimeSave(usecase.RuntimeSaveIntentSaveOnly)
}

func (m *Model) requestSaveAndQuit() (tea.Model, tea.Cmd) {
	if !m.nonBlockingRuntimeCommandContextActive() {
		return m, nil
	}
	m.ui.pendingDatabaseTransition = nil
	return m.requestRuntimeSave(usecase.RuntimeSaveIntentSaveAndQuit)
}

func (m *Model) confirmSaveChanges() (tea.Model, tea.Cmd) {
	return m.startRuntimeSave(usecase.RuntimeSaveIntentSaveOnly)
}

func (m *Model) requestRuntimeSave(intent usecase.RuntimeSaveIntent) (tea.Model, tea.Cmd) {
	decision := m.saveWorkflowUseCase().PlanRequest(intent, m.hasDirtyEdits())
	if !decision.StartSave {
		m.ui.pendingSaveSuccessAction = usecase.RuntimeSaveSuccessActionNone
		return m.applyRuntimeSaveRequestDecision(decision)
	}
	if m.saveChanges == nil {
		m.ui.pendingSaveSuccessAction = usecase.RuntimeSaveSuccessActionNone
		m.ui.statusMessage = "Error: save use case unavailable"
		return m, nil
	}
	return m.startRuntimeSave(intent)
}

func (m *Model) startRuntimeSave(intent usecase.RuntimeSaveIntent) (tea.Model, tea.Cmd) {
	changes, err := m.buildTableChanges()
	if err != nil {
		m.ui.pendingSaveSuccessAction = usecase.RuntimeSaveSuccessActionNone
		m.ui.statusMessage = "Error: " + err.Error()
		return m, nil
	}
	decision := m.saveWorkflowUseCase().PlanStart(intent, hasEffectiveTableChanges(changes))
	if !decision.StartSave {
		m.ui.pendingSaveSuccessAction = usecase.RuntimeSaveSuccessActionNone
		return m.applyRuntimeSaveRequestDecision(decision)
	}
	m.ui.saveInFlight = true
	m.ui.pendingSaveSuccessAction = decision.SuccessAction
	m.ui.statusMessage = "Saving changes..."
	return m, saveChangesCmd(m.ctx, m.saveChanges, m.currentTableName(), changes)
}

func (m *Model) confirmDiscardTableSwitch() (tea.Model, tea.Cmd) {
	if m.ui.pendingTableIndex < 0 || m.ui.pendingTableIndex >= len(m.read.tables) {
		m.ui.pendingTableIndex = -1
		return m, nil
	}
	target := m.ui.pendingTableIndex
	m.ui.pendingTableIndex = -1
	m.clearStagedState()
	m.read.selectedTable = target
	m.resetTableContext()
	return m, m.loadViewForSelection()
}

func (m *Model) confirmDiscardQuit() (tea.Model, tea.Cmd) {
	m.ui.pendingSaveSuccessAction = usecase.RuntimeSaveSuccessActionNone
	m.ui.pendingDatabaseTransition = nil
	m.clearStagedState()
	return m, tea.Quit
}

func (m *Model) applyRuntimeSaveRequestDecision(decision usecase.RuntimeSaveRequestDecision) (tea.Model, tea.Cmd) {
	if decision.ImmediateStatus != "" {
		m.ui.statusMessage = decision.ImmediateStatus
	}
	if decision.ImmediateExit {
		return m, tea.Quit
	}
	return m, nil
}

func hasEffectiveTableChanges(changes dto.TableChanges) bool {
	return len(changes.Inserts) > 0 || len(changes.Updates) > 0 || len(changes.Deletes) > 0
}
