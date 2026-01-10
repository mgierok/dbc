package tui

import (
	"context"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/dto"
	"github.com/mgierok/dbc/internal/application/usecase"
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

type Model struct {
	ctx           context.Context
	listTables    *usecase.ListTables
	getSchema     *usecase.GetSchema
	listRecords   *usecase.ListRecords
	listOperators *usecase.ListOperators

	width  int
	height int

	focus    PanelFocus
	viewMode ViewMode

	tables        []dto.Table
	selectedTable int

	schema      dto.Schema
	schemaIndex int

	records         []dto.RecordRow
	recordOffset    int
	recordHasMore   bool
	recordSelection int
	recordColumn    int
	recordRequestID int
	recordLoading   bool

	currentFilter     *dto.Filter
	popup             filterPopup
	pendingFilterOpen bool
	pendingCtrlW      bool
	pendingG          bool

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

type errMsg struct {
	err error
}

func NewModel(ctx context.Context, listTables *usecase.ListTables, getSchema *usecase.GetSchema, listRecords *usecase.ListRecords, listOperators *usecase.ListOperators) *Model {
	if ctx == nil {
		ctx = context.Background()
	}
	return &Model{
		ctx:           ctx,
		listTables:    listTables,
		getSchema:     getSchema,
		listRecords:   listRecords,
		listOperators: listOperators,
		focus:         FocusTables,
		viewMode:      ViewSchema,
		recordHasMore: true,
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

	if m.popup.active {
		return m.handlePopupKey(msg)
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
		return m.switchToRecords()
	case "F":
		return m.startFilterPopup()
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

func (m *Model) handlePopupKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
		if m.popup.step == filterInputValue {
			m.popup.cursor = clamp(m.popup.cursor-1, 0, len(m.popup.input))
		}
		return m, nil
	case "right":
		if m.popup.step == filterInputValue {
			m.popup.cursor = clamp(m.popup.cursor+1, 0, len(m.popup.input))
		}
		return m, nil
	case "backspace":
		if m.popup.step == filterInputValue && m.popup.input != "" {
			m.popup.input, m.popup.cursor = deleteAtCursor(m.popup.input, m.popup.cursor)
		}
		return m, nil
	}

	if m.popup.step == filterInputValue && msg.Type == tea.KeyRunes {
		insert := string(msg.Runes)
		m.popup.input, m.popup.cursor = insertAtCursor(m.popup.input, insert, m.popup.cursor)
	}
	return m, nil
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
	if m.focus != FocusContent || m.viewMode != ViewRecords {
		return m, nil
	}
	if m.recordColumn > 0 {
		m.recordColumn--
	}
	return m, nil
}

func (m *Model) moveRight() (tea.Model, tea.Cmd) {
	if m.focus != FocusContent || m.viewMode != ViewRecords {
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
	m.popup = filterPopup{
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
	m.popup = filterPopup{}
}

func (m *Model) confirmPopupSelection() (tea.Model, tea.Cmd) {
	switch m.popup.step {
	case filterSelectColumn:
		if len(m.schema.Columns) == 0 {
			return m, nil
		}
		column := m.schema.Columns[m.popup.columnIndex]
		operators, err := m.listOperators.Execute(m.ctx, column.Type)
		if err != nil {
			m.statusMessage = "Error: " + err.Error()
			return m, nil
		}
		m.popup.operators = operators
		m.popup.operatorIndex = 0
		m.popup.step = filterSelectOperator
		return m, nil
	case filterSelectOperator:
		if len(m.popup.operators) == 0 {
			return m, nil
		}
		operator := m.popup.operators[m.popup.operatorIndex]
		if operator.RequiresValue {
			m.popup.input = ""
			m.popup.cursor = 0
			m.popup.step = filterInputValue
			return m, nil
		}
		return m.applyFilter(operator, "")
	case filterInputValue:
		operator := m.popup.operators[m.popup.operatorIndex]
		return m.applyFilter(operator, m.popup.input)
	default:
		return m, nil
	}
}

func (m *Model) movePopupSelection(delta int) {
	switch m.popup.step {
	case filterSelectColumn:
		if len(m.schema.Columns) == 0 {
			return
		}
		m.popup.columnIndex = clamp(m.popup.columnIndex+delta, 0, len(m.schema.Columns)-1)
	case filterSelectOperator:
		if len(m.popup.operators) == 0 {
			return
		}
		m.popup.operatorIndex = clamp(m.popup.operatorIndex+delta, 0, len(m.popup.operators)-1)
	}
}

func (m *Model) applyFilter(operator dto.Operator, value string) (tea.Model, tea.Cmd) {
	if len(m.schema.Columns) == 0 {
		return m, nil
	}
	column := m.schema.Columns[m.popup.columnIndex]
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
	m.popup = filterPopup{}
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
