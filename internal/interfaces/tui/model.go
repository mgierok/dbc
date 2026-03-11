package tui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/dto"
	"github.com/mgierok/dbc/internal/application/usecase"
	"github.com/mgierok/dbc/internal/interfaces/tui/internal/primitives"
)

const (
	defaultRecordPageLimit = 20
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

type Model struct {
	ctx            context.Context
	listTables     listTablesUseCase
	getSchema      getSchemaUseCase
	listRecords    listRecordsUseCase
	listOperators  listOperatorsUseCase
	saveChanges    saveChangesUseCase
	translator     *usecase.StagedChangesTranslator
	stagingPolicy  *usecase.StagingPolicy
	dirtyNavPolicy *usecase.DirtyNavigationPolicy
	runtimeSession *RuntimeSessionState
	styles         primitives.RenderStyles

	staging stagingState
	read    runtimeReadState
	overlay runtimeOverlayState
	ui      runtimeUIState
}

var _ tea.Model = (*Model)(nil)

var detectRenderStyles = primitives.ResolveRenderStylesFromEnv

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

func NewModel(ctx context.Context, listTables listTablesUseCase, getSchema getSchemaUseCase, listRecords listRecordsUseCase, listOperators listOperatorsUseCase, saveChanges saveChangesUseCase, translator *usecase.StagedChangesTranslator, runtimeSession *RuntimeSessionState) *Model {
	if ctx == nil {
		ctx = context.Background()
	}
	if runtimeSession == nil {
		runtimeSession = &RuntimeSessionState{}
	}
	return &Model{
		ctx:            ctx,
		listTables:     listTables,
		getSchema:      getSchema,
		listRecords:    listRecords,
		listOperators:  listOperators,
		saveChanges:    saveChanges,
		translator:     translator,
		runtimeSession: runtimeSession,
		styles:         detectRenderStyles(),
		read: runtimeReadState{
			focus:            FocusTables,
			viewMode:         ViewSchema,
			recordTotalPages: 1,
		},
		ui: runtimeUIState{
			pendingTableIndex: -1,
		},
	}
}

func (m *Model) Init() tea.Cmd {
	return loadTablesCmd(m.ctx, m.listTables)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.ui.width = msg.Width
		m.ui.height = msg.Height
		return m, nil
	case tablesMsg:
		m.read.tables = msg.tables
		if len(m.read.tables) == 0 {
			m.ui.statusMessage = "No tables found"
			return m, nil
		}
		m.read.selectedTable = 0
		return m, m.loadSchemaCmd()
	case schemaMsg:
		if msg.tableName != m.currentTableName() {
			return m, nil
		}
		m.read.schema = msg.schema
		m.read.schemaIndex = 0
		if m.read.recordColumn >= len(m.read.schema.Columns) {
			m.read.recordColumn = 0
		}
		if m.overlay.pendingFilterOpen {
			m.overlay.pendingFilterOpen = false
			m.openFilterPopup()
		}
		if m.overlay.pendingSortOpen {
			m.overlay.pendingSortOpen = false
			m.openSortPopup()
		}
		return m, nil
	case recordsMsg:
		if msg.tableName != m.currentTableName() {
			return m, nil
		}
		if msg.requestID != m.read.recordRequestID {
			return m, nil
		}
		m.read.recordLoading = false
		m.read.records = msg.page.Rows
		m.read.recordTotalCount = msg.page.TotalCount
		m.read.recordTotalPages = m.computeTotalPages(msg.page.TotalCount)
		m.read.recordPageIndex = clamp(m.read.recordPageIndex, 0, m.read.recordTotalPages-1)
		m.normalizeRecordSelection()
		return m, nil
	case saveChangesMsg:
		if msg.err != nil {
			m.ui.pendingConfigOpen = false
			m.ui.statusMessage = "Error: " + msg.err.Error()
			return m, nil
		}
		m.clearStagedState()
		if m.ui.pendingConfigOpen {
			m.ui.pendingConfigOpen = false
			m.ui.openConfigSelector = true
			m.ui.statusMessage = "Opening config manager"
			return m, tea.Quit
		}
		m.ui.statusMessage = fmt.Sprintf("Saved %d changes", msg.count)
		return m, m.loadRecordsCmd(true)
	case errMsg:
		m.read.recordLoading = false
		m.ui.statusMessage = "Error: " + msg.err.Error()
		return m, nil
	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m *Model) resetTableContext() {
	m.read.currentFilter = nil
	m.read.currentSort = nil
	m.read.schema = dto.Schema{}
	m.read.schemaIndex = 0
	m.read.records = nil
	m.read.recordPageIndex = 0
	m.read.recordTotalPages = 1
	m.read.recordTotalCount = 0
	m.read.recordSelection = 0
	m.read.recordColumn = 0
	m.read.recordLoading = false
	m.read.recordFieldFocus = false
	m.overlay.filterPopup = filterPopup{}
	m.overlay.sortPopup = sortPopup{}
	m.overlay.helpPopup = helpPopup{}
	m.overlay.recordDetail = recordDetailState{}
	m.overlay.editPopup = editPopup{}
	m.overlay.confirmPopup = confirmPopup{}
	m.clearStagedState()
	m.ui.pendingTableIndex = -1
	m.overlay.pendingFilterOpen = false
	m.overlay.pendingSortOpen = false
}

func (m *Model) loadViewForSelection() tea.Cmd {
	if m.read.viewMode == ViewRecords {
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
		m.read.recordPageIndex = 0
	}
	if m.read.recordLoading {
		return nil
	}
	m.read.records = nil
	m.read.recordSelection = 0
	m.read.recordFieldFocus = false
	m.closeRecordDetail()
	m.read.recordLoading = true
	m.read.recordRequestID++
	recordLimit := m.effectiveRecordLimit()
	offset := m.read.recordPageIndex * recordLimit
	return loadRecordsCmd(m.ctx, m.listRecords, tableName, offset, recordLimit, m.read.currentFilter, m.read.currentSort, m.read.recordRequestID)
}

func (m *Model) computeTotalPages(totalCount int) int {
	if totalCount <= 0 {
		return 1
	}
	recordLimit := m.effectiveRecordLimit()
	pages := totalCount / recordLimit
	if totalCount%recordLimit != 0 {
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
	if m.read.focus == FocusContent && m.read.viewMode == ViewRecords {
		return height - 2
	}
	return height - 1
}

func (m *Model) effectiveRecordLimit() int {
	if m.runtimeSession == nil {
		return defaultRecordPageLimit
	}
	return m.runtimeSession.effectiveRecordsPageLimit()
}

func (m *Model) contentHeight() int {
	if m.ui.height <= 0 {
		return 16
	}
	if m.ui.height <= 6 {
		return 1
	}
	return m.ui.height - 5
}

func (m *Model) contentSelection() int {
	switch m.read.viewMode {
	case ViewSchema:
		return m.read.schemaIndex
	case ViewRecords:
		return m.read.recordSelection
	default:
		return 0
	}
}

func (m *Model) contentMaxIndex() int {
	switch m.read.viewMode {
	case ViewSchema:
		return len(m.read.schema.Columns) - 1
	case ViewRecords:
		return m.totalRecordRows() - 1
	default:
		return 0
	}
}

func (m *Model) currentTableName() string {
	if len(m.read.tables) == 0 {
		return ""
	}
	if m.read.selectedTable < 0 || m.read.selectedTable >= len(m.read.tables) {
		return ""
	}
	return m.read.tables[m.read.selectedTable].Name
}

func (m *Model) commandPrompt() string {
	if !m.overlay.commandInput.active {
		return ""
	}
	cursor := clamp(m.overlay.commandInput.cursor, 0, len(m.overlay.commandInput.value))
	value := m.overlay.commandInput.value[:cursor] + "|" + m.overlay.commandInput.value[cursor:]
	return ":" + value
}

func (m *Model) ShouldOpenConfigSelector() bool {
	return m.ui.openConfigSelector
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
