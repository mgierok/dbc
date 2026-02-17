package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/mgierok/dbc/internal/application/usecase"
	"github.com/mgierok/dbc/internal/infrastructure/config"
	"github.com/mgierok/dbc/internal/infrastructure/engine"
	"github.com/mgierok/dbc/internal/interfaces/tui"
)

func main() {
	options, err := parseStartupOptions(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid startup arguments: %v\n", err)
		os.Exit(1)
	}

	cfgPath, err := config.DefaultPath()
	if err != nil {
		log.Fatalf("failed to resolve config path: %v", err)
	}

	configStore := config.NewStore(cfgPath)
	connectionChecker := engine.NewSQLiteConnectionChecker()
	listConfiguredDatabases := usecase.NewListConfiguredDatabases(configStore)
	createConfiguredDatabase := usecase.NewCreateConfiguredDatabase(configStore, connectionChecker)
	updateConfiguredDatabase := usecase.NewUpdateConfiguredDatabase(configStore, connectionChecker)
	deleteConfiguredDatabase := usecase.NewDeleteConfiguredDatabase(configStore)
	getActiveConfigPath := usecase.NewGetActiveConfigPath(configStore)
	selectorState := tui.SelectorLaunchState{}
	directLaunchPending := options.directLaunchConnString != ""

	for {
		currentStartupOptions := startupOptions{}
		if directLaunchPending {
			currentStartupOptions = options
		}

		selected, startupPath, err := resolveStartupSelection(
			currentStartupOptions,
			func() ([]tui.DatabaseOption, error) {
				return listConfiguredDatabaseOptions(context.Background(), listConfiguredDatabases)
			},
			func() (tui.DatabaseOption, error) {
				return tui.SelectDatabaseWithState(
					context.Background(),
					listConfiguredDatabases,
					createConfiguredDatabase,
					updateConfiguredDatabase,
					deleteConfiguredDatabase,
					getActiveConfigPath,
					selectorState,
				)
			},
		)
		if err != nil {
			if errors.Is(err, tui.ErrDatabaseSelectionCanceled) {
				return
			}
			log.Fatalf("failed to select database: %v", err)
		}
		directLaunchPending = false

		db, err := connectSelectedDatabase(selected)
		if err != nil {
			if startupPath == startupPathDirectLaunch {
				fmt.Fprintln(os.Stderr, buildDirectLaunchFailureMessage(selected.ConnString, err.Error()))
				os.Exit(1)
			}

			selectorState = tui.SelectorLaunchState{
				StatusMessage:    buildConnectionFailureStatus(selected, err.Error()),
				PreferConnString: selected.ConnString,
			}
			continue
		}
		selectorState = tui.SelectorLaunchState{}

		engine := engine.NewSQLiteEngine(db)
		listTables := usecase.NewListTables(engine)
		getSchema := usecase.NewGetSchema(engine)
		listRecords := usecase.NewListRecords(engine)
		listOperators := usecase.NewListOperators(engine)
		saveChanges := usecase.NewSaveTableChanges(engine)

		runErr := tui.Run(context.Background(), listTables, getSchema, listRecords, listOperators, saveChanges)
		if closeErr := db.Close(); closeErr != nil {
			log.Printf("failed to close database: %v", closeErr)
		}
		if errors.Is(runErr, tui.ErrOpenConfigSelector) {
			continue
		}
		if runErr != nil {
			fmt.Printf("application error: %v\n", runErr)
		}
		return
	}
}

type startupPath int

const (
	startupPathSelector startupPath = iota
	startupPathDirectLaunch
)

type startupOptions struct {
	directLaunchConnString string
}

func parseStartupOptions(args []string) (startupOptions, error) {
	options := startupOptions{}

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-d", "--database":
			if options.directLaunchConnString != "" {
				return startupOptions{}, errors.New("direct-launch parameter was provided more than once; use exactly one of -d or --database")
			}
			if i+1 >= len(args) {
				return startupOptions{}, errors.New("missing value for -d/--database; usage: dbc -d <sqlite-db-path>")
			}
			next := strings.TrimSpace(args[i+1])
			if next == "" {
				return startupOptions{}, errors.New("empty value for -d/--database; provide a non-empty SQLite database path")
			}

			options.directLaunchConnString = next
			i++
		default:
			return startupOptions{}, fmt.Errorf(
				"unsupported startup argument %q; supported direct-launch options: -d <sqlite-db-path> or --database <sqlite-db-path>",
				args[i],
			)
		}
	}

	return options, nil
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
		}
		configuredOptions, err := listConfiguredDatabases()
		if err != nil {
			return tui.DatabaseOption{}, startupPathDirectLaunch, err
		}
		if matched, ok := resolveConfiguredDirectLaunchIdentity(directLaunchSelection.ConnString, configuredOptions); ok {
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
	if runtime.GOOS == "windows" {
		return strings.EqualFold(left, right)
	}
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
