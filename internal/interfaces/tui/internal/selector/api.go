package selector

import (
	"context"
	"errors"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/dto"
)

var (
	ErrDatabaseSelectionCanceled   = errors.New("database selection canceled")
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

	managerIndex int
}

type SelectorLaunchState struct {
	StatusMessage     string
	PreferConnString  string
	AdditionalOptions []DatabaseOption
}

type Manager interface {
	List(ctx context.Context) ([]dto.ConfigDatabase, error)
	Create(ctx context.Context, entry dto.ConfigDatabase) error
	Update(ctx context.Context, index int, entry dto.ConfigDatabase) error
	Delete(ctx context.Context, index int) error
	ActivePath(ctx context.Context) (string, error)
}

type selectorManager = Manager

func SelectDatabase(ctx context.Context, manager Manager) (DatabaseOption, error) {
	return SelectDatabaseWithState(ctx, manager, SelectorLaunchState{})
}

func SelectDatabaseWithState(ctx context.Context, manager Manager, state SelectorLaunchState) (DatabaseOption, error) {
	if manager == nil {
		return DatabaseOption{}, errors.New("selector manager is required")
	}

	model, err := newDatabaseSelectorModel(ctx, manager, state)
	if err != nil {
		return DatabaseOption{}, err
	}

	program := tea.NewProgram(model, tea.WithAltScreen())
	final, err := program.Run()
	if err != nil {
		return DatabaseOption{}, err
	}
	selector, ok := final.(*databaseSelectorModel)
	if !ok {
		return DatabaseOption{}, errors.New("unexpected selector state")
	}
	if selector.canceled {
		return DatabaseOption{}, ErrDatabaseSelectionCanceled
	}
	if !selector.chosen {
		return DatabaseOption{}, ErrDatabaseSelectionUnfinished
	}
	if selector.selected < 0 || selector.selected >= len(selector.options) {
		return DatabaseOption{}, ErrDatabaseSelectionUnfinished
	}
	return selector.options[selector.selected], nil
}
