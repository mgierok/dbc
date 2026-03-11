package tui

import "github.com/mgierok/dbc/internal/application/dto"

// runtimeReadState keeps the mutable read-side runtime state together so
// navigation and data-browsing concerns stay distinct from overlays and UI.
type runtimeReadState struct {
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

	currentFilter *dto.Filter
	currentSort   *dto.Sort
}

// runtimeOverlayState keeps popup/input state and overlay-specific deferred
// actions together so modal workflow ownership stays explicit.
type runtimeOverlayState struct {
	filterPopup  filterPopup
	sortPopup    sortPopup
	commandInput commandInput
	helpPopup    helpPopup
	recordDetail recordDetailState
	editPopup    editPopup
	confirmPopup confirmPopup

	pendingFilterOpen bool
	pendingSortOpen   bool
	pendingG          bool
}

// runtimeUIState keeps terminal/session-shell state together so display sizing
// and runtime handoff flags are owned outside the read/write workflows.
type runtimeUIState struct {
	width  int
	height int

	statusMessage      string
	openConfigSelector bool
	pendingTableIndex  int
	pendingConfigOpen  bool
}

// Keep the staged refactor lint-clean while later chunks rewire Model to these
// grouped state holders.
var (
	_ = runtimeReadState{
		focus:            FocusTables,
		viewMode:         ViewSchema,
		tables:           nil,
		selectedTable:    0,
		schema:           dto.Schema{},
		schemaIndex:      0,
		records:          nil,
		recordPageIndex:  0,
		recordTotalPages: 0,
		recordTotalCount: 0,
		recordSelection:  0,
		recordColumn:     0,
		recordRequestID:  0,
		recordLoading:    false,
		recordFieldFocus: false,
		currentFilter:    nil,
		currentSort:      nil,
	}
	_ = runtimeOverlayState{
		filterPopup:       filterPopup{},
		sortPopup:         sortPopup{},
		commandInput:      commandInput{},
		helpPopup:         helpPopup{},
		recordDetail:      recordDetailState{},
		editPopup:         editPopup{},
		confirmPopup:      confirmPopup{},
		pendingFilterOpen: false,
		pendingSortOpen:   false,
		pendingG:          false,
	}
	_ = runtimeUIState{
		width:              0,
		height:             0,
		statusMessage:      "",
		openConfigSelector: false,
		pendingTableIndex:  0,
		pendingConfigOpen:  false,
	}
)
