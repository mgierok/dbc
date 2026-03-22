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
	sessionScopedOptions  []tui.DatabaseOption
	directLaunchPending   bool
	pendingRuntimeTarget  *tui.DatabaseOption
	selectDatabaseFn      func() (tui.DatabaseOption, startupPath, error)
	runSelectedDatabaseFn func(tui.DatabaseOption, startupPath) (bool, error)
	connectDatabaseFn     func(tui.DatabaseOption) (*sql.DB, error)
	openRuntimeRunDepsFn  func(tui.DatabaseOption) (tui.RuntimeRunDeps, error)
	runRuntimeSessionFn   func(tui.RuntimeRunDeps) (tui.RuntimeExitResult, error)
}

func newRuntimeStartupOrchestrator(options startupOptions, deps runtimeStartupDependencies) *runtimeStartupOrchestrator {
	orchestrator := &runtimeStartupOrchestrator{
		options:              options,
		deps:                 deps,
		directLaunchPending:  options.directLaunchConnString != "",
		connectDatabaseFn:    connectSelectedDatabase,
		openRuntimeRunDepsFn: nil,
		runRuntimeSessionFn:  runRuntimeSession,
	}
	orchestrator.selectDatabaseFn = orchestrator.selectDatabase
	orchestrator.runSelectedDatabaseFn = orchestrator.runSelectedDatabase
	orchestrator.openRuntimeRunDepsFn = orchestrator.openRuntimeRunDeps
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
			if errors.Is(err, tui.ErrDatabaseSelectionDismissed) {
				return nil
			}
			return fmt.Errorf("failed to select database: %w", err)
		}

		o.sessionScopedOptions = trackSessionScopedCLIOption(o.sessionScopedOptions, selected)
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
	if o.pendingRuntimeTarget != nil {
		selected := *o.pendingRuntimeTarget
		o.pendingRuntimeTarget = nil
		return selected, startupPathRuntimeSwitch, nil
	}

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
	openRuntimeRunDepsFn := o.openRuntimeRunDepsFn
	if openRuntimeRunDepsFn == nil {
		openRuntimeRunDepsFn = o.openRuntimeRunDeps
	}

	runtimeDeps, err := openRuntimeRunDepsFn(selected)
	if err != nil {
		if selectedStartupPath == startupPathDirectLaunch {
			return false, newPresentedStartupFailure(
				startupExitCodeRuntimeFailure,
				buildDirectLaunchFailureMessage(selected.ConnString, err.Error()),
			)
		}

		retryState := tui.SelectorLaunchState{
			StatusMessage:     buildConnectionFailureStatus(selected, err.Error()),
			PreferConnString:  selected.ConnString,
			AdditionalOptions: cloneDatabaseOptions(o.sessionScopedOptions),
		}
		o.selectorState = retryState
		return true, nil
	}
	o.selectorState = tui.SelectorLaunchState{}

	runRuntimeSessionFn := o.runRuntimeSessionFn
	if runRuntimeSessionFn == nil {
		runRuntimeSessionFn = runRuntimeSession
	}

	runtimeExitResult, runErr := runRuntimeSessionFn(runtimeDeps)
	if runErr != nil {
		return false, newPresentedStartupFailure(
			startupExitCodeRuntimeFailure,
			fmt.Sprintf("application error: %v", runErr),
		)
	}

	if runtimeExitResult.Action == tui.RuntimeExitActionOpenDatabaseNext {
		nextTarget := runtimeExitResult.NextDatabase
		o.pendingRuntimeTarget = &nextTarget
		return true, nil
	}
	return false, nil
}

func (o *runtimeStartupOrchestrator) openRuntimeRunDeps(selected tui.DatabaseOption) (tui.RuntimeRunDeps, error) {
	connectDatabaseFn := o.connectDatabaseFn
	if connectDatabaseFn == nil {
		connectDatabaseFn = connectSelectedDatabase
	}

	db, err := connectDatabaseFn(selected)
	if err != nil {
		return tui.RuntimeRunDeps{}, err
	}

	sqliteEngine := engine.NewSQLiteEngine(db)
	return tui.RuntimeRunDeps{
		ListTables:    usecase.NewListTables(sqliteEngine),
		GetSchema:     usecase.NewGetSchema(sqliteEngine),
		ListRecords:   usecase.NewListRecords(sqliteEngine),
		ListOperators: usecase.NewListOperators(sqliteEngine),
		SaveChanges:   usecase.NewSaveTableChanges(sqliteEngine),
		Translator:    usecase.NewStagedChangesTranslator(),
		DatabaseSelector: &tui.RuntimeDatabaseSelectorDeps{
			ListConfiguredDatabases:  o.deps.listConfiguredDatabases,
			CreateConfiguredDatabase: o.deps.createConfiguredDB,
			UpdateConfiguredDatabase: o.deps.updateConfiguredDB,
			DeleteConfiguredDatabase: o.deps.deleteConfiguredDB,
			GetActiveConfigPath:      o.deps.getActiveConfigPath,
			CurrentDatabase:          selected,
			AdditionalOptions:        cloneDatabaseOptions(o.sessionScopedOptions),
		},
		Close: func() {
			if closeErr := closeDatabaseFn(db); closeErr != nil {
				logPrintfFn("failed to close database: %v", closeErr)
			}
		},
	}, nil
}

func runRuntimeSession(runtimeDeps tui.RuntimeRunDeps) (tui.RuntimeExitResult, error) {
	return tuiRunFn(context.Background(), runtimeDeps)
}
