package tui

import (
	"github.com/mgierok/dbc/internal/application/dto"
	"github.com/mgierok/dbc/internal/application/usecase"
)

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
	filterPopup      filterPopup
	sortPopup        sortPopup
	commandInput     commandInput
	helpPopup        helpPopup
	recordDetail     recordDetailState
	editPopup        editPopup
	confirmPopup     confirmPopup
	databaseSelector runtimeDatabaseSelectorPopup

	pendingFilterOpen bool
	pendingSortOpen   bool
	pendingG          bool
}

// runtimeUIState keeps terminal/session-shell state together so display sizing
// and runtime handoff flags are owned outside the read/write workflows.
type runtimeUIState struct {
	width  int
	height int

	statusMessage            string
	saveInFlight             bool
	pendingSaveSuccessAction usecase.RuntimeSaveSuccessAction
	runtimeSwitchInFlight    bool
	openConfigSelector       bool
	pendingNavigation        *usecase.PendingRuntimeNavigation
	pendingCommandInput      string
}
