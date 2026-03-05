package main

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/mgierok/dbc/internal/application/usecase"
	"github.com/mgierok/dbc/internal/infrastructure/engine"
	"github.com/mgierok/dbc/internal/interfaces/tui"
)

type startupPath int

const (
	startupPathSelector startupPath = iota
	startupPathDirectLaunch
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
		directLaunchSelection := tui.DatabaseOption{
			Name:       options.directLaunchConnString,
			ConnString: options.directLaunchConnString,
			Source:     tui.DatabaseOptionSourceCLI,
		}
		configuredOptions, err := listConfiguredDatabases()
		if err != nil {
			return tui.DatabaseOption{}, startupPathDirectLaunch, err
		}
		if matched, ok := resolveConfiguredDirectLaunchIdentity(directLaunchSelection.ConnString, configuredOptions); ok {
			matched.Source = tui.DatabaseOptionSourceConfig
			return matched, startupPathDirectLaunch, nil
		}
		return directLaunchSelection, startupPathDirectLaunch, nil
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

func resolveConfiguredDirectLaunchIdentity(directLaunchConnString string, configuredOptions []tui.DatabaseOption) (tui.DatabaseOption, bool) {
	normalizedDirectLaunch := normalizeSQLiteConnectionIdentity(directLaunchConnString)
	if normalizedDirectLaunch == "" {
		return tui.DatabaseOption{}, false
	}

	for _, option := range configuredOptions {
		normalizedConfigured := normalizeSQLiteConnectionIdentity(option.ConnString)
		if normalizedConfigured == "" {
			continue
		}
		if sqliteConnectionIdentityEqual(normalizedDirectLaunch, normalizedConfigured) {
			return option, true
		}
	}

	return tui.DatabaseOption{}, false
}

func normalizeSQLiteConnectionIdentity(connString string) string {
	normalized := strings.TrimSpace(connString)
	if normalized == "" {
		return ""
	}

	normalized = filepath.Clean(normalized)
	if !filepath.IsAbs(normalized) {
		absPath, err := filepath.Abs(normalized)
		if err == nil {
			normalized = absPath
		}
	}
	return normalized
}

func sqliteConnectionIdentityEqual(left string, right string) bool {
	return left == right
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

func trackSessionScopedDirectLaunchOption(existing []tui.DatabaseOption, selectedStartupPath startupPath, selected tui.DatabaseOption) []tui.DatabaseOption {
	if selectedStartupPath != startupPathDirectLaunch || selected.Source != tui.DatabaseOptionSourceCLI {
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

func cloneDatabaseOptions(options []tui.DatabaseOption) []tui.DatabaseOption {
	if len(options) == 0 {
		return nil
	}
	cloned := make([]tui.DatabaseOption, len(options))
	copy(cloned, options)
	return cloned
}
