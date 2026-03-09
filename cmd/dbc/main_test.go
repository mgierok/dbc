package main

import (
	"bytes"
	"errors"
	"runtime/debug"
	"strings"
	"testing"
)

func TestValidateSupportedOS_AcceptsMacOSAndLinux(t *testing.T) {
	t.Parallel()

	cases := []string{"darwin", "linux"}
	for _, goos := range cases {
		goos := goos
		t.Run(goos, func(t *testing.T) {
			t.Parallel()

			// Act
			err := validateSupportedOS(goos)

			// Assert
			if err != nil {
				t.Fatalf("expected supported OS %q to pass validation, got %v", goos, err)
			}
		})
	}
}

func TestValidateSupportedOS_RejectsWindows(t *testing.T) {
	t.Parallel()

	// Act
	err := validateSupportedOS("windows")

	// Assert
	if err == nil {
		t.Fatal("expected unsupported-OS error, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported operating system") {
		t.Fatalf("expected unsupported-OS token, got %q", err.Error())
	}
	if !strings.Contains(err.Error(), "windows") {
		t.Fatalf("expected OS name in error, got %q", err.Error())
	}
}

func TestParseStartupOptions_AcceptsDirectLaunchAliases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		args []string
	}{
		{
			name: "short alias",
			args: []string{"-d", "/tmp/direct.sqlite"},
		},
		{
			name: "long alias",
			args: []string{"--database", "/tmp/direct.sqlite"},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act
			options, err := parseStartupOptions(tc.args)

			// Assert
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if options.directLaunchConnString != "/tmp/direct.sqlite" {
				t.Fatalf("expected direct-launch connection string to be parsed, got %q", options.directLaunchConnString)
			}
		})
	}
}

func TestParseStartupOptions_AcceptsInformationalAliases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		args     []string
		expected startupInformationalCommand
	}{
		{
			name:     "long help alias",
			args:     []string{"--help"},
			expected: startupInformationalHelp,
		},
		{
			name:     "short help alias",
			args:     []string{"-h"},
			expected: startupInformationalHelp,
		},
		{
			name:     "long version alias",
			args:     []string{"--version"},
			expected: startupInformationalVersion,
		},
		{
			name:     "short version alias",
			args:     []string{"-v"},
			expected: startupInformationalVersion,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act
			options, err := parseStartupOptions(tc.args)

			// Assert
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if options.informationalCommand != tc.expected {
				t.Fatalf("expected informational command %v, got %v", tc.expected, options.informationalCommand)
			}
			if options.directLaunchConnString != "" {
				t.Fatalf("expected direct-launch connection string to stay empty, got %q", options.directLaunchConnString)
			}
		})
	}
}

func TestParseStartupOptions_ReturnsErrorForRepeatedLogicalInformationalAliases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		args []string
	}{
		{
			name: "help aliases repeated logically",
			args: []string{"--help", "-h"},
		},
		{
			name: "version aliases repeated logically",
			args: []string{"--version", "-v"},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act
			_, err := parseStartupOptions(tc.args)

			// Assert
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), "more than once") {
				t.Fatalf("expected duplicate-flag guidance, got %q", err.Error())
			}
		})
	}
}

func TestParseStartupOptions_ReturnsErrorForMixedInformationalAndDirectLaunchFlags(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		args []string
	}{
		{
			name: "help mixed with direct launch",
			args: []string{"--help", "-d", "/tmp/direct.sqlite"},
		},
		{
			name: "version mixed with direct launch",
			args: []string{"--version", "--database", "/tmp/direct.sqlite"},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act
			_, err := parseStartupOptions(tc.args)

			// Assert
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), "cannot be combined") {
				t.Fatalf("expected mixed-mode guidance, got %q", err.Error())
			}
		})
	}
}

func TestParseStartupOptions_ReturnsErrorForMixedInformationalFlags(t *testing.T) {
	t.Parallel()

	// Act
	_, err := parseStartupOptions([]string{"--help", "--version"})

	// Assert
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "cannot be combined") {
		t.Fatalf("expected mixed informational guidance, got %q", err.Error())
	}
}

func TestParseStartupOptions_ReturnsErrorForMissingDirectLaunchValue(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		args []string
	}{
		{
			name: "short alias missing value",
			args: []string{"-d"},
		},
		{
			name: "long alias missing value",
			args: []string{"--database"},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act
			_, err := parseStartupOptions(tc.args)

			// Assert
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), "missing value") {
				t.Fatalf("expected missing-value guidance, got %q", err.Error())
			}
			if !strings.Contains(err.Error(), "-d/--database") {
				t.Fatalf("expected argument hint in error, got %q", err.Error())
			}
		})
	}
}

func TestParseStartupOptions_ReturnsErrorForUnsupportedArgument(t *testing.T) {
	t.Parallel()

	// Act
	_, err := parseStartupOptions([]string{"--unknown"})

	// Assert
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported startup argument") {
		t.Fatalf("expected unsupported-argument error, got %q", err.Error())
	}
	if !strings.Contains(err.Error(), "-d <sqlite-db-path>") {
		t.Fatalf("expected supported usage hint, got %q", err.Error())
	}
}

func TestParseStartupOptions_ReturnsUsageErrorTypeForValidationFailures(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		args []string
	}{
		{
			name: "unsupported argument",
			args: []string{"--unknown"},
		},
		{
			name: "missing value",
			args: []string{"--database"},
		},
		{
			name: "mixed informational and direct launch",
			args: []string{"--help", "--database", "/tmp/direct.sqlite"},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act
			_, err := parseStartupOptions(tc.args)

			// Assert
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			var usageErr startupUsageError
			if !errors.As(err, &usageErr) {
				t.Fatalf("expected startupUsageError type, got %T", err)
			}
		})
	}
}

func TestClassifyStartupFailure_MapsUsageErrorsToExitCodeTwoWithUsageContract(t *testing.T) {
	t.Parallel()

	// Arrange
	err := startupUsageError{message: "unsupported startup argument \"--bad\""}

	// Act
	failure := classifyStartupFailure(err)

	// Assert
	if failure.exitCode != 2 {
		t.Fatalf("expected exit code 2 for usage error, got %d", failure.exitCode)
	}
	requiredTokens := []string{
		"Error:",
		"Hint:",
		"Usage:",
		"dbc [options]",
		"dbc --help",
	}
	for _, token := range requiredTokens {
		if !strings.Contains(failure.stderrOutput, token) {
			t.Fatalf("expected usage contract token %q in output %q", token, failure.stderrOutput)
		}
	}
}

func TestClassifyStartupFailure_MapsRuntimeErrorsToExitCodeOne(t *testing.T) {
	t.Parallel()

	// Arrange
	err := errors.New("stdout write failed")

	// Act
	failure := classifyStartupFailure(err)

	// Assert
	if failure.exitCode != 1 {
		t.Fatalf("expected exit code 1 for runtime failure, got %d", failure.exitCode)
	}
	if !strings.Contains(failure.stderrOutput, "Startup error:") {
		t.Fatalf("expected runtime error prefix, got %q", failure.stderrOutput)
	}
	if !strings.Contains(failure.stderrOutput, "stdout write failed") {
		t.Fatalf("expected runtime error details, got %q", failure.stderrOutput)
	}
}

func TestRunMain_ReturnsRuntimeFailureExitCodeWhenRuntimeStartupFails(t *testing.T) {
	t.Parallel()

	// Arrange
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	// Act
	exitCode := runMain(
		[]string{"-d", "/tmp/direct.sqlite"},
		&stdout,
		&stderr,
		"linux",
		func() (*debug.BuildInfo, bool) {
			return nil, false
		},
		func(_ startupOptions) error {
			return errors.New("runtime bootstrap failed")
		},
	)

	// Assert
	if exitCode != startupExitCodeRuntimeFailure {
		t.Fatalf("expected runtime failure exit code %d, got %d", startupExitCodeRuntimeFailure, exitCode)
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected runtime failure path to avoid stdout output, got %q", stdout.String())
	}
	rendered := stderr.String()
	if !strings.Contains(rendered, "Startup error: runtime bootstrap failed") {
		t.Fatalf("expected runtime failure rendering in stderr, got %q", rendered)
	}
}

func TestRunMain_UsesExplicitOperationalFailureOutputWhenProvided(t *testing.T) {
	t.Parallel()

	// Arrange
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	expectedMessage := buildDirectLaunchFailureMessage("/tmp/missing.sqlite", "database file does not exist")

	// Act
	exitCode := runMain(
		[]string{"-d", "/tmp/missing.sqlite"},
		&stdout,
		&stderr,
		"linux",
		func() (*debug.BuildInfo, bool) {
			return nil, false
		},
		func(_ startupOptions) error {
			return newPresentedStartupFailure(startupExitCodeRuntimeFailure, expectedMessage)
		},
	)

	// Assert
	if exitCode != startupExitCodeRuntimeFailure {
		t.Fatalf("expected runtime failure exit code %d, got %d", startupExitCodeRuntimeFailure, exitCode)
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected explicit operational failure path to avoid stdout output, got %q", stdout.String())
	}
	if strings.TrimSpace(stderr.String()) != expectedMessage {
		t.Fatalf("expected explicit operational failure message %q, got %q", expectedMessage, stderr.String())
	}
}

func TestRunStartupDispatch_UsesInformationalHandlerWithoutRuntimeStartup(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name            string
		args            []string
		expectedCommand startupInformationalCommand
	}{
		{
			name:            "long help alias",
			args:            []string{"--help"},
			expectedCommand: startupInformationalHelp,
		},
		{
			name:            "short help alias",
			args:            []string{"-h"},
			expectedCommand: startupInformationalHelp,
		},
		{
			name:            "long version alias",
			args:            []string{"--version"},
			expectedCommand: startupInformationalVersion,
		},
		{
			name:            "short version alias",
			args:            []string{"-v"},
			expectedCommand: startupInformationalVersion,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			runtimeCalled := false
			handledCommand := startupInformationalNone

			// Act
			err := runStartupDispatch(
				tc.args,
				func(command startupInformationalCommand) error {
					handledCommand = command
					return nil
				},
				func(_ startupOptions) error {
					runtimeCalled = true
					return nil
				},
			)

			// Assert
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if runtimeCalled {
				t.Fatal("expected runtime startup handler to stay skipped for informational dispatch")
			}
			if handledCommand != tc.expectedCommand {
				t.Fatalf("expected informational command %v, got %v", tc.expectedCommand, handledCommand)
			}
		})
	}
}

func TestRunStartupDispatch_UsesRuntimeStartupWhenInformationalFlagsAreAbsent(t *testing.T) {
	t.Parallel()

	// Arrange
	informationalCalled := false
	runtimeCalled := false
	capturedOptions := startupOptions{}

	// Act
	err := runStartupDispatch(
		[]string{"-d", "/tmp/direct.sqlite"},
		func(_ startupInformationalCommand) error {
			informationalCalled = true
			return nil
		},
		func(options startupOptions) error {
			runtimeCalled = true
			capturedOptions = options
			return nil
		},
	)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if informationalCalled {
		t.Fatal("expected informational handler to stay skipped when flags are absent")
	}
	if !runtimeCalled {
		t.Fatal("expected runtime startup handler to be called")
	}
	if capturedOptions.directLaunchConnString != "/tmp/direct.sqlite" {
		t.Fatalf("expected parsed direct-launch value %q, got %q", "/tmp/direct.sqlite", capturedOptions.directLaunchConnString)
	}
}

func TestRunStartupDispatch_HelpAliasesProduceEquivalentRenderedOutput(t *testing.T) {
	t.Parallel()

	renderHelp := func(args []string) (string, error) {
		rendered := ""
		err := runStartupDispatch(
			args,
			func(command startupInformationalCommand) error {
				rendered = renderStartupInformationalOutput(command)
				return nil
			},
			func(_ startupOptions) error {
				t.Fatal("expected runtime startup handler to stay skipped for help dispatch")
				return nil
			},
		)
		return rendered, err
	}

	// Act
	longHelpOutput, longErr := renderHelp([]string{"--help"})
	shortHelpOutput, shortErr := renderHelp([]string{"-h"})

	// Assert
	if longErr != nil {
		t.Fatalf("expected no error for --help, got %v", longErr)
	}
	if shortErr != nil {
		t.Fatalf("expected no error for -h, got %v", shortErr)
	}
	if longHelpOutput != shortHelpOutput {
		t.Fatalf("expected equivalent help output for aliases, got --help=%q and -h=%q", longHelpOutput, shortHelpOutput)
	}
}

func TestRunStartupDispatch_VersionAliasesProduceEquivalentRenderedOutput(t *testing.T) {
	t.Parallel()

	renderVersion := func(args []string) (string, error) {
		rendered := ""
		err := runStartupDispatch(
			args,
			func(command startupInformationalCommand) error {
				rendered = renderStartupInformationalOutput(command)
				return nil
			},
			func(_ startupOptions) error {
				t.Fatal("expected runtime startup handler to stay skipped for version dispatch")
				return nil
			},
		)
		return rendered, err
	}

	// Act
	longVersionOutput, longErr := renderVersion([]string{"--version"})
	shortVersionOutput, shortErr := renderVersion([]string{"-v"})

	// Assert
	if longErr != nil {
		t.Fatalf("expected no error for --version, got %v", longErr)
	}
	if shortErr != nil {
		t.Fatalf("expected no error for -v, got %v", shortErr)
	}
	if longVersionOutput != shortVersionOutput {
		t.Fatalf("expected equivalent version output for aliases, got --version=%q and -v=%q", longVersionOutput, shortVersionOutput)
	}
}

func TestRenderStartupInformationalOutput_HelpContainsRequiredContractTokens(t *testing.T) {
	t.Parallel()

	// Act
	helpOutput := renderStartupInformationalOutput(startupInformationalHelp)

	// Assert
	requiredTokens := []string{
		"DBC is a terminal-first SQLite database browser.",
		"Usage:",
		"dbc [options]",
		"Options:",
		"-h, --help",
		"-v, --version",
		"-d, --database <sqlite-db-path>",
		"Examples:",
		"dbc --database ./data/app.sqlite",
		"dbc --version",
	}

	for _, token := range requiredTokens {
		if !strings.Contains(helpOutput, token) {
			t.Fatalf("expected help output to include token %q, got %q", token, helpOutput)
		}
	}
}

func TestRenderStartupInformationalOutput_HelpIsDeterministic(t *testing.T) {
	t.Parallel()

	// Act
	first := renderStartupInformationalOutput(startupInformationalHelp)
	second := renderStartupInformationalOutput(startupInformationalHelp)

	// Assert
	if first == "" {
		t.Fatal("expected non-empty help output")
	}
	if first != second {
		t.Fatalf("expected deterministic help output, got first=%q second=%q", first, second)
	}
}

func TestResolveStartupVersionToken_ReturnsShortCommitHashWhenRevisionMetadataAvailable(t *testing.T) {
	t.Parallel()

	// Arrange
	buildInfo := &debug.BuildInfo{
		Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: "0123456789abcdef0123456789abcdef01234567"},
		},
	}

	// Act
	version := resolveStartupVersionToken(func() (*debug.BuildInfo, bool) {
		return buildInfo, true
	})

	// Assert
	if version != "0123456789ab" {
		t.Fatalf("expected short revision token %q, got %q", "0123456789ab", version)
	}
}

func TestResolveStartupVersionToken_ReturnsDevWhenRevisionMetadataIsUnavailable(t *testing.T) {
	t.Parallel()

	// Act
	version := resolveStartupVersionToken(func() (*debug.BuildInfo, bool) {
		return nil, false
	})

	// Assert
	if version != "dev" {
		t.Fatalf("expected fallback token %q, got %q", "dev", version)
	}
}

func TestResolveStartupVersionToken_IsDeterministicForSameBuildMetadata(t *testing.T) {
	t.Parallel()

	// Arrange
	buildInfo := &debug.BuildInfo{
		Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: "abcdef0123456789abcdef0123456789abcdef01"},
		},
	}

	readBuildInfo := func() (*debug.BuildInfo, bool) {
		return buildInfo, true
	}

	// Act
	first := resolveStartupVersionToken(readBuildInfo)
	second := resolveStartupVersionToken(readBuildInfo)

	// Assert
	if first == "" {
		t.Fatal("expected non-empty version token")
	}
	if first != second {
		t.Fatalf("expected deterministic version token, got first=%q second=%q", first, second)
	}
}

func TestRenderStartupInformationalOutput_VersionIsSingleToken(t *testing.T) {
	t.Parallel()

	// Act
	versionOutput := renderStartupInformationalOutput(startupInformationalVersion)

	// Assert
	if len(strings.Fields(versionOutput)) != 1 {
		t.Fatalf("expected single-token version output, got %q", versionOutput)
	}
	if strings.TrimSpace(versionOutput) != versionOutput {
		t.Fatalf("expected trimmed version token, got %q", versionOutput)
	}
}
