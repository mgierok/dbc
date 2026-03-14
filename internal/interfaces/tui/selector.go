package tui

import (
	"context"
	"errors"

	"github.com/mgierok/dbc/internal/application/dto"
	"github.com/mgierok/dbc/internal/application/usecase"
	selectorpkg "github.com/mgierok/dbc/internal/interfaces/tui/internal/selector"
)

var (
	ErrDatabaseSelectionCanceled   = selectorpkg.ErrDatabaseSelectionCanceled
	ErrDatabaseSelectionDismissed  = selectorpkg.ErrDatabaseSelectionDismissed
	ErrDatabaseSelectionUnfinished = selectorpkg.ErrDatabaseSelectionUnfinished
)

var selectDatabaseWithStateFn = selectorpkg.SelectDatabaseWithState

type DatabaseOptionSource = selectorpkg.DatabaseOptionSource
type SelectorBrowseEscBehavior = selectorpkg.SelectorBrowseEscBehavior

const (
	DatabaseOptionSourceConfig = selectorpkg.DatabaseOptionSourceConfig
	DatabaseOptionSourceCLI    = selectorpkg.DatabaseOptionSourceCLI

	SelectorBrowseEscBehaviorStartupExit   = selectorpkg.SelectorBrowseEscBehaviorStartupExit
	SelectorBrowseEscBehaviorRuntimeResume = selectorpkg.SelectorBrowseEscBehaviorRuntimeResume
)

type DatabaseOption = selectorpkg.DatabaseOption
type SelectorLaunchState = selectorpkg.SelectorLaunchState

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

	return selectDatabaseWithStateFn(ctx, selectorUseCaseAdapter{
		list:   listConfiguredDatabases,
		create: createConfiguredDatabase,
		update: updateConfiguredDatabase,
		del:    deleteConfiguredDatabase,
		active: getActiveConfigPath,
	}, state)
}
