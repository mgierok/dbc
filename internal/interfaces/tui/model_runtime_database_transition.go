package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/usecase"
)

func (m *Model) requestRuntimeDatabaseTransition(target usecase.RuntimeDatabaseTarget, force bool, pendingCommandInput string) (tea.Model, tea.Cmd) {
	plan := m.navigationWorkflowUseCase().PlanDatabaseTransition(target, m.hasDirtyEdits() && !force, m.dirtyEditCount())
	return m.applyRuntimeNavigationPlan(plan, pendingCommandInput)
}

func (m *Model) executeRuntimeNavigationNextAction(action usecase.RuntimeNavigationNextAction) (tea.Model, tea.Cmd) {
	switch action.Kind {
	case usecase.RuntimeNavigationNextActionSwitchTable:
		targetIndex := m.indexOfTableByName(action.TargetTableName)
		if targetIndex < 0 {
			m.ui.statusMessage = fmt.Sprintf("Error: target table %q is no longer available", action.TargetTableName)
			return m, nil
		}
		if action.ClearDirtyState {
			m.clearStagedState()
		}
		m.read.selectedTable = targetIndex
		m.resetTableContext()
		return m, m.loadViewForSelection()
	case usecase.RuntimeNavigationNextActionOpenDatabase:
		if action.ClearDirtyState {
			m.clearStagedState()
		}
		m.overlay.commandInput = commandInput{}
		m.exitResult = runtimeExitResultOpenDatabaseNext(databaseOptionFromRuntimeDatabaseOption(action.DatabaseTarget.Option))
		return m, tea.Quit
	case usecase.RuntimeNavigationNextActionQuitRuntime:
		m.ui.pendingSaveSuccessAction = usecase.RuntimeSaveSuccessActionNone
		if action.ClearDirtyState {
			m.clearStagedState()
		}
		m.ui.pendingNavigation = nil
		m.ui.pendingCommandInput = ""
		return m, tea.Quit
	case usecase.RuntimeNavigationNextActionStayInRuntime, usecase.RuntimeNavigationNextActionNone:
		return m, nil
	default:
		return m, nil
	}
}

func (m *Model) resolveRuntimeDatabaseTransitionTargetFromOption(selected DatabaseOption) (usecase.RuntimeDatabaseTarget, error) {
	configuredOptions, err := m.configuredRuntimeDatabaseOptions()
	if err != nil {
		return usecase.RuntimeDatabaseTarget{}, err
	}

	return m.databaseTargetResolverUseCase().Resolve(
		runtimeDatabaseOptionFromSelectorOption(m.currentRuntimeDatabaseOption()),
		configuredOptions,
		runtimeDatabaseOptionFromSelectorOption(selected),
	)
}

func (m *Model) resolveRuntimeDatabaseTransitionTargetFromConnString(connString string) (usecase.RuntimeDatabaseTarget, error) {
	configuredOptions, err := m.configuredRuntimeDatabaseOptions()
	if err != nil {
		return usecase.RuntimeDatabaseTarget{}, err
	}

	return m.databaseTargetResolverUseCase().Resolve(
		runtimeDatabaseOptionFromSelectorOption(m.currentRuntimeDatabaseOption()),
		configuredOptions,
		runtimeDatabaseRequestOptionFromConnString(connString),
	)
}

func (m *Model) configuredRuntimeDatabaseOptions() ([]usecase.RuntimeDatabaseOption, error) {
	if m.runtimeDatabaseSelectorDeps == nil || m.runtimeDatabaseSelectorDeps.ListConfiguredDatabases == nil {
		return nil, fmt.Errorf("database selector unavailable")
	}

	entries, err := m.runtimeDatabaseSelectorDeps.ListConfiguredDatabases.Execute(m.ctx)
	if err != nil {
		return nil, err
	}

	options := make([]usecase.RuntimeDatabaseOption, len(entries))
	for i, entry := range entries {
		options[i] = usecase.RuntimeDatabaseOption{
			Name:       entry.Name,
			ConnString: entry.Path,
			Source:     usecase.RuntimeDatabaseOptionSourceConfig,
		}
	}
	return options, nil
}

func runtimeDatabaseOptionFromSelectorOption(option DatabaseOption) usecase.RuntimeDatabaseOption {
	source := usecase.RuntimeDatabaseOptionSourceConfig
	if option.Source == DatabaseOptionSourceCLI {
		source = usecase.RuntimeDatabaseOptionSourceCLI
	}
	return usecase.RuntimeDatabaseOption{
		Name:       option.Name,
		ConnString: option.ConnString,
		Source:     source,
	}
}

func runtimeDatabaseRequestOptionFromConnString(connString string) usecase.RuntimeDatabaseOption {
	return usecase.RuntimeDatabaseOption{
		Name:       connString,
		ConnString: connString,
		Source:     usecase.RuntimeDatabaseOptionSourceCLI,
	}
}

func databaseOptionFromRuntimeDatabaseOption(option usecase.RuntimeDatabaseOption) DatabaseOption {
	source := DatabaseOptionSourceConfig
	if option.Source == usecase.RuntimeDatabaseOptionSourceCLI {
		source = DatabaseOptionSourceCLI
	}
	return DatabaseOption{
		Name:       option.Name,
		ConnString: option.ConnString,
		Source:     source,
	}
}

func (m *Model) indexOfTableByName(target string) int {
	for i, table := range m.read.tables {
		if table.Name == target {
			return i
		}
	}
	return -1
}
