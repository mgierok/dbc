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
