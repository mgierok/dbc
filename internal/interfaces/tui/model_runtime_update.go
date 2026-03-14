package tui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/dto"
)

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
			m.ui.pendingQuitAfterSave = false
			m.ui.statusMessage = "Error: " + msg.err.Error()
			return m, nil
		}
		m.clearStagedState()
		if m.ui.pendingQuitAfterSave {
			m.ui.pendingQuitAfterSave = false
			return m, tea.Quit
		}
		if m.ui.pendingConfigOpen {
			m.ui.pendingConfigOpen = false
			m.ui.openConfigSelector = true
			m.ui.statusMessage = "Opening database selector"
			return m, tea.Quit
		}
		m.ui.statusMessage = formatSavedRowsMessage(msg.count)
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
	m.resetReadRecordBrowsingState()
	m.resetTableOverlayState()
	m.clearStagedState()
	m.ui.pendingTableIndex = -1
}

func (m *Model) resetReadRecordBrowsingState() {
	m.read.records = nil
	m.read.recordPageIndex = 0
	m.read.recordTotalPages = 1
	m.read.recordTotalCount = 0
	m.read.recordSelection = 0
	m.read.recordColumn = 0
	m.read.recordLoading = false
	m.read.recordFieldFocus = false
	m.closeRecordDetail()
}

func (m *Model) resetReadRecordReloadState() {
	m.read.records = nil
	m.read.recordSelection = 0
	m.read.recordFieldFocus = false
	m.closeRecordDetail()
}

func (m *Model) resetTableOverlayState() {
	m.overlay.filterPopup = filterPopup{}
	m.overlay.sortPopup = sortPopup{}
	m.overlay.helpPopup = helpPopup{}
	m.overlay.recordDetail = recordDetailState{}
	m.overlay.editPopup = editPopup{}
	m.overlay.confirmPopup = confirmPopup{}
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
	m.resetReadRecordReloadState()
	m.read.recordLoading = true
	m.read.recordRequestID++
	recordLimit := m.effectiveRecordLimit()
	offset := m.read.recordPageIndex * recordLimit
	return loadRecordsCmd(
		m.ctx,
		m.listRecords,
		tableName,
		offset,
		recordLimit,
		m.read.currentFilter,
		m.read.currentSort,
		m.read.recordRequestID,
	)
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

func saveChangesCmd(ctx context.Context, uc saveChangesUseCase, tableName string, changes dto.TableChanges) tea.Cmd {
	return func() tea.Msg {
		count, err := uc.ExecuteDTO(ctx, tableName, changes)
		return saveChangesMsg{count: count, err: err}
	}
}

func formatSavedRowsMessage(count int) string {
	return fmt.Sprintf("Affected rows: %d", count)
}
