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
	"runtime/debug"
	"strings"

	"github.com/mgierok/dbc/internal/application/usecase"
	"github.com/mgierok/dbc/internal/infrastructure/config"
	"github.com/mgierok/dbc/internal/infrastructure/engine"
	"github.com/mgierok/dbc/internal/interfaces/tui"
)

func main() {
	err := runStartupDispatch(
		os.Args[1:],
		func(command startupInformationalCommand) error {
			_, err := fmt.Fprintln(os.Stdout, renderStartupInformationalOutput(command))
			return err
		},
		func(options startupOptions) error {
			runRuntimeStartup(options)
			return nil
		},
	)
	if err != nil {
		failure := classifyStartupFailure(err)
		fmt.Fprintln(os.Stderr, failure.stderrOutput)
		os.Exit(failure.exitCode)
	}
}

func runRuntimeStartup(options startupOptions) {
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
	sessionScopedOptions := []tui.DatabaseOption{}
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
		sessionScopedOptions = trackSessionScopedDirectLaunchOption(sessionScopedOptions, startupPath, selected)
		directLaunchPending = false

		db, err := connectSelectedDatabase(selected)
		if err != nil {
			if startupPath == startupPathDirectLaunch {
				fmt.Fprintln(os.Stderr, buildDirectLaunchFailureMessage(selected.ConnString, err.Error()))
				os.Exit(1)
			}

			selectorState = tui.SelectorLaunchState{
				StatusMessage:     buildConnectionFailureStatus(selected, err.Error()),
				PreferConnString:  selected.ConnString,
				AdditionalOptions: cloneDatabaseOptions(sessionScopedOptions),
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
			selectorState = tui.SelectorLaunchState{
				PreferConnString:  selected.ConnString,
				AdditionalOptions: cloneDatabaseOptions(sessionScopedOptions),
			}
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

type startupInformationalCommand int

const (
	startupInformationalNone startupInformationalCommand = iota
	startupInformationalHelp
	startupInformationalVersion
)

const (
	startupVersionFallbackToken   = "dev"
	startupVersionShortHashLength = 12
	startupExitCodeRuntimeFailure = 1
	startupExitCodeInvalidUsage   = 2
)

type startupOptions struct {
	directLaunchConnString string
	informationalCommand   startupInformationalCommand
}

type startupFailure struct {
	exitCode     int
	stderrOutput string
}

type startupUsageError struct {
	message string
}

func (e startupUsageError) Error() string {
	return e.message
}

func newStartupUsageError(message string) error {
	return startupUsageError{message: message}
}

func newStartupUsageErrorf(format string, args ...any) error {
	return startupUsageError{message: fmt.Sprintf(format, args...)}
}

func classifyStartupFailure(err error) startupFailure {
	var usageErr startupUsageError
	if errors.As(err, &usageErr) {
		return startupFailure{
			exitCode:     startupExitCodeInvalidUsage,
			stderrOutput: renderStartupUsageFailureOutput(usageErr),
		}
	}

	return startupFailure{
		exitCode:     startupExitCodeRuntimeFailure,
		stderrOutput: fmt.Sprintf("Startup error: %v", err),
	}
}

func renderStartupUsageFailureOutput(err error) string {
	detail := strings.TrimSpace(err.Error())
	detail = strings.TrimSuffix(detail, ".")

	lines := []string{
		"Error: invalid startup arguments.",
		fmt.Sprintf("Hint: %s. Run 'dbc --help' for full startup usage.", detail),
		"Usage: dbc [options].",
	}

	return strings.Join(lines, "\n")
}

func runStartupDispatch(
	args []string,
	handleInformational func(startupInformationalCommand) error,
	runRuntime func(startupOptions) error,
) error {
	options, err := parseStartupOptions(args)
	if err != nil {
		return err
	}

	if options.informationalCommand != startupInformationalNone {
		return handleInformational(options.informationalCommand)
	}

	return runRuntime(options)
}

func parseStartupOptions(args []string) (startupOptions, error) {
	options := startupOptions{}
	var helpFlagCount int
	var versionFlagCount int

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-h", "--help":
			if options.directLaunchConnString != "" {
				return startupOptions{}, newStartupUsageError("informational flag cannot be combined with -d/--database in the same startup invocation")
			}
			helpFlagCount++
			if helpFlagCount > 1 {
				return startupOptions{}, newStartupUsageError("help flag was provided more than once; use exactly one of -h or --help")
			}
			if options.informationalCommand == startupInformationalVersion {
				return startupOptions{}, newStartupUsageError("help and version informational flags cannot be combined in the same startup invocation")
			}
			options.informationalCommand = startupInformationalHelp
		case "-v", "--version":
			if options.directLaunchConnString != "" {
				return startupOptions{}, newStartupUsageError("informational flag cannot be combined with -d/--database in the same startup invocation")
			}
			versionFlagCount++
			if versionFlagCount > 1 {
				return startupOptions{}, newStartupUsageError("version flag was provided more than once; use exactly one of -v or --version")
			}
			if options.informationalCommand == startupInformationalHelp {
				return startupOptions{}, newStartupUsageError("help and version informational flags cannot be combined in the same startup invocation")
			}
			options.informationalCommand = startupInformationalVersion
		case "-d", "--database":
			if options.directLaunchConnString != "" {
				return startupOptions{}, newStartupUsageError("direct-launch parameter was provided more than once; use exactly one of -d or --database")
			}
			if options.informationalCommand != startupInformationalNone {
				return startupOptions{}, newStartupUsageError("informational flag cannot be combined with -d/--database in the same startup invocation")
			}
			if i+1 >= len(args) {
				return startupOptions{}, newStartupUsageError("missing value for -d/--database; usage: dbc -d <sqlite-db-path>")
			}
			next := strings.TrimSpace(args[i+1])
			if next == "" {
				return startupOptions{}, newStartupUsageError("empty value for -d/--database; provide a non-empty SQLite database path")
			}

			options.directLaunchConnString = next
			i++
		default:
			return startupOptions{}, newStartupUsageErrorf(
				"unsupported startup argument %q; supported options: -d <sqlite-db-path>, --database <sqlite-db-path>, -h/--help, -v/--version",
				args[i],
			)
		}
	}

	return options, nil
}

func renderStartupInformationalOutput(command startupInformationalCommand) string {
	switch command {
	case startupInformationalHelp:
		return renderStartupHelpOutput()
	case startupInformationalVersion:
		return resolveStartupVersionToken(debug.ReadBuildInfo)
	default:
		return ""
	}
}

func resolveStartupVersionToken(readBuildInfo func() (*debug.BuildInfo, bool)) string {
	if readBuildInfo == nil {
		return startupVersionFallbackToken
	}

	buildInfo, ok := readBuildInfo()
	if !ok || buildInfo == nil {
		return startupVersionFallbackToken
	}

	for _, setting := range buildInfo.Settings {
		if setting.Key != "vcs.revision" {
			continue
		}

		revisionFields := strings.Fields(setting.Value)
		if len(revisionFields) == 0 {
			return startupVersionFallbackToken
		}

		return shortRevisionToken(revisionFields[0])
	}

	return startupVersionFallbackToken
}

func shortRevisionToken(revision string) string {
	if len(revision) <= startupVersionShortHashLength {
		return revision
	}
	return revision[:startupVersionShortHashLength]
}

func renderStartupHelpOutput() string {
	lines := []string{
		"DBC is a terminal-first SQLite database browser.",
		"",
		"Usage:",
		"  dbc [options]",
		"",
		"Options:",
		"  -h, --help                      Show startup help and exit.",
		"  -v, --version                   Print build version token and exit.",
		"  -d, --database <sqlite-db-path> Launch directly with a SQLite database path.",
		"",
		"Examples:",
		"  dbc --database ./data/app.sqlite",
		"  dbc --version",
	}

	return strings.Join(lines, "\n")
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
