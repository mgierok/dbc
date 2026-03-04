package tui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/dto"
	"github.com/mgierok/dbc/internal/application/usecase"
)

const (
	recordPageSize = 20
)

type PanelFocus int

const (
	FocusTables PanelFocus = iota
	FocusContent
)

type ViewMode int

const (
	ViewSchema ViewMode = iota
	ViewRecords
)

type filterStep int

const (
	filterSelectColumn filterStep = iota
	filterSelectOperator
	filterInputValue
)

type sortStep int

const (
	sortSelectColumn sortStep = iota
	sortSelectDirection
)

type filterPopup struct {
	active        bool
	step          filterStep
	columnIndex   int
	operatorIndex int
	input         string
	operators     []dto.Operator
	cursor        int
}

type sortPopup struct {
	active         bool
	step           sortStep
	columnIndex    int
	directionIndex int
}

type commandInput struct {
	active bool
	value  string
	cursor int
}

type helpPopup struct {
	active       bool
	scrollOffset int
	context      helpPopupContext
}

type helpPopupContext int

const (
	helpPopupContextUnknown helpPopupContext = iota
	helpPopupContextTables
	helpPopupContextSchema
	helpPopupContextRecords
	helpPopupContextRecordDetail
	helpPopupContextFilterPopup
	helpPopupContextSortPopup
	helpPopupContextEditPopup
	helpPopupContextConfirmPopup
	helpPopupContextCommandInput
	helpPopupContextHelpPopup
)

type recordDetailState struct {
	active       bool
	scrollOffset int
}

type stagedEdit struct {
	Value dto.StagedValue
}

type recordEdits struct {
	identity dto.RecordIdentity
	changes  map[int]stagedEdit
}

type pendingInsertRow struct {
	values       map[int]stagedEdit
	explicitAuto map[int]bool
	showAuto     bool
}

type recordDelete struct {
	identity dto.RecordIdentity
}

type stagedOperationKind int

const (
	opInsertAdded stagedOperationKind = iota
	opInsertRemoved
	opCellEdited
	opDeleteToggled
)

type cellEditTarget int

const (
	cellEditPersisted cellEditTarget = iota
	cellEditInsert
)

type insertOperation struct {
	index int
	row   pendingInsertRow
}

type cellEditOperation struct {
	target             cellEditTarget
	insertIndex        int
	recordKey          string
	identity           dto.RecordIdentity
	columnIndex        int
	before             stagedEdit
	beforeExists       bool
	after              stagedEdit
	afterExists        bool
	beforeExplicitAuto bool
	afterExplicitAuto  bool
}

type deleteToggleOperation struct {
	key          string
	identity     dto.RecordIdentity
	beforeMarked bool
	afterMarked  bool
}

type stagedOperation struct {
	kind   stagedOperationKind
	insert insertOperation
	cell   cellEditOperation
	del    deleteToggleOperation
}

type editPopup struct {
	active       bool
	rowIndex     int
	columnIndex  int
	input        string
	cursor       int
	optionIndex  int
	isNull       bool
	errorMessage string
}

type confirmAction int

const (
	confirmSave confirmAction = iota + 1
	confirmDiscardTable
	confirmCancelTableSwitch
	confirmConfigSaveAndOpen
	confirmConfigDiscardAndOpen
	confirmConfigCancel
)

type confirmOption struct {
	label  string
	action confirmAction
}

type confirmPopup struct {
	active   bool
	title    string
	action   confirmAction
	message  string
	options  []confirmOption
	selected int
	modal    bool
}

type pkColumn struct {
	index  int
	column dto.SchemaColumn
}

type Model struct {
	ctx           context.Context
	listTables    listTablesUseCase
	getSchema     getSchemaUseCase
	listRecords   listRecordsUseCase
	listOperators listOperatorsUseCase
	saveChanges   saveChangesUseCase
	translator    *usecase.StagedChangesTranslator

	width  int
	height int

	focus    PanelFocus
	viewMode ViewMode

	tables        []dto.Table
	selectedTable int

	schema      dto.Schema
	schemaIndex int

	records          []dto.RecordRow
	recordPageIndex  int
	recordTotalPages int
	recordTotalCount int
	recordSelection  int
	recordColumn     int
	recordRequestID  int
	recordLoading    bool
	recordFieldFocus bool
	pendingInserts   []pendingInsertRow
	pendingUpdates   map[string]recordEdits
	pendingDeletes   map[string]recordDelete
	history          []stagedOperation
	future           []stagedOperation

	currentFilter     *dto.Filter
	currentSort       *dto.Sort
	filterPopup       filterPopup
	sortPopup         sortPopup
	commandInput      commandInput
	helpPopup         helpPopup
	recordDetail      recordDetailState
	editPopup         editPopup
	confirmPopup      confirmPopup
	pendingFilterOpen bool
	pendingSortOpen   bool
	pendingG          bool
	pendingTableIndex int
	pendingConfigOpen bool

	openConfigSelector bool
	statusMessage      string
}

var _ tea.Model = (*Model)(nil)

type tablesMsg struct {
	tables []dto.Table
}

type schemaMsg struct {
	tableName string
	schema    dto.Schema
}

type recordsMsg struct {
	tableName string
	requestID int
	page      dto.RecordPage
}

type saveChangesMsg struct {
	count int
	err   error
}

type errMsg struct {
	err error
}

type listTablesUseCase interface {
	Execute(ctx context.Context) ([]dto.Table, error)
}

type getSchemaUseCase interface {
	Execute(ctx context.Context, tableName string) (dto.Schema, error)
}

type listRecordsUseCase interface {
	Execute(ctx context.Context, tableName string, offset, limit int, filter *dto.Filter, sort *dto.Sort) (dto.RecordPage, error)
}

type listOperatorsUseCase interface {
	Execute(ctx context.Context, columnType string) ([]dto.Operator, error)
}

type saveChangesUseCase interface {
	ExecuteDTO(ctx context.Context, tableName string, changes dto.TableChanges) error
}

func NewModel(ctx context.Context, listTables listTablesUseCase, getSchema getSchemaUseCase, listRecords listRecordsUseCase, listOperators listOperatorsUseCase, saveChanges saveChangesUseCase, translator *usecase.StagedChangesTranslator) *Model {
	if ctx == nil {
		ctx = context.Background()
	}
	return &Model{
		ctx:               ctx,
		listTables:        listTables,
		getSchema:         getSchema,
		listRecords:       listRecords,
		listOperators:     listOperators,
		saveChanges:       saveChanges,
		translator:        translator,
		focus:             FocusTables,
		viewMode:          ViewSchema,
		recordTotalPages:  1,
		pendingTableIndex: -1,
	}
}

func (m *Model) Init() tea.Cmd {
	return loadTablesCmd(m.ctx, m.listTables)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tablesMsg:
		m.tables = msg.tables
		if len(m.tables) == 0 {
			m.statusMessage = "No tables found"
			return m, nil
		}
		m.selectedTable = 0
		return m, m.loadSchemaCmd()
	case schemaMsg:
		if msg.tableName != m.currentTableName() {
			return m, nil
		}
		m.schema = msg.schema
		m.schemaIndex = 0
		if m.recordColumn >= len(m.schema.Columns) {
			m.recordColumn = 0
		}
		if m.pendingFilterOpen {
			m.pendingFilterOpen = false
			m.openFilterPopup()
		}
		if m.pendingSortOpen {
			m.pendingSortOpen = false
			m.openSortPopup()
		}
		return m, nil
	case recordsMsg:
		if msg.tableName != m.currentTableName() {
			return m, nil
		}
		if msg.requestID != m.recordRequestID {
			return m, nil
		}
		m.recordLoading = false
		m.records = msg.page.Rows
		m.recordTotalCount = msg.page.TotalCount
		m.recordTotalPages = m.computeTotalPages(msg.page.TotalCount)
		m.recordPageIndex = clamp(m.recordPageIndex, 0, m.recordTotalPages-1)
		m.normalizeRecordSelection()
		return m, nil
	case saveChangesMsg:
		if msg.err != nil {
			m.pendingConfigOpen = false
			m.statusMessage = "Error: " + msg.err.Error()
			return m, nil
		}
		m.clearStagedState()
		if m.pendingConfigOpen {
			m.pendingConfigOpen = false
			m.openConfigSelector = true
			m.statusMessage = "Opening config manager"
			return m, tea.Quit
		}
		m.statusMessage = fmt.Sprintf("Saved %d changes", msg.count)
		return m, m.loadRecordsCmd(true)
	case errMsg:
		m.recordLoading = false
		m.statusMessage = "Error: " + msg.err.Error()
		return m, nil
	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

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
		m.pendingTableIndex = index
		m.openModalConfirmPopupWithOptions(
			"Switch Table",
			fmt.Sprintf(
				"Switching tables will cause loss of unsaved data (%d changes). Are you sure you want to discard unsaved data?",
				m.dirtyEditCount(),
			),
			[]confirmOption{
				{label: "(y) Yes, discard changes and switch table", action: confirmDiscardTable},
				{label: "(n) No, continue editing", action: confirmCancelTableSwitch},
			},
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
			m.openModalConfirmPopupWithOptions(
				"Config",
				"Unsaved changes detected. Choose save, discard, or cancel.",
				[]confirmOption{
					{label: "Save and open config", action: confirmConfigSaveAndOpen},
					{label: "Discard and open config", action: confirmConfigDiscardAndOpen},
					{label: "Cancel", action: confirmConfigCancel},
				},
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

func (m *Model) resetTableContext() {
	m.currentFilter = nil
	m.currentSort = nil
	m.schema = dto.Schema{}
	m.schemaIndex = 0
	m.records = nil
	m.recordPageIndex = 0
	m.recordTotalPages = 1
	m.recordTotalCount = 0
	m.recordSelection = 0
	m.recordColumn = 0
	m.recordLoading = false
	m.recordFieldFocus = false
	m.filterPopup = filterPopup{}
	m.sortPopup = sortPopup{}
	m.helpPopup = helpPopup{}
	m.recordDetail = recordDetailState{}
	m.editPopup = editPopup{}
	m.confirmPopup = confirmPopup{}
	m.clearStagedState()
	m.pendingTableIndex = -1
	m.pendingFilterOpen = false
	m.pendingSortOpen = false
}

func (m *Model) loadViewForSelection() tea.Cmd {
	if m.viewMode == ViewRecords {
		return tea.Batch(m.loadSchemaCmd(), m.loadRecordsCmd(true))
	}
	return m.loadSchemaCmd()
}

func (m *Model) loadSchemaCmd() tea.Cmd {
	tableName := m.currentTableName()
	if strings.TrimSpace(tableName) == "" {
		return nil
	}
	return loadSchemaCmd(m.ctx, m.getSchema, tableName)
}

func (m *Model) loadRecordsCmd(reset bool) tea.Cmd {
	tableName := m.currentTableName()
	if strings.TrimSpace(tableName) == "" {
		return nil
	}
	if reset {
		m.recordPageIndex = 0
	}
	if m.recordLoading {
		return nil
	}
	m.records = nil
	m.recordSelection = 0
	m.recordFieldFocus = false
	m.closeRecordDetail()
	m.recordLoading = true
	m.recordRequestID++
	offset := m.recordPageIndex * recordPageSize
	return loadRecordsCmd(m.ctx, m.listRecords, tableName, offset, recordPageSize, m.currentFilter, m.currentSort, m.recordRequestID)
}

func (m *Model) computeTotalPages(totalCount int) int {
	if totalCount <= 0 {
		return 1
	}
	pages := totalCount / recordPageSize
	if totalCount%recordPageSize != 0 {
		pages++
	}
	if pages < 1 {
		return 1
	}
	return pages
}

func (m *Model) pageSize() int {
	height := m.contentHeight()
	if height < 4 {
		return 1
	}
	if m.focus == FocusContent && m.viewMode == ViewRecords {
		return height - 2
	}
	return height - 1
}

func (m *Model) contentHeight() int {
	if m.height <= 0 {
		return 16
	}
	if m.height <= 6 {
		return 1
	}
	return m.height - 5
}

func (m *Model) contentSelection() int {
	switch m.viewMode {
	case ViewSchema:
		return m.schemaIndex
	case ViewRecords:
		return m.recordSelection
	default:
		return 0
	}
}

func (m *Model) contentMaxIndex() int {
	switch m.viewMode {
	case ViewSchema:
		return len(m.schema.Columns) - 1
	case ViewRecords:
		return m.totalRecordRows() - 1
	default:
		return 0
	}
}

func (m *Model) currentTableName() string {
	if len(m.tables) == 0 {
		return ""
	}
	if m.selectedTable < 0 || m.selectedTable >= len(m.tables) {
		return ""
	}
	return m.tables[m.selectedTable].Name
}

func (m *Model) commandPrompt() string {
	if !m.commandInput.active {
		return ""
	}
	cursor := clamp(m.commandInput.cursor, 0, len(m.commandInput.value))
	value := m.commandInput.value[:cursor] + "|" + m.commandInput.value[cursor:]
	return ":" + value
}

func (m *Model) ShouldOpenConfigSelector() bool {
	return m.openConfigSelector
}

func optionIndex(options []string, value string) int {
	for i, option := range options {
		if strings.EqualFold(option, value) {
			return i
		}
	}
	return 0
}

func sortDirections() []dto.SortDirection {
	return []dto.SortDirection{dto.SortDirectionAsc, dto.SortDirectionDesc}
}

func containsInt(values []int, target int) bool {
	return indexOfInt(values, target) >= 0
}

func indexOfInt(values []int, target int) int {
	for i, value := range values {
		if value == target {
			return i
		}
	}
	return -1
}

func clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func insertAtCursor(value, insert string, cursor int) (string, int) {
	if insert == "" {
		return value, cursor
	}
	cursor = clamp(cursor, 0, len(value))
	updated := value[:cursor] + insert + value[cursor:]
	return updated, cursor + len(insert)
}

func deleteAtCursor(value string, cursor int) (string, int) {
	if value == "" || cursor <= 0 {
		return value, 0
	}
	cursor = clamp(cursor, 0, len(value))
	updated := value[:cursor-1] + value[cursor:]
	return updated, cursor - 1
}

func loadTablesCmd(ctx context.Context, uc listTablesUseCase) tea.Cmd {
	return func() tea.Msg {
		tables, err := uc.Execute(ctx)
		if err != nil {
			return errMsg{err: err}
		}
		return tablesMsg{tables: tables}
	}
}

func loadSchemaCmd(ctx context.Context, uc getSchemaUseCase, tableName string) tea.Cmd {
	return func() tea.Msg {
		schema, err := uc.Execute(ctx, tableName)
		if err != nil {
			return errMsg{err: err}
		}
		return schemaMsg{tableName: tableName, schema: schema}
	}
}

func loadRecordsCmd(ctx context.Context, uc listRecordsUseCase, tableName string, offset, limit int, filter *dto.Filter, sort *dto.Sort, requestID int) tea.Cmd {
	return func() tea.Msg {
		page, err := uc.Execute(ctx, tableName, offset, limit, filter, sort)
		if err != nil {
			return errMsg{err: err}
		}
		return recordsMsg{tableName: tableName, requestID: requestID, page: page}
	}
}

func saveChangesCmd(ctx context.Context, uc saveChangesUseCase, tableName string, changes dto.TableChanges, count int) tea.Cmd {
	return func() tea.Msg {
		err := uc.ExecuteDTO(ctx, tableName, changes)
		return saveChangesMsg{count: count, err: err}
	}
}
