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
	if deps.getActiveConfigPath == nil {
		t.Fatal("expected getActiveConfigPath use case to be initialized")
	}

	activePath, err := deps.getActiveConfigPath.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected active config path to resolve, got %v", err)
	}
	if activePath != expectedPath {
		t.Fatalf("expected active config path %q, got %q", expectedPath, activePath)
	}
}

func TestNewRuntimeStartupOrchestrator_SetsDirectLaunchPendingFromOptions(t *testing.T) {
	t.Parallel()

	// Arrange
	options := startupOptions{directLaunchConnString: "/tmp/direct.sqlite"}

	// Act
	orchestrator := newRuntimeStartupOrchestrator(options, runtimeStartupDependencies{})

	// Assert
	if !orchestrator.directLaunchPending {
		t.Fatal("expected direct launch to start in pending state")
	}
	if orchestrator.selectDatabaseFn == nil {
		t.Fatal("expected selectDatabaseFn to be initialized")
	}
	if orchestrator.runSelectedDatabaseFn == nil {
		t.Fatal("expected runSelectedDatabaseFn to be initialized")
	}
	if orchestrator.connectDatabaseFn == nil {
		t.Fatal("expected connectDatabaseFn to be initialized")
	}
	if orchestrator.runRuntimeSessionFn == nil {
		t.Fatal("expected runRuntimeSessionFn to be initialized")
	}
}

func TestRunRuntimeStartup_WrapsDependencyResolutionFailure(t *testing.T) {
	restore := snapshotStartupRuntimeTestHooks()
	defer restore()

	// Arrange
	expectedErr := errors.New("config path unavailable")
	newRuntimeStartupDependenciesFn = func() (runtimeStartupDependencies, error) {
		return runtimeStartupDependencies{}, expectedErr
	}

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
	restore := snapshotStartupRuntimeTestHooks()
	defer restore()

	// Arrange
	options := startupOptions{directLaunchConnString: "/tmp/direct.sqlite"}
	expectedDeps := runtimeStartupDependencies{}
	orchestratorCalled := false
	newRuntimeStartupDependenciesFn = func() (runtimeStartupDependencies, error) {
		return expectedDeps, nil
	}
	newRuntimeStartupOrchestratorFn = func(gotOptions startupOptions, gotDeps runtimeStartupDependencies) *runtimeStartupOrchestrator {
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
	}

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

func TestRuntimeStartupOrchestratorRun_ReturnsNilWhenSelectorIsCanceled(t *testing.T) {
	t.Parallel()

	// Arrange
	orchestrator := &runtimeStartupOrchestrator{
		selectDatabaseFn: func() (tui.DatabaseOption, startupPath, error) {
			return tui.DatabaseOption{}, startupPathSelector, tui.ErrDatabaseSelectionCanceled
		},
		runSelectedDatabaseFn: func(tui.DatabaseOption, startupPath) (bool, error) {
			t.Fatal("expected selected database runner to stay unused after selector cancel")
			return false, nil
		},
	}

	// Act
	err := orchestrator.run()

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
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

func TestRuntimeStartupOrchestratorRun_TracksDirectLaunchOptionAcrossConfigRetry(t *testing.T) {
	t.Parallel()

	// Arrange
	selected := tui.DatabaseOption{
		Name:       "/tmp/direct.sqlite",
		ConnString: "/tmp/direct.sqlite",
		Source:     tui.DatabaseOptionSourceCLI,
	}
	orchestrator := newRuntimeStartupOrchestrator(
		startupOptions{directLaunchConnString: selected.ConnString},
		runtimeStartupDependencies{},
	)
	selectCalls := 0
	runCalls := 0
	orchestrator.selectDatabaseFn = func() (tui.DatabaseOption, startupPath, error) {
		selectCalls++
		switch selectCalls {
		case 1:
			if !orchestrator.directLaunchPending {
				t.Fatal("expected direct launch to remain pending before the first selection")
			}
			return selected, startupPathDirectLaunch, nil
		case 2:
			if orchestrator.directLaunchPending {
				t.Fatal("expected direct launch pending state to clear after first selection")
			}
			if orchestrator.selectorState.PreferConnString != selected.ConnString {
				t.Fatalf("expected selector retry to prefer %q, got %q", selected.ConnString, orchestrator.selectorState.PreferConnString)
			}
			if len(orchestrator.selectorState.AdditionalOptions) != 1 {
				t.Fatalf("expected one session-scoped option, got %d", len(orchestrator.selectorState.AdditionalOptions))
			}
			if orchestrator.selectorState.AdditionalOptions[0].ConnString != selected.ConnString {
				t.Fatalf("expected retry option conn string %q, got %q", selected.ConnString, orchestrator.selectorState.AdditionalOptions[0].ConnString)
			}
			return selected, startupPathSelector, nil
		default:
			t.Fatalf("unexpected selection call %d", selectCalls)
			return tui.DatabaseOption{}, startupPathSelector, nil
		}
	}
	orchestrator.connectDatabaseFn = func(tui.DatabaseOption) (*sql.DB, error) {
		return &sql.DB{}, nil
	}
	var firstRuntimeSession *tui.RuntimeSessionState
	orchestrator.runRuntimeSessionFn = func(_ *sql.DB, runtimeSession *tui.RuntimeSessionState) error {
		runCalls++
		if runtimeSession == nil {
			t.Fatal("expected shared runtime session state")
		}
		if runCalls == 1 {
			firstRuntimeSession = runtimeSession
			if runtimeSession.RecordsPageLimit != 0 {
				t.Fatalf("expected zero-value runtime session limit on first run, got %d", runtimeSession.RecordsPageLimit)
			}
			runtimeSession.RecordsPageLimit = 55
			return tui.ErrOpenConfigSelector
		}
		if runtimeSession != firstRuntimeSession {
			t.Fatal("expected runtime session state pointer to be reused across runs")
		}
		if runtimeSession.RecordsPageLimit != 55 {
			t.Fatalf("expected runtime session limit to survive config round-trip, got %d", runtimeSession.RecordsPageLimit)
		}
		return nil
	}

	// Act
	err := orchestrator.run()

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if selectCalls != 2 {
		t.Fatalf("expected two selection attempts, got %d", selectCalls)
	}
	if runCalls != 2 {
		t.Fatalf("expected two runtime session attempts, got %d", runCalls)
	}
}

func TestRuntimeStartupOrchestratorSelectDatabase_UsesSelectorState(t *testing.T) {
	restore := snapshotStartupRuntimeTestHooks()
	defer restore()

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
	orchestrator := newRuntimeStartupOrchestrator(startupOptions{}, runtimeStartupDependencies{})
	orchestrator.selectorState = expectedState
	selectDatabaseWithStateFn = func(
		ctx context.Context,
		listConfiguredDatabases *usecase.ListConfiguredDatabases,
		createConfiguredDatabase *usecase.CreateConfiguredDatabase,
		updateConfiguredDatabase *usecase.UpdateConfiguredDatabase,
		deleteConfiguredDatabase *usecase.DeleteConfiguredDatabase,
		getActiveConfigPath *usecase.GetActiveConfigPath,
		state tui.SelectorLaunchState,
	) (tui.DatabaseOption, error) {
		if !reflect.DeepEqual(state, expectedState) {
			t.Fatalf("expected selector state %+v, got %+v", expectedState, state)
		}
		return expectedOption, nil
	}

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
	orchestrator.runRuntimeSessionFn = func(_ *sql.DB, runtimeSession *tui.RuntimeSessionState) error {
		if runtimeSession == nil {
			t.Fatal("expected runtime session state to be provided")
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
		return nil
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
	orchestrator.runRuntimeSessionFn = func(*sql.DB, *tui.RuntimeSessionState) error {
		return errors.New("runtime exploded")
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

func TestRunRuntimeSession_PropagatesRunErrorAndClosesDatabase(t *testing.T) {
	restore := snapshotStartupRuntimeTestHooks()
	defer restore()

	// Arrange
	expectedErr := errors.New("tui failed")
	expectedDB := &sql.DB{}
	closed := false
	tuiRunFn = func(
		_ context.Context,
		_ *usecase.ListTables,
		_ *usecase.GetSchema,
		_ *usecase.ListRecords,
		_ *usecase.ListOperators,
		_ *usecase.SaveDatabaseChanges,
		_ *usecase.StagedChangesTranslator,
		_ *tui.RuntimeSessionState,
	) error {
		return expectedErr
	}
	closeDatabaseFn = func(db *sql.DB) error {
		if db != expectedDB {
			t.Fatalf("expected close on db %p, got %p", expectedDB, db)
		}
		closed = true
		return nil
	}

	// Act
	err := runRuntimeSession(expectedDB, nil)

	// Assert
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected run error %v, got %v", expectedErr, err)
	}
	if !closed {
		t.Fatal("expected database close after runtime session")
	}
}

func TestRunRuntimeSession_PassesRuntimeSessionStateToTUI(t *testing.T) {
	restore := snapshotStartupRuntimeTestHooks()
	defer restore()

	// Arrange
	runtimeSession := &tui.RuntimeSessionState{RecordsPageLimit: 33}
	tuiRunFn = func(
		_ context.Context,
		_ *usecase.ListTables,
		_ *usecase.GetSchema,
		_ *usecase.ListRecords,
		_ *usecase.ListOperators,
		_ *usecase.SaveDatabaseChanges,
		_ *usecase.StagedChangesTranslator,
		gotRuntimeSession *tui.RuntimeSessionState,
	) error {
		if gotRuntimeSession != runtimeSession {
			t.Fatalf("expected runtime session pointer %p, got %p", runtimeSession, gotRuntimeSession)
		}
		if gotRuntimeSession.RecordsPageLimit != 33 {
			t.Fatalf("expected runtime session limit 33, got %d", gotRuntimeSession.RecordsPageLimit)
		}
		return nil
	}
	closeDatabaseFn = func(*sql.DB) error {
		return nil
	}

	// Act
	err := runRuntimeSession(&sql.DB{}, runtimeSession)

	// Assert
	if err != nil {
		t.Fatalf("expected no runtime error, got %v", err)
	}
}

func TestRunRuntimeSession_LogsCloseFailure(t *testing.T) {
	restore := snapshotStartupRuntimeTestHooks()
	defer restore()

	// Arrange
	logged := ""
	tuiRunFn = func(
		_ context.Context,
		_ *usecase.ListTables,
		_ *usecase.GetSchema,
		_ *usecase.ListRecords,
		_ *usecase.ListOperators,
		_ *usecase.SaveDatabaseChanges,
		_ *usecase.StagedChangesTranslator,
		_ *tui.RuntimeSessionState,
	) error {
		return nil
	}
	closeDatabaseFn = func(*sql.DB) error {
		return errors.New("close failed")
	}
	logPrintfFn = func(format string, args ...any) {
		logged = fmt.Sprintf(format, args...)
	}

	// Act
	err := runRuntimeSession(&sql.DB{}, nil)

	// Assert
	if err != nil {
		t.Fatalf("expected no runtime error, got %v", err)
	}
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

func snapshotStartupRuntimeTestHooks() func() {
	oldDepsFactory := newRuntimeStartupDependenciesFn
	oldOrchestratorFactory := newRuntimeStartupOrchestratorFn
	oldSelectDatabase := selectDatabaseWithStateFn
	oldTUIRun := tuiRunFn
	oldCloseDatabase := closeDatabaseFn
	oldLogPrintf := logPrintfFn

	return func() {
		newRuntimeStartupDependenciesFn = oldDepsFactory
		newRuntimeStartupOrchestratorFn = oldOrchestratorFactory
		selectDatabaseWithStateFn = oldSelectDatabase
		tuiRunFn = oldTUIRun
		closeDatabaseFn = oldCloseDatabase
		logPrintfFn = oldLogPrintf
	}
}
