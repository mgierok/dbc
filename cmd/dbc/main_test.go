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
