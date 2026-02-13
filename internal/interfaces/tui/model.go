package tui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/dto"
	"github.com/mgierok/dbc/internal/application/usecase"
	"github.com/mgierok/dbc/internal/domain/model"
	"github.com/mgierok/dbc/internal/domain/service"
)

const (
	recordPageSize     = 50
	recordLoadDistance = 5
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

type filterPopup struct {
	active        bool
	step          filterStep
	columnIndex   int
	operatorIndex int
	input         string
	operators     []dto.Operator
	cursor        int
}

type stagedEdit struct {
	Value model.Value
}

type recordEdits struct {
	identity model.RecordIdentity
	changes  map[int]stagedEdit
}

type pendingInsertRow struct {
	values       map[int]stagedEdit
	explicitAuto map[int]bool
	showAuto     bool
}

type recordDelete struct {
	identity model.RecordIdentity
}

type stagedOperation struct {
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
)

type confirmPopup struct {
	active  bool
	action  confirmAction
	message string
}

type pkColumn struct {
	index  int
	column dto.SchemaColumn
}

type Model struct {
	ctx           context.Context
	listTables    *usecase.ListTables
	getSchema     *usecase.GetSchema
	listRecords   *usecase.ListRecords
	listOperators *usecase.ListOperators
	saveChanges   *usecase.SaveTableChanges

	width  int
	height int

	focus    PanelFocus
	viewMode ViewMode

	tables        []dto.Table
	selectedTable int

	schema      dto.Schema
	schemaIndex int

	records          []dto.RecordRow
	recordOffset     int
	recordHasMore    bool
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
	filterPopup       filterPopup
	editPopup         editPopup
	confirmPopup      confirmPopup
	pendingFilterOpen bool
	pendingCtrlW      bool
	pendingG          bool
	pendingTableIndex int

	statusMessage string
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

func NewModel(ctx context.Context, listTables *usecase.ListTables, getSchema *usecase.GetSchema, listRecords *usecase.ListRecords, listOperators *usecase.ListOperators, saveChanges *usecase.SaveTableChanges) *Model {
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
		focus:             FocusTables,
		viewMode:          ViewSchema,
		recordHasMore:     true,
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
		return m, nil
	case recordsMsg:
		if msg.tableName != m.currentTableName() {
			return m, nil
		}
		if msg.requestID != m.recordRequestID {
			return m, nil
		}
		m.recordLoading = false
		m.records = append(m.records, msg.page.Rows...)
		m.recordOffset += len(msg.page.Rows)
		m.recordHasMore = msg.page.HasMore
		return m, nil
	case saveChangesMsg:
		if msg.err != nil {
			m.statusMessage = "Error: " + msg.err.Error()
			return m, nil
		}
		m.clearStagedState()
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

	if m.editPopup.active {
		return m.handleEditPopupKey(msg)
	}
	if m.confirmPopup.active {
		return m.handleConfirmPopupKey(msg)
	}
	if m.filterPopup.active {
		return m.handleFilterPopupKey(msg)
	}

	if m.pendingCtrlW {
		m.pendingCtrlW = false
		switch key {
		case "h":
			m.focus = FocusTables
			return m, nil
		case "l":
			m.focus = FocusContent
			return m, nil
		case "w":
			m.toggleFocus()
			return m, nil
		default:
			return m, nil
		}
	}

	if m.pendingG {
		if key == "g" {
			m.pendingG = false
			return m.jumpTop()
		}
		m.pendingG = false
	}

	switch key {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "ctrl+w":
		m.pendingCtrlW = true
		return m, nil
	case "g":
		m.pendingG = true
		return m, nil
	case "G":
		return m.jumpBottom()
	case "enter":
		if m.viewMode == ViewRecords && m.focus == FocusContent {
			if !m.recordFieldFocus {
				return m.enableRecordFieldFocus()
			}
			return m.openEditPopup()
		}
		return m.switchToRecords()
	case "esc":
		if m.viewMode == ViewRecords && m.recordFieldFocus {
			m.recordFieldFocus = false
			return m, nil
		}
		return m, nil
	case "F":
		return m.startFilterPopup()
	case "w":
		return m.requestSaveChanges()
	case "i":
		return m.addPendingInsert()
	case "d":
		return m.toggleDeleteSelection()
	case "ctrl+a":
		return m.toggleInsertAutoFields()
	case "j":
		return m.moveDown()
	case "k":
		return m.moveUp()
	case "h":
		return m.moveLeft()
	case "l":
		return m.moveRight()
	case "ctrl+f":
		return m.pageDown()
	case "ctrl+b":
		return m.pageUp()
	default:
		return m, nil
	}
}

func (m *Model) handleFilterPopupKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "esc":
		m.closeFilterPopup()
		return m, nil
	case "enter":
		return m.confirmPopupSelection()
	case "j":
		m.movePopupSelection(1)
		return m, nil
	case "k":
		m.movePopupSelection(-1)
		return m, nil
	case "left":
		if m.filterPopup.step == filterInputValue {
			m.filterPopup.cursor = clamp(m.filterPopup.cursor-1, 0, len(m.filterPopup.input))
		}
		return m, nil
	case "right":
		if m.filterPopup.step == filterInputValue {
			m.filterPopup.cursor = clamp(m.filterPopup.cursor+1, 0, len(m.filterPopup.input))
		}
		return m, nil
	case "backspace":
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

func (m *Model) handleEditPopupKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	column, ok := m.editColumn()
	if !ok {
		m.closeEditPopup()
		return m, nil
	}

	switch key {
	case "esc":
		m.closeEditPopup()
		return m, nil
	case "ctrl+n":
		if !column.Nullable {
			m.editPopup.errorMessage = "Column is not nullable"
			return m, nil
		}
		m.editPopup.isNull = true
		m.editPopup.errorMessage = ""
		return m, nil
	case "enter":
		return m.confirmEditPopup()
	case "j":
		if column.Input.Kind == dto.ColumnInputSelect {
			if len(column.Input.Options) > 0 {
				m.editPopup.optionIndex = clamp(m.editPopup.optionIndex+1, 0, len(column.Input.Options)-1)
				m.editPopup.isNull = false
				m.editPopup.errorMessage = ""
			}
		}
		return m, nil
	case "k":
		if column.Input.Kind == dto.ColumnInputSelect {
			if len(column.Input.Options) > 0 {
				m.editPopup.optionIndex = clamp(m.editPopup.optionIndex-1, 0, len(column.Input.Options)-1)
				m.editPopup.isNull = false
				m.editPopup.errorMessage = ""
			}
		}
		return m, nil
	case "left":
		if column.Input.Kind == dto.ColumnInputText {
			m.editPopup.cursor = clamp(m.editPopup.cursor-1, 0, len(m.editPopup.input))
			m.editPopup.errorMessage = ""
		}
		return m, nil
	case "right":
		if column.Input.Kind == dto.ColumnInputText {
			m.editPopup.cursor = clamp(m.editPopup.cursor+1, 0, len(m.editPopup.input))
			m.editPopup.errorMessage = ""
		}
		return m, nil
	case "backspace":
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
	switch key {
	case "esc", "n":
		m.closeConfirmPopup()
		m.pendingTableIndex = -1
		return m, nil
	case "enter", "y":
		action := m.confirmPopup.action
		m.closeConfirmPopup()
		switch action {
		case confirmSave:
			return m.confirmSaveChanges()
		case confirmDiscardTable:
			return m.confirmDiscardTableSwitch()
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

	value, err := service.ParseValue(column.Type, input, m.editPopup.isNull, column.Nullable)
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
	m.confirmPopup = confirmPopup{active: true, action: action, message: message}
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

func (m *Model) toggleFocus() {
	if m.focus == FocusTables {
		m.focus = FocusContent
		return
	}
	m.focus = FocusTables
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
		m.openConfirmPopup(confirmDiscardTable, "Discard changes and switch tables?")
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
		return m, m.maybeLoadMoreRecords()
	default:
		return m, nil
	}
}

func (m *Model) switchToSchema() (tea.Model, tea.Cmd) {
	m.viewMode = ViewSchema
	m.recordFieldFocus = false
	if m.currentTableName() == "" {
		return m, nil
	}
	if len(m.schema.Columns) == 0 {
		return m, m.loadSchemaCmd()
	}
	return m, nil
}

func (m *Model) switchToRecords() (tea.Model, tea.Cmd) {
	m.viewMode = ViewRecords
	m.recordFieldFocus = false
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

func (m *Model) applyFilter(operator dto.Operator, value string) (tea.Model, tea.Cmd) {
	if len(m.schema.Columns) == 0 {
		return m, nil
	}
	column := m.schema.Columns[m.filterPopup.columnIndex]
	filter := &dto.Filter{
		Column: column.Name,
		Operator: dto.Operator{
			Name:          operator.Name,
			SQL:           operator.SQL,
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

func (m *Model) resetTableContext() {
	m.currentFilter = nil
	m.schema = dto.Schema{}
	m.schemaIndex = 0
	m.records = nil
	m.recordOffset = 0
	m.recordSelection = 0
	m.recordColumn = 0
	m.recordHasMore = true
	m.recordLoading = false
	m.recordFieldFocus = false
	m.filterPopup = filterPopup{}
	m.editPopup = editPopup{}
	m.confirmPopup = confirmPopup{}
	m.clearStagedState()
	m.pendingTableIndex = -1
	m.pendingFilterOpen = false
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
		m.records = nil
		m.recordOffset = 0
		m.recordSelection = 0
		m.recordFieldFocus = false
		m.recordHasMore = true
	}
	if !m.recordHasMore || m.recordLoading {
		return nil
	}
	m.recordLoading = true
	m.recordRequestID++
	return loadRecordsCmd(m.ctx, m.listRecords, tableName, m.recordOffset, recordPageSize, m.currentFilter, m.recordRequestID)
}

func (m *Model) maybeLoadMoreRecords() tea.Cmd {
	if m.viewMode != ViewRecords || !m.recordHasMore {
		return nil
	}
	if len(m.records) == 0 {
		return m.loadRecordsCmd(true)
	}
	persistedIndex := m.persistedRowIndex(m.recordSelection)
	if persistedIndex < 0 {
		return nil
	}
	if persistedIndex >= len(m.records)-recordLoadDistance {
		return m.loadRecordsCmd(false)
	}
	return nil
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
		return 20
	}
	if m.height <= 2 {
		return 1
	}
	return m.height - 1
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

func (m *Model) schemaColumns() []string {
	if len(m.schema.Columns) == 0 {
		return nil
	}
	columns := make([]string, len(m.schema.Columns))
	for i, column := range m.schema.Columns {
		columns[i] = column.Name
	}
	return columns
}

func (m *Model) requestSaveChanges() (tea.Model, tea.Cmd) {
	if m.viewMode != ViewRecords || !m.hasDirtyEdits() {
		return m, nil
	}
	if m.saveChanges == nil {
		m.statusMessage = "Error: save use case unavailable"
		return m, nil
	}
	m.openConfirmPopup(confirmSave, "Save staged changes?")
	return m, nil
}

func (m *Model) confirmSaveChanges() (tea.Model, tea.Cmd) {
	changes, err := m.buildTableChanges()
	if err != nil {
		m.statusMessage = "Error: " + err.Error()
		return m, nil
	}
	if len(changes.Inserts) == 0 && len(changes.Updates) == 0 && len(changes.Deletes) == 0 {
		return m, nil
	}
	count := m.dirtyEditCount()
	return m, saveChangesCmd(m.ctx, m.saveChanges, m.currentTableName(), changes, count)
}

func (m *Model) confirmDiscardTableSwitch() (tea.Model, tea.Cmd) {
	if m.pendingTableIndex < 0 || m.pendingTableIndex >= len(m.tables) {
		m.pendingTableIndex = -1
		return m, nil
	}
	target := m.pendingTableIndex
	m.pendingTableIndex = -1
	m.clearStagedState()
	m.selectedTable = target
	m.resetTableContext()
	return m, m.loadViewForSelection()
}

func (m *Model) addPendingInsert() (tea.Model, tea.Cmd) {
	if m.viewMode != ViewRecords || m.focus != FocusContent {
		return m, nil
	}
	if len(m.schema.Columns) == 0 {
		m.statusMessage = "Error: no schema loaded"
		return m, nil
	}
	row := pendingInsertRow{
		values:       make(map[int]stagedEdit, len(m.schema.Columns)),
		explicitAuto: make(map[int]bool),
	}
	for index, column := range m.schema.Columns {
		row.values[index] = stagedEdit{Value: initialInsertValue(column)}
	}
	m.pendingInserts = append([]pendingInsertRow{row}, m.pendingInserts...)
	m.recordSelection = 0
	m.recordColumn = m.defaultRecordColumnForRow(0)
	m.recordFieldFocus = true
	return m, nil
}

func (m *Model) toggleDeleteSelection() (tea.Model, tea.Cmd) {
	if m.viewMode != ViewRecords || m.focus != FocusContent {
		return m, nil
	}
	if m.recordSelection < 0 || m.recordSelection >= m.totalRecordRows() {
		return m, nil
	}
	if insertIndex, isInsert := m.pendingInsertIndexForSelection(); isInsert {
		m.removePendingInsert(insertIndex)
		return m, nil
	}
	if !m.canEditRecords() {
		m.statusMessage = "Error: table has no primary key"
		return m, nil
	}
	key, identity, err := m.recordIdentityForVisibleRow(m.recordSelection)
	if err != nil {
		m.statusMessage = "Error: " + err.Error()
		return m, nil
	}
	if m.pendingDeletes == nil {
		m.pendingDeletes = make(map[string]recordDelete)
	}
	if _, exists := m.pendingDeletes[key]; exists {
		delete(m.pendingDeletes, key)
	} else {
		m.pendingDeletes[key] = recordDelete{identity: identity}
	}
	return m, nil
}

func (m *Model) toggleInsertAutoFields() (tea.Model, tea.Cmd) {
	if m.viewMode != ViewRecords || m.focus != FocusContent {
		return m, nil
	}
	insertIndex, isInsert := m.pendingInsertIndexForSelection()
	if !isInsert {
		return m, nil
	}
	row := m.pendingInserts[insertIndex]
	row.showAuto = !row.showAuto
	m.pendingInserts[insertIndex] = row
	visibleColumns := m.visibleColumnIndicesForSelection()
	if len(visibleColumns) == 0 {
		m.recordColumn = 0
		m.recordFieldFocus = false
		return m, nil
	}
	if !containsInt(visibleColumns, m.recordColumn) {
		m.recordColumn = visibleColumns[0]
	}
	return m, nil
}

func (m *Model) stageEdit(rowIndex, columnIndex int, value model.Value) error {
	if columnIndex < 0 || columnIndex >= len(m.schema.Columns) {
		return fmt.Errorf("column index out of range")
	}
	if rowIndex < 0 || rowIndex >= m.totalRecordRows() {
		return fmt.Errorf("record index out of range")
	}
	if insertIndex, isInsert := m.pendingInsertIndex(rowIndex); isInsert {
		return m.stageInsertEdit(insertIndex, columnIndex, value)
	}
	return m.stagePersistedEdit(rowIndex, columnIndex, value)
}

func (m *Model) stagePersistedEdit(visibleRowIndex, columnIndex int, value model.Value) error {
	key, identity, err := m.recordIdentityForVisibleRow(visibleRowIndex)
	if err != nil {
		return err
	}
	if m.pendingUpdates == nil {
		m.pendingUpdates = make(map[string]recordEdits)
	}
	edits := m.pendingUpdates[key]
	if edits.changes == nil {
		edits.changes = make(map[int]stagedEdit)
	}
	edits.identity = identity

	original := m.visibleRowValue(visibleRowIndex, columnIndex)
	if displayValue(value) == original {
		delete(edits.changes, columnIndex)
		if len(edits.changes) == 0 {
			delete(m.pendingUpdates, key)
			return nil
		}
		m.pendingUpdates[key] = edits
		return nil
	}

	edits.changes[columnIndex] = stagedEdit{Value: value}
	m.pendingUpdates[key] = edits
	return nil
}

func (m *Model) stageInsertEdit(insertIndex, columnIndex int, value model.Value) error {
	if insertIndex < 0 || insertIndex >= len(m.pendingInserts) {
		return fmt.Errorf("insert index out of range")
	}
	if columnIndex < 0 || columnIndex >= len(m.schema.Columns) {
		return fmt.Errorf("column index out of range")
	}
	row := m.pendingInserts[insertIndex]
	if row.values == nil {
		row.values = make(map[int]stagedEdit, len(m.schema.Columns))
	}
	if row.explicitAuto == nil {
		row.explicitAuto = make(map[int]bool)
	}
	row.values[columnIndex] = stagedEdit{Value: value}
	if m.schema.Columns[columnIndex].AutoIncrement {
		row.explicitAuto[columnIndex] = true
	}
	m.pendingInserts[insertIndex] = row
	return nil
}

func (m *Model) canEditRecords() bool {
	for _, column := range m.schema.Columns {
		if column.PrimaryKey {
			return true
		}
	}
	return false
}

func (m *Model) editColumn() (dto.SchemaColumn, bool) {
	index := m.editPopup.columnIndex
	if index < 0 || index >= len(m.schema.Columns) {
		return dto.SchemaColumn{}, false
	}
	return m.schema.Columns[index], true
}

func (m *Model) recordValue(rowIndex, columnIndex int) string {
	if rowIndex < 0 || rowIndex >= len(m.records) {
		return ""
	}
	values := m.records[rowIndex].Values
	if columnIndex < 0 || columnIndex >= len(values) {
		return ""
	}
	return values[columnIndex]
}

func (m *Model) stagedEditForRow(rowIndex, columnIndex int) (stagedEdit, bool) {
	persistedIndex := m.persistedRowIndex(rowIndex)
	if persistedIndex < 0 {
		return stagedEdit{}, false
	}
	key, ok := m.recordKeyForPersistedRow(persistedIndex)
	if !ok {
		return stagedEdit{}, false
	}
	edits, ok := m.pendingUpdates[key]
	if !ok {
		return stagedEdit{}, false
	}
	edit, ok := edits.changes[columnIndex]
	return edit, ok
}

func (m *Model) recordKeyForPersistedRow(rowIndex int) (string, bool) {
	pkColumns := m.primaryKeyColumns()
	if len(pkColumns) == 0 || rowIndex < 0 || rowIndex >= len(m.records) {
		return "", false
	}
	values := m.records[rowIndex].Values
	parts := make([]string, 0, len(pkColumns))
	for _, pk := range pkColumns {
		if pk.index < 0 || pk.index >= len(values) {
			return "", false
		}
		parts = append(parts, fmt.Sprintf("%s=%s", pk.column.Name, values[pk.index]))
	}
	return strings.Join(parts, "|"), true
}

func (m *Model) recordIdentityForVisibleRow(rowIndex int) (string, model.RecordIdentity, error) {
	persistedIndex := m.persistedRowIndex(rowIndex)
	if persistedIndex < 0 {
		return "", model.RecordIdentity{}, fmt.Errorf("record index out of range")
	}
	return m.recordIdentityForPersistedRow(persistedIndex)
}

func (m *Model) recordIdentityForPersistedRow(rowIndex int) (string, model.RecordIdentity, error) {
	pkColumns := m.primaryKeyColumns()
	if len(pkColumns) == 0 {
		return "", model.RecordIdentity{}, fmt.Errorf("table has no primary key")
	}
	if rowIndex < 0 || rowIndex >= len(m.records) {
		return "", model.RecordIdentity{}, fmt.Errorf("record index out of range")
	}
	values := m.records[rowIndex].Values
	keys := make([]model.ColumnValue, 0, len(pkColumns))
	parts := make([]string, 0, len(pkColumns))
	for _, pk := range pkColumns {
		if pk.index < 0 || pk.index >= len(values) {
			return "", model.RecordIdentity{}, fmt.Errorf("primary key index out of range")
		}
		rawValue := values[pk.index]
		isNull := strings.EqualFold(rawValue, "NULL")
		nullable := pk.column.Nullable && !pk.column.PrimaryKey
		parsed, err := service.ParseValue(pk.column.Type, rawValue, isNull, nullable)
		if err != nil {
			return "", model.RecordIdentity{}, err
		}
		keys = append(keys, model.ColumnValue{Column: pk.column.Name, Value: parsed})
		parts = append(parts, fmt.Sprintf("%s=%s", pk.column.Name, rawValue))
	}
	return strings.Join(parts, "|"), model.RecordIdentity{Keys: keys}, nil
}

func (m *Model) primaryKeyColumns() []pkColumn {
	if len(m.schema.Columns) == 0 {
		return nil
	}
	var pkColumns []pkColumn
	for i, column := range m.schema.Columns {
		if column.PrimaryKey {
			pkColumns = append(pkColumns, pkColumn{index: i, column: column})
		}
	}
	return pkColumns
}

func (m *Model) buildTableChanges() (model.TableChanges, error) {
	changes := model.TableChanges{}

	for _, row := range m.pendingInserts {
		insert, err := m.buildInsertChange(row)
		if err != nil {
			return model.TableChanges{}, err
		}
		changes.Inserts = append(changes.Inserts, insert)
	}

	deleteKeys := make(map[string]struct{}, len(m.pendingDeletes))
	for key, deleteChange := range m.pendingDeletes {
		deleteKeys[key] = struct{}{}
		changes.Deletes = append(changes.Deletes, model.RecordDelete{
			Identity: deleteChange.identity,
		})
	}

	for key, edits := range m.pendingUpdates {
		if _, deleted := deleteKeys[key]; deleted {
			continue
		}
		if len(edits.changes) == 0 {
			continue
		}
		if edits.identity.RowID == nil && len(edits.identity.Keys) == 0 {
			return model.TableChanges{}, fmt.Errorf("record identity missing")
		}
		updateChanges := make([]model.ColumnValue, 0, len(edits.changes))
		for colIndex, change := range edits.changes {
			if colIndex < 0 || colIndex >= len(m.schema.Columns) {
				return model.TableChanges{}, fmt.Errorf("column index out of range")
			}
			column := m.schema.Columns[colIndex]
			updateChanges = append(updateChanges, model.ColumnValue{Column: column.Name, Value: change.Value})
		}
		changes.Updates = append(changes.Updates, model.RecordUpdate{
			Identity: edits.identity,
			Changes:  updateChanges,
		})
	}

	return changes, nil
}

func (m *Model) dirtyEditCount() int {
	count := 0
	for _, edits := range m.pendingUpdates {
		count += len(edits.changes)
	}
	count += len(m.pendingInserts)
	count += len(m.pendingDeletes)
	return count
}

func (m *Model) hasDirtyEdits() bool {
	return m.dirtyEditCount() > 0
}

func (m *Model) clearStagedState() {
	m.pendingInserts = nil
	m.pendingUpdates = nil
	m.pendingDeletes = nil
	m.history = nil
	m.future = nil
}

func (m *Model) totalRecordRows() int {
	return len(m.pendingInserts) + len(m.records)
}

func (m *Model) pendingInsertIndex(rowIndex int) (int, bool) {
	if rowIndex < 0 || rowIndex >= len(m.pendingInserts) {
		return -1, false
	}
	return rowIndex, true
}

func (m *Model) pendingInsertIndexForSelection() (int, bool) {
	return m.pendingInsertIndex(m.recordSelection)
}

func (m *Model) persistedRowIndex(rowIndex int) int {
	persisted := rowIndex - len(m.pendingInserts)
	if persisted < 0 || persisted >= len(m.records) {
		return -1
	}
	return persisted
}

func (m *Model) visibleRowValue(rowIndex, columnIndex int) string {
	if insertIndex, isInsert := m.pendingInsertIndex(rowIndex); isInsert {
		row := m.pendingInserts[insertIndex]
		if value, ok := row.values[columnIndex]; ok {
			return displayValue(value.Value)
		}
		return ""
	}
	persistedIndex := m.persistedRowIndex(rowIndex)
	if persistedIndex < 0 {
		return ""
	}
	return m.recordValue(persistedIndex, columnIndex)
}

func (m *Model) isRowMarkedDelete(rowIndex int) bool {
	persistedIndex := m.persistedRowIndex(rowIndex)
	if persistedIndex < 0 {
		return false
	}
	key, ok := m.recordKeyForPersistedRow(persistedIndex)
	if !ok {
		return false
	}
	_, marked := m.pendingDeletes[key]
	return marked
}

func (m *Model) removePendingInsert(index int) {
	if index < 0 || index >= len(m.pendingInserts) {
		return
	}
	m.pendingInserts = append(m.pendingInserts[:index], m.pendingInserts[index+1:]...)
	if m.recordSelection >= m.totalRecordRows() {
		m.recordSelection = maxInt(0, m.totalRecordRows()-1)
	}
	if m.totalRecordRows() == 0 {
		m.recordSelection = 0
		m.recordFieldFocus = false
	}
	visibleColumns := m.visibleColumnIndicesForSelection()
	if len(visibleColumns) > 0 && !containsInt(visibleColumns, m.recordColumn) {
		m.recordColumn = visibleColumns[0]
	}
}

func (m *Model) visibleColumnIndicesForSelection() []int {
	if len(m.schema.Columns) == 0 {
		return nil
	}
	insertIndex, isInsert := m.pendingInsertIndexForSelection()
	columns := make([]int, 0, len(m.schema.Columns))
	for idx, column := range m.schema.Columns {
		if isInsert && !m.pendingInserts[insertIndex].showAuto && column.AutoIncrement {
			continue
		}
		columns = append(columns, idx)
	}
	return columns
}

func (m *Model) defaultRecordColumnForRow(rowIndex int) int {
	columns := m.visibleColumnIndicesForRow(rowIndex)
	if len(columns) == 0 {
		return 0
	}
	return columns[0]
}

func (m *Model) visibleColumnIndicesForRow(rowIndex int) []int {
	if len(m.schema.Columns) == 0 {
		return nil
	}
	insertIndex, isInsert := m.pendingInsertIndex(rowIndex)
	columns := make([]int, 0, len(m.schema.Columns))
	for idx, column := range m.schema.Columns {
		if isInsert && !m.pendingInserts[insertIndex].showAuto && column.AutoIncrement {
			continue
		}
		columns = append(columns, idx)
	}
	return columns
}

func (m *Model) buildInsertChange(row pendingInsertRow) (model.RecordInsert, error) {
	insert := model.RecordInsert{}
	for colIndex, column := range m.schema.Columns {
		value, ok := row.values[colIndex]
		if !ok {
			continue
		}
		textValue := displayValue(value.Value)
		if !column.Nullable && column.DefaultValue == nil && !column.AutoIncrement && strings.TrimSpace(textValue) == "" {
			return model.RecordInsert{}, fmt.Errorf("value for column %q is required", column.Name)
		}
		columnValue := model.ColumnValue{
			Column: column.Name,
			Value:  value.Value,
		}
		if column.AutoIncrement {
			if row.explicitAuto[colIndex] {
				insert.ExplicitAutoValues = append(insert.ExplicitAutoValues, columnValue)
			}
			continue
		}
		insert.Values = append(insert.Values, columnValue)
	}
	if len(insert.Values) == 0 && len(insert.ExplicitAutoValues) == 0 {
		return model.RecordInsert{}, model.ErrMissingInsertValues
	}
	return insert, nil
}

func initialInsertValue(column dto.SchemaColumn) model.Value {
	if column.DefaultValue != nil {
		return model.Value{Text: *column.DefaultValue, Raw: *column.DefaultValue}
	}
	if column.Nullable {
		return model.Value{IsNull: true, Text: "NULL"}
	}
	return model.Value{Text: "", Raw: ""}
}

func displayValue(value model.Value) string {
	if value.IsNull {
		return "NULL"
	}
	if strings.TrimSpace(value.Text) != "" {
		return value.Text
	}
	if value.Raw != nil {
		return fmt.Sprint(value.Raw)
	}
	return ""
}

func optionIndex(options []string, value string) int {
	for i, option := range options {
		if strings.EqualFold(option, value) {
			return i
		}
	}
	return 0
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

func loadTablesCmd(ctx context.Context, uc *usecase.ListTables) tea.Cmd {
	return func() tea.Msg {
		tables, err := uc.Execute(ctx)
		if err != nil {
			return errMsg{err: err}
		}
		return tablesMsg{tables: tables}
	}
}

func loadSchemaCmd(ctx context.Context, uc *usecase.GetSchema, tableName string) tea.Cmd {
	return func() tea.Msg {
		schema, err := uc.Execute(ctx, tableName)
		if err != nil {
			return errMsg{err: err}
		}
		return schemaMsg{tableName: tableName, schema: schema}
	}
}

func loadRecordsCmd(ctx context.Context, uc *usecase.ListRecords, tableName string, offset, limit int, filter *dto.Filter, requestID int) tea.Cmd {
	return func() tea.Msg {
		page, err := uc.Execute(ctx, tableName, offset, limit, filter)
		if err != nil {
			return errMsg{err: err}
		}
		return recordsMsg{tableName: tableName, requestID: requestID, page: page}
	}
}

func saveChangesCmd(ctx context.Context, uc *usecase.SaveTableChanges, tableName string, changes model.TableChanges, count int) tea.Cmd {
	return func() tea.Msg {
		err := uc.Execute(ctx, tableName, changes)
		return saveChangesMsg{count: count, err: err}
	}
}
