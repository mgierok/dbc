package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/dto"
)

func (m *Model) enableRecordFieldFocus() (tea.Model, tea.Cmd) {
	if m.viewMode != ViewRecords || m.totalRecordRows() == 0 {
		return m, nil
	}
	visibleColumns := m.visibleColumnIndicesForSelection()
	if len(visibleColumns) == 0 {
		return m, nil
	}
	if !containsInt(visibleColumns, m.recordColumn) {
		m.recordColumn = visibleColumns[0]
	}
	m.recordFieldFocus = true
	return m, nil
}

func (m *Model) openEditPopup() (tea.Model, tea.Cmd) {
	if m.recordSelection < 0 || m.recordSelection >= m.totalRecordRows() {
		return m, nil
	}
	if insertIndex, isInsert := m.pendingInsertIndexForSelection(); !isInsert && !m.canEditRecords() {
		m.statusMessage = "Error: table has no primary key"
		return m, nil
	} else if isInsert && (insertIndex < 0 || insertIndex >= len(m.staging.pendingInserts)) {
		return m, nil
	}
	visibleColumns := m.visibleColumnIndicesForSelection()
	if len(visibleColumns) == 0 {
		return m, nil
	}
	if !containsInt(visibleColumns, m.recordColumn) {
		m.recordColumn = visibleColumns[0]
	}
	if m.recordColumn < 0 || m.recordColumn >= len(m.schema.Columns) {
		return m, nil
	}
	column := m.schema.Columns[m.recordColumn]
	currentValue := m.visibleRowValue(m.recordSelection, m.recordColumn)
	if insertIndex, isInsert := m.pendingInsertIndexForSelection(); isInsert {
		if value, ok := m.staging.pendingInserts[insertIndex].values[m.recordColumn]; ok {
			currentValue = displayValue(value.Value)
		}
	} else if staged, ok := m.stagedEditForRow(m.recordSelection, m.recordColumn); ok {
		currentValue = displayValue(staged.Value)
	}

	popup := editPopup{
		active:      true,
		rowIndex:    m.recordSelection,
		columnIndex: m.recordColumn,
		input:       currentValue,
		cursor:      len(currentValue),
	}
	if strings.EqualFold(currentValue, "NULL") {
		popup.isNull = true
		popup.input = ""
		popup.cursor = 0
	}
	if column.Input.Kind == dto.ColumnInputSelect {
		popup.optionIndex = optionIndex(column.Input.Options, currentValue)
	}
	m.editPopup = popup
	return m, nil
}

func (m *Model) confirmEditPopup() (tea.Model, tea.Cmd) {
	column, ok := m.editColumn()
	if !ok {
		m.closeEditPopup()
		return m, nil
	}
	input := m.editPopup.input
	if column.Input.Kind == dto.ColumnInputSelect && len(column.Input.Options) > 0 {
		input = column.Input.Options[clamp(m.editPopup.optionIndex, 0, len(column.Input.Options)-1)]
	}

	value, err := m.translatorUseCase().ParseStagedValue(column, input, m.editPopup.isNull)
	if err != nil {
		m.editPopup.errorMessage = err.Error()
		return m, nil
	}
	if err := m.stageEdit(m.editPopup.rowIndex, m.editPopup.columnIndex, value); err != nil {
		m.editPopup.errorMessage = err.Error()
		return m, nil
	}
	m.closeEditPopup()
	return m, nil
}

func (m *Model) closeEditPopup() {
	m.editPopup = editPopup{}
}

func (m *Model) openConfirmPopup(action confirmAction, message string) {
	m.confirmPopup = confirmPopup{
		active:  true,
		title:   "Confirm",
		action:  action,
		message: message,
	}
}

func (m *Model) openModalConfirmPopupWithOptions(title, message string, options []confirmOption, selected int) {
	if len(options) == 0 {
		m.confirmPopup = confirmPopup{
			active:  true,
			title:   title,
			action:  confirmConfigCancel,
			message: message,
			modal:   true,
		}
		return
	}
	m.confirmPopup = confirmPopup{
		active:   true,
		title:    title,
		message:  message,
		options:  options,
		selected: clamp(selected, 0, len(options)-1),
		modal:    true,
	}
}

func (m *Model) closeConfirmPopup() {
	m.confirmPopup = confirmPopup{}
}

func (m *Model) startFilterPopup() (tea.Model, tea.Cmd) {
	if m.viewMode != ViewRecords || m.focus != FocusContent {
		return m, nil
	}
	if m.currentTableName() == "" {
		return m, nil
	}
	if len(m.schema.Columns) == 0 {
		m.pendingFilterOpen = true
		return m, m.loadSchemaCmd()
	}
	m.openFilterPopup()
	return m, nil
}

func (m *Model) startSortPopup() (tea.Model, tea.Cmd) {
	if m.viewMode != ViewRecords || m.focus != FocusContent {
		return m, nil
	}
	if m.currentTableName() == "" {
		return m, nil
	}
	if len(m.schema.Columns) == 0 {
		m.pendingSortOpen = true
		return m, m.loadSchemaCmd()
	}
	m.openSortPopup()
	return m, nil
}

func (m *Model) openFilterPopup() {
	m.filterPopup = filterPopup{
		active:        true,
		step:          filterSelectColumn,
		columnIndex:   0,
		operatorIndex: 0,
		input:         "",
		operators:     nil,
		cursor:        0,
	}
}

func (m *Model) closeFilterPopup() {
	m.filterPopup = filterPopup{}
}

func (m *Model) openSortPopup() {
	directionIndex := 0
	columnIndex := 0
	if m.currentSort != nil {
		for i, column := range m.schema.Columns {
			if column.Name == m.currentSort.Column {
				columnIndex = i
				break
			}
		}
		if m.currentSort.Direction == dto.SortDirectionDesc {
			directionIndex = 1
		}
	}
	m.sortPopup = sortPopup{
		active:         true,
		step:           sortSelectColumn,
		columnIndex:    columnIndex,
		directionIndex: directionIndex,
	}
}

func (m *Model) closeSortPopup() {
	m.sortPopup = sortPopup{}
}

func (m *Model) startCommandInput() (tea.Model, tea.Cmd) {
	m.commandInput = commandInput{
		active: true,
		value:  "",
		cursor: 0,
	}
	return m, nil
}

func (m *Model) submitCommandInput() (tea.Model, tea.Cmd) {
	command := ":" + strings.TrimSpace(m.commandInput.value)
	m.commandInput = commandInput{}

	commandSpec, found := resolveRuntimeCommand(command)
	if !found {
		m.statusMessage = fmt.Sprintf("Unknown command: %s", command)
		return m, nil
	}

	switch commandSpec.action {
	case runtimeCommandActionOpenHelp:
		m.openHelpPopup(m.currentHelpPopupContext())
		return m, nil
	case runtimeCommandActionQuit:
		return m, tea.Quit
	case runtimeCommandActionOpenConfig:
		if m.hasDirtyEdits() {
			prompt := m.dirtyNavigationPolicyUseCase().BuildConfigPrompt()
			m.openModalConfirmPopupWithOptions(
				prompt.Title,
				prompt.Message,
				m.confirmOptionsFromDirtyPrompt(prompt, true),
				0,
			)
			return m, nil
		}
		m.openConfigSelector = true
		m.statusMessage = "Opening config manager"
		return m, tea.Quit
	}

	return m, nil
}

func (m *Model) openHelpPopup(context helpPopupContext) {
	m.commandInput = commandInput{}
	m.helpPopup = helpPopup{
		active:       true,
		scrollOffset: 0,
		context:      context,
	}
}

func (m *Model) closeHelpPopup() {
	m.helpPopup = helpPopup{}
}

func (m *Model) openRecordDetail() (tea.Model, tea.Cmd) {
	if m.viewMode != ViewRecords || m.focus != FocusContent {
		return m, nil
	}
	if m.totalRecordRows() == 0 {
		return m, nil
	}
	m.recordDetail = recordDetailState{
		active:       true,
		scrollOffset: 0,
	}
	return m, nil
}

func (m *Model) closeRecordDetail() {
	m.recordDetail = recordDetailState{}
}

func (m *Model) moveRecordDetailScroll(delta int) {
	maxOffset := m.recordDetailMaxOffset()
	m.recordDetail.scrollOffset = clamp(m.recordDetail.scrollOffset+delta, 0, maxOffset)
}

func (m *Model) recordDetailVisibleLines() int {
	visible := m.contentHeight() - 1
	if visible < 1 {
		return 1
	}
	return visible
}

func (m *Model) recordDetailMaxOffset() int {
	_, rightWidth := m.panelWidths()
	maxOffset := len(m.recordDetailContentLines(rightWidth)) - m.recordDetailVisibleLines()
	if maxOffset < 0 {
		return 0
	}
	return maxOffset
}

func (m *Model) moveHelpPopupScroll(delta int) {
	maxOffset := m.helpPopupMaxOffset()
	m.helpPopup.scrollOffset = clamp(m.helpPopup.scrollOffset+delta, 0, maxOffset)
}

func (m *Model) helpPopupVisibleLines() int {
	const minVisibleLines = 6
	const maxVisibleLines = 12

	visible := m.contentHeight() - 10
	if visible < minVisibleLines {
		return minVisibleLines
	}
	if visible > maxVisibleLines {
		return maxVisibleLines
	}
	return visible
}

func (m *Model) helpPopupMaxOffset() int {
	maxOffset := len(m.helpPopupContentLines()) - m.helpPopupVisibleLines()
	if maxOffset < 0 {
		return 0
	}
	return maxOffset
}

func (m *Model) currentHelpPopupContext() helpPopupContext {
	switch {
	case m.editPopup.active:
		return helpPopupContextEditPopup
	case m.confirmPopup.active:
		return helpPopupContextConfirmPopup
	case m.filterPopup.active:
		return helpPopupContextFilterPopup
	case m.sortPopup.active:
		return helpPopupContextSortPopup
	case m.helpPopup.active:
		return helpPopupContextHelpPopup
	case m.commandInput.active:
		return helpPopupContextCommandInput
	case m.recordDetail.active:
		return helpPopupContextRecordDetail
	case m.focus == FocusTables:
		return helpPopupContextTables
	case m.focus == FocusContent && m.viewMode == ViewSchema:
		return helpPopupContextSchema
	case m.focus == FocusContent && m.viewMode == ViewRecords:
		return helpPopupContextRecords
	default:
		return helpPopupContextUnknown
	}
}

func (m *Model) helpPopupContextTitle() string {
	switch m.helpPopup.context {
	case helpPopupContextTables:
		return "Context Help: Tables"
	case helpPopupContextSchema:
		return "Context Help: Schema"
	case helpPopupContextRecords:
		return "Context Help: Records"
	case helpPopupContextRecordDetail:
		return "Context Help: Record Detail"
	case helpPopupContextFilterPopup:
		return "Context Help: Filter Popup"
	case helpPopupContextSortPopup:
		return "Context Help: Sort Popup"
	case helpPopupContextEditPopup:
		return "Context Help: Edit Popup"
	case helpPopupContextConfirmPopup:
		return "Context Help: Confirm Popup"
	case helpPopupContextCommandInput:
		return "Context Help: Command Input"
	case helpPopupContextHelpPopup:
		return "Context Help: Help Popup"
	default:
		return "Context Help"
	}
}

func (m *Model) helpPopupContentLines() []string {
	shortcuts := m.helpPopupContextShortcuts()
	if strings.TrimSpace(shortcuts) == "" {
		return []string{"No keybindings available in this context."}
	}
	parts := strings.Split(shortcuts, frameSegmentSeparator)
	lines := make([]string, 0, len(parts))
	for _, part := range parts {
		line := strings.TrimSpace(part)
		if line == "" {
			continue
		}
		lines = append(lines, line)
	}
	if len(lines) == 0 {
		return []string{"No keybindings available in this context."}
	}
	return lines
}

func (m *Model) helpPopupContextShortcuts() string {
	switch m.helpPopup.context {
	case helpPopupContextEditPopup:
		return runtimeStatusEditShortcuts()
	case helpPopupContextConfirmPopup:
		return runtimeStatusConfirmShortcuts(len(m.confirmPopup.options) > 0)
	case helpPopupContextFilterPopup:
		return runtimeStatusFilterPopupShortcuts()
	case helpPopupContextSortPopup:
		return runtimeStatusSortPopupShortcuts()
	case helpPopupContextHelpPopup:
		return runtimeStatusHelpPopupShortcuts()
	case helpPopupContextCommandInput:
		return runtimeStatusCommandInputShortcuts()
	case helpPopupContextRecordDetail:
		return runtimeStatusRecordDetailShortcuts()
	case helpPopupContextTables:
		return runtimeStatusTablesShortcuts()
	case helpPopupContextSchema:
		return runtimeStatusSchemaShortcuts()
	case helpPopupContextRecords:
		return runtimeStatusRecordsShortcuts()
	default:
		return ""
	}
}
