package selector

import (
	"context"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/interfaces/tui/internal/primitives"
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
	active       bool
	optionIndex  int
	managerIndex int
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

type databaseSelectorModel struct {
	ctx     context.Context
	manager selectorManager
	styles  primitives.RenderStyles

	options  []DatabaseOption
	width    int
	height   int
	chosen   bool
	canceled bool

	mode          selectorMode
	browse        selectorBrowseState
	form          selectorForm
	confirmDelete selectorDeleteConfirm
	helpPopup     selectorHelpPopup

	activeConfigPath string

	launchAdditionalOptions []DatabaseOption
	configOptionCount       int
	requiresFirstEntry      bool
}

var detectRenderStyles = primitives.ResolveRenderStylesFromEnv

func newDatabaseSelectorModel(ctx context.Context, manager selectorManager, launchState ...SelectorLaunchState) (*databaseSelectorModel, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	state := SelectorLaunchState{}
	if len(launchState) > 0 {
		state = launchState[0]
	}
	model := &databaseSelectorModel{
		ctx:                     ctx,
		manager:                 manager,
		styles:                  detectRenderStyles(),
		mode:                    selectorModeBrowse,
		launchAdditionalOptions: normalizeAdditionalOptions(state.AdditionalOptions),
	}
	if err := model.refreshOptions(); err != nil {
		return nil, err
	}
	if err := model.refreshActivePath(); err != nil {
		return nil, err
	}
	model.applyLaunchState(state)
	if len(model.options) == 0 {
		model.requiresFirstEntry = true
		model.openAddForm()
		model.browse.statusMessage = "First database entry is required"
	}
	return model, nil
}

func (m *databaseSelectorModel) applyLaunchState(state SelectorLaunchState) {
	if strings.TrimSpace(state.StatusMessage) != "" {
		m.browse.statusMessage = state.StatusMessage
	}
	preferredConnString := strings.TrimSpace(state.PreferConnString)
	if preferredConnString == "" || len(m.options) == 0 {
		return
	}
	for i, option := range m.options {
		if strings.TrimSpace(option.ConnString) == preferredConnString {
			m.browse.selected = i
			return
		}
	}
}

func (m *databaseSelectorModel) refreshOptions() error {
	entries, err := m.manager.List(m.ctx)
	if err != nil {
		return err
	}
	configOptions := make([]DatabaseOption, len(entries))
	for i, entry := range entries {
		configOptions[i] = DatabaseOption{
			Name:         entry.Name,
			ConnString:   entry.Path,
			Source:       DatabaseOptionSourceConfig,
			managerIndex: i,
		}
	}
	m.configOptionCount = len(configOptions)
	m.options = mergeConfigAndAdditionalOptions(configOptions, m.launchAdditionalOptions)
	if len(m.options) == 0 {
		m.browse.selected = 0
		return nil
	}
	m.browse.selected = clamp(m.browse.selected, 0, len(m.options)-1)
	return nil
}

func (m *databaseSelectorModel) refreshActivePath() error {
	path, err := m.manager.ActivePath(m.ctx)
	if err != nil {
		return err
	}
	m.activeConfigPath = path
	return nil
}

func (m *databaseSelectorModel) Init() tea.Cmd {
	return nil
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
			m.canceled = true
			return m, tea.Quit
		case primitives.KeyMatches(primitives.KeySelectorEnter, key):
			if len(m.options) == 0 {
				if m.requiresFirstEntry {
					m.openAddForm()
					m.browse.statusMessage = "First database entry is required"
				}
				return m, nil
			}
			m.chosen = true
			return m, tea.Quit
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
