package tui

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	selectorpkg "github.com/mgierok/dbc/internal/interfaces/tui/internal/selector"
)

func (m *Model) openRuntimeDatabaseSelectorPopup() {
	if m.runtimeDatabaseSelectorDeps == nil {
		m.ui.statusMessage = "Error: database selector unavailable"
		return
	}
	deps := m.runtimeDatabaseSelectorDeps
	if deps.ListConfiguredDatabases == nil || deps.CreateConfiguredDatabase == nil || deps.UpdateConfiguredDatabase == nil || deps.DeleteConfiguredDatabase == nil || deps.GetActiveConfigPath == nil {
		m.ui.statusMessage = "Error: database selector unavailable"
		return
	}

	controller, err := selectorpkg.NewRuntimeController(m.ctx, selectorUseCaseAdapter{
		list:   deps.ListConfiguredDatabases,
		create: deps.CreateConfiguredDatabase,
		update: deps.UpdateConfiguredDatabase,
		del:    deps.DeleteConfiguredDatabase,
		active: deps.GetActiveConfigPath,
	}, selectorpkg.SelectorLaunchState{
		PreferConnString:  deps.CurrentDatabase.ConnString,
		AdditionalOptions: cloneDatabaseOptions(deps.AdditionalOptions),
	})
	if err != nil {
		m.ui.statusMessage = "Error: " + err.Error()
		return
	}

	m.overlay.databaseSelector = runtimeDatabaseSelectorPopup{
		active:     true,
		controller: controller,
	}
	m.ui.openConfigSelector = true
}

func (m *Model) closeRuntimeDatabaseSelectorPopup() {
	m.overlay.databaseSelector = runtimeDatabaseSelectorPopup{}
	m.ui.openConfigSelector = false
}

func (m *Model) handleRuntimeDatabaseSelectorKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if !m.overlay.databaseSelector.active || m.overlay.databaseSelector.controller == nil {
		return m, nil
	}

	cmd := m.overlay.databaseSelector.controller.Handle(msg)
	intent := m.overlay.databaseSelector.controller.ConsumeIntent()
	switch intent.Type {
	case selectorpkg.IntentTypeClose:
		m.closeRuntimeDatabaseSelectorPopup()
		return m, nil
	case selectorpkg.IntentTypeSelect:
		return m.handleRuntimeDatabaseSelection(intent.Option)
	default:
		return m, cmd
	}
}

func (m *Model) handleRuntimeDatabaseSelection(selected DatabaseOption) (tea.Model, tea.Cmd) {
	target, err := m.resolveRuntimeDatabaseTransitionTargetFromOption(selected)
	if err != nil {
		m.ui.statusMessage = "Error: " + err.Error()
		return m, nil
	}
	return m.requestRuntimeDatabaseTransition(runtimeDatabaseTransitionRequest{
		Target: target,
		Origin: runtimeDatabaseTransitionOriginConfigSelector,
	})
}

func (m *Model) runtimeDatabaseSwitcher() RuntimeDatabaseSwitcher {
	if m.runtimeDatabaseSelectorDeps == nil {
		return nil
	}
	return m.runtimeDatabaseSelectorDeps.SwitchDatabase
}

func (m *Model) currentRuntimeDatabaseConnString() string {
	if m.runtimeDatabaseSelectorDeps == nil {
		return ""
	}
	return m.runtimeDatabaseSelectorDeps.CurrentDatabase.ConnString
}

func (m *Model) currentRuntimeDatabaseOption() DatabaseOption {
	if m.runtimeDatabaseSelectorDeps == nil {
		return DatabaseOption{}
	}
	return m.runtimeDatabaseSelectorDeps.CurrentDatabase
}

func formatRuntimeDatabaseSelectionFailure(selected DatabaseOption, reason string) string {
	return fmt.Sprintf(
		"Connection failed for %q: %s. Choose another database or edit selected entry.",
		selected.Name,
		reason,
	)
}

type runtimeDatabaseSwitchCompletedMsg struct {
	request  runtimeDatabaseTransitionRequest
	snapshot runtimeDatabaseTransitionSnapshot
	deps     RuntimeRunDeps
	err      error
}

func switchRuntimeDatabaseCmd(
	ctx context.Context,
	switcher RuntimeDatabaseSwitcher,
	request runtimeDatabaseTransitionRequest,
	snapshot runtimeDatabaseTransitionSnapshot,
) tea.Cmd {
	return func() tea.Msg {
		if switcher == nil {
			return runtimeDatabaseSwitchCompletedMsg{
				request:  request,
				snapshot: snapshot,
				err:      fmt.Errorf("database selector unavailable"),
			}
		}
		deps, err := switcher.Switch(ctx, request.Target.Option)
		return runtimeDatabaseSwitchCompletedMsg{
			request:  request,
			snapshot: snapshot,
			deps:     deps,
			err:      err,
		}
	}
}
