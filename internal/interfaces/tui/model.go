package tui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/dto"
	"github.com/mgierok/dbc/internal/application/usecase"
	"github.com/mgierok/dbc/internal/interfaces/tui/internal/primitives"
	selectorpkg "github.com/mgierok/dbc/internal/interfaces/tui/internal/selector"
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
	active        bool
	mode          commandInputMode
	value         string
	cursor        int
	pendingStatus string
}

type commandInputMode int

const (
	commandInputModeEditing commandInputMode = iota
	commandInputModePending
)

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

type dirtyConfirmFlow int

const (
	dirtyConfirmFlowTableSwitch dirtyConfirmFlow = iota + 1
	dirtyConfirmFlowDatabaseTransition
	dirtyConfirmFlowQuit
)

type confirmAction int

const (
	confirmDiscardTable confirmAction = iota + 1
	confirmCancelTableSwitch
	confirmDatabaseTransitionSave
	confirmDatabaseTransitionDiscard
	confirmDatabaseTransitionCancel
	confirmDiscardQuit
	confirmCancelQuit
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

type runtimeDatabaseSelectorPopup struct {
	active     bool
	controller *selectorpkg.Controller
}

type DatabaseConnectionValidator interface {
	Execute(ctx context.Context, connString string) error
}

type RuntimeDatabaseSelectorDeps struct {
	ListConfiguredDatabases  *usecase.ListConfiguredDatabases
	CreateConfiguredDatabase *usecase.CreateConfiguredDatabase
	UpdateConfiguredDatabase *usecase.UpdateConfiguredDatabase
	DeleteConfiguredDatabase *usecase.DeleteConfiguredDatabase
	GetActiveConfigPath      *usecase.GetActiveConfigPath
	CurrentDatabase          DatabaseOption
	AdditionalOptions        []DatabaseOption
}

type Model struct {
	ctx                         context.Context
	runtimeBundleCtx            context.Context
	runtimeBundleCancel         context.CancelFunc
	runtimeBundleToken          int
	listTables                  listTablesUseCase
	getSchema                   getSchemaUseCase
	listRecords                 listRecordsUseCase
	listOperators               listOperatorsUseCase
	saveChanges                 saveChangesUseCase
	translator                  *usecase.StagedChangesTranslator
	stagingPolicy               *usecase.StagingPolicy
	dirtyNavPolicy              *usecase.DirtyNavigationPolicy
	runtimeSession              *RuntimeSessionState
	runtimeDatabaseSelectorDeps *RuntimeDatabaseSelectorDeps
	runtimeClose                func()
	styles                      primitives.RenderStyles
	exitResult                  RuntimeExitResult

	staging stagingState
	read    runtimeReadState
	overlay runtimeOverlayState
	ui      runtimeUIState
}

var _ tea.Model = (*Model)(nil)

var detectRenderStyles = primitives.ResolveRenderStylesFromEnv

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
	ExecuteDTO(ctx context.Context, tableName string, changes dto.TableChanges) (int, error)
}

func NewModel(ctx context.Context, runtimeDeps RuntimeRunDeps, runtimeSession *RuntimeSessionState) *Model {
	if ctx == nil {
		ctx = context.Background()
	}
	if runtimeSession == nil {
		runtimeSession = &RuntimeSessionState{}
	}
	model := &Model{
		ctx:            ctx,
		runtimeSession: runtimeSession,
		styles:         detectRenderStyles(),
		exitResult:     runtimeExitResultQuit(),
		read: runtimeReadState{
			focus:            FocusTables,
			viewMode:         ViewSchema,
			recordTotalPages: 1,
		},
		ui: runtimeUIState{
			pendingTableIndex: -1,
		},
	}
	model.initializeRuntimeBundle()
	model.applyRuntimeRunDeps(runtimeDeps)
	return model
}

func (m *Model) Init() tea.Cmd {
	return loadTablesCmd(m.runtimeReadContext(), m.listTables, m.runtimeBundleToken)
}
