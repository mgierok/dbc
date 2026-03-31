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
	load   *usecase.LoadDatabaseSelectorState
	create *usecase.CreateConfiguredDatabase
	update *usecase.UpdateConfiguredDatabase
	del    *usecase.DeleteConfiguredDatabase
}

func (a selectorUseCaseAdapter) LoadState(ctx context.Context, input dto.DatabaseSelectorLoadInput) (dto.DatabaseSelectorState, error) {
	return a.load.Execute(ctx, input)
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

func SelectDatabase(
	ctx context.Context,
	loadDatabaseSelectorState *usecase.LoadDatabaseSelectorState,
	createConfiguredDatabase *usecase.CreateConfiguredDatabase,
	updateConfiguredDatabase *usecase.UpdateConfiguredDatabase,
	deleteConfiguredDatabase *usecase.DeleteConfiguredDatabase,
) (DatabaseOption, error) {
	return SelectDatabaseWithState(
		ctx,
		loadDatabaseSelectorState,
		createConfiguredDatabase,
		updateConfiguredDatabase,
		deleteConfiguredDatabase,
		SelectorLaunchState{},
	)
}

func SelectDatabaseWithState(
	ctx context.Context,
	loadDatabaseSelectorState *usecase.LoadDatabaseSelectorState,
	createConfiguredDatabase *usecase.CreateConfiguredDatabase,
	updateConfiguredDatabase *usecase.UpdateConfiguredDatabase,
	deleteConfiguredDatabase *usecase.DeleteConfiguredDatabase,
	state SelectorLaunchState,
) (DatabaseOption, error) {
	if loadDatabaseSelectorState == nil || createConfiguredDatabase == nil || updateConfiguredDatabase == nil || deleteConfiguredDatabase == nil {
		return DatabaseOption{}, errors.New("selector config management use cases are required")
	}

	return selectDatabaseWithStateFn(ctx, selectorUseCaseAdapter{
		load:   loadDatabaseSelectorState,
		create: createConfiguredDatabase,
		update: updateConfiguredDatabase,
		del:    deleteConfiguredDatabase,
	}, state)
}

func cloneDatabaseOptions(options []DatabaseOption) []DatabaseOption {
	if len(options) == 0 {
		return nil
	}
	cloned := make([]DatabaseOption, len(options))
	copy(cloned, options)
	return cloned
}
