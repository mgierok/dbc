package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/dto"
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
	Target runtimeDatabaseTransitionTarget
	Force  bool
	Origin runtimeDatabaseTransitionOrigin
}

type runtimeDatabaseTransitionSnapshot struct {
	SelectedTableName string
	Focus             PanelFocus
	ViewMode          ViewMode
	Filter            *dto.Filter
	Sort              *dto.Sort
	PageIndex         int
}

type runtimeDatabaseReloadRestoreState struct {
	snapshot           runtimeDatabaseTransitionSnapshot
	requestedPageIndex int
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
	switcher := m.runtimeDatabaseSwitcher()
	if switcher == nil {
		m.ui.statusMessage = "Error: database selector unavailable"
		return m, nil
	}

	snapshot := runtimeDatabaseTransitionSnapshot{}
	if request.Target.Kind == reloadCurrentDatabase {
		snapshot = m.captureRuntimeDatabaseTransitionSnapshot()
	}

	m.ui.pendingDatabaseTransition = nil
	m.ui.runtimeSwitchInFlight = true
	if request.Origin == runtimeDatabaseTransitionOriginConfigSelector && m.overlay.databaseSelector.controller != nil {
		m.overlay.databaseSelector.controller.SetStatusMessage(runtimeDatabaseTransitionInFlightStatus(request))
	}
	return m, switchRuntimeDatabaseCmd(m.ctx, switcher, request, snapshot)
}

func (m *Model) confirmDatabaseTransitionSave() (tea.Model, tea.Cmd) {
	if m.ui.pendingDatabaseTransition == nil {
		return m, nil
	}

	updatedModel, cmd := m.confirmSaveChanges()
	if cmd == nil {
		m.ui.pendingDatabaseTransition = nil
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

func runtimeDatabaseTransitionInFlightStatus(request runtimeDatabaseTransitionRequest) string {
	switch request.Target.Kind {
	case reloadCurrentDatabase:
		return fmt.Sprintf("Reloading %q...", request.Target.Option.Name)
	default:
		return fmt.Sprintf("Opening %q...", request.Target.Option.Name)
	}
}

func (m *Model) captureRuntimeDatabaseTransitionSnapshot() runtimeDatabaseTransitionSnapshot {
	return runtimeDatabaseTransitionSnapshot{
		SelectedTableName: m.currentTableName(),
		Focus:             m.read.focus,
		ViewMode:          m.read.viewMode,
		Filter:            cloneFilter(m.read.currentFilter),
		Sort:              cloneSort(m.read.currentSort),
		PageIndex:         m.read.recordPageIndex,
	}
}

func cloneFilter(filter *dto.Filter) *dto.Filter {
	if filter == nil {
		return nil
	}
	cloned := *filter
	return &cloned
}

func cloneSort(sort *dto.Sort) *dto.Sort {
	if sort == nil {
		return nil
	}
	cloned := *sort
	return &cloned
}

func cloneRuntimeDatabaseTransitionRequest(request runtimeDatabaseTransitionRequest) runtimeDatabaseTransitionRequest {
	return runtimeDatabaseTransitionRequest{
		Target: runtimeDatabaseTransitionTarget{
			Option: request.Target.Option,
			Kind:   request.Target.Kind,
		},
		Force:  request.Force,
		Origin: request.Origin,
	}
}

func cloneRuntimeDatabaseReloadRestoreState(snapshot runtimeDatabaseTransitionSnapshot) *runtimeDatabaseReloadRestoreState {
	return &runtimeDatabaseReloadRestoreState{
		snapshot: runtimeDatabaseTransitionSnapshot{
			SelectedTableName: snapshot.SelectedTableName,
			Focus:             snapshot.Focus,
			ViewMode:          snapshot.ViewMode,
			Filter:            cloneFilter(snapshot.Filter),
			Sort:              cloneSort(snapshot.Sort),
			PageIndex:         snapshot.PageIndex,
		},
		requestedPageIndex: snapshot.PageIndex,
	}
}

func (m *Model) applyPendingDatabaseReloadRestoreAfterTables() tea.Cmd {
	restore := m.ui.pendingDatabaseReloadRestore
	if restore == nil {
		return m.loadSchemaCmd()
	}

	selectedTableIndex := indexTableByName(m.read.tables, restore.snapshot.SelectedTableName)
	if selectedTableIndex < 0 {
		m.ui.pendingDatabaseReloadRestore = nil
		return m.loadSchemaCmd()
	}

	m.read.selectedTable = selectedTableIndex
	m.read.focus = restore.snapshot.Focus
	m.read.viewMode = restore.snapshot.ViewMode
	m.read.currentFilter = cloneFilter(restore.snapshot.Filter)
	m.read.currentSort = cloneSort(restore.snapshot.Sort)

	if restore.snapshot.ViewMode != ViewRecords {
		m.ui.pendingDatabaseReloadRestore = nil
		return m.loadSchemaCmd()
	}

	m.read.recordPageIndex = maxInt(restore.snapshot.PageIndex, 0)
	return tea.Batch(m.loadSchemaCmd(), m.loadRecordsCmd(false))
}

func indexTableByName(tables []dto.Table, name string) int {
	for i, table := range tables {
		if table.Name == name {
			return i
		}
	}
	return -1
}

func maxInt(left, right int) int {
	if left > right {
		return left
	}
	return right
}
