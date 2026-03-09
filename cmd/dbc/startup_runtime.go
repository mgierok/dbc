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

var (
	newRuntimeStartupDependenciesFn = newRuntimeStartupDependencies
	newRuntimeStartupOrchestratorFn = newRuntimeStartupOrchestrator
	selectDatabaseWithStateFn       = tui.SelectDatabaseWithState
	tuiRunFn                        = tui.Run
	closeDatabaseFn                 = func(db *sql.DB) error { return db.Close() }
	logPrintfFn                     = log.Printf
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
	options               startupOptions
	deps                  runtimeStartupDependencies
	selectorState         tui.SelectorLaunchState
	runtimeSessionState   tui.RuntimeSessionState
	sessionScopedOptions  []tui.DatabaseOption
	directLaunchPending   bool
	selectDatabaseFn      func() (tui.DatabaseOption, startupPath, error)
	runSelectedDatabaseFn func(tui.DatabaseOption, startupPath) (bool, error)
	connectDatabaseFn     func(tui.DatabaseOption) (*sql.DB, error)
	runRuntimeSessionFn   func(*sql.DB, *tui.RuntimeSessionState) error
}

func newRuntimeStartupOrchestrator(options startupOptions, deps runtimeStartupDependencies) *runtimeStartupOrchestrator {
	orchestrator := &runtimeStartupOrchestrator{
		options:             options,
		deps:                deps,
		directLaunchPending: options.directLaunchConnString != "",
		connectDatabaseFn:   connectSelectedDatabase,
		runRuntimeSessionFn: runRuntimeSession,
	}
	orchestrator.selectDatabaseFn = orchestrator.selectDatabase
	orchestrator.runSelectedDatabaseFn = orchestrator.runSelectedDatabase
	return orchestrator
}

func runRuntimeStartup(options startupOptions) error {
	deps, err := newRuntimeStartupDependenciesFn()
	if err != nil {
		return fmt.Errorf("failed to resolve config path: %w", err)
	}

	orchestrator := newRuntimeStartupOrchestratorFn(options, deps)
	return orchestrator.run()
}

func (o *runtimeStartupOrchestrator) run() error {
	selectDatabaseFn := o.selectDatabaseFn
	if selectDatabaseFn == nil {
		selectDatabaseFn = o.selectDatabase
	}
	runSelectedDatabaseFn := o.runSelectedDatabaseFn
	if runSelectedDatabaseFn == nil {
		runSelectedDatabaseFn = o.runSelectedDatabase
	}

	for {
		selected, selectedStartupPath, err := selectDatabaseFn()
		if err != nil {
			if errors.Is(err, tui.ErrDatabaseSelectionCanceled) {
				return nil
			}
			return fmt.Errorf("failed to select database: %w", err)
		}

		o.sessionScopedOptions = trackSessionScopedDirectLaunchOption(o.sessionScopedOptions, selectedStartupPath, selected)
		o.directLaunchPending = false

		shouldContinue, err := runSelectedDatabaseFn(selected, selectedStartupPath)
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
			return selectDatabaseWithStateFn(
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
	connectDatabaseFn := o.connectDatabaseFn
	if connectDatabaseFn == nil {
		connectDatabaseFn = connectSelectedDatabase
	}

	db, err := connectDatabaseFn(selected)
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

	runRuntimeSessionFn := o.runRuntimeSessionFn
	if runRuntimeSessionFn == nil {
		runRuntimeSessionFn = runRuntimeSession
	}

	runErr := runRuntimeSessionFn(db, &o.runtimeSessionState)
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

func runRuntimeSession(db *sql.DB, runtimeSession *tui.RuntimeSessionState) error {
	sqliteEngine := engine.NewSQLiteEngine(db)
	listTables := usecase.NewListTables(sqliteEngine)
	getSchema := usecase.NewGetSchema(sqliteEngine)
	listRecords := usecase.NewListRecords(sqliteEngine)
	listOperators := usecase.NewListOperators(sqliteEngine)
	saveChanges := usecase.NewSaveTableChanges(sqliteEngine)
	translator := usecase.NewStagedChangesTranslator()

	runErr := tuiRunFn(context.Background(), listTables, getSchema, listRecords, listOperators, saveChanges, translator, runtimeSession)
	if closeErr := closeDatabaseFn(db); closeErr != nil {
		logPrintfFn("failed to close database: %v", closeErr)
	}
	return runErr
}
