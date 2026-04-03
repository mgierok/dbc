package main

import (
	"bytes"
	"errors"
	"runtime/debug"
	"strings"
	"testing"
)

func TestValidateSupportedOS(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		goos      string
		wantErr   bool
		errTokens []string
	}{
		{
			name:    "accepts darwin",
			goos:    "darwin",
			wantErr: false,
		},
		{
			name:    "accepts linux",
			goos:    "linux",
			wantErr: false,
		},
		{
			name:      "rejects windows",
			goos:      "windows",
			wantErr:   true,
			errTokens: []string{"unsupported operating system", "windows"},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act
			err := validateSupportedOS(tc.goos)

			// Assert
			if !tc.wantErr {
				if err != nil {
					t.Fatalf("expected supported OS %q to pass validation, got %v", tc.goos, err)
				}
				return
			}

			assertErrorContains(t, err, tc.errTokens...)
		})
	}
}

func TestParseStartupOptions_SuccessCases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		args []string
		want startupOptions
	}{
		{
			name: "accepts short direct launch alias",
			args: []string{"-d", "/tmp/direct.sqlite"},
			want: startupOptions{directLaunchConnString: "/tmp/direct.sqlite"},
		},
		{
			name: "accepts long direct launch alias",
			args: []string{"--database", "/tmp/direct.sqlite"},
			want: startupOptions{directLaunchConnString: "/tmp/direct.sqlite"},
		},
		{
			name: "accepts long help alias",
			args: []string{"--help"},
			want: startupOptions{informationalCommand: startupInformationalHelp},
		},
		{
			name: "accepts short help alias",
			args: []string{"-h"},
			want: startupOptions{informationalCommand: startupInformationalHelp},
		},
		{
			name: "accepts long version alias",
			args: []string{"--version"},
			want: startupOptions{informationalCommand: startupInformationalVersion},
		},
		{
			name: "accepts short version alias",
			args: []string{"-v"},
			want: startupOptions{informationalCommand: startupInformationalVersion},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act
			got, err := parseStartupOptions(tc.args)

			// Assert
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if got != tc.want {
				t.Fatalf("expected startup options %+v, got %+v", tc.want, got)
			}
		})
	}
}

func TestParseStartupOptions_ValidationFailures(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		args      []string
		errTokens []string
	}{
		{
			name:      "rejects repeated help aliases",
			args:      []string{"--help", "-h"},
			errTokens: []string{"more than once"},
		},
		{
			name:      "rejects repeated version aliases",
			args:      []string{"--version", "-v"},
			errTokens: []string{"more than once"},
		},
		{
			name:      "rejects mixed informational and direct launch flags",
			args:      []string{"--help", "-d", "/tmp/direct.sqlite"},
			errTokens: []string{"cannot be combined"},
		},
		{
			name:      "rejects mixed informational flags",
			args:      []string{"--help", "--version"},
			errTokens: []string{"cannot be combined"},
		},
		{
			name:      "rejects short direct launch alias without value",
			args:      []string{"-d"},
			errTokens: []string{"missing value", "-d/--database"},
		},
		{
			name:      "rejects long direct launch alias without value",
			args:      []string{"--database"},
			errTokens: []string{"missing value", "-d/--database"},
		},
		{
			name:      "rejects whitespace-only direct launch value",
			args:      []string{"--database", "\n  "},
			errTokens: []string{"empty value", "-d/--database"},
		},
		{
			name:      "rejects unsupported argument",
			args:      []string{"--unknown"},
			errTokens: []string{"unsupported startup argument", "-d <sqlite-db-path>"},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act
			_, err := parseStartupOptions(tc.args)

			// Assert
			_ = assertUsageError(t, err, tc.errTokens...)
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

func TestRunStartupDispatchWithOS_ShortCircuitsUnsupportedOS(t *testing.T) {
	t.Parallel()

	// Arrange
	informationalCalled := false
	runtimeCalled := false

	// Act
	err := runStartupDispatchWithOS(
		"windows",
		[]string{"--help"},
		func(startupInformationalCommand) error {
			informationalCalled = true
			return nil
		},
		func(startupOptions) error {
			runtimeCalled = true
			return nil
		},
	)

	// Assert
	assertErrorContains(t, err, "unsupported operating system", "windows")
	if informationalCalled {
		t.Fatal("expected informational handler to stay skipped for unsupported OS")
	}
	if runtimeCalled {
		t.Fatal("expected runtime handler to stay skipped for unsupported OS")
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

func TestRunStartupDispatch_InformationalAliasesProduceEquivalentRenderedOutput(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		longArgs  []string
		shortArgs []string
	}{
		{
			name:      "help aliases",
			longArgs:  []string{"--help"},
			shortArgs: []string{"-h"},
		},
		{
			name:      "version aliases",
			longArgs:  []string{"--version"},
			shortArgs: []string{"-v"},
		},
	}

	renderOutput := func(t *testing.T, args []string) string {
		t.Helper()

		rendered := ""
		err := runStartupDispatch(
			args,
			func(command startupInformationalCommand) error {
				rendered = renderStartupInformationalOutput(command)
				return nil
			},
			func(_ startupOptions) error {
				t.Fatal("expected runtime startup handler to stay skipped for informational dispatch")
				return nil
			},
		)
		if err != nil {
			t.Fatalf("expected no error for %v, got %v", args, err)
		}

		return rendered
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act
			longOutput := renderOutput(t, tc.longArgs)
			shortOutput := renderOutput(t, tc.shortArgs)

			// Assert
			if longOutput != shortOutput {
				t.Fatalf("expected equivalent rendered output for aliases, got long=%q short=%q", longOutput, shortOutput)
			}
		})
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
