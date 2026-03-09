package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mgierok/dbc/internal/interfaces/tui"
)

func TestBuildConnectionFailureStatus_IncludesGuidanceAndDatabaseName(t *testing.T) {
	t.Parallel()

	// Arrange
	selected := tui.DatabaseOption{Name: "analytics"}

	// Act
	status := buildConnectionFailureStatus(selected, "ping failed")

	// Assert
	if !strings.Contains(status, "analytics") {
		t.Fatalf("expected selected database name in status, got %q", status)
	}
	if !strings.Contains(status, "Choose another database or edit selected entry") {
		t.Fatalf("expected user guidance in status, got %q", status)
	}
	if !strings.Contains(status, "ping failed") {
		t.Fatalf("expected error detail in status, got %q", status)
	}
}

func TestResolveStartupSelection_UsesDirectLaunchWithoutSelectorCall(t *testing.T) {
	t.Parallel()

	// Arrange
	options := startupOptions{directLaunchConnString: "/tmp/direct.sqlite"}
	listCalled := false
	selectorCalled := false

	// Act
	selected, path, err := resolveStartupSelection(
		options,
		func() ([]tui.DatabaseOption, error) {
			listCalled = true
			return []tui.DatabaseOption{}, nil
		},
		func() (tui.DatabaseOption, error) {
			selectorCalled = true
			return tui.DatabaseOption{}, nil
		},
	)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !listCalled {
		t.Fatal("expected configured database list callback for direct-launch startup")
	}
	if selectorCalled {
		t.Fatal("expected selector callback to be skipped for direct-launch startup")
	}
	if path != startupPathDirectLaunch {
		t.Fatalf("expected startup path %v, got %v", startupPathDirectLaunch, path)
	}
	if selected.ConnString != "/tmp/direct.sqlite" {
		t.Fatalf("expected direct-launch conn string, got %q", selected.ConnString)
	}
}

func TestResolveStartupSelection_UsesSelectorWhenDirectLaunchMissing(t *testing.T) {
	t.Parallel()

	// Arrange
	options := startupOptions{}
	listCalled := false
	selectorCalled := false
	expected := tui.DatabaseOption{Name: "analytics", ConnString: "/tmp/analytics.sqlite"}

	// Act
	selected, path, err := resolveStartupSelection(
		options,
		func() ([]tui.DatabaseOption, error) {
			listCalled = true
			return []tui.DatabaseOption{}, nil
		},
		func() (tui.DatabaseOption, error) {
			selectorCalled = true
			return expected, nil
		},
	)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if listCalled {
		t.Fatal("expected configured database list callback to be skipped when direct-launch is missing")
	}
	if !selectorCalled {
		t.Fatal("expected selector callback when direct-launch flag is not provided")
	}
	if path != startupPathSelector {
		t.Fatalf("expected startup path %v, got %v", startupPathSelector, path)
	}
	if selected != expected {
		t.Fatalf("expected selected option %+v, got %+v", expected, selected)
	}
}

func TestResolveStartupSelection_ReturnsSelectorError(t *testing.T) {
	t.Parallel()

	// Arrange
	options := startupOptions{}
	expectedErr := errors.New("selector failed")

	// Act
	_, _, err := resolveStartupSelection(
		options,
		func() ([]tui.DatabaseOption, error) {
			return []tui.DatabaseOption{}, nil
		},
		func() (tui.DatabaseOption, error) {
			return tui.DatabaseOption{}, expectedErr
		},
	)

	// Assert
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected selector error %v, got %v", expectedErr, err)
	}
}

func TestNewStartupSelectionStrategy_UsesDirectLaunchWhenPending(t *testing.T) {
	t.Parallel()

	// Arrange
	options := startupOptions{directLaunchConnString: "/tmp/direct.sqlite"}
	strategy := newStartupSelectionStrategy(options, true)
	listCalled := false
	selectorCalled := false

	// Act
	selected, path, err := strategy.resolve(
		func() ([]tui.DatabaseOption, error) {
			listCalled = true
			return nil, nil
		},
		func() (tui.DatabaseOption, error) {
			selectorCalled = true
			return tui.DatabaseOption{}, nil
		},
	)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !listCalled {
		t.Fatal("expected direct-launch strategy to list configured options")
	}
	if selectorCalled {
		t.Fatal("expected direct-launch strategy to skip selector callback")
	}
	if path != startupPathDirectLaunch {
		t.Fatalf("expected startup path %v, got %v", startupPathDirectLaunch, path)
	}
	if selected.ConnString != "/tmp/direct.sqlite" {
		t.Fatalf("expected direct-launch conn string %q, got %q", "/tmp/direct.sqlite", selected.ConnString)
	}
}

func TestNewStartupSelectionStrategy_FallsBackToSelectorWhenDirectLaunchIsNotPending(t *testing.T) {
	t.Parallel()

	// Arrange
	options := startupOptions{directLaunchConnString: "/tmp/direct.sqlite"}
	strategy := newStartupSelectionStrategy(options, false)
	listCalled := false
	selectorCalled := false
	expected := tui.DatabaseOption{Name: "analytics", ConnString: "/tmp/analytics.sqlite"}

	// Act
	selected, path, err := strategy.resolve(
		func() ([]tui.DatabaseOption, error) {
			listCalled = true
			return nil, nil
		},
		func() (tui.DatabaseOption, error) {
			selectorCalled = true
			return expected, nil
		},
	)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if listCalled {
		t.Fatal("expected selector strategy to skip configured database listing")
	}
	if !selectorCalled {
		t.Fatal("expected selector callback to be used when direct launch is not pending")
	}
	if path != startupPathSelector {
		t.Fatalf("expected startup path %v, got %v", startupPathSelector, path)
	}
	if selected != expected {
		t.Fatalf("expected selected option %+v, got %+v", expected, selected)
	}
}

func TestResolveStartupSelection_ReusesConfiguredIdentityWhenNormalizedPathsMatch(t *testing.T) {
	t.Parallel()

	// Arrange
	configured := filepath.Join(t.TempDir(), "direct.sqlite")
	directLaunch := configured + string(os.PathSeparator) + "."
	options := startupOptions{directLaunchConnString: directLaunch}
	selectorCalled := false

	// Act
	selected, path, err := resolveStartupSelection(
		options,
		func() ([]tui.DatabaseOption, error) {
			return []tui.DatabaseOption{
				{Name: "local", ConnString: configured},
			}, nil
		},
		func() (tui.DatabaseOption, error) {
			selectorCalled = true
			return tui.DatabaseOption{}, nil
		},
	)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if selectorCalled {
		t.Fatal("expected selector callback to be skipped for direct-launch startup")
	}
	if path != startupPathDirectLaunch {
		t.Fatalf("expected startup path %v, got %v", startupPathDirectLaunch, path)
	}
	if selected.Name != "local" {
		t.Fatalf("expected configured identity name %q, got %q", "local", selected.Name)
	}
	if selected.ConnString != configured {
		t.Fatalf("expected configured conn string %q, got %q", configured, selected.ConnString)
	}
	if selected.Source != tui.DatabaseOptionSourceConfig {
		t.Fatalf("expected configured source %q, got %q", tui.DatabaseOptionSourceConfig, selected.Source)
	}
}

func TestResolveStartupSelection_ReturnsDirectIdentityWhenNoNormalizedConfiguredMatchExists(t *testing.T) {
	t.Parallel()

	// Arrange
	options := startupOptions{directLaunchConnString: "/tmp/direct.sqlite"}

	// Act
	selected, path, err := resolveStartupSelection(
		options,
		func() ([]tui.DatabaseOption, error) {
			return []tui.DatabaseOption{
				{Name: "local", ConnString: "/tmp/configured.sqlite"},
			}, nil
		},
		func() (tui.DatabaseOption, error) {
			return tui.DatabaseOption{}, nil
		},
	)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if path != startupPathDirectLaunch {
		t.Fatalf("expected startup path %v, got %v", startupPathDirectLaunch, path)
	}
	if selected.Name != "/tmp/direct.sqlite" {
		t.Fatalf("expected direct-launch identity name %q, got %q", "/tmp/direct.sqlite", selected.Name)
	}
	if selected.ConnString != "/tmp/direct.sqlite" {
		t.Fatalf("expected direct-launch conn string %q, got %q", "/tmp/direct.sqlite", selected.ConnString)
	}
	if selected.Source != tui.DatabaseOptionSourceCLI {
		t.Fatalf("expected direct-launch source %q, got %q", tui.DatabaseOptionSourceCLI, selected.Source)
	}
}

func TestResolveStartupSelection_UsesFirstConfiguredIdentityForDeterministicNormalizedMatch(t *testing.T) {
	t.Parallel()

	// Arrange
	configured := filepath.Join(t.TempDir(), "direct.sqlite")
	options := startupOptions{directLaunchConnString: configured + string(os.PathSeparator) + "."}

	// Act
	selected, _, err := resolveStartupSelection(
		options,
		func() ([]tui.DatabaseOption, error) {
			return []tui.DatabaseOption{
				{Name: "first", ConnString: configured},
				{Name: "second", ConnString: configured + string(os.PathSeparator) + "."},
			}, nil
		},
		func() (tui.DatabaseOption, error) {
			return tui.DatabaseOption{}, nil
		},
	)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if selected.Name != "first" {
		t.Fatalf("expected first configured identity to be selected, got %q", selected.Name)
	}
	if selected.Source != tui.DatabaseOptionSourceConfig {
		t.Fatalf("expected configured source %q, got %q", tui.DatabaseOptionSourceConfig, selected.Source)
	}
}

func TestBuildDirectLaunchFailureMessage_ContainsReasonAndGuidance(t *testing.T) {
	t.Parallel()

	// Act
	message := buildDirectLaunchFailureMessage("/tmp/missing.sqlite", "database file does not exist")

	// Assert
	if !strings.Contains(message, "database file does not exist") {
		t.Fatalf("expected failure reason in message, got %q", message)
	}
	if !strings.Contains(message, "retry with -d/--database") {
		t.Fatalf("expected corrective guidance in message, got %q", message)
	}
}

func TestTrackSessionScopedDirectLaunchOption_AddsDirectCLISelection(t *testing.T) {
	t.Parallel()

	// Arrange
	var existing []tui.DatabaseOption
	selected := tui.DatabaseOption{
		Name:       "/tmp/direct.sqlite",
		ConnString: "/tmp/direct.sqlite",
		Source:     tui.DatabaseOptionSourceCLI,
	}

	// Act
	updated := trackSessionScopedDirectLaunchOption(existing, startupPathDirectLaunch, selected)

	// Assert
	if len(updated) != 1 {
		t.Fatalf("expected one session-scoped option, got %d", len(updated))
	}
	if updated[0].ConnString != "/tmp/direct.sqlite" {
		t.Fatalf("expected session option conn string %q, got %q", "/tmp/direct.sqlite", updated[0].ConnString)
	}
}

func TestTrackSessionScopedDirectLaunchOption_SkipsConfiguredMatchReuse(t *testing.T) {
	t.Parallel()

	// Arrange
	var existing []tui.DatabaseOption
	selected := tui.DatabaseOption{
		Name:       "local",
		ConnString: "/tmp/local.sqlite",
		Source:     tui.DatabaseOptionSourceConfig,
	}

	// Act
	updated := trackSessionScopedDirectLaunchOption(existing, startupPathDirectLaunch, selected)

	// Assert
	if len(updated) != 0 {
		t.Fatalf("expected no session option for configured identity reuse, got %d", len(updated))
	}
}

func TestTrackSessionScopedDirectLaunchOption_DeduplicatesNormalizedIdentity(t *testing.T) {
	t.Parallel()

	// Arrange
	basePath := filepath.Join(t.TempDir(), "direct.sqlite")
	existing := []tui.DatabaseOption{
		{
			Name:       basePath,
			ConnString: basePath,
			Source:     tui.DatabaseOptionSourceCLI,
		},
	}
	selected := tui.DatabaseOption{
		Name:       basePath + string(os.PathSeparator) + ".",
		ConnString: basePath + string(os.PathSeparator) + ".",
		Source:     tui.DatabaseOptionSourceCLI,
	}

	// Act
	updated := trackSessionScopedDirectLaunchOption(existing, startupPathDirectLaunch, selected)

	// Assert
	if len(updated) != 1 {
		t.Fatalf("expected one deduplicated session option, got %d", len(updated))
	}
}
