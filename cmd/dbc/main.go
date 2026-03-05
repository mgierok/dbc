package main

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
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
	if err := validateSupportedOS(runtime.GOOS); err != nil {
		return err
	}

	options, err := parseStartupOptions(args)
	if err != nil {
		return err
	}

	if options.informationalCommand != startupInformationalNone {
		return handleInformational(options.informationalCommand)
	}

	return runRuntime(options)
}

func validateSupportedOS(goos string) error {
	switch goos {
	case "darwin", "linux":
		return nil
	default:
		return fmt.Errorf(
			"unsupported operating system %q: supported operating systems are macOS and Linux",
			goos,
		)
	}
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
