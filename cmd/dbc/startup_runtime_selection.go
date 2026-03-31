package main

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/mgierok/dbc/internal/application/usecase"
	"github.com/mgierok/dbc/internal/infrastructure/engine"
	"github.com/mgierok/dbc/internal/interfaces/tui"
	"github.com/mgierok/dbc/internal/sqliteidentity"
)

type startupPath int

const (
	startupPathSelector startupPath = iota
	startupPathDirectLaunch
	startupPathRuntimeSwitch
)

type startupSelectionStrategy interface {
	resolve(
		listConfiguredDatabases func() ([]tui.DatabaseOption, error),
		selectDatabase func() (tui.DatabaseOption, error),
	) (tui.DatabaseOption, startupPath, error)
}

type selectorStartupSelectionStrategy struct{}

func (selectorStartupSelectionStrategy) resolve(
	listConfiguredDatabases func() ([]tui.DatabaseOption, error),
	selectDatabase func() (tui.DatabaseOption, error),
) (tui.DatabaseOption, startupPath, error) {
	return resolveStartupSelection(startupOptions{}, listConfiguredDatabases, selectDatabase)
}

type directLaunchStartupSelectionStrategy struct {
	options startupOptions
}

func (s directLaunchStartupSelectionStrategy) resolve(
	listConfiguredDatabases func() ([]tui.DatabaseOption, error),
	selectDatabase func() (tui.DatabaseOption, error),
) (tui.DatabaseOption, startupPath, error) {
	return resolveStartupSelection(s.options, listConfiguredDatabases, selectDatabase)
}

func newStartupSelectionStrategy(options startupOptions, directLaunchPending bool) startupSelectionStrategy {
	if directLaunchPending && strings.TrimSpace(options.directLaunchConnString) != "" {
		return directLaunchStartupSelectionStrategy{options: options}
	}

	return selectorStartupSelectionStrategy{}
}

func resolveStartupSelection(
	options startupOptions,
	listConfiguredDatabases func() ([]tui.DatabaseOption, error),
	selectDatabase func() (tui.DatabaseOption, error),
) (tui.DatabaseOption, startupPath, error) {
	if options.directLaunchConnString != "" {
		configuredOptions, err := listConfiguredDatabases()
		if err != nil {
			return tui.DatabaseOption{}, startupPathDirectLaunch, err
		}
		target, err := usecase.NewRuntimeDatabaseTargetResolver().Resolve(
			usecase.RuntimeDatabaseOption{},
			runtimeDatabaseOptionsFromSelectorOptions(configuredOptions),
			runtimeDatabaseRequestOptionFromConnString(options.directLaunchConnString),
		)
		if err != nil {
			return tui.DatabaseOption{}, startupPathDirectLaunch, err
		}
		return selectorOptionFromRuntimeDatabaseOption(target.Option), startupPathDirectLaunch, nil
	}

	selected, err := selectDatabase()
	if err != nil {
		return tui.DatabaseOption{}, startupPathSelector, err
	}

	return selected, startupPathSelector, nil
}

func listConfiguredDatabaseOptions(ctx context.Context, listConfiguredDatabases *usecase.ListConfiguredDatabases) ([]tui.DatabaseOption, error) {
	entries, err := listConfiguredDatabases.Execute(ctx)
	if err != nil {
		return nil, err
	}

	options := make([]tui.DatabaseOption, len(entries))
	for i, entry := range entries {
		options[i] = tui.DatabaseOption{
			Name:       entry.Name,
			ConnString: entry.Path,
			Source:     tui.DatabaseOptionSourceConfig,
		}
	}
	return options, nil
}

func normalizeSQLiteConnectionIdentity(connString string) string {
	return sqliteidentity.Normalize(connString)
}

func sqliteConnectionIdentityEqual(left string, right string) bool {
	return sqliteidentity.Equivalent(left, right)
}

func connectSelectedDatabase(selected tui.DatabaseOption) (*sql.DB, error) {
	return engine.OpenSQLiteDatabase(context.Background(), selected.ConnString)
}

func buildConnectionFailureStatus(selected tui.DatabaseOption, reason string) string {
	return fmt.Sprintf(
		"Connection failed for %q: %s. Choose another database or edit selected entry.",
		selected.Name,
		reason,
	)
}

func buildDirectLaunchFailureMessage(connString, reason string) string {
	return fmt.Sprintf(
		"Cannot start DBC with direct launch target %q: %s. Check that the SQLite path is valid and reachable, then retry with -d/--database.",
		connString,
		reason,
	)
}

func trackSessionScopedCLIOption(existing []tui.DatabaseOption, selected tui.DatabaseOption) []tui.DatabaseOption {
	if selected.Source != tui.DatabaseOptionSourceCLI {
		return existing
	}

	normalizedSelected := normalizeSQLiteConnectionIdentity(selected.ConnString)
	if normalizedSelected == "" {
		return existing
	}
	for _, option := range existing {
		normalizedExisting := normalizeSQLiteConnectionIdentity(option.ConnString)
		if normalizedExisting == "" {
			continue
		}
		if sqliteConnectionIdentityEqual(normalizedExisting, normalizedSelected) {
			return existing
		}
	}

	sessionOption := selected
	sessionOption.Source = tui.DatabaseOptionSourceCLI
	return append(existing, sessionOption)
}

func trackSessionScopedDirectLaunchOption(existing []tui.DatabaseOption, _ startupPath, selected tui.DatabaseOption) []tui.DatabaseOption {
	return trackSessionScopedCLIOption(existing, selected)
}

func cloneDatabaseOptions(options []tui.DatabaseOption) []tui.DatabaseOption {
	if len(options) == 0 {
		return nil
	}
	cloned := make([]tui.DatabaseOption, len(options))
	copy(cloned, options)
	return cloned
}

func runtimeDatabaseOptionsFromSelectorOptions(options []tui.DatabaseOption) []usecase.RuntimeDatabaseOption {
	if len(options) == 0 {
		return nil
	}
	converted := make([]usecase.RuntimeDatabaseOption, len(options))
	for i, option := range options {
		source := usecase.RuntimeDatabaseOptionSourceConfig
		if option.Source == tui.DatabaseOptionSourceCLI {
			source = usecase.RuntimeDatabaseOptionSourceCLI
		}
		converted[i] = usecase.RuntimeDatabaseOption{
			Name:       option.Name,
			ConnString: option.ConnString,
			Source:     source,
		}
	}
	return converted
}

func selectorOptionFromRuntimeDatabaseOption(option usecase.RuntimeDatabaseOption) tui.DatabaseOption {
	source := tui.DatabaseOptionSourceConfig
	if option.Source == usecase.RuntimeDatabaseOptionSourceCLI {
		source = tui.DatabaseOptionSourceCLI
	}
	return tui.DatabaseOption{
		Name:       option.Name,
		ConnString: option.ConnString,
		Source:     source,
	}
}

func runtimeDatabaseRequestOptionFromConnString(connString string) usecase.RuntimeDatabaseOption {
	return usecase.RuntimeDatabaseOption{
		Name:       connString,
		ConnString: connString,
		Source:     usecase.RuntimeDatabaseOptionSourceCLI,
	}
}
