package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/dto"
	"github.com/mgierok/dbc/internal/application/usecase"
)

func (m *Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	if keyMatches(keyRuntimeOpenContextHelp, key) {
		if m.helpPopup.active {
			return m, nil
		}
		m.openHelpPopup(m.currentHelpPopupContext())
		return m, nil
	}

	if m.helpPopup.active {
		return m.handleHelpPopupKey(msg)
	}
	if m.editPopup.active {
		return m.handleEditPopupKey(msg)
	}
	if m.confirmPopup.active {
		return m.handleConfirmPopupKey(msg)
	}
	if m.filterPopup.active {
		return m.handleFilterPopupKey(msg)
	}
	if m.sortPopup.active {
		return m.handleSortPopupKey(msg)
	}
	if m.commandInput.active {
		return m.handleCommandInputKey(msg)
	}
	if m.recordDetail.active {
		return m.handleRecordDetailKey(msg)
	}

	if m.pendingG {
		if keyMatches(keyRuntimeJumpTopPending, key) {
			m.pendingG = false
			return m.jumpTop()
		}
		m.pendingG = false
	}

	switch {
	case keyMatches(keyRuntimeOpenCommandInput, key):
		return m.startCommandInput()
	case keyMatches(keyRuntimeJumpTopPending, key):
		m.pendingG = true
		return m, nil
	case keyMatches(keyRuntimeJumpBottom, key):
		return m.jumpBottom()
	case keyMatches(keyRuntimeEnter, key):
		if m.viewMode == ViewRecords && m.focus == FocusContent {
			return m.openRecordDetail()
		}
		return m.switchToRecords()
	case keyMatches(keyRuntimeEdit, key):
		if m.viewMode == ViewRecords && m.focus == FocusContent {
			if !m.recordFieldFocus {
				return m.enableRecordFieldFocus()
			}
			return m.openEditPopup()
		}
		return m, nil
	case keyMatches(keyRuntimeEsc, key):
		if m.viewMode == ViewRecords && m.recordFieldFocus {
			m.recordFieldFocus = false
			return m, nil
		}
		if m.focus == FocusContent {
			m.focus = FocusTables
			m.viewMode = ViewSchema
			return m, nil
		}
		return m, nil
	case keyMatches(keyRuntimeFilter, key):
		return m.startFilterPopup()
	case keyMatches(keyRuntimeSort, key):
		return m.startSortPopup()
	case keyMatches(keyRuntimeRecordDetail, key):
		return m.openRecordDetail()
	case keyMatches(keyRuntimeSave, key):
		return m.requestSaveChanges()
	case keyMatches(keyRuntimeInsert, key):
		return m.addPendingInsert()
	case keyMatches(keyRuntimeDelete, key):
		return m.toggleDeleteSelection()
	case keyMatches(keyRuntimeUndo, key):
		return m.undoStagedAction()
	case keyMatches(keyRuntimeRedo, key):
		return m.redoStagedAction()
	case keyMatches(keyRuntimeToggleAutoFields, key):
		return m.toggleInsertAutoFields()
	case keyMatches(keyRuntimeMoveDown, key):
		return m.moveDown()
	case keyMatches(keyRuntimeMoveUp, key):
		return m.moveUp()
	case keyMatches(keyRuntimeMoveLeft, key):
		return m.moveLeft()
	case keyMatches(keyRuntimeMoveRight, key):
		return m.moveRight()
	case keyMatches(keyRuntimePageDown, key):
		return m.pageDown()
	case keyMatches(keyRuntimePageUp, key):
		return m.pageUp()
	default:
		return m, nil
	}
}

func (m *Model) handleFilterPopupKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch {
	case keyMatches(keyRuntimeEsc, key):
		m.closeFilterPopup()
		return m, nil
	case keyMatches(keyRuntimeEnter, key):
		return m.confirmPopupSelection()
	case keyMatches(keyRuntimeMoveDown, key):
		m.movePopupSelection(1)
		return m, nil
	case keyMatches(keyRuntimeMoveUp, key):
		m.movePopupSelection(-1)
		return m, nil
	case keyMatches(keyInputMoveLeft, key):
		if m.filterPopup.step == filterInputValue {
			m.filterPopup.cursor = clamp(m.filterPopup.cursor-1, 0, len(m.filterPopup.input))
		}
		return m, nil
	case keyMatches(keyInputMoveRight, key):
		if m.filterPopup.step == filterInputValue {
			m.filterPopup.cursor = clamp(m.filterPopup.cursor+1, 0, len(m.filterPopup.input))
		}
		return m, nil
	case keyMatches(keyInputBackspace, key):
		if m.filterPopup.step == filterInputValue && m.filterPopup.input != "" {
			m.filterPopup.input, m.filterPopup.cursor = deleteAtCursor(m.filterPopup.input, m.filterPopup.cursor)
		}
		return m, nil
	}

	if m.filterPopup.step == filterInputValue && msg.Type == tea.KeyRunes {
		insert := string(msg.Runes)
		m.filterPopup.input, m.filterPopup.cursor = insertAtCursor(m.filterPopup.input, insert, m.filterPopup.cursor)
	}
	return m, nil
}

func (m *Model) handleSortPopupKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch {
	case keyMatches(keyRuntimeEsc, key):
		m.closeSortPopup()
		return m, nil
	case keyMatches(keyRuntimeEnter, key):
		return m.confirmSortPopupSelection()
	case keyMatches(keyRuntimeMoveDown, key):
		m.moveSortPopupSelection(1)
		return m, nil
	case keyMatches(keyRuntimeMoveUp, key):
		m.moveSortPopupSelection(-1)
		return m, nil
	default:
		return m, nil
	}
}

func (m *Model) handleHelpPopupKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	if m.commandInput.active {
		return m.handleCommandInputKey(msg)
	}

	switch {
	case keyMatches(keyRuntimeEsc, key):
		m.closeHelpPopup()
		return m, nil
	case keyMatches(keyRuntimeOpenCommandInput, key):
		return m.startCommandInput()
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

func (m *Model) handleCommandInputKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch {
	case keyMatches(keyRuntimeEsc, key):
		m.commandInput = commandInput{}
		return m, nil
	case keyMatches(keyRuntimeEnter, key):
		return m.submitCommandInput()
	case keyMatches(keyInputMoveLeft, key):
		m.commandInput.cursor = clamp(m.commandInput.cursor-1, 0, len(m.commandInput.value))
		return m, nil
	case keyMatches(keyInputMoveRight, key):
		m.commandInput.cursor = clamp(m.commandInput.cursor+1, 0, len(m.commandInput.value))
		return m, nil
	case keyMatches(keyInputBackspace, key):
		if m.commandInput.value != "" {
			m.commandInput.value, m.commandInput.cursor = deleteAtCursor(m.commandInput.value, m.commandInput.cursor)
		}
		return m, nil
	}

	if msg.Type == tea.KeyRunes {
		insert := string(msg.Runes)
		m.commandInput.value, m.commandInput.cursor = insertAtCursor(m.commandInput.value, insert, m.commandInput.cursor)
	}
	return m, nil
}

func (m *Model) handleEditPopupKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	column, ok := m.editColumn()
	if !ok {
		m.closeEditPopup()
		return m, nil
	}

	switch {
	case keyMatches(keyRuntimeEsc, key):
		m.closeEditPopup()
		return m, nil
	case keyMatches(keyEditSetNull, key):
		if !column.Nullable {
			m.editPopup.errorMessage = "Column is not nullable"
			return m, nil
		}
		m.editPopup.isNull = true
		m.editPopup.errorMessage = ""
		return m, nil
	case keyMatches(keyRuntimeEnter, key):
		return m.confirmEditPopup()
	case keyMatches(keyRuntimeMoveDown, key):
		if column.Input.Kind == dto.ColumnInputSelect {
			if len(column.Input.Options) > 0 {
				m.editPopup.optionIndex = clamp(m.editPopup.optionIndex+1, 0, len(column.Input.Options)-1)
				m.editPopup.isNull = false
				m.editPopup.errorMessage = ""
			}
		}
		return m, nil
	case keyMatches(keyRuntimeMoveUp, key):
		if column.Input.Kind == dto.ColumnInputSelect {
			if len(column.Input.Options) > 0 {
				m.editPopup.optionIndex = clamp(m.editPopup.optionIndex-1, 0, len(column.Input.Options)-1)
				m.editPopup.isNull = false
				m.editPopup.errorMessage = ""
			}
		}
		return m, nil
	case keyMatches(keyInputMoveLeft, key):
		if column.Input.Kind == dto.ColumnInputText {
			m.editPopup.cursor = clamp(m.editPopup.cursor-1, 0, len(m.editPopup.input))
			m.editPopup.errorMessage = ""
		}
		return m, nil
	case keyMatches(keyInputMoveRight, key):
		if column.Input.Kind == dto.ColumnInputText {
			m.editPopup.cursor = clamp(m.editPopup.cursor+1, 0, len(m.editPopup.input))
			m.editPopup.errorMessage = ""
		}
		return m, nil
	case keyMatches(keyInputBackspace, key):
		if column.Input.Kind == dto.ColumnInputText && m.editPopup.input != "" {
			m.editPopup.input, m.editPopup.cursor = deleteAtCursor(m.editPopup.input, m.editPopup.cursor)
			m.editPopup.isNull = false
			m.editPopup.errorMessage = ""
		}
		return m, nil
	}

	if column.Input.Kind == dto.ColumnInputText && msg.Type == tea.KeyRunes {
		insert := string(msg.Runes)
		m.editPopup.input, m.editPopup.cursor = insertAtCursor(m.editPopup.input, insert, m.editPopup.cursor)
		m.editPopup.isNull = false
		m.editPopup.errorMessage = ""
	}
	return m, nil
}

func (m *Model) handleConfirmPopupKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch {
	case keyMatches(keyPopupMoveDown, key):
		if len(m.confirmPopup.options) > 0 {
			m.confirmPopup.selected = clamp(m.confirmPopup.selected+1, 0, len(m.confirmPopup.options)-1)
		}
		return m, nil
	case keyMatches(keyPopupMoveUp, key):
		if len(m.confirmPopup.options) > 0 {
			m.confirmPopup.selected = clamp(m.confirmPopup.selected-1, 0, len(m.confirmPopup.options)-1)
		}
		return m, nil
	case keyMatches(keyConfirmCancel, key):
		m.closeConfirmPopup()
		m.pendingTableIndex = -1
		m.pendingConfigOpen = false
		return m, nil
	case keyMatches(keyConfirmAccept, key):
		action := m.confirmPopup.action
		if len(m.confirmPopup.options) > 0 {
			action = m.confirmPopup.options[clamp(m.confirmPopup.selected, 0, len(m.confirmPopup.options)-1)].action
		}
		m.closeConfirmPopup()
		switch action {
		case confirmSave:
			return m.confirmSaveChanges()
		case confirmDiscardTable:
			return m.confirmDiscardTableSwitch()
		case confirmCancelTableSwitch:
			m.pendingTableIndex = -1
			return m, nil
		case confirmConfigSaveAndOpen:
			return m.confirmConfigSaveAndOpen()
		case confirmConfigDiscardAndOpen:
			return m.confirmConfigDiscardAndOpen()
		case confirmConfigCancel:
			m.pendingConfigOpen = false
			return m, nil
		default:
			return m, nil
		}
	default:
		return m, nil
	}
}

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
	} else if isInsert && (insertIndex < 0 || insertIndex >= len(m.pendingInserts)) {
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
		if value, ok := m.pendingInserts[insertIndex].values[m.recordColumn]; ok {
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

func (m *Model) moveDown() (tea.Model, tea.Cmd) {
	switch m.focus {
	case FocusTables:
		return m.moveTableSelection(1)
	case FocusContent:
		return m.moveContentSelection(1)
	default:
		return m, nil
	}
}

func (m *Model) moveUp() (tea.Model, tea.Cmd) {
	switch m.focus {
	case FocusTables:
		return m.moveTableSelection(-1)
	case FocusContent:
		return m.moveContentSelection(-1)
	default:
		return m, nil
	}
}

func (m *Model) moveLeft() (tea.Model, tea.Cmd) {
	if m.focus != FocusContent || m.viewMode != ViewRecords || !m.recordFieldFocus {
		return m, nil
	}
	visibleColumns := m.visibleColumnIndicesForSelection()
	if len(visibleColumns) == 0 {
		return m, nil
	}
	current := indexOfInt(visibleColumns, m.recordColumn)
	if current == -1 {
		m.recordColumn = visibleColumns[0]
		return m, nil
	}
	if current > 0 {
		m.recordColumn = visibleColumns[current-1]
	}
	return m, nil
}

func (m *Model) moveRight() (tea.Model, tea.Cmd) {
	if m.focus != FocusContent || m.viewMode != ViewRecords || !m.recordFieldFocus {
		return m, nil
	}
	visibleColumns := m.visibleColumnIndicesForSelection()
	if len(visibleColumns) == 0 {
		return m, nil
	}
	current := indexOfInt(visibleColumns, m.recordColumn)
	if current == -1 {
		m.recordColumn = visibleColumns[0]
		return m, nil
	}
	if current < len(visibleColumns)-1 {
		m.recordColumn = visibleColumns[current+1]
	}
	return m, nil
}

func (m *Model) pageDown() (tea.Model, tea.Cmd) {
	if m.focus == FocusContent && m.viewMode == ViewRecords {
		return m.nextRecordPage()
	}
	page := m.pageSize()
	switch m.focus {
	case FocusTables:
		return m.moveTableSelection(page)
	case FocusContent:
		return m.moveContentSelection(page)
	default:
		return m, nil
	}
}

func (m *Model) pageUp() (tea.Model, tea.Cmd) {
	if m.focus == FocusContent && m.viewMode == ViewRecords {
		return m.prevRecordPage()
	}
	page := m.pageSize()
	switch m.focus {
	case FocusTables:
		return m.moveTableSelection(-page)
	case FocusContent:
		return m.moveContentSelection(-page)
	default:
		return m, nil
	}
}

func (m *Model) nextRecordPage() (tea.Model, tea.Cmd) {
	if m.recordTotalPages <= 1 {
		return m, nil
	}
	if m.recordPageIndex >= m.recordTotalPages-1 {
		return m, nil
	}
	m.recordPageIndex++
	return m, m.loadRecordsCmd(false)
}

func (m *Model) prevRecordPage() (tea.Model, tea.Cmd) {
	if m.recordTotalPages <= 1 {
		return m, nil
	}
	if m.recordPageIndex <= 0 {
		return m, nil
	}
	m.recordPageIndex--
	return m, m.loadRecordsCmd(false)
}

func (m *Model) jumpTop() (tea.Model, tea.Cmd) {
	switch m.focus {
	case FocusTables:
		return m.setTableSelection(0)
	case FocusContent:
		return m.setContentSelection(0)
	default:
		return m, nil
	}
}

func (m *Model) jumpBottom() (tea.Model, tea.Cmd) {
	switch m.focus {
	case FocusTables:
		return m.setTableSelection(len(m.tables) - 1)
	case FocusContent:
		return m.setContentSelection(m.contentMaxIndex())
	default:
		return m, nil
	}
}

func (m *Model) moveTableSelection(delta int) (tea.Model, tea.Cmd) {
	target := m.selectedTable + delta
	return m.setTableSelection(target)
}

func (m *Model) setTableSelection(index int) (tea.Model, tea.Cmd) {
	if len(m.tables) == 0 {
		return m, nil
	}
	index = clamp(index, 0, len(m.tables)-1)
	if index == m.selectedTable {
		return m, nil
	}
	if m.hasDirtyEdits() {
		prompt := m.dirtyNavigationPolicyUseCase().BuildTableSwitchPrompt(m.dirtyEditCount())
		m.pendingTableIndex = index
		m.openModalConfirmPopupWithOptions(
			prompt.Title,
			prompt.Message,
			m.confirmOptionsFromDirtyPrompt(prompt, false),
			0,
		)
		return m, nil
	}
	m.selectedTable = index
	m.resetTableContext()
	return m, m.loadViewForSelection()
}

func (m *Model) moveContentSelection(delta int) (tea.Model, tea.Cmd) {
	target := m.contentSelection() + delta
	return m.setContentSelection(target)
}

func (m *Model) setContentSelection(index int) (tea.Model, tea.Cmd) {
	maxIndex := m.contentMaxIndex()
	if maxIndex < 0 {
		return m, nil
	}
	index = clamp(index, 0, maxIndex)
	switch m.viewMode {
	case ViewSchema:
		m.schemaIndex = index
		return m, nil
	case ViewRecords:
		m.recordSelection = index
		m.syncRecordColumnForSelection()
		return m, nil
	default:
		return m, nil
	}
}

func (m *Model) switchToRecords() (tea.Model, tea.Cmd) {
	m.viewMode = ViewRecords
	m.focus = FocusContent
	m.recordFieldFocus = false
	m.closeRecordDetail()
	if m.currentTableName() == "" {
		return m, nil
	}

	var cmds []tea.Cmd
	if len(m.schema.Columns) == 0 {
		cmds = append(cmds, m.loadSchemaCmd())
	}
	if len(m.records) == 0 {
		cmds = append(cmds, m.loadRecordsCmd(true))
	}
	if len(cmds) == 0 {
		return m, nil
	}
	return m, tea.Batch(cmds...)
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

func (m *Model) handleRecordDetailKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch {
	case keyMatches(keyRuntimeEsc, key):
		m.closeRecordDetail()
		return m, nil
	case keyMatches(keyPopupMoveDown, key):
		m.moveRecordDetailScroll(1)
		return m, nil
	case keyMatches(keyPopupMoveUp, key):
		m.moveRecordDetailScroll(-1)
		return m, nil
	case keyMatches(keyRuntimePageDown, key):
		m.moveRecordDetailScroll(m.recordDetailVisibleLines())
		return m, nil
	case keyMatches(keyRuntimePageUp, key):
		m.moveRecordDetailScroll(-m.recordDetailVisibleLines())
		return m, nil
	case keyMatches(keyPopupJumpTop, key):
		m.recordDetail.scrollOffset = 0
		return m, nil
	case keyMatches(keyPopupJumpBottom, key):
		m.recordDetail.scrollOffset = m.recordDetailMaxOffset()
		return m, nil
	default:
		return m, nil
	}
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

func (m *Model) confirmPopupSelection() (tea.Model, tea.Cmd) {
	switch m.filterPopup.step {
	case filterSelectColumn:
		if len(m.schema.Columns) == 0 {
			return m, nil
		}
		column := m.schema.Columns[m.filterPopup.columnIndex]
		operators, err := m.listOperators.Execute(m.ctx, column.Type)
		if err != nil {
			m.statusMessage = "Error: " + err.Error()
			return m, nil
		}
		m.filterPopup.operators = operators
		m.filterPopup.operatorIndex = 0
		m.filterPopup.step = filterSelectOperator
		return m, nil
	case filterSelectOperator:
		if len(m.filterPopup.operators) == 0 {
			return m, nil
		}
		operator := m.filterPopup.operators[m.filterPopup.operatorIndex]
		if operator.RequiresValue {
			m.filterPopup.input = ""
			m.filterPopup.cursor = 0
			m.filterPopup.step = filterInputValue
			return m, nil
		}
		return m.applyFilter(operator, "")
	case filterInputValue:
		operator := m.filterPopup.operators[m.filterPopup.operatorIndex]
		return m.applyFilter(operator, m.filterPopup.input)
	default:
		return m, nil
	}
}

func (m *Model) confirmSortPopupSelection() (tea.Model, tea.Cmd) {
	switch m.sortPopup.step {
	case sortSelectColumn:
		if len(m.schema.Columns) == 0 {
			return m, nil
		}
		m.sortPopup.step = sortSelectDirection
		return m, nil
	case sortSelectDirection:
		if len(m.schema.Columns) == 0 {
			return m, nil
		}
		directions := sortDirections()
		if len(directions) == 0 {
			return m, nil
		}
		column := m.schema.Columns[m.sortPopup.columnIndex]
		direction := directions[clamp(m.sortPopup.directionIndex, 0, len(directions)-1)]
		return m.applySort(column.Name, direction)
	default:
		return m, nil
	}
}

func (m *Model) movePopupSelection(delta int) {
	switch m.filterPopup.step {
	case filterSelectColumn:
		if len(m.schema.Columns) == 0 {
			return
		}
		m.filterPopup.columnIndex = clamp(m.filterPopup.columnIndex+delta, 0, len(m.schema.Columns)-1)
	case filterSelectOperator:
		if len(m.filterPopup.operators) == 0 {
			return
		}
		m.filterPopup.operatorIndex = clamp(m.filterPopup.operatorIndex+delta, 0, len(m.filterPopup.operators)-1)
	}
}

func (m *Model) moveSortPopupSelection(delta int) {
	switch m.sortPopup.step {
	case sortSelectColumn:
		if len(m.schema.Columns) == 0 {
			return
		}
		m.sortPopup.columnIndex = clamp(m.sortPopup.columnIndex+delta, 0, len(m.schema.Columns)-1)
	case sortSelectDirection:
		directions := sortDirections()
		if len(directions) == 0 {
			return
		}
		m.sortPopup.directionIndex = clamp(m.sortPopup.directionIndex+delta, 0, len(directions)-1)
	}
}

func (m *Model) applyFilter(operator dto.Operator, value string) (tea.Model, tea.Cmd) {
	if len(m.schema.Columns) == 0 {
		return m, nil
	}
	column := m.schema.Columns[m.filterPopup.columnIndex]
	filter := &dto.Filter{
		Column: column.Name,
		Operator: dto.Operator{
			Name:          operator.Name,
			Kind:          operator.Kind,
			RequiresValue: operator.RequiresValue,
		},
		Value: value,
	}
	m.currentFilter = filter
	m.closeFilterPopup()

	if m.viewMode == ViewRecords {
		return m, m.loadRecordsCmd(true)
	}
	return m, nil
}

func (m *Model) applySort(column string, direction dto.SortDirection) (tea.Model, tea.Cmd) {
	m.currentSort = &dto.Sort{
		Column:    column,
		Direction: direction,
	}
	m.closeSortPopup()
	return m, m.loadRecordsCmd(true)
}

func (m *Model) dirtyNavigationPolicyUseCase() *usecase.DirtyNavigationPolicy {
	if m.dirtyNavPolicy != nil {
		return m.dirtyNavPolicy
	}
	return usecase.NewDirtyNavigationPolicy()
}

func (m *Model) confirmOptionsFromDirtyPrompt(prompt usecase.DirtyDecisionPrompt, configFlow bool) []confirmOption {
	options := make([]confirmOption, 0, len(prompt.Options))
	for _, option := range prompt.Options {
		options = append(options, confirmOption{
			label:  option.Label,
			action: mapDirtyDecisionToConfirmAction(option.ID, configFlow),
		})
	}
	return options
}

func mapDirtyDecisionToConfirmAction(decisionID string, configFlow bool) confirmAction {
	if !configFlow {
		switch decisionID {
		case usecase.DirtyDecisionDiscard:
			return confirmDiscardTable
		case usecase.DirtyDecisionCancel:
			return confirmCancelTableSwitch
		default:
			return confirmCancelTableSwitch
		}
	}

	switch decisionID {
	case usecase.DirtyDecisionSave:
		return confirmConfigSaveAndOpen
	case usecase.DirtyDecisionDiscard:
		return confirmConfigDiscardAndOpen
	case usecase.DirtyDecisionCancel:
		return confirmConfigCancel
	default:
		return confirmConfigCancel
	}
}
