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
	if orchestrator.openRuntimeRunDepsFn == nil {
		t.Fatal("expected openRuntimeRunDepsFn to be initialized")
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

func TestRuntimeStartupOrchestratorRun_UsesSingleStartupSelectionAndProvidesRuntimeSwitcher(t *testing.T) {
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
	selectCalls := 0
	orchestrator.selectDatabaseFn = func() (tui.DatabaseOption, startupPath, error) {
		selectCalls++
		switch selectCalls {
		case 1:
			return selectedA, startupPathSelector, nil
		default:
			t.Fatalf("unexpected selection call %d", selectCalls)
			return tui.DatabaseOption{}, startupPathSelector, nil
		}
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
	var switchedRuntimeDeps tui.RuntimeRunDeps
	orchestrator.runRuntimeSessionFn = func(runtimeDeps tui.RuntimeRunDeps, runtimeSession *tui.RuntimeSessionState) error {
		runCalls++
		if runtimeSession == nil {
			t.Fatal("expected shared runtime session state")
		}
		if runtimeDeps.DatabaseSelector == nil {
			t.Fatal("expected runtime database selector deps")
		}
		if runtimeDeps.DatabaseSelector.CurrentDatabase.ConnString != selectedA.ConnString {
			t.Fatalf("expected current runtime database %q, got %q", selectedA.ConnString, runtimeDeps.DatabaseSelector.CurrentDatabase.ConnString)
		}
		if runtimeDeps.DatabaseSelector.SwitchDatabase == nil {
			t.Fatal("expected runtime switcher in runtime deps")
		}
		runtimeSession.RecordsPageLimit = 55
		var err error
		switchedRuntimeDeps, err = runtimeDeps.DatabaseSelector.SwitchDatabase.Switch(context.Background(), selectedB)
		if err != nil {
			t.Fatalf("expected runtime switcher to build new runtime deps, got %v", err)
		}
		return nil
	}

	// Act
	err := orchestrator.run()

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if selectCalls != 1 {
		t.Fatalf("expected one startup selector call, got %d", selectCalls)
	}
	if connectCalls != 2 {
		t.Fatalf("expected two connection attempts, got %d", connectCalls)
	}
	if runCalls != 1 {
		t.Fatalf("expected one runtime session run, got %d", runCalls)
	}
	if switchedRuntimeDeps.DatabaseSelector == nil {
		t.Fatal("expected switched runtime deps to include selector dependencies")
	}
	if switchedRuntimeDeps.DatabaseSelector.CurrentDatabase.ConnString != selectedB.ConnString {
		t.Fatalf("expected switched runtime database %q, got %q", selectedB.ConnString, switchedRuntimeDeps.DatabaseSelector.CurrentDatabase.ConnString)
	}
	if orchestrator.runtimeSessionState.RecordsPageLimit != 55 {
		t.Fatalf("expected runtime session state to survive inside orchestrator, got %d", orchestrator.runtimeSessionState.RecordsPageLimit)
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
	orchestrator.runRuntimeSessionFn = func(runtimeDeps tui.RuntimeRunDeps, runtimeSession *tui.RuntimeSessionState) error {
		if runtimeSession == nil {
			t.Fatal("expected runtime session state to be provided")
		}
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
	orchestrator.runRuntimeSessionFn = func(tui.RuntimeRunDeps, *tui.RuntimeSessionState) error {
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

func TestRunRuntimeSession_PropagatesRunError(t *testing.T) {
	restore := snapshotStartupRuntimeTestHooks()
	defer restore()

	// Arrange
	expectedErr := errors.New("tui failed")
	tuiRunFn = func(_ context.Context, _ tui.RuntimeRunDeps, _ *tui.RuntimeSessionState) error {
		return expectedErr
	}

	// Act
	err := runRuntimeSession(tui.RuntimeRunDeps{}, nil)

	// Assert
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected run error %v, got %v", expectedErr, err)
	}
}

func TestRunRuntimeSession_PassesRuntimeSessionStateToTUI(t *testing.T) {
	restore := snapshotStartupRuntimeTestHooks()
	defer restore()

	// Arrange
	runtimeSession := &tui.RuntimeSessionState{RecordsPageLimit: 33}
	tuiRunFn = func(_ context.Context, _ tui.RuntimeRunDeps, gotRuntimeSession *tui.RuntimeSessionState) error {
		if gotRuntimeSession != runtimeSession {
			t.Fatalf("expected runtime session pointer %p, got %p", runtimeSession, gotRuntimeSession)
		}
		if gotRuntimeSession.RecordsPageLimit != 33 {
			t.Fatalf("expected runtime session limit 33, got %d", gotRuntimeSession.RecordsPageLimit)
		}
		return nil
	}

	// Act
	err := runRuntimeSession(tui.RuntimeRunDeps{}, runtimeSession)

	// Assert
	if err != nil {
		t.Fatalf("expected no runtime error, got %v", err)
	}
}

func TestRuntimeStartupOrchestratorOpenRuntimeRunDeps_SwitcherBuildsNewBundleWithoutClosingCurrentDB(t *testing.T) {
	restore := snapshotStartupRuntimeTestHooks()
	defer restore()

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
	currentClosed := false
	nextClosed := false
	dbA := &sql.DB{}
	dbB := &sql.DB{}
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
		switch got.ConnString {
		case selectedA.ConnString:
			return dbA, nil
		case selectedB.ConnString:
			return dbB, nil
		default:
			t.Fatalf("unexpected connection target %q", got.ConnString)
			return nil, nil
		}
	}
	closeDatabaseFn = func(db *sql.DB) error {
		switch db {
		case dbA:
			currentClosed = true
		case dbB:
			nextClosed = true
		default:
			t.Fatalf("unexpected close target %p", db)
		}
		return nil
	}

	// Act
	runtimeDeps, err := orchestrator.openRuntimeRunDeps(selectedA)
	if err != nil {
		t.Fatalf("expected initial runtime deps, got %v", err)
	}
	switchedRuntimeDeps, err := runtimeDeps.DatabaseSelector.SwitchDatabase.Switch(context.Background(), selectedB)
	if err != nil {
		t.Fatalf("expected switched runtime deps, got %v", err)
	}

	// Assert
	if connectCalls != 2 {
		t.Fatalf("expected two runtime bundle connection attempts, got %d", connectCalls)
	}
	if currentClosed {
		t.Fatal("expected current database to stay open while building replacement runtime deps")
	}
	if switchedRuntimeDeps.DatabaseSelector == nil {
		t.Fatal("expected switched runtime deps to include selector deps")
	}
	if switchedRuntimeDeps.DatabaseSelector.CurrentDatabase.ConnString != selectedB.ConnString {
		t.Fatalf("expected switched runtime database %q, got %q", selectedB.ConnString, switchedRuntimeDeps.DatabaseSelector.CurrentDatabase.ConnString)
	}
	if len(switchedRuntimeDeps.DatabaseSelector.AdditionalOptions) != 1 {
		t.Fatalf("expected session-scoped options to carry into switched runtime deps, got %d", len(switchedRuntimeDeps.DatabaseSelector.AdditionalOptions))
	}
	runtimeDeps.Close()
	if !currentClosed {
		t.Fatal("expected current database close after explicit runtime deps cleanup")
	}
	if nextClosed {
		t.Fatal("expected switched runtime database to remain open until its own cleanup")
	}
}

func TestRuntimeStartupOrchestratorOpenRuntimeRunDeps_CloseLogsFailure(t *testing.T) {
	restore := snapshotStartupRuntimeTestHooks()
	defer restore()

	// Arrange
	logged := ""
	db := &sql.DB{}
	orchestrator := newRuntimeStartupOrchestrator(startupOptions{}, runtimeStartupDependencies{})
	orchestrator.connectDatabaseFn = func(tui.DatabaseOption) (*sql.DB, error) {
		return db, nil
	}
	closeDatabaseFn = func(got *sql.DB) error {
		if got != db {
			t.Fatalf("expected close on db %p, got %p", db, got)
		}
		return errors.New("close failed")
	}
	logPrintfFn = func(format string, args ...any) {
		logged = fmt.Sprintf(format, args...)
	}

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
