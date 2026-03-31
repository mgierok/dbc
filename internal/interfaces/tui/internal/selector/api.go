package selector

import (
	"context"
	"errors"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/dto"
)

var (
	ErrDatabaseSelectionCanceled   = errors.New("database selection canceled")
	ErrDatabaseSelectionDismissed  = errors.New("database selection dismissed")
	ErrDatabaseSelectionUnfinished = errors.New("database selection not confirmed")
)

type DatabaseOptionSource string

const (
	DatabaseOptionSourceConfig DatabaseOptionSource = "config"
	DatabaseOptionSourceCLI    DatabaseOptionSource = "cli"
)

type DatabaseOption struct {
	Name       string
	ConnString string
	Source     DatabaseOptionSource

	configIndex int
	canEdit     bool
	canDelete   bool
}

type SelectorBrowseEscBehavior int

const (
	SelectorBrowseEscBehaviorStartupExit SelectorBrowseEscBehavior = iota
	SelectorBrowseEscBehaviorRuntimeResume
)

type SelectorLaunchState struct {
	StatusMessage     string
	PreferConnString  string
	AdditionalOptions []DatabaseOption
	BrowseEscBehavior SelectorBrowseEscBehavior
}

type Manager interface {
	LoadState(ctx context.Context, input dto.DatabaseSelectorLoadInput) (dto.DatabaseSelectorState, error)
	Create(ctx context.Context, entry dto.ConfigDatabase) error
	Update(ctx context.Context, index int, entry dto.ConfigDatabase) error
	Delete(ctx context.Context, index int) error
}

type selectorManager = Manager

type selectorProgram interface {
	Run() (tea.Model, error)
}

var newSelectorProgram = func(model tea.Model, options ...tea.ProgramOption) selectorProgram {
	return tea.NewProgram(model, options...)
}

type IntentType int

const (
	IntentTypeNone IntentType = iota
	IntentTypeClose
	IntentTypeSelect
)

type Intent struct {
	Type   IntentType
	Option DatabaseOption
}

type Controller struct {
	model *databaseSelectorModel
}

func NewStartupController(ctx context.Context, manager Manager, state SelectorLaunchState) (*Controller, error) {
	model, err := newDatabaseSelectorModelWithHost(ctx, manager, controllerHostConfig{
		browseEscActionLabel:     "quit",
		firstSetupEscActionLabel: "exit app",
	}, state)
	if err != nil {
		return nil, err
	}
	return &Controller{model: model}, nil
}

func NewRuntimeController(ctx context.Context, manager Manager, state SelectorLaunchState) (*Controller, error) {
	model, err := newDatabaseSelectorModelWithHost(ctx, manager, controllerHostConfig{
		browseEscActionLabel:     "close",
		firstSetupEscActionLabel: "close",
	}, state)
	if err != nil {
		return nil, err
	}
	return &Controller{model: model}, nil
}

func (c *Controller) Handle(msg tea.Msg) tea.Cmd {
	if c == nil || c.model == nil {
		return nil
	}
	_, cmd := c.model.Update(msg)
	return cmd
}

func (c *Controller) ConsumeIntent() Intent {
	if c == nil || c.model == nil {
		return Intent{}
	}
	return c.model.consumeIntent()
}

func (c *Controller) PopupLines(totalWidth, totalHeight int) []string {
	if c == nil || c.model == nil {
		return nil
	}
	return c.model.popupLines(totalWidth, totalHeight)
}

func (c *Controller) View() string {
	if c == nil || c.model == nil {
		return ""
	}
	return c.model.View()
}

func (c *Controller) SetStatusMessage(message string) {
	if c == nil || c.model == nil {
		return
	}
	c.model.browse.statusMessage = message
}

func (c *Controller) SetSelectedIndex(index int) {
	if c == nil || c.model == nil || len(c.model.options) == 0 {
		return
	}
	c.model.browse.selected = clamp(index, 0, len(c.model.options)-1)
}

type startupHostModel struct {
	controller *Controller
}

func (m *startupHostModel) Init() tea.Cmd {
	if m == nil || m.controller == nil || m.controller.model == nil {
		return nil
	}
	return m.controller.model.Init()
}

func (m *startupHostModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m == nil || m.controller == nil || m.controller.model == nil {
		return m, nil
	}
	cmd := m.controller.Handle(msg)
	if m.controller.model.peekIntent().Type != IntentTypeNone {
		return m, tea.Quit
	}
	return m, cmd
}

func (m *startupHostModel) View() string {
	if m == nil || m.controller == nil || m.controller.model == nil {
		return ""
	}
	return m.controller.model.View()
}

func SelectDatabase(ctx context.Context, manager Manager) (DatabaseOption, error) {
	return SelectDatabaseWithState(ctx, manager, SelectorLaunchState{})
}

func SelectDatabaseWithState(ctx context.Context, manager Manager, state SelectorLaunchState) (DatabaseOption, error) {
	if manager == nil {
		return DatabaseOption{}, errors.New("selector manager is required")
	}

	controller, err := NewStartupController(ctx, manager, state)
	if err != nil {
		return DatabaseOption{}, err
	}

	program := newSelectorProgram(&startupHostModel{controller: controller}, tea.WithAltScreen())
	final, err := program.Run()
	if err != nil {
		return DatabaseOption{}, err
	}
	hostModel, ok := final.(*startupHostModel)
	if !ok {
		return DatabaseOption{}, errors.New("unexpected selector state")
	}
	intent := hostModel.controller.model.peekIntent()
	if hostModel.controller.model.dismissed {
		return DatabaseOption{}, ErrDatabaseSelectionDismissed
	}
	if intent.Type == IntentTypeClose {
		return DatabaseOption{}, ErrDatabaseSelectionCanceled
	}
	if intent.Type != IntentTypeSelect {
		return DatabaseOption{}, ErrDatabaseSelectionUnfinished
	}
	if intent.Option.ConnString == "" {
		return DatabaseOption{}, ErrDatabaseSelectionUnfinished
	}
	return intent.Option, nil
}
