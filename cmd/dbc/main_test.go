package main

import (
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/mgierok/dbc/internal/interfaces/tui"
)

func TestConnectSelectedDatabase_ReturnsErrorForInvalidPath(t *testing.T) {
	// Arrange
	selected := tui.DatabaseOption{
		Name:       "invalid",
		ConnString: t.TempDir(),
	}

	// Act
	db, err := connectSelectedDatabase(selected)

	// Assert
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if db != nil {
		t.Fatal("expected nil database on error")
	}
}

func TestConnectSelectedDatabase_ReturnsErrorForMissingDatabaseFile(t *testing.T) {
	// Arrange
	missingPath := filepath.Join(t.TempDir(), "missing.sqlite")
	selected := tui.DatabaseOption{
		Name:       "missing",
		ConnString: missingPath,
	}

	// Act
	db, err := connectSelectedDatabase(selected)

	// Assert
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
	if db != nil {
		t.Fatal("expected nil database on error")
	}
	if _, statErr := os.Stat(missingPath); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("expected missing file to remain missing, got stat error %v", statErr)
	}
}

func TestConnectSelectedDatabase_ReturnsDatabaseForExistingReachableConnection(t *testing.T) {
	// Arrange
	dbPath := filepath.Join(t.TempDir(), "existing.sqlite")
	seed, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open seed sqlite database: %v", err)
	}
	if _, err := seed.Exec(`CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY);`); err != nil {
		t.Fatalf("failed to initialize seed sqlite database: %v", err)
	}
	if err := seed.Close(); err != nil {
		t.Fatalf("failed to close seed sqlite database: %v", err)
	}

	selected := tui.DatabaseOption{
		Name:       "existing",
		ConnString: dbPath,
	}

	// Act
	db, err := connectSelectedDatabase(selected)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if db == nil {
		t.Fatal("expected database instance, got nil")
	}
	if closeErr := db.Close(); closeErr != nil {
		t.Fatalf("expected close without error, got %v", closeErr)
	}
}

func TestBuildConnectionFailureStatus_IncludesGuidanceAndDatabaseName(t *testing.T) {
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

func TestResolveStartupSelection_UsesDirectLaunchWithoutSelectorCall(t *testing.T) {
	t.Parallel()

	// Arrange
	options := startupOptions{directLaunchConnString: "/tmp/direct.sqlite"}
	selectorCalled := false

	// Act
	selected, path, err := resolveStartupSelection(options, func() (tui.DatabaseOption, error) {
		selectorCalled = true
		return tui.DatabaseOption{}, nil
	})

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
	if selected.ConnString != "/tmp/direct.sqlite" {
		t.Fatalf("expected direct-launch conn string, got %q", selected.ConnString)
	}
}

func TestResolveStartupSelection_UsesSelectorWhenDirectLaunchMissing(t *testing.T) {
	t.Parallel()

	// Arrange
	options := startupOptions{}
	selectorCalled := false
	expected := tui.DatabaseOption{Name: "analytics", ConnString: "/tmp/analytics.sqlite"}

	// Act
	selected, path, err := resolveStartupSelection(options, func() (tui.DatabaseOption, error) {
		selectorCalled = true
		return expected, nil
	})

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
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
	_, _, err := resolveStartupSelection(options, func() (tui.DatabaseOption, error) {
		return tui.DatabaseOption{}, expectedErr
	})

	// Assert
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected selector error %v, got %v", expectedErr, err)
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
