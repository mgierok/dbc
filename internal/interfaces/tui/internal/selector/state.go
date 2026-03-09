package selector

import (
	"context"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/dto"
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

type databaseSelectorModel struct {
	ctx     context.Context
	manager selectorManager
	styles  renderStyles

	options  []DatabaseOption
	selected int
	width    int
	height   int
	chosen   bool
	canceled bool

	mode          selectorMode
	form          selectorForm
	confirmDelete selectorDeleteConfirm
	helpPopup     selectorHelpPopup

	activeConfigPath string
	statusMessage    string

	launchAdditionalOptions []DatabaseOption
	configOptionCount       int
	requiresFirstEntry      bool
}

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
		model.statusMessage = "First database entry is required"
	}
	return model, nil
}

func (m *databaseSelectorModel) applyLaunchState(state SelectorLaunchState) {
	if strings.TrimSpace(state.StatusMessage) != "" {
		m.statusMessage = state.StatusMessage
	}
	preferredConnString := strings.TrimSpace(state.PreferConnString)
	if preferredConnString == "" || len(m.options) == 0 {
		return
	}
	for i, option := range m.options {
		if strings.TrimSpace(option.ConnString) == preferredConnString {
			m.selected = i
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
		m.selected = 0
		return nil
	}
	m.selected = clamp(m.selected, 0, len(m.options)-1)
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
		if keyMatches(keySelectorOpenContextHelp, key) {
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
		case keyMatches(keySelectorCancel, key):
			m.canceled = true
			return m, tea.Quit
		case keyMatches(keySelectorEnter, key):
			if len(m.options) == 0 {
				if m.requiresFirstEntry {
					m.openAddForm()
					m.statusMessage = "First database entry is required"
				}
				return m, nil
			}
			m.chosen = true
			return m, tea.Quit
		case keyMatches(keySelectorMoveDown, key):
			m.moveSelection(1)
			return m, nil
		case keyMatches(keySelectorMoveUp, key):
			m.moveSelection(-1)
			return m, nil
		case keyMatches(keySelectorJumpTop, key):
			m.selectTop()
			return m, nil
		case keyMatches(keySelectorJumpBottom, key):
			m.selectBottom()
			return m, nil
		case keyMatches(keySelectorPageDown, key):
			m.page(1)
			return m, nil
		case keyMatches(keySelectorPageUp, key):
			m.page(-1)
			return m, nil
		case keyMatches(keySelectorAdd, key):
			m.openAddForm()
			return m, nil
		case keyMatches(keySelectorEdit, key):
			if m.requiresFirstEntry {
				m.statusMessage = "Edit is unavailable during first setup"
				return m, nil
			}
			m.openEditForm()
			return m, nil
		case keyMatches(keySelectorDelete, key):
			if m.requiresFirstEntry {
				m.statusMessage = "Delete is unavailable during first setup"
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

func (m *databaseSelectorModel) openHelpPopup() {
	m.helpPopup = selectorHelpPopup{
		active:       true,
		scrollOffset: 0,
		context:      m.mode,
	}
}

func (m *databaseSelectorModel) closeHelpPopup() {
	m.helpPopup = selectorHelpPopup{}
}

func (m *databaseSelectorModel) handleHelpPopupKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch {
	case keyMatches(keyRuntimeEsc, key):
		m.closeHelpPopup()
		return m, nil
	case keyMatches(keyPopupMoveDown, key):
		m.moveHelpPopupScroll(1)
		return m, nil
	case keyMatches(keyPopupMoveUp, key):
		m.moveHelpPopupScroll(-1)
		return m, nil
	case keyMatches(keyRuntimePageDown, key):
		m.moveHelpPopupScroll(m.helpPopupVisibleLines())
		return m, nil
	case keyMatches(keyRuntimePageUp, key):
		m.moveHelpPopupScroll(-m.helpPopupVisibleLines())
		return m, nil
	case keyMatches(keyPopupJumpTop, key):
		m.helpPopup.scrollOffset = 0
		return m, nil
	case keyMatches(keyPopupJumpBottom, key):
		m.helpPopup.scrollOffset = m.helpPopupMaxOffset()
		return m, nil
	default:
		return m, nil
	}
}

func (m *databaseSelectorModel) helpPopupVisibleLines() int {
	height := m.height
	if height <= 0 {
		height = 24
	}
	visible := height - 8
	if visible < 1 {
		visible = 1
	}
	return visible
}

func (m *databaseSelectorModel) helpPopupMaxOffset() int {
	rows := m.helpPopupContentLines()
	maxOffset := len(rows) - m.helpPopupVisibleLines()
	if maxOffset < 0 {
		maxOffset = 0
	}
	return maxOffset
}

func (m *databaseSelectorModel) moveHelpPopupScroll(delta int) {
	maxOffset := m.helpPopupMaxOffset()
	m.helpPopup.scrollOffset = clamp(m.helpPopup.scrollOffset+delta, 0, maxOffset)
}

func (m *databaseSelectorModel) moveSelection(delta int) {
	if len(m.options) == 0 {
		return
	}
	m.selected = clamp(m.selected+delta, 0, len(m.options)-1)
}

func (m *databaseSelectorModel) selectTop() {
	if len(m.options) == 0 {
		return
	}
	m.selected = 0
}

func (m *databaseSelectorModel) selectBottom() {
	if len(m.options) == 0 {
		return
	}
	m.selected = len(m.options) - 1
}

func (m *databaseSelectorModel) page(delta int) {
	page := m.listHeight(m.height)
	if page < 1 {
		page = 1
	}
	m.moveSelection(delta * page)
}

func (m *databaseSelectorModel) openAddForm() {
	m.mode = selectorModeAdd
	m.form = selectorForm{
		editIndex:   -1,
		activeField: selectorInputName,
	}
}

func (m *databaseSelectorModel) openEditForm() {
	if len(m.options) == 0 {
		m.statusMessage = "No database selected to edit"
		return
	}
	selected := m.options[m.selected]
	if !selected.isConfigBacked() {
		m.statusMessage = "CLI session entry cannot be edited"
		return
	}
	m.mode = selectorModeEdit
	m.form = selectorForm{
		editIndex:   selected.managerIndex,
		activeField: selectorInputName,
		nameValue:   selected.Name,
		pathValue:   selected.ConnString,
	}
}

func (m *databaseSelectorModel) openDeleteConfirmation() {
	if len(m.options) == 0 {
		m.statusMessage = "No database selected to delete"
		return
	}
	selected := m.options[m.selected]
	if !selected.isConfigBacked() {
		m.statusMessage = "CLI session entry cannot be deleted"
		return
	}
	m.mode = selectorModeConfirmDelete
	m.confirmDelete = selectorDeleteConfirm{
		active:       true,
		optionIndex:  m.selected,
		managerIndex: selected.managerIndex,
	}
}

func (m *databaseSelectorModel) handleFormKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch {
	case keyMatches(keySelectorFormEsc, key):
		if m.requiresFirstEntry && len(m.options) == 0 {
			m.canceled = true
			return m, tea.Quit
		}
		m.mode = selectorModeBrowse
		m.form = selectorForm{}
		m.statusMessage = "Config update canceled"
		return m, nil
	case keyMatches(keySelectorFormSwitch, key):
		if m.form.activeField == selectorInputName {
			m.form.activeField = selectorInputPath
		} else {
			m.form.activeField = selectorInputName
		}
		return m, nil
	case keyMatches(keySelectorFormClear, key):
		m.setActiveFormValue("")
		m.form.errorMessage = ""
		return m, nil
	case keyMatches(keySelectorFormBackspace, key):
		value := m.activeFormValue()
		if value == "" {
			return m, nil
		}
		m.setActiveFormValue(value[:len(value)-1])
		m.form.errorMessage = ""
		return m, nil
	case keyMatches(keySelectorEnter, key):
		return m.submitForm()
	default:
		if len(msg.Runes) == 0 {
			return m, nil
		}
		m.setActiveFormValue(m.activeFormValue() + string(msg.Runes))
		m.form.errorMessage = ""
		return m, nil
	}
}

func (m *databaseSelectorModel) handleDeleteConfirmationKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch {
	case keyMatches(keySelectorDeleteCancel, key):
		m.mode = selectorModeBrowse
		m.confirmDelete = selectorDeleteConfirm{}
		m.statusMessage = "Delete canceled"
		return m, nil
	case keyMatches(keySelectorDeleteConfirm, key):
		optionIndex := m.confirmDelete.optionIndex
		managerIndex := m.confirmDelete.managerIndex
		if optionIndex < 0 || optionIndex >= len(m.options) || managerIndex < 0 {
			m.mode = selectorModeBrowse
			m.confirmDelete = selectorDeleteConfirm{}
			m.statusMessage = "Invalid selection for delete"
			return m, nil
		}
		if err := m.manager.Delete(m.ctx, managerIndex); err != nil {
			m.mode = selectorModeBrowse
			m.confirmDelete = selectorDeleteConfirm{}
			m.statusMessage = "Delete failed: " + err.Error()
			return m, nil
		}

		m.mode = selectorModeBrowse
		m.confirmDelete = selectorDeleteConfirm{}
		if err := m.refreshOptions(); err != nil {
			m.statusMessage = "Delete succeeded but refresh failed: " + err.Error()
			return m, nil
		}
		m.statusMessage = "Database deleted"
		return m, nil
	default:
		return m, nil
	}
}

func (m *databaseSelectorModel) submitForm() (tea.Model, tea.Cmd) {
	name := strings.TrimSpace(m.form.nameValue)
	path := strings.TrimSpace(m.form.pathValue)
	if name == "" {
		m.form.errorMessage = "Name is required"
		return m, nil
	}
	if path == "" {
		m.form.errorMessage = "Path is required"
		return m, nil
	}

	entry := dto.ConfigDatabase{
		Name: name,
		Path: path,
	}

	var err error
	switch m.mode {
	case selectorModeAdd:
		err = m.manager.Create(m.ctx, entry)
	case selectorModeEdit:
		err = m.manager.Update(m.ctx, m.form.editIndex, entry)
	default:
		return m, nil
	}
	if err != nil {
		m.form.errorMessage = err.Error()
		return m, nil
	}

	previousMode := m.mode
	previousEditIndex := m.form.editIndex
	m.mode = selectorModeBrowse
	m.form = selectorForm{}

	if err := m.refreshOptions(); err != nil {
		m.statusMessage = "Config updated but refresh failed: " + err.Error()
		return m, nil
	}
	if previousMode == selectorModeAdd && len(m.options) > 0 {
		m.selected = m.configOptionCount - 1
		if m.requiresFirstEntry {
			m.statusMessage = "Database added. Press Enter to continue or a to add another"
			return m, nil
		}
		m.statusMessage = "Database added"
		return m, nil
	}
	if previousMode == selectorModeEdit && m.configOptionCount > 0 {
		m.selected = clamp(previousEditIndex, 0, m.configOptionCount-1)
		m.statusMessage = "Database updated"
	}
	return m, nil
}

func (m *databaseSelectorModel) activeFormValue() string {
	if m.form.activeField == selectorInputPath {
		return m.form.pathValue
	}
	return m.form.nameValue
}

func (m *databaseSelectorModel) setActiveFormValue(value string) {
	if m.form.activeField == selectorInputPath {
		m.form.pathValue = value
		return
	}
	m.form.nameValue = value
}
