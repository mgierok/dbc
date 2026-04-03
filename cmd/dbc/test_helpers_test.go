package main

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	"github.com/mgierok/dbc/internal/application/usecase"
	"github.com/mgierok/dbc/internal/interfaces/tui"
)

type startupRuntimeHookOverrides struct {
	newDependencies         func() (runtimeStartupDependencies, error)
	newOrchestrator         func(startupOptions, runtimeStartupDependencies) *runtimeStartupOrchestrator
	selectDatabaseWithState func(
		context.Context,
		*usecase.LoadDatabaseSelectorState,
		*usecase.CreateConfiguredDatabase,
		*usecase.UpdateConfiguredDatabase,
		*usecase.DeleteConfiguredDatabase,
		tui.SelectorLaunchState,
	) (tui.DatabaseOption, error)
	runTUI        func(context.Context, tui.RuntimeRunDeps) (tui.RuntimeExitResult, error)
	closeDatabase func(*sql.DB) error
	logPrintf     func(string, ...any)
}

func withStartupRuntimeHooks(t *testing.T, overrides startupRuntimeHookOverrides) {
	t.Helper()

	oldDepsFactory := newRuntimeStartupDependenciesFn
	oldOrchestratorFactory := newRuntimeStartupOrchestratorFn
	oldSelectDatabase := selectDatabaseWithStateFn
	oldTUIRun := tuiRunFn
	oldCloseDatabase := closeDatabaseFn
	oldLogPrintf := logPrintfFn

	t.Cleanup(func() {
		newRuntimeStartupDependenciesFn = oldDepsFactory
		newRuntimeStartupOrchestratorFn = oldOrchestratorFactory
		selectDatabaseWithStateFn = oldSelectDatabase
		tuiRunFn = oldTUIRun
		closeDatabaseFn = oldCloseDatabase
		logPrintfFn = oldLogPrintf
	})

	if overrides.newDependencies != nil {
		newRuntimeStartupDependenciesFn = overrides.newDependencies
	}
	if overrides.newOrchestrator != nil {
		newRuntimeStartupOrchestratorFn = overrides.newOrchestrator
	}
	if overrides.selectDatabaseWithState != nil {
		selectDatabaseWithStateFn = overrides.selectDatabaseWithState
	}
	if overrides.runTUI != nil {
		tuiRunFn = overrides.runTUI
	}
	if overrides.closeDatabase != nil {
		closeDatabaseFn = overrides.closeDatabase
	}
	if overrides.logPrintf != nil {
		logPrintfFn = overrides.logPrintf
	}
}

func assertUsageError(t *testing.T, err error, requiredTokens ...string) startupUsageError {
	t.Helper()

	if err == nil {
		t.Fatal("expected startup usage error, got nil")
	}

	var usageErr startupUsageError
	if !errors.As(err, &usageErr) {
		t.Fatalf("expected startupUsageError, got %T", err)
	}

	for _, token := range requiredTokens {
		if !strings.Contains(err.Error(), token) {
			t.Fatalf("expected error %q to include token %q", err.Error(), token)
		}
	}

	return usageErr
}

func assertErrorContains(t *testing.T, err error, requiredTokens ...string) {
	t.Helper()

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	for _, token := range requiredTokens {
		if !strings.Contains(err.Error(), token) {
			t.Fatalf("expected error %q to include token %q", err.Error(), token)
		}
	}
}

func assertDatabaseOption(t *testing.T, got tui.DatabaseOption, want tui.DatabaseOption) {
	t.Helper()

	if got != want {
		t.Fatalf("expected database option %+v, got %+v", want, got)
	}
}
