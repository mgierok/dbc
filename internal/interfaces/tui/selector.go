package tui

import (
	"context"
	"errors"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/dto"
	"github.com/mgierok/dbc/internal/application/usecase"
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

type selectorManager interface {
	List(ctx context.Context) ([]dto.ConfigDatabase, error)
	Create(ctx context.Context, entry dto.ConfigDatabase) error
	Update(ctx context.Context, index int, entry dto.ConfigDatabase) error
	Delete(ctx context.Context, index int) error
	ActivePath(ctx context.Context) (string, error)
}

type selectorUseCaseAdapter struct {
	list   *usecase.ListConfiguredDatabases
	create *usecase.CreateConfiguredDatabase
	update *usecase.UpdateConfiguredDatabase
	del    *usecase.DeleteConfiguredDatabase
	active *usecase.GetActiveConfigPath
}

func (a selectorUseCaseAdapter) List(ctx context.Context) ([]dto.ConfigDatabase, error) {
	return a.list.Execute(ctx)
}

func (a selectorUseCaseAdapter) Create(ctx context.Context, entry dto.ConfigDatabase) error {
	return a.create.Execute(ctx, entry)
}

func (a selectorUseCaseAdapter) Update(ctx context.Context, index int, entry dto.ConfigDatabase) error {
	return a.update.Execute(ctx, index, entry)
}

func (a selectorUseCaseAdapter) Delete(ctx context.Context, index int) error {
	return a.del.Execute(ctx, index)
}

func (a selectorUseCaseAdapter) ActivePath(ctx context.Context) (string, error) {
	return a.active.Execute(ctx)
}

func SelectDatabase(
	ctx context.Context,
	listConfiguredDatabases *usecase.ListConfiguredDatabases,
	createConfiguredDatabase *usecase.CreateConfiguredDatabase,
	updateConfiguredDatabase *usecase.UpdateConfiguredDatabase,
	deleteConfiguredDatabase *usecase.DeleteConfiguredDatabase,
	getActiveConfigPath *usecase.GetActiveConfigPath,
) (DatabaseOption, error) {
	return SelectDatabaseWithState(
		ctx,
		listConfiguredDatabases,
		createConfiguredDatabase,
		updateConfiguredDatabase,
		deleteConfiguredDatabase,
		getActiveConfigPath,
		SelectorLaunchState{},
	)
}

func SelectDatabaseWithState(
	ctx context.Context,
	listConfiguredDatabases *usecase.ListConfiguredDatabases,
	createConfiguredDatabase *usecase.CreateConfiguredDatabase,
	updateConfiguredDatabase *usecase.UpdateConfiguredDatabase,
	deleteConfiguredDatabase *usecase.DeleteConfiguredDatabase,
	getActiveConfigPath *usecase.GetActiveConfigPath,
	state SelectorLaunchState,
) (DatabaseOption, error) {
	if listConfiguredDatabases == nil || createConfiguredDatabase == nil || updateConfiguredDatabase == nil || deleteConfiguredDatabase == nil || getActiveConfigPath == nil {
		return DatabaseOption{}, errors.New("selector config management use cases are required")
	}

	model, err := newDatabaseSelectorModel(ctx, selectorUseCaseAdapter{
		list:   listConfiguredDatabases,
		create: createConfiguredDatabase,
		update: updateConfiguredDatabase,
		del:    deleteConfiguredDatabase,
		active: getActiveConfigPath,
	}, state)
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
