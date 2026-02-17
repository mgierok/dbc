package tui

import (
	"context"
	"errors"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/dto"
	"github.com/mgierok/dbc/internal/application/usecase"
)

var (
	ErrDatabaseSelectionCanceled   = errors.New("database selection canceled")
	ErrDatabaseSelectionUnfinished = errors.New("database selection not confirmed")
)

type DatabaseOptionSource string

const (
	DatabaseOptionSourceConfig DatabaseOptionSource = "config"
	DatabaseOptionSourceCLI    DatabaseOptionSource = "cli"
)

type DatabaseOption struct {
	Name       string
	ConnString string
	Source     DatabaseOptionSource

	managerIndex int
}

type SelectorLaunchState struct {
	StatusMessage     string
	PreferConnString  string
	AdditionalOptions []DatabaseOption
}

type selectorManager interface {
	List(ctx context.Context) ([]dto.ConfigDatabase, error)
	Create(ctx context.Context, entry dto.ConfigDatabase) error
	Update(ctx context.Context, index int, entry dto.ConfigDatabase) error
	Delete(ctx context.Context, index int) error
	ActivePath(ctx context.Context) (string, error)
}

type selectorUseCaseAdapter struct {
	list   *usecase.ListConfiguredDatabases
	create *usecase.CreateConfiguredDatabase
	update *usecase.UpdateConfiguredDatabase
	del    *usecase.DeleteConfiguredDatabase
	active *usecase.GetActiveConfigPath
}

func (a selectorUseCaseAdapter) List(ctx context.Context) ([]dto.ConfigDatabase, error) {
	return a.list.Execute(ctx)
}

func (a selectorUseCaseAdapter) Create(ctx context.Context, entry dto.ConfigDatabase) error {
	return a.create.Execute(ctx, entry)
}

func (a selectorUseCaseAdapter) Update(ctx context.Context, index int, entry dto.ConfigDatabase) error {
	return a.update.Execute(ctx, index, entry)
}

func (a selectorUseCaseAdapter) Delete(ctx context.Context, index int) error {
	return a.del.Execute(ctx, index)
}

func (a selectorUseCaseAdapter) ActivePath(ctx context.Context) (string, error) {
	return a.active.Execute(ctx)
}

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

type databaseSelectorModel struct {
	ctx     context.Context
	manager selectorManager

	options  []DatabaseOption
	selected int
	width    int
	height   int
	chosen   bool
	canceled bool

	mode          selectorMode
	form          selectorForm
	confirmDelete selectorDeleteConfirm

	activeConfigPath string
	statusMessage    string

	launchAdditionalOptions []DatabaseOption
	configOptionCount       int
	requiresFirstEntry      bool
}

func SelectDatabase(
	ctx context.Context,
	listConfiguredDatabases *usecase.ListConfiguredDatabases,
	createConfiguredDatabase *usecase.CreateConfiguredDatabase,
	updateConfiguredDatabase *usecase.UpdateConfiguredDatabase,
	deleteConfiguredDatabase *usecase.DeleteConfiguredDatabase,
	getActiveConfigPath *usecase.GetActiveConfigPath,
) (DatabaseOption, error) {
	return SelectDatabaseWithState(
		ctx,
		listConfiguredDatabases,
		createConfiguredDatabase,
		updateConfiguredDatabase,
		deleteConfiguredDatabase,
		getActiveConfigPath,
		SelectorLaunchState{},
	)
}

func SelectDatabaseWithState(
	ctx context.Context,
	listConfiguredDatabases *usecase.ListConfiguredDatabases,
	createConfiguredDatabase *usecase.CreateConfiguredDatabase,
	updateConfiguredDatabase *usecase.UpdateConfiguredDatabase,
	deleteConfiguredDatabase *usecase.DeleteConfiguredDatabase,
	getActiveConfigPath *usecase.GetActiveConfigPath,
	state SelectorLaunchState,
) (DatabaseOption, error) {
	if listConfiguredDatabases == nil || createConfiguredDatabase == nil || updateConfiguredDatabase == nil || deleteConfiguredDatabase == nil || getActiveConfigPath == nil {
		return DatabaseOption{}, errors.New("selector config management use cases are required")
	}

	model, err := newDatabaseSelectorModel(ctx, selectorUseCaseAdapter{
		list:   listConfiguredDatabases,
		create: createConfiguredDatabase,
		update: updateConfiguredDatabase,
		del:    deleteConfiguredDatabase,
		active: getActiveConfigPath,
	}, state)
	if err != nil {
		return DatabaseOption{}, err
	}

	program := tea.NewProgram(model, tea.WithAltScreen())
	final, err := program.Run()
	if err != nil {
		return DatabaseOption{}, err
	}
	selector, ok := final.(*databaseSelectorModel)
	if !ok {
		return DatabaseOption{}, errors.New("unexpected selector state")
	}
	if selector.canceled {
		return DatabaseOption{}, ErrDatabaseSelectionCanceled
	}
	if !selector.chosen {
		return DatabaseOption{}, ErrDatabaseSelectionUnfinished
	}
	if selector.selected < 0 || selector.selected >= len(selector.options) {
		return DatabaseOption{}, ErrDatabaseSelectionUnfinished
	}
	return selector.options[selector.selected], nil
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
		if m.mode == selectorModeAdd || m.mode == selectorModeEdit {
			return m.handleFormKey(msg)
		}
		if m.mode == selectorModeConfirmDelete {
			return m.handleDeleteConfirmationKey(msg)
		}

		key := msg.String()
		switch key {
		case "ctrl+c", "q", "esc":
			m.canceled = true
			return m, tea.Quit
		case "enter":
			if len(m.options) == 0 {
				if m.requiresFirstEntry {
					m.openAddForm()
					m.statusMessage = "First database entry is required"
				}
				return m, nil
			}
			m.chosen = true
			return m, tea.Quit
		case "j", "down":
			m.moveSelection(1)
			return m, nil
		case "k", "up":
			m.moveSelection(-1)
			return m, nil
		case "g", "home":
			m.selectTop()
			return m, nil
		case "G", "end":
			m.selectBottom()
			return m, nil
		case "ctrl+f", "pgdown":
			m.page(1)
			return m, nil
		case "ctrl+b", "pgup":
			m.page(-1)
			return m, nil
		case "a":
			m.openAddForm()
			return m, nil
		case "e":
			if m.requiresFirstEntry {
				m.statusMessage = "Edit is unavailable during first setup"
				return m, nil
			}
			m.openEditForm()
			return m, nil
		case "d":
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

func (m *databaseSelectorModel) View() string {
	width := m.width
	height := m.height
	if width <= 0 {
		width = 80
	}
	if height <= 0 {
		height = 24
	}

	listHeight := m.listHeight(height)
	lines := m.boxLines(listHeight, width)
	boxHeight := len(lines)
	if boxHeight > height {
		boxHeight = height
		lines = lines[:boxHeight]
	}

	leftPad := 0
	boxWidth := len(lines[0])
	if width > boxWidth {
		leftPad = (width - boxWidth) / 2
	}
	topPad := 0
	if height > boxHeight {
		topPad = (height - boxHeight) / 2
	}

	full := make([]string, 0, height)
	for i := 0; i < topPad; i++ {
		full = append(full, strings.Repeat(" ", width))
	}
	for _, line := range lines {
		full = append(full, padRight(strings.Repeat(" ", leftPad)+line, width))
	}
	for len(full) < height {
		full = append(full, strings.Repeat(" ", width))
	}
	return strings.Join(full, "\n")
}

func (m *databaseSelectorModel) listHeight(totalHeight int) int {
	if totalHeight <= 0 {
		totalHeight = 24
	}
	maxListHeight := totalHeight - 8
	if maxListHeight < 1 {
		maxListHeight = 1
	}
	if len(m.options) < maxListHeight {
		return len(m.options)
	}
	return maxListHeight
}

func (m *databaseSelectorModel) boxLines(listHeight, totalWidth int) []string {
	title := "Select database"
	configPath := m.activeConfigPath
	if strings.TrimSpace(configPath) == "" {
		configPath = "unavailable"
	}
	pathLine := "Config: " + configPath

	items := m.optionLines()
	contentWidth := len(title)
	if len(pathLine) > contentWidth {
		contentWidth = len(pathLine)
	}
	for _, item := range items {
		itemWidth := len(item) + 2
		if itemWidth > contentWidth {
			contentWidth = itemWidth
		}
	}
	for _, line := range m.contextLines() {
		if len(line) > contentWidth {
			contentWidth = len(line)
		}
	}
	for _, line := range m.formLines() {
		if len(line) > contentWidth {
			contentWidth = len(line)
		}
	}
	for _, line := range m.deleteConfirmationLines() {
		if len(line) > contentWidth {
			contentWidth = len(line)
		}
	}
	if contentWidth < 1 {
		contentWidth = 1
	}
	maxInner := totalWidth - 2
	if maxInner < 1 {
		maxInner = 1
	}
	if contentWidth > maxInner {
		contentWidth = maxInner
	}

	border := "+" + strings.Repeat("-", contentWidth) + "+"
	lines := []string{
		border,
		"|" + padRight(title, contentWidth) + "|",
		"|" + padRight(pathLine, contentWidth) + "|",
		"|" + strings.Repeat("-", contentWidth) + "|",
	}

	if m.mode == selectorModeAdd || m.mode == selectorModeEdit {
		for _, line := range m.formLines() {
			lines = append(lines, "|"+padRight(line, contentWidth)+"|")
		}
		lines = append(lines, border)
		return lines
	}

	if m.mode == selectorModeConfirmDelete {
		for _, line := range m.deleteConfirmationLines() {
			lines = append(lines, "|"+padRight(line, contentWidth)+"|")
		}
		lines = append(lines, border)
		return lines
	}

	if len(items) == 0 {
		lines = append(lines, "|"+padRight("No databases configured.", contentWidth)+"|")
	} else {
		start := scrollStart(m.selected, listHeight, len(items))
		end := minInt(len(items), start+listHeight)
		for i := start; i < end; i++ {
			prefix := "  "
			if i == m.selected {
				prefix = "> "
			}
			lines = append(lines, "|"+padRight(prefix+items[i], contentWidth)+"|")
		}
	}

	for _, line := range m.contextLines() {
		lines = append(lines, "|"+padRight(line, contentWidth)+"|")
	}
	lines = append(lines, border)
	return lines
}

func (m *databaseSelectorModel) optionLines() []string {
	items := make([]string, len(m.options))
	for i, option := range m.options {
		items[i] = fmt.Sprintf("%s %s | %s", option.marker(), option.Name, option.ConnString)
	}
	return items
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
	switch msg.String() {
	case "esc":
		if m.requiresFirstEntry && len(m.options) == 0 {
			m.canceled = true
			return m, tea.Quit
		}
		m.mode = selectorModeBrowse
		m.form = selectorForm{}
		m.statusMessage = "Config update canceled"
		return m, nil
	case "tab", "shift+tab":
		if m.form.activeField == selectorInputName {
			m.form.activeField = selectorInputPath
		} else {
			m.form.activeField = selectorInputName
		}
		return m, nil
	case "ctrl+u":
		m.setActiveFormValue("")
		m.form.errorMessage = ""
		return m, nil
	case "backspace", "ctrl+h":
		value := m.activeFormValue()
		if value == "" {
			return m, nil
		}
		m.setActiveFormValue(value[:len(value)-1])
		m.form.errorMessage = ""
		return m, nil
	case "enter":
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
	switch msg.String() {
	case "esc":
		m.mode = selectorModeBrowse
		m.confirmDelete = selectorDeleteConfirm{}
		m.statusMessage = "Delete canceled"
		return m, nil
	case "enter":
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

func (m *databaseSelectorModel) contextLines() []string {
	if m.mode != selectorModeBrowse {
		return nil
	}
	lines := []string{""}
	if m.requiresFirstEntry {
		lines = append(lines, "First setup: Enter continue | a add database")
		lines = append(lines, "j/k navigate | q quit")
	} else {
		lines = append(lines, "j/k navigate | Enter select | a add | e edit | d delete")
		lines = append(lines, "Esc cancel | q quit")
		lines = append(lines, "Legend: ⚙ config | ⌨ CLI session")
	}
	if strings.TrimSpace(m.statusMessage) != "" {
		lines = append(lines, "Status: "+m.statusMessage)
	}
	return lines
}

func (m *databaseSelectorModel) formLines() []string {
	if m.mode != selectorModeAdd && m.mode != selectorModeEdit {
		return nil
	}
	title := "Add database"
	if m.mode == selectorModeEdit {
		title = "Edit database"
	}
	namePrefix := "  "
	pathPrefix := "  "
	nameValue := m.form.nameValue
	pathValue := m.form.pathValue
	if m.form.activeField == selectorInputName {
		namePrefix = "> "
		nameValue += "|"
	} else {
		pathPrefix = "> "
		pathValue += "|"
	}

	escLabel := "Esc cancel"
	if m.requiresFirstEntry && len(m.options) == 0 && m.mode == selectorModeAdd {
		escLabel = "Esc exit app"
	}

	lines := []string{
		title,
		"",
		namePrefix + "Name: " + nameValue,
		pathPrefix + "Path: " + pathValue,
		"",
		"Tab switch field | Ctrl+u clear field",
		"Enter save | " + escLabel,
	}
	if strings.TrimSpace(m.form.errorMessage) != "" {
		lines = append(lines, "Error: "+m.form.errorMessage)
	}
	return lines
}

func (m *databaseSelectorModel) deleteConfirmationLines() []string {
	if m.mode != selectorModeConfirmDelete {
		return nil
	}
	if m.confirmDelete.optionIndex < 0 || m.confirmDelete.optionIndex >= len(m.options) {
		return []string{
			"Cannot delete: invalid selection.",
			"Press Esc to return.",
		}
	}
	selected := m.options[m.confirmDelete.optionIndex]
	return []string{
		"Delete database entry?",
		"",
		selected.Name + " | " + selected.ConnString,
		"",
		"Enter confirm delete | Esc cancel",
	}
}

func (o DatabaseOption) source() DatabaseOptionSource {
	if o.Source == DatabaseOptionSourceCLI {
		return DatabaseOptionSourceCLI
	}
	return DatabaseOptionSourceConfig
}

func (o DatabaseOption) marker() string {
	if o.source() == DatabaseOptionSourceCLI {
		return "⌨"
	}
	return "⚙"
}

func (o DatabaseOption) isConfigBacked() bool {
	return o.source() == DatabaseOptionSourceConfig && o.managerIndex >= 0
}

func normalizeAdditionalOptions(options []DatabaseOption) []DatabaseOption {
	if len(options) == 0 {
		return nil
	}

	normalized := make([]DatabaseOption, 0, len(options))
	seen := make(map[string]struct{}, len(options))
	for _, option := range options {
		connString := strings.TrimSpace(option.ConnString)
		if connString == "" {
			continue
		}
		if _, exists := seen[connString]; exists {
			continue
		}
		seen[connString] = struct{}{}

		name := strings.TrimSpace(option.Name)
		if name == "" {
			name = connString
		}
		normalized = append(normalized, DatabaseOption{
			Name:         name,
			ConnString:   connString,
			Source:       DatabaseOptionSourceCLI,
			managerIndex: -1,
		})
	}
	return normalized
}

func mergeConfigAndAdditionalOptions(configOptions []DatabaseOption, additionalOptions []DatabaseOption) []DatabaseOption {
	if len(additionalOptions) == 0 {
		return configOptions
	}

	merged := make([]DatabaseOption, 0, len(configOptions)+len(additionalOptions))
	merged = append(merged, configOptions...)

	seen := make(map[string]struct{}, len(configOptions)+len(additionalOptions))
	for _, option := range configOptions {
		key := strings.TrimSpace(option.ConnString)
		if key == "" {
			continue
		}
		seen[key] = struct{}{}
	}

	for _, option := range additionalOptions {
		key := strings.TrimSpace(option.ConnString)
		if key != "" {
			if _, exists := seen[key]; exists {
				continue
			}
			seen[key] = struct{}{}
		}
		merged = append(merged, option)
	}

	return merged
}
