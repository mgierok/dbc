package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/mgierok/dbc/internal/application/usecase"
	"github.com/mgierok/dbc/internal/infrastructure/config"
	"github.com/mgierok/dbc/internal/infrastructure/engine"
	"github.com/mgierok/dbc/internal/interfaces/tui"
)

type runtimeStartupDependencies struct {
	listConfiguredDatabases *usecase.ListConfiguredDatabases
	createConfiguredDB      *usecase.CreateConfiguredDatabase
	updateConfiguredDB      *usecase.UpdateConfiguredDatabase
	deleteConfiguredDB      *usecase.DeleteConfiguredDatabase
	getActiveConfigPath     *usecase.GetActiveConfigPath
}

func newRuntimeStartupDependencies() (runtimeStartupDependencies, error) {
	cfgPath, err := config.DefaultPath()
	if err != nil {
		return runtimeStartupDependencies{}, err
	}

	configStore := config.NewStore(cfgPath)
	connectionChecker := engine.NewSQLiteConnectionChecker()
	return runtimeStartupDependencies{
		listConfiguredDatabases: usecase.NewListConfiguredDatabases(configStore),
		createConfiguredDB:      usecase.NewCreateConfiguredDatabase(configStore, connectionChecker),
		updateConfiguredDB:      usecase.NewUpdateConfiguredDatabase(configStore, connectionChecker),
		deleteConfiguredDB:      usecase.NewDeleteConfiguredDatabase(configStore),
		getActiveConfigPath:     usecase.NewGetActiveConfigPath(configStore),
	}, nil
}

type runtimeStartupOrchestrator struct {
	options              startupOptions
	deps                 runtimeStartupDependencies
	selectorState        tui.SelectorLaunchState
	sessionScopedOptions []tui.DatabaseOption
	directLaunchPending  bool
}

func newRuntimeStartupOrchestrator(options startupOptions, deps runtimeStartupDependencies) *runtimeStartupOrchestrator {
	return &runtimeStartupOrchestrator{
		options:             options,
		deps:                deps,
		directLaunchPending: options.directLaunchConnString != "",
	}
}

func runRuntimeStartup(options startupOptions) error {
	deps, err := newRuntimeStartupDependencies()
	if err != nil {
		return fmt.Errorf("failed to resolve config path: %w", err)
	}

	orchestrator := newRuntimeStartupOrchestrator(options, deps)
	return orchestrator.run()
}

func (o *runtimeStartupOrchestrator) run() error {
	for {
		selected, selectedStartupPath, err := o.selectDatabase()
		if err != nil {
			if errors.Is(err, tui.ErrDatabaseSelectionCanceled) {
				return nil
			}
			return fmt.Errorf("failed to select database: %w", err)
		}

		o.sessionScopedOptions = trackSessionScopedDirectLaunchOption(o.sessionScopedOptions, selectedStartupPath, selected)
		o.directLaunchPending = false

		shouldContinue, err := o.runSelectedDatabase(selected, selectedStartupPath)
		if err != nil {
			return err
		}
		if !shouldContinue {
			return nil
		}
	}
}

func (o *runtimeStartupOrchestrator) selectDatabase() (tui.DatabaseOption, startupPath, error) {
	strategy := newStartupSelectionStrategy(o.options, o.directLaunchPending)
	return strategy.resolve(
		func() ([]tui.DatabaseOption, error) {
			return listConfiguredDatabaseOptions(context.Background(), o.deps.listConfiguredDatabases)
		},
		func() (tui.DatabaseOption, error) {
			return tui.SelectDatabaseWithState(
				context.Background(),
				o.deps.listConfiguredDatabases,
				o.deps.createConfiguredDB,
				o.deps.updateConfiguredDB,
				o.deps.deleteConfiguredDB,
				o.deps.getActiveConfigPath,
				o.selectorState,
			)
		},
	)
}

func (o *runtimeStartupOrchestrator) runSelectedDatabase(selected tui.DatabaseOption, selectedStartupPath startupPath) (bool, error) {
	db, err := connectSelectedDatabase(selected)
	if err != nil {
		if selectedStartupPath == startupPathDirectLaunch {
			return false, newPresentedStartupFailure(
				startupExitCodeRuntimeFailure,
				buildDirectLaunchFailureMessage(selected.ConnString, err.Error()),
			)
		}

		o.selectorState = tui.SelectorLaunchState{
			StatusMessage:     buildConnectionFailureStatus(selected, err.Error()),
			PreferConnString:  selected.ConnString,
			AdditionalOptions: cloneDatabaseOptions(o.sessionScopedOptions),
		}
		return true, nil
	}
	o.selectorState = tui.SelectorLaunchState{}

	runErr := runRuntimeSession(db)
	if errors.Is(runErr, tui.ErrOpenConfigSelector) {
		o.selectorState = tui.SelectorLaunchState{
			PreferConnString:  selected.ConnString,
			AdditionalOptions: cloneDatabaseOptions(o.sessionScopedOptions),
		}
		return true, nil
	}
	if runErr != nil {
		return false, newPresentedStartupFailure(
			startupExitCodeRuntimeFailure,
			fmt.Sprintf("application error: %v", runErr),
		)
	}
	return false, nil
}

func runRuntimeSession(db *sql.DB) error {
	sqliteEngine := engine.NewSQLiteEngine(db)
	listTables := usecase.NewListTables(sqliteEngine)
	getSchema := usecase.NewGetSchema(sqliteEngine)
	listRecords := usecase.NewListRecords(sqliteEngine)
	listOperators := usecase.NewListOperators(sqliteEngine)
	saveChanges := usecase.NewSaveTableChanges(sqliteEngine)
	translator := usecase.NewStagedChangesTranslator()

	runErr := tui.Run(context.Background(), listTables, getSchema, listRecords, listOperators, saveChanges, translator)
	if closeErr := db.Close(); closeErr != nil {
		log.Printf("failed to close database: %v", closeErr)
	}
	return runErr
}
