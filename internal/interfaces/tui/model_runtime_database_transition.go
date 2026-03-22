package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/usecase"
	"github.com/mgierok/dbc/internal/sqliteidentity"
)

type runtimeDatabaseTransitionKind int

const (
	reloadCurrentDatabase runtimeDatabaseTransitionKind = iota + 1
	switchDifferentDatabase
)

type runtimeDatabaseTransitionOrigin string

const (
	runtimeDatabaseTransitionOriginConfigSelector runtimeDatabaseTransitionOrigin = "config-selector"
	runtimeDatabaseTransitionOriginEditCommand    runtimeDatabaseTransitionOrigin = "edit-command"
)

type runtimeDatabaseTransitionTarget struct {
	Option DatabaseOption
	Kind   runtimeDatabaseTransitionKind
}

type runtimeDatabaseTransitionRequest struct {
	Target  runtimeDatabaseTransitionTarget
	Force   bool
	Origin  runtimeDatabaseTransitionOrigin
	Command string
}

func (m *Model) requestRuntimeDatabaseTransition(request runtimeDatabaseTransitionRequest) (tea.Model, tea.Cmd) {
	if m.hasDirtyEdits() && !request.Force {
		clonedRequest := cloneRuntimeDatabaseTransitionRequest(request)
		m.ui.pendingDatabaseTransition = &clonedRequest
		prompt := m.databaseTransitionDirtyPrompt(request.Target.Kind)
		m.openModalConfirmPopupWithOptions(
			prompt.Title,
			prompt.Message,
			m.confirmOptionsFromDirtyPrompt(prompt, dirtyConfirmFlowDatabaseTransition),
			0,
		)
		return m, nil
	}

	return m.executeRuntimeDatabaseTransition(request)
}

func (m *Model) executeRuntimeDatabaseTransition(request runtimeDatabaseTransitionRequest) (tea.Model, tea.Cmd) {
	m.ui.pendingDatabaseTransition = nil
	if request.Origin == runtimeDatabaseTransitionOriginEditCommand {
		m.overlay.commandInput = commandInput{}
	}
	m.exitResult = runtimeExitResultOpenDatabaseNext(request.Target.Option)
	return m, tea.Quit
}

func (m *Model) confirmDatabaseTransitionSave() (tea.Model, tea.Cmd) {
	request := m.ui.pendingDatabaseTransition
	if request == nil {
		return m, nil
	}

	if request.Origin == runtimeDatabaseTransitionOriginEditCommand {
		m.overlay.commandInput = commandInput{}
	}
	updatedModel, cmd := m.confirmSaveChanges()
	if cmd == nil {
		m.ui.pendingDatabaseTransition = nil
		if request.Origin == runtimeDatabaseTransitionOriginEditCommand {
			m.restoreEditingCommandInput(request.Command)
		}
	}
	return updatedModel, cmd
}

func (m *Model) confirmDatabaseTransitionDiscard() (tea.Model, tea.Cmd) {
	request := m.ui.pendingDatabaseTransition
	if request == nil {
		return m, nil
	}

	cloned := cloneRuntimeDatabaseTransitionRequest(*request)
	m.ui.pendingDatabaseTransition = nil
	m.clearStagedState()
	return m.executeRuntimeDatabaseTransition(cloned)
}

func (m *Model) resolveRuntimeDatabaseTransitionTargetFromOption(selected DatabaseOption) (runtimeDatabaseTransitionTarget, error) {
	return m.resolveRuntimeDatabaseTransitionTargetFromConnString(selected.ConnString)
}

func (m *Model) resolveRuntimeDatabaseTransitionTargetFromConnString(connString string) (runtimeDatabaseTransitionTarget, error) {
	trimmedConnString := strings.TrimSpace(connString)
	if trimmedConnString == "" {
		current := m.currentRuntimeDatabaseOption()
		if strings.TrimSpace(current.ConnString) == "" {
			return runtimeDatabaseTransitionTarget{}, fmt.Errorf("current database unavailable")
		}
		return runtimeDatabaseTransitionTarget{
			Option: current,
			Kind:   reloadCurrentDatabase,
		}, nil
	}

	configuredOptions, err := m.configuredRuntimeDatabaseOptions()
	if err != nil {
		return runtimeDatabaseTransitionTarget{}, err
	}

	resolvedOption := DatabaseOption{
		Name:       trimmedConnString,
		ConnString: trimmedConnString,
		Source:     DatabaseOptionSourceCLI,
	}
	if matched, ok := resolveConfiguredRuntimeDatabaseIdentity(trimmedConnString, configuredOptions); ok {
		matched.Source = DatabaseOptionSourceConfig
		resolvedOption = matched
	}

	kind := switchDifferentDatabase
	if sqliteidentity.Equivalent(resolvedOption.ConnString, m.currentRuntimeDatabaseConnString()) {
		kind = reloadCurrentDatabase
	}
	return runtimeDatabaseTransitionTarget{
		Option: resolvedOption,
		Kind:   kind,
	}, nil
}

func resolveConfiguredRuntimeDatabaseIdentity(connString string, configuredOptions []DatabaseOption) (DatabaseOption, bool) {
	normalizedConnString := sqliteidentity.Normalize(connString)
	if normalizedConnString == "" {
		return DatabaseOption{}, false
	}

	for _, option := range configuredOptions {
		if sqliteidentity.Equivalent(normalizedConnString, option.ConnString) {
			return option, true
		}
	}

	return DatabaseOption{}, false
}

func (m *Model) configuredRuntimeDatabaseOptions() ([]DatabaseOption, error) {
	if m.runtimeDatabaseSelectorDeps == nil || m.runtimeDatabaseSelectorDeps.ListConfiguredDatabases == nil {
		return nil, fmt.Errorf("database selector unavailable")
	}

	entries, err := m.runtimeDatabaseSelectorDeps.ListConfiguredDatabases.Execute(m.ctx)
	if err != nil {
		return nil, err
	}

	options := make([]DatabaseOption, len(entries))
	for i, entry := range entries {
		options[i] = DatabaseOption{
			Name:       entry.Name,
			ConnString: entry.Path,
			Source:     DatabaseOptionSourceConfig,
		}
	}
	return options, nil
}

func (m *Model) databaseTransitionDirtyPrompt(kind runtimeDatabaseTransitionKind) usecase.DirtyDecisionPrompt {
	switch kind {
	case reloadCurrentDatabase:
		return m.dirtyNavigationPolicyUseCase().BuildDatabaseReloadPrompt(m.dirtyEditCount())
	default:
		return m.dirtyNavigationPolicyUseCase().BuildDatabaseOpenPrompt(m.dirtyEditCount())
	}
}

func cloneRuntimeDatabaseTransitionRequest(request runtimeDatabaseTransitionRequest) runtimeDatabaseTransitionRequest {
	return runtimeDatabaseTransitionRequest{
		Target: runtimeDatabaseTransitionTarget{
			Option: request.Target.Option,
			Kind:   request.Target.Kind,
		},
		Force:   request.Force,
		Origin:  request.Origin,
		Command: request.Command,
	}
}
