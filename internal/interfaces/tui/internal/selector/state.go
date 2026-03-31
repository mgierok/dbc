package selector

import (
	"context"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/dto"
	"github.com/mgierok/dbc/internal/interfaces/tui/internal/primitives"
	"github.com/mgierok/dbc/internal/sqliteidentity"
)

type selectorMode int

const (
	selectorModeBrowse selectorMode = iota
	selectorModeAdd
	selectorModeEdit
	selectorModeConfirmDelete
)

type selectorInputField int

const (
	selectorInputName selectorInputField = iota
	selectorInputPath
)

type selectorForm struct {
	editIndex    int
	activeField  selectorInputField
	nameValue    string
	pathValue    string
	errorMessage string
}

type selectorDeleteConfirm struct {
	active      bool
	optionIndex int
	configIndex int
}

type selectorHelpPopup struct {
	active       bool
	scrollOffset int
	context      selectorMode
}

type selectorBrowseState struct {
	selected      int
	statusMessage string
}

type controllerHostConfig struct {
	browseEscActionLabel     string
	firstSetupEscActionLabel string
}

type databaseSelectorModel struct {
	ctx     context.Context
	manager selectorManager
	styles  primitives.RenderStyles

	options   []DatabaseOption
	width     int
	height    int
	chosen    bool
	canceled  bool
	dismissed bool
	intent    Intent

	mode          selectorMode
	browse        selectorBrowseState
	form          selectorForm
	confirmDelete selectorDeleteConfirm
	helpPopup     selectorHelpPopup

	activeConfigPath string

	launchAdditionalOptions []dto.DatabaseSelectorAdditionalOption
	browseEscBehavior       SelectorBrowseEscBehavior
	hostConfig              controllerHostConfig
	configOptionCount       int
	requiresFirstEntry      bool
}

var detectRenderStyles = primitives.ResolveRenderStylesFromEnv

func newDatabaseSelectorModel(ctx context.Context, manager selectorManager, launchState ...SelectorLaunchState) (*databaseSelectorModel, error) {
	state := SelectorLaunchState{}
	if len(launchState) > 0 {
		state = launchState[0]
	}
	hostConfig := controllerHostConfig{
		browseEscActionLabel:     "quit",
		firstSetupEscActionLabel: "exit app",
	}
	if state.BrowseEscBehavior == SelectorBrowseEscBehaviorRuntimeResume {
		hostConfig.browseEscActionLabel = "close"
		hostConfig.firstSetupEscActionLabel = "close"
	}
	return newDatabaseSelectorModelWithHost(ctx, manager, controllerHostConfig{
		browseEscActionLabel:     hostConfig.browseEscActionLabel,
		firstSetupEscActionLabel: hostConfig.firstSetupEscActionLabel,
	}, state)
}

func newDatabaseSelectorModelWithHost(ctx context.Context, manager selectorManager, hostConfig controllerHostConfig, launchState SelectorLaunchState) (*databaseSelectorModel, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	model := &databaseSelectorModel{
		ctx:                     ctx,
		manager:                 manager,
		styles:                  detectRenderStyles(),
		mode:                    selectorModeBrowse,
		launchAdditionalOptions: mapLaunchAdditionalOptions(launchState.AdditionalOptions),
		browseEscBehavior:       launchState.BrowseEscBehavior,
		hostConfig:              hostConfig,
	}
	if err := model.refreshOptions(); err != nil {
		return nil, err
	}
	model.applyLaunchState(launchState)
	if model.requiresFirstEntry && len(model.options) == 0 {
		model.openAddForm()
		model.browse.statusMessage = "First database entry is required"
	}
	return model, nil
}

func (m *databaseSelectorModel) browseEscActionLabel() string {
	label := strings.TrimSpace(m.hostConfig.browseEscActionLabel)
	if label == "" {
		return "quit"
	}
	return label
}

func (m *databaseSelectorModel) firstSetupEscActionLabel() string {
	label := strings.TrimSpace(m.hostConfig.firstSetupEscActionLabel)
	if label == "" {
		return "exit app"
	}
	return label
}

func (m *databaseSelectorModel) applyLaunchState(state SelectorLaunchState) {
	if strings.TrimSpace(state.StatusMessage) != "" {
		m.browse.statusMessage = state.StatusMessage
	}
	preferredConnString := sqliteidentity.Normalize(state.PreferConnString)
	if preferredConnString == "" || len(m.options) == 0 {
		return
	}
	for i, option := range m.options {
		if sqliteidentity.Equivalent(option.ConnString, preferredConnString) {
			m.browse.selected = i
			return
		}
	}
}

func (m *databaseSelectorModel) refreshOptions() error {
	state, err := m.manager.LoadState(m.ctx, dto.DatabaseSelectorLoadInput{
		AdditionalOptions: cloneDatabaseSelectorAdditionalOptions(m.launchAdditionalOptions),
	})
	if err != nil {
		return err
	}
	m.activeConfigPath = state.ActiveConfigPath
	m.requiresFirstEntry = state.RequiresFirstEntry
	m.options = mapDatabaseSelectorOptions(state.Options)
	m.configOptionCount = countConfigBackedOptions(m.options)
	if len(m.options) == 0 {
		m.browse.selected = 0
		return nil
	}
	m.browse.selected = clamp(m.browse.selected, 0, len(m.options)-1)
	return nil
}

func (m *databaseSelectorModel) Init() tea.Cmd {
	return nil
}

func (m *databaseSelectorModel) peekIntent() Intent {
	return m.intent
}

func (m *databaseSelectorModel) consumeIntent() Intent {
	intent := m.intent
	m.intent = Intent{}
	return intent
}

func (m *databaseSelectorModel) requestClose() {
	if m.browseEscBehavior == SelectorBrowseEscBehaviorRuntimeResume {
		m.dismissed = true
	} else {
		m.canceled = true
	}
	m.intent = Intent{Type: IntentTypeClose}
}

func (m *databaseSelectorModel) requestSelect() {
	m.chosen = true
	if m.browse.selected < 0 || m.browse.selected >= len(m.options) {
		m.intent = Intent{Type: IntentTypeSelect}
		return
	}
	m.intent = Intent{
		Type:   IntentTypeSelect,
		Option: m.options[m.browse.selected],
	}
}

func (m *databaseSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		key := msg.String()
		if m.helpPopup.active {
			return m.handleHelpPopupKey(msg)
		}
		if primitives.KeyMatches(primitives.KeySelectorOpenContextHelp, key) {
			m.openHelpPopup()
			return m, nil
		}

		if m.mode == selectorModeAdd || m.mode == selectorModeEdit {
			return m.handleFormKey(msg)
		}
		if m.mode == selectorModeConfirmDelete {
			return m.handleDeleteConfirmationKey(msg)
		}

		switch {
		case primitives.KeyMatches(primitives.KeySelectorCancel, key):
			m.requestClose()
			return m, nil
		case primitives.KeyMatches(primitives.KeySelectorEnter, key):
			if len(m.options) == 0 {
				if m.requiresFirstEntry {
					m.openAddForm()
					m.browse.statusMessage = "First database entry is required"
				}
				return m, nil
			}
			m.requestSelect()
			return m, nil
		case primitives.KeyMatches(primitives.KeySelectorMoveDown, key):
			m.moveSelection(1)
			return m, nil
		case primitives.KeyMatches(primitives.KeySelectorMoveUp, key):
			m.moveSelection(-1)
			return m, nil
		case primitives.KeyMatches(primitives.KeySelectorJumpTop, key):
			m.selectTop()
			return m, nil
		case primitives.KeyMatches(primitives.KeySelectorJumpBottom, key):
			m.selectBottom()
			return m, nil
		case primitives.KeyMatches(primitives.KeySelectorPageDown, key):
			m.page(1)
			return m, nil
		case primitives.KeyMatches(primitives.KeySelectorPageUp, key):
			m.page(-1)
			return m, nil
		case primitives.KeyMatches(primitives.KeySelectorAdd, key):
			m.openAddForm()
			return m, nil
		case primitives.KeyMatches(primitives.KeySelectorEdit, key):
			if m.requiresFirstEntry {
				m.browse.statusMessage = "Edit is unavailable during first setup"
				return m, nil
			}
			m.openEditForm()
			return m, nil
		case primitives.KeyMatches(primitives.KeySelectorDelete, key):
			if m.requiresFirstEntry {
				m.browse.statusMessage = "Delete is unavailable during first setup"
				return m, nil
			}
			m.openDeleteConfirmation()
			return m, nil
		default:
			return m, nil
		}
	default:
		return m, nil
	}
}

func mapLaunchAdditionalOptions(options []DatabaseOption) []dto.DatabaseSelectorAdditionalOption {
	if len(options) == 0 {
		return nil
	}
	mapped := make([]dto.DatabaseSelectorAdditionalOption, len(options))
	for i, option := range options {
		source := dto.DatabaseSelectorOptionSourceConfig
		if option.Source == DatabaseOptionSourceCLI {
			source = dto.DatabaseSelectorOptionSourceCLI
		}
		mapped[i] = dto.DatabaseSelectorAdditionalOption{
			Name:       option.Name,
			ConnString: option.ConnString,
			Source:     source,
		}
	}
	return mapped
}

func cloneDatabaseSelectorAdditionalOptions(options []dto.DatabaseSelectorAdditionalOption) []dto.DatabaseSelectorAdditionalOption {
	if len(options) == 0 {
		return nil
	}
	cloned := make([]dto.DatabaseSelectorAdditionalOption, len(options))
	copy(cloned, options)
	return cloned
}

func mapDatabaseSelectorOptions(options []dto.DatabaseSelectorOption) []DatabaseOption {
	if len(options) == 0 {
		return nil
	}
	mapped := make([]DatabaseOption, len(options))
	for i, option := range options {
		source := DatabaseOptionSourceConfig
		if option.Source == dto.DatabaseSelectorOptionSourceCLI {
			source = DatabaseOptionSourceCLI
		}
		mapped[i] = DatabaseOption{
			Name:        option.Name,
			ConnString:  option.ConnString,
			Source:      source,
			configIndex: option.ConfigIndex,
			canEdit:     option.CanEdit,
			canDelete:   option.CanDelete,
		}
	}
	return mapped
}

func countConfigBackedOptions(options []DatabaseOption) int {
	count := 0
	for _, option := range options {
		if option.configIndex >= 0 {
			count++
		}
	}
	return count
}
