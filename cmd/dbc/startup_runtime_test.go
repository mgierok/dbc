package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/mgierok/dbc/internal/application/dto"
	"github.com/mgierok/dbc/internal/application/port"
	"github.com/mgierok/dbc/internal/application/usecase"
	"github.com/mgierok/dbc/internal/interfaces/tui"
)

func TestNewRuntimeStartupDependencies_UsesDefaultConfigPathAndBuildsUseCases(t *testing.T) {
	// Arrange
	home := t.TempDir()
	t.Setenv("HOME", home)
	expectedPath := filepath.Join(home, ".config", "dbc", "config.json")

	// Act
	deps, err := newRuntimeStartupDependencies()

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if deps.listConfiguredDatabases == nil {
		t.Fatal("expected listConfiguredDatabases use case to be initialized")
	}
	if deps.createConfiguredDB == nil {
		t.Fatal("expected createConfiguredDB use case to be initialized")
	}
	if deps.updateConfiguredDB == nil {
		t.Fatal("expected updateConfiguredDB use case to be initialized")
	}
	if deps.deleteConfiguredDB == nil {
		t.Fatal("expected deleteConfiguredDB use case to be initialized")
	}
	if deps.loadDatabaseSelectorState == nil {
		t.Fatal("expected loadDatabaseSelectorState use case to be initialized")
	}

	state, err := deps.loadDatabaseSelectorState.Execute(context.Background(), dto.DatabaseSelectorLoadInput{})
	if err != nil {
		t.Fatalf("expected selector state to resolve, got %v", err)
	}
	if state.ActiveConfigPath != expectedPath {
		t.Fatalf("expected active config path %q, got %q", expectedPath, state.ActiveConfigPath)
	}
}

func TestNewRuntimeStartupOrchestrator_TracksDirectLaunchPendingFromOptions(t *testing.T) {
	t.Parallel()

	// Arrange
	options := startupOptions{directLaunchConnString: "/tmp/direct.sqlite"}

	// Act
	orchestrator := newRuntimeStartupOrchestrator(options, runtimeStartupDependencies{})

	// Assert
	if !orchestrator.directLaunchPending {
		t.Fatal("expected direct launch to start in pending state")
	}
}

func TestRunRuntimeStartup_WrapsDependencyResolutionFailure(t *testing.T) {
	// Arrange
	expectedErr := errors.New("config path unavailable")
	withStartupRuntimeHooks(t, startupRuntimeHookOverrides{
		newDependencies: func() (runtimeStartupDependencies, error) {
			return runtimeStartupDependencies{}, expectedErr
		},
	})

	// Act
	err := runRuntimeStartup(startupOptions{})

	// Assert
	if err == nil {
		t.Fatal("expected dependency-resolution error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to resolve config path") {
		t.Fatalf("expected wrapped config-path failure, got %v", err)
	}
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected wrapped error %v, got %v", expectedErr, err)
	}
}

func TestRunRuntimeStartup_UsesResolvedDependenciesAndOrchestrator(t *testing.T) {
	// Arrange
	options := startupOptions{directLaunchConnString: "/tmp/direct.sqlite"}
	expectedDeps := runtimeStartupDependencies{}
	orchestratorCalled := false
	withStartupRuntimeHooks(t, startupRuntimeHookOverrides{
		newDependencies: func() (runtimeStartupDependencies, error) {
			return expectedDeps, nil
		},
		newOrchestrator: func(gotOptions startupOptions, gotDeps runtimeStartupDependencies) *runtimeStartupOrchestrator {
			orchestratorCalled = true
			if gotOptions != options {
				t.Fatalf("expected options %+v, got %+v", options, gotOptions)
			}
			if gotDeps != expectedDeps {
				t.Fatalf("expected dependencies %+v, got %+v", expectedDeps, gotDeps)
			}
			return &runtimeStartupOrchestrator{
				selectDatabaseFn: func() (tui.DatabaseOption, startupPath, error) {
					return tui.DatabaseOption{}, startupPathSelector, tui.ErrDatabaseSelectionCanceled
				},
			}
		},
	})

	// Act
	err := runRuntimeStartup(options)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !orchestratorCalled {
		t.Fatal("expected runtime startup orchestrator to be constructed")
	}
}

func TestRuntimeStartupOrchestratorRun_ReturnsNilWhenSelectionEndsWithoutDatabase(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		err  error
	}{
		{
			name: "selector canceled",
			err:  tui.ErrDatabaseSelectionCanceled,
		},
		{
			name: "selector dismissed",
			err:  tui.ErrDatabaseSelectionDismissed,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			orchestrator := &runtimeStartupOrchestrator{
				selectDatabaseFn: func() (tui.DatabaseOption, startupPath, error) {
					return tui.DatabaseOption{}, startupPathSelector, tc.err
				},
				runSelectedDatabaseFn: func(tui.DatabaseOption, startupPath) (bool, error) {
					t.Fatal("expected selected database runner to stay unused after selection end")
					return false, nil
				},
			}

			// Act
			err := orchestrator.run()

			// Assert
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}

func TestRuntimeStartupOrchestratorRun_WrapsSelectionFailure(t *testing.T) {
	t.Parallel()

	// Arrange
	expectedErr := errors.New("selector exploded")
	orchestrator := &runtimeStartupOrchestrator{
		selectDatabaseFn: func() (tui.DatabaseOption, startupPath, error) {
			return tui.DatabaseOption{}, startupPathSelector, expectedErr
		},
	}

	// Act
	err := orchestrator.run()

	// Assert
	if err == nil {
		t.Fatal("expected selection error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to select database") {
		t.Fatalf("expected selection error wrapper, got %v", err)
	}
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected wrapped error %v, got %v", expectedErr, err)
	}
}

func TestRuntimeStartupOrchestratorRun_ReopensRequestedDatabaseWithoutSelectorRoundTrip(t *testing.T) {
	t.Parallel()

	// Arrange
	selectedA := tui.DatabaseOption{
		Name:       "primary",
		ConnString: "/tmp/primary.sqlite",
		Source:     tui.DatabaseOptionSourceConfig,
	}
	selectedB := tui.DatabaseOption{
		Name:       "analytics",
		ConnString: "/tmp/analytics.sqlite",
		Source:     tui.DatabaseOptionSourceConfig,
	}
	orchestrator := newRuntimeStartupOrchestrator(startupOptions{}, runtimeStartupDependencies{})
	selectionCalls := 0
	orchestrator.selectDatabaseFn = func() (tui.DatabaseOption, startupPath, error) {
		if orchestrator.pendingRuntimeTarget != nil {
			selected := *orchestrator.pendingRuntimeTarget
			orchestrator.pendingRuntimeTarget = nil
			return selected, startupPathRuntimeSwitch, nil
		}
		selectionCalls++
		if selectionCalls == 1 {
			return selectedA, startupPathSelector, nil
		}
		t.Fatalf("unexpected selector call %d", selectionCalls)
		return tui.DatabaseOption{}, startupPathSelector, nil
	}
	connectCalls := 0
	orchestrator.connectDatabaseFn = func(got tui.DatabaseOption) (*sql.DB, error) {
		connectCalls++
		switch connectCalls {
		case 1:
			if got.ConnString != selectedA.ConnString {
				t.Fatalf("expected first connect for %q, got %q", selectedA.ConnString, got.ConnString)
			}
		case 2:
			if got.ConnString != selectedB.ConnString {
				t.Fatalf("expected switch connect for %q, got %q", selectedB.ConnString, got.ConnString)
			}
		default:
			t.Fatalf("unexpected connect call %d", connectCalls)
		}
		return &sql.DB{}, nil
	}
	runCalls := 0
	orchestrator.runRuntimeSessionFn = func(runtimeDeps tui.RuntimeRunDeps) (tui.RuntimeExitResult, error) {
		runCalls++
		if runtimeDeps.DatabaseSelector == nil {
			t.Fatal("expected runtime database selector deps")
		}
		switch runCalls {
		case 1:
			if runtimeDeps.DatabaseSelector.CurrentDatabase.ConnString != selectedA.ConnString {
				t.Fatalf("expected first runtime database %q, got %q", selectedA.ConnString, runtimeDeps.DatabaseSelector.CurrentDatabase.ConnString)
			}
			return tui.RuntimeExitResult{
				Action:       tui.RuntimeExitActionOpenDatabaseNext,
				NextDatabase: selectedB,
			}, nil
		case 2:
			if runtimeDeps.DatabaseSelector.CurrentDatabase.ConnString != selectedB.ConnString {
				t.Fatalf("expected reopened runtime database %q, got %q", selectedB.ConnString, runtimeDeps.DatabaseSelector.CurrentDatabase.ConnString)
			}
			return tui.RuntimeExitResult{Action: tui.RuntimeExitActionQuit}, nil
		default:
			t.Fatalf("unexpected runtime session call %d", runCalls)
			return tui.RuntimeExitResult{}, nil
		}
	}

	// Act
	err := orchestrator.run()

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if connectCalls != 2 {
		t.Fatalf("expected two connection attempts, got %d", connectCalls)
	}
	if selectionCalls != 1 {
		t.Fatalf("expected one selector-origin selection, got %d", selectionCalls)
	}
	if runCalls != 2 {
		t.Fatalf("expected two runtime session runs, got %d", runCalls)
	}
}

func TestRuntimeStartupOrchestratorSelectDatabase_UsesSelectorState(t *testing.T) {
	// Arrange
	expectedState := tui.SelectorLaunchState{
		StatusMessage:    "Connection failed for \"analytics\": ping failed",
		PreferConnString: "/tmp/analytics.sqlite",
		AdditionalOptions: []tui.DatabaseOption{
			{
				Name:       "/tmp/direct.sqlite",
				ConnString: "/tmp/direct.sqlite",
				Source:     tui.DatabaseOptionSourceCLI,
			},
		},
	}
	expectedOption := tui.DatabaseOption{
		Name:       "analytics",
		ConnString: "/tmp/analytics.sqlite",
		Source:     tui.DatabaseOptionSourceConfig,
	}
	store := &fakeStartupSelectionConfigStore{
		entries: []port.ConfigEntry{
			{Name: "local", DBPath: "/tmp/local.sqlite"},
		},
	}
	orchestrator := newRuntimeStartupOrchestrator(startupOptions{}, runtimeStartupDependencies{
		loadDatabaseSelectorState: usecase.NewLoadDatabaseSelectorState(store),
	})
	orchestrator.selectorState = expectedState
	withStartupRuntimeHooks(t, startupRuntimeHookOverrides{
		selectDatabaseWithState: func(
			ctx context.Context,
			loadDatabaseSelectorState *usecase.LoadDatabaseSelectorState,
			createConfiguredDatabase *usecase.CreateConfiguredDatabase,
			updateConfiguredDatabase *usecase.UpdateConfiguredDatabase,
			deleteConfiguredDatabase *usecase.DeleteConfiguredDatabase,
			state tui.SelectorLaunchState,
		) (tui.DatabaseOption, error) {
			if loadDatabaseSelectorState == nil {
				t.Fatal("expected loadDatabaseSelectorState to be wired")
			}
			if !reflect.DeepEqual(state, expectedState) {
				t.Fatalf("expected selector state %+v, got %+v", expectedState, state)
			}
			return expectedOption, nil
		},
	})

	// Act
	selected, path, err := orchestrator.selectDatabase()

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if path != startupPathSelector {
		t.Fatalf("expected startup path %v, got %v", startupPathSelector, path)
	}
	if selected != expectedOption {
		t.Fatalf("expected selected option %+v, got %+v", expectedOption, selected)
	}
}

func TestRuntimeStartupOrchestratorRunSelectedDatabase_ReturnsPresentedFailureForDirectLaunchConnectError(t *testing.T) {
	t.Parallel()

	// Arrange
	orchestrator := newRuntimeStartupOrchestrator(startupOptions{}, runtimeStartupDependencies{})
	orchestrator.connectDatabaseFn = func(tui.DatabaseOption) (*sql.DB, error) {
		return nil, errors.New("database file does not exist")
	}
	selected := tui.DatabaseOption{
		Name:       "/tmp/missing.sqlite",
		ConnString: "/tmp/missing.sqlite",
	}

	// Act
	shouldContinue, err := orchestrator.runSelectedDatabase(selected, startupPathDirectLaunch)

	// Assert
	if shouldContinue {
		t.Fatal("expected direct-launch connection failure to stop startup")
	}
	if err == nil {
		t.Fatal("expected presented startup failure, got nil")
	}
	if failure := classifyStartupFailure(err); failure.stderrOutput != buildDirectLaunchFailureMessage(selected.ConnString, "database file does not exist") {
		t.Fatalf("expected direct-launch failure output, got %q", failure.stderrOutput)
	}
}

func TestRuntimeStartupOrchestratorRunSelectedDatabase_SetsSelectorStateOnSelectorConnectError(t *testing.T) {
	t.Parallel()

	// Arrange
	sessionOption := tui.DatabaseOption{
		Name:       "/tmp/session.sqlite",
		ConnString: "/tmp/session.sqlite",
		Source:     tui.DatabaseOptionSourceCLI,
	}
	orchestrator := newRuntimeStartupOrchestrator(startupOptions{}, runtimeStartupDependencies{})
	orchestrator.sessionScopedOptions = []tui.DatabaseOption{sessionOption}
	orchestrator.connectDatabaseFn = func(tui.DatabaseOption) (*sql.DB, error) {
		return nil, errors.New("ping failed")
	}
	selected := tui.DatabaseOption{
		Name:       "analytics",
		ConnString: "/tmp/analytics.sqlite",
	}

	// Act
	shouldContinue, err := orchestrator.runSelectedDatabase(selected, startupPathSelector)

	// Assert
	if err != nil {
		t.Fatalf("expected selector retry path, got %v", err)
	}
	if !shouldContinue {
		t.Fatal("expected selector connection failure to reopen selection flow")
	}
	if !strings.Contains(orchestrator.selectorState.StatusMessage, "Connection failed for \"analytics\": ping failed") {
		t.Fatalf("expected retry status message, got %q", orchestrator.selectorState.StatusMessage)
	}
	if orchestrator.selectorState.PreferConnString != selected.ConnString {
		t.Fatalf("expected preferred conn string %q, got %q", selected.ConnString, orchestrator.selectorState.PreferConnString)
	}
	if len(orchestrator.selectorState.AdditionalOptions) != 1 {
		t.Fatalf("expected one cloned session option, got %d", len(orchestrator.selectorState.AdditionalOptions))
	}
	if orchestrator.selectorState.AdditionalOptions[0].ConnString != sessionOption.ConnString {
		t.Fatalf("expected cloned session option conn string %q, got %q", sessionOption.ConnString, orchestrator.selectorState.AdditionalOptions[0].ConnString)
	}
}

func TestRuntimeStartupOrchestratorRunSelectedDatabase_ResetsSelectorStateBeforeRunningSession(t *testing.T) {
	t.Parallel()

	// Arrange
	orchestrator := newRuntimeStartupOrchestrator(startupOptions{}, runtimeStartupDependencies{})
	orchestrator.selectorState = tui.SelectorLaunchState{
		StatusMessage:    "stale retry state",
		PreferConnString: "/tmp/stale.sqlite",
	}
	orchestrator.connectDatabaseFn = func(tui.DatabaseOption) (*sql.DB, error) {
		return &sql.DB{}, nil
	}
	orchestrator.runRuntimeSessionFn = func(runtimeDeps tui.RuntimeRunDeps) (tui.RuntimeExitResult, error) {
		if runtimeDeps.DatabaseSelector == nil {
			t.Fatal("expected runtime run deps to include database selector deps")
		}
		if runtimeDeps.DatabaseSelector.CurrentDatabase.ConnString != "/tmp/analytics.sqlite" {
			t.Fatalf("expected current runtime database /tmp/analytics.sqlite, got %q", runtimeDeps.DatabaseSelector.CurrentDatabase.ConnString)
		}
		if orchestrator.selectorState.StatusMessage != "" {
			t.Fatalf("expected selector status to reset before session run, got %q", orchestrator.selectorState.StatusMessage)
		}
		if orchestrator.selectorState.PreferConnString != "" {
			t.Fatalf("expected selector preferred conn string to reset before session run, got %q", orchestrator.selectorState.PreferConnString)
		}
		if len(orchestrator.selectorState.AdditionalOptions) != 0 {
			t.Fatalf("expected selector state to reset before session run, got %+v", orchestrator.selectorState)
		}
		return tui.RuntimeExitResult{Action: tui.RuntimeExitActionQuit}, nil
	}

	// Act
	shouldContinue, err := orchestrator.runSelectedDatabase(
		tui.DatabaseOption{Name: "analytics", ConnString: "/tmp/analytics.sqlite"},
		startupPathSelector,
	)

	// Assert
	if err != nil {
		t.Fatalf("expected successful runtime session, got %v", err)
	}
	if shouldContinue {
		t.Fatal("expected successful runtime session to finish startup loop")
	}
}

func TestRuntimeStartupOrchestratorRunSelectedDatabase_ReturnsPresentedFailureForRuntimeSessionError(t *testing.T) {
	t.Parallel()

	// Arrange
	orchestrator := newRuntimeStartupOrchestrator(startupOptions{}, runtimeStartupDependencies{})
	orchestrator.connectDatabaseFn = func(tui.DatabaseOption) (*sql.DB, error) {
		return &sql.DB{}, nil
	}
	orchestrator.runRuntimeSessionFn = func(tui.RuntimeRunDeps) (tui.RuntimeExitResult, error) {
		return tui.RuntimeExitResult{}, errors.New("runtime exploded")
	}

	// Act
	shouldContinue, err := orchestrator.runSelectedDatabase(
		tui.DatabaseOption{Name: "analytics", ConnString: "/tmp/analytics.sqlite"},
		startupPathSelector,
	)

	// Assert
	if shouldContinue {
		t.Fatal("expected runtime error to stop startup")
	}
	if err == nil {
		t.Fatal("expected presented startup failure, got nil")
	}
	if failure := classifyStartupFailure(err); failure.stderrOutput != "application error: runtime exploded" {
		t.Fatalf("expected runtime failure output, got %q", failure.stderrOutput)
	}
}

func TestRunRuntimeSession_DelegatesToTUI(t *testing.T) {
	cases := []struct {
		name    string
		result  tui.RuntimeExitResult
		runErr  error
		wantErr bool
	}{
		{
			name:    "propagates run error",
			runErr:  errors.New("tui failed"),
			wantErr: true,
		},
		{
			name: "returns exit result from tui",
			result: tui.RuntimeExitResult{
				Action: tui.RuntimeExitActionOpenDatabaseNext,
				NextDatabase: tui.DatabaseOption{
					Name:       "analytics",
					ConnString: "/tmp/analytics.sqlite",
					Source:     tui.DatabaseOptionSourceCLI,
				},
			},
			wantErr: false,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			withStartupRuntimeHooks(t, startupRuntimeHookOverrides{
				runTUI: func(context.Context, tui.RuntimeRunDeps) (tui.RuntimeExitResult, error) {
					return tc.result, tc.runErr
				},
			})

			// Act
			result, err := runRuntimeSession(tui.RuntimeRunDeps{})

			// Assert
			if tc.wantErr {
				if !errors.Is(err, tc.runErr) {
					t.Fatalf("expected run error %v, got %v", tc.runErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("expected no runtime error, got %v", err)
			}
			if result != tc.result {
				t.Fatalf("expected runtime exit result %+v, got %+v", tc.result, result)
			}
		})
	}
}

func TestRuntimeStartupOrchestratorOpenRuntimeRunDeps_ExposesCurrentDatabaseAndAdditionalOptions(t *testing.T) {
	// Arrange
	selectedA := tui.DatabaseOption{
		Name:       "primary",
		ConnString: "/tmp/primary.sqlite",
		Source:     tui.DatabaseOptionSourceConfig,
	}
	currentClosed := false
	nextClosed := false
	dbA := &sql.DB{}
	orchestrator := newRuntimeStartupOrchestrator(startupOptions{}, runtimeStartupDependencies{})
	orchestrator.sessionScopedOptions = []tui.DatabaseOption{
		{
			Name:       "/tmp/direct.sqlite",
			ConnString: "/tmp/direct.sqlite",
			Source:     tui.DatabaseOptionSourceCLI,
		},
	}
	connectCalls := 0
	orchestrator.connectDatabaseFn = func(got tui.DatabaseOption) (*sql.DB, error) {
		connectCalls++
		if got.ConnString != selectedA.ConnString {
			t.Fatalf("expected connection target %q, got %q", selectedA.ConnString, got.ConnString)
		}
		return dbA, nil
	}
	withStartupRuntimeHooks(t, startupRuntimeHookOverrides{
		closeDatabase: func(db *sql.DB) error {
			switch db {
			case dbA:
				currentClosed = true
			default:
				t.Fatalf("unexpected close target %p", db)
			}
			return nil
		},
	})

	// Act
	runtimeDeps, err := orchestrator.openRuntimeRunDeps(selectedA)
	if err != nil {
		t.Fatalf("expected initial runtime deps, got %v", err)
	}

	// Assert
	if connectCalls != 1 {
		t.Fatalf("expected one runtime connection attempt, got %d", connectCalls)
	}
	if currentClosed {
		t.Fatal("expected current database to stay open until explicit cleanup")
	}
	if runtimeDeps.DatabaseSelector == nil {
		t.Fatal("expected runtime deps to include selector dependencies")
	}
	if runtimeDeps.DatabaseSelector.CurrentDatabase.ConnString != selectedA.ConnString {
		t.Fatalf("expected runtime database %q, got %q", selectedA.ConnString, runtimeDeps.DatabaseSelector.CurrentDatabase.ConnString)
	}
	if len(runtimeDeps.DatabaseSelector.AdditionalOptions) != 1 {
		t.Fatalf("expected one additional option, got %d", len(runtimeDeps.DatabaseSelector.AdditionalOptions))
	}
	if runtimeDeps.DatabaseSelector.AdditionalOptions[0].ConnString != "/tmp/direct.sqlite" {
		t.Fatalf("expected CLI additional option /tmp/direct.sqlite, got %q", runtimeDeps.DatabaseSelector.AdditionalOptions[0].ConnString)
	}
	runtimeDeps.Close()
	if !currentClosed {
		t.Fatal("expected current database close after explicit runtime deps cleanup")
	}
	if nextClosed {
		t.Fatal("expected no secondary database to be closed")
	}
}

func TestRuntimeStartupOrchestratorRun_RuntimeInitiatedCLIReopenFailureReturnsToSelectorWithTrackedOption(t *testing.T) {
	// Arrange
	selectedCLI := tui.DatabaseOption{
		Name:       "/tmp/runtime-cli.sqlite",
		ConnString: "/tmp/runtime-cli.sqlite",
		Source:     tui.DatabaseOptionSourceCLI,
	}
	orchestrator := newRuntimeStartupOrchestrator(startupOptions{}, runtimeStartupDependencies{})
	orchestrator.selectDatabaseFn = orchestrator.selectDatabase
	orchestrator.connectDatabaseFn = func(got tui.DatabaseOption) (*sql.DB, error) {
		if got.ConnString != selectedCLI.ConnString {
			t.Fatalf("expected runtime reopen target %q, got %q", selectedCLI.ConnString, got.ConnString)
		}
		return nil, errors.New("ping failed")
	}
	orchestrator.runRuntimeSessionFn = func(tui.RuntimeRunDeps) (tui.RuntimeExitResult, error) {
		return tui.RuntimeExitResult{
			Action:       tui.RuntimeExitActionOpenDatabaseNext,
			NextDatabase: selectedCLI,
		}, nil
	}
	orchestrator.pendingRuntimeTarget = &selectedCLI

	// Act
	shouldContinue, err := orchestrator.runSelectedDatabase(selectedCLI, startupPathRuntimeSwitch)

	// Assert
	if err != nil {
		t.Fatalf("expected selector retry after failed runtime reopen, got %v", err)
	}
	if !shouldContinue {
		t.Fatal("expected runtime reopen failure to return to selector")
	}
	if len(orchestrator.selectorState.AdditionalOptions) != 0 {
		t.Fatalf("expected selector state to be populated by outer loop tracking, got %+v", orchestrator.selectorState.AdditionalOptions)
	}
	orchestrator.sessionScopedOptions = trackSessionScopedCLIOption(orchestrator.sessionScopedOptions, selectedCLI)
	if len(orchestrator.sessionScopedOptions) != 1 {
		t.Fatalf("expected orchestrator to retain one CLI session option, got %d", len(orchestrator.sessionScopedOptions))
	}
	if orchestrator.sessionScopedOptions[0].ConnString != selectedCLI.ConnString {
		t.Fatalf("expected orchestrator to retain CLI option %q, got %q", selectedCLI.ConnString, orchestrator.sessionScopedOptions[0].ConnString)
	}
	if orchestrator.selectorState.PreferConnString != selectedCLI.ConnString {
		t.Fatalf("expected selector preferred conn string %q, got %q", selectedCLI.ConnString, orchestrator.selectorState.PreferConnString)
	}
}

func TestRuntimeStartupOrchestratorOpenRuntimeRunDeps_CloseLogsFailure(t *testing.T) {
	// Arrange
	logged := ""
	db := &sql.DB{}
	orchestrator := newRuntimeStartupOrchestrator(startupOptions{}, runtimeStartupDependencies{})
	orchestrator.connectDatabaseFn = func(tui.DatabaseOption) (*sql.DB, error) {
		return db, nil
	}
	withStartupRuntimeHooks(t, startupRuntimeHookOverrides{
		closeDatabase: func(got *sql.DB) error {
			if got != db {
				t.Fatalf("expected close on db %p, got %p", db, got)
			}
			return errors.New("close failed")
		},
		logPrintf: func(format string, args ...any) {
			logged = fmt.Sprintf(format, args...)
		},
	})

	// Act
	runtimeDeps, err := orchestrator.openRuntimeRunDeps(tui.DatabaseOption{
		Name:       "analytics",
		ConnString: "/tmp/analytics.sqlite",
	})
	if err != nil {
		t.Fatalf("expected runtime deps, got %v", err)
	}
	runtimeDeps.Close()

	// Assert
	if !strings.Contains(logged, "failed to close database: close failed") {
		t.Fatalf("expected close failure to be logged, got %q", logged)
	}
}

func TestConnectSelectedDatabase_OpensReachableSQLiteFile(t *testing.T) {
	t.Parallel()

	// Arrange
	file, err := os.CreateTemp("", "dbc-startup-runtime-*.sqlite")
	if err != nil {
		t.Fatalf("expected sqlite file to be created, got %v", err)
	}
	t.Cleanup(func() {
		if removeErr := os.Remove(file.Name()); removeErr != nil && !os.IsNotExist(removeErr) {
			t.Fatalf("expected temp sqlite file removal to succeed, got %v", removeErr)
		}
	})
	if err := file.Close(); err != nil {
		t.Fatalf("expected sqlite file close to succeed, got %v", err)
	}
	dbPath := file.Name()

	// Act
	db, err := connectSelectedDatabase(tui.DatabaseOption{
		Name:       "analytics",
		ConnString: dbPath,
	})

	// Assert
	if err != nil {
		t.Fatalf("expected sqlite connection to open, got %v", err)
	}
	if db == nil {
		t.Fatal("expected sqlite connection handle, got nil")
	}
	if err := db.Close(); err != nil {
		t.Fatalf("expected sqlite connection close to succeed, got %v", err)
	}
}
