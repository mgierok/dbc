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
	rowIndex int
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
	saveEdits     *usecase.SaveRecordEdits

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
	stagedEdits      map[string]recordEdits

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

type saveEditsMsg struct {
	count int
	err   error
}

type errMsg struct {
	err error
}

func NewModel(ctx context.Context, listTables *usecase.ListTables, getSchema *usecase.GetSchema, listRecords *usecase.ListRecords, listOperators *usecase.ListOperators, saveEdits *usecase.SaveRecordEdits) *Model {
	if ctx == nil {
		ctx = context.Background()
	}
	return &Model{
		ctx:               ctx,
		listTables:        listTables,
		getSchema:         getSchema,
		listRecords:       listRecords,
		listOperators:     listOperators,
		saveEdits:         saveEdits,
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
	case saveEditsMsg:
		if msg.err != nil {
			m.statusMessage = "Error: " + msg.err.Error()
			return m, nil
		}
		m.applyStagedEdits()
		m.clearStagedEdits()
		m.statusMessage = fmt.Sprintf("Saved %d changes", msg.count)
		return m, nil
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
		return m.requestSaveEdits()
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
			return m.confirmSaveEdits()
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
	if m.viewMode != ViewRecords || len(m.records) == 0 {
		return m, nil
	}
	if m.recordColumn >= len(m.schema.Columns) {
		m.recordColumn = 0
	}
	m.recordFieldFocus = true
	return m, nil
}

func (m *Model) openEditPopup() (tea.Model, tea.Cmd) {
	if !m.canEditRecords() {
		m.statusMessage = "Error: table has no primary key"
		return m, nil
	}
	if m.recordSelection < 0 || m.recordSelection >= len(m.records) {
		return m, nil
	}
	if m.recordColumn < 0 || m.recordColumn >= len(m.schema.Columns) {
		return m, nil
	}
	column := m.schema.Columns[m.recordColumn]
	currentValue := m.recordValue(m.recordSelection, m.recordColumn)
	if staged, ok := m.stagedEditForRow(m.recordSelection, m.recordColumn); ok {
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
	if m.recordColumn > 0 {
		m.recordColumn--
	}
	return m, nil
}

func (m *Model) moveRight() (tea.Model, tea.Cmd) {
	if m.focus != FocusContent || m.viewMode != ViewRecords || !m.recordFieldFocus {
		return m, nil
	}
	columns := m.schemaColumns()
	if len(columns) == 0 {
		return m, nil
	}
	if m.recordColumn < len(columns)-1 {
		m.recordColumn++
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
	m.stagedEdits = nil
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
	if m.recordSelection >= len(m.records)-recordLoadDistance {
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
		return len(m.records) - 1
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

func (m *Model) requestSaveEdits() (tea.Model, tea.Cmd) {
	if m.viewMode != ViewRecords || !m.hasDirtyEdits() {
		return m, nil
	}
	if m.saveEdits == nil {
		m.statusMessage = "Error: save use case unavailable"
		return m, nil
	}
	m.openConfirmPopup(confirmSave, "Save staged changes?")
	return m, nil
}

func (m *Model) confirmSaveEdits() (tea.Model, tea.Cmd) {
	updates, err := m.buildRecordUpdates()
	if err != nil {
		m.statusMessage = "Error: " + err.Error()
		return m, nil
	}
	if len(updates) == 0 {
		return m, nil
	}
	count := m.dirtyEditCount()
	return m, saveEditsCmd(m.ctx, m.saveEdits, m.currentTableName(), updates, count)
}

func (m *Model) confirmDiscardTableSwitch() (tea.Model, tea.Cmd) {
	if m.pendingTableIndex < 0 || m.pendingTableIndex >= len(m.tables) {
		m.pendingTableIndex = -1
		return m, nil
	}
	target := m.pendingTableIndex
	m.pendingTableIndex = -1
	m.clearStagedEdits()
	m.selectedTable = target
	m.resetTableContext()
	return m, m.loadViewForSelection()
}

func (m *Model) stageEdit(rowIndex, columnIndex int, value model.Value) error {
	if rowIndex < 0 || rowIndex >= len(m.records) {
		return fmt.Errorf("record index out of range")
	}
	if columnIndex < 0 || columnIndex >= len(m.schema.Columns) {
		return fmt.Errorf("column index out of range")
	}
	key, identity, err := m.recordIdentityForRow(rowIndex)
	if err != nil {
		return err
	}
	if m.stagedEdits == nil {
		m.stagedEdits = make(map[string]recordEdits)
	}
	edits := m.stagedEdits[key]
	if edits.changes == nil {
		edits.changes = make(map[int]stagedEdit)
	}
	edits.identity = identity
	edits.rowIndex = rowIndex

	original := m.recordValue(rowIndex, columnIndex)
	if displayValue(value) == original {
		delete(edits.changes, columnIndex)
		if len(edits.changes) == 0 {
			delete(m.stagedEdits, key)
			return nil
		}
		m.stagedEdits[key] = edits
		return nil
	}

	edits.changes[columnIndex] = stagedEdit{Value: value}
	m.stagedEdits[key] = edits
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
	key, ok := m.recordKeyForRow(rowIndex)
	if !ok {
		return stagedEdit{}, false
	}
	edits, ok := m.stagedEdits[key]
	if !ok {
		return stagedEdit{}, false
	}
	edit, ok := edits.changes[columnIndex]
	return edit, ok
}

func (m *Model) recordKeyForRow(rowIndex int) (string, bool) {
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

func (m *Model) recordIdentityForRow(rowIndex int) (string, model.RecordIdentity, error) {
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

func (m *Model) buildRecordUpdates() ([]model.RecordUpdate, error) {
	if len(m.stagedEdits) == 0 {
		return nil, nil
	}
	updates := make([]model.RecordUpdate, 0, len(m.stagedEdits))
	for _, edits := range m.stagedEdits {
		if len(edits.changes) == 0 {
			continue
		}
		if edits.identity.RowID == nil && len(edits.identity.Keys) == 0 {
			return nil, fmt.Errorf("record identity missing")
		}
		changes := make([]model.ColumnValue, 0, len(edits.changes))
		for index := range edits.changes {
			if index < 0 || index >= len(m.schema.Columns) {
				return nil, fmt.Errorf("column index out of range")
			}
		}
		for colIndex, change := range edits.changes {
			column := m.schema.Columns[colIndex]
			changes = append(changes, model.ColumnValue{Column: column.Name, Value: change.Value})
		}
		updates = append(updates, model.RecordUpdate{
			Identity: edits.identity,
			Changes:  changes,
		})
	}
	return updates, nil
}

func (m *Model) applyStagedEdits() {
	for _, edits := range m.stagedEdits {
		if edits.rowIndex < 0 || edits.rowIndex >= len(m.records) {
			continue
		}
		for colIndex, change := range edits.changes {
			if colIndex < 0 || colIndex >= len(m.records[edits.rowIndex].Values) {
				continue
			}
			m.records[edits.rowIndex].Values[colIndex] = displayValue(change.Value)
		}
	}
}

func (m *Model) clearStagedEdits() {
	m.stagedEdits = nil
}

func (m *Model) dirtyEditCount() int {
	count := 0
	for _, edits := range m.stagedEdits {
		count += len(edits.changes)
	}
	return count
}

func (m *Model) hasDirtyEdits() bool {
	return m.dirtyEditCount() > 0
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

func saveEditsCmd(ctx context.Context, uc *usecase.SaveRecordEdits, tableName string, updates []model.RecordUpdate, count int) tea.Cmd {
	return func() tea.Msg {
		err := uc.Execute(ctx, tableName, updates)
		return saveEditsMsg{count: count, err: err}
	}
}
