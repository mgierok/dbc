package engine

import (
	"context"
	"database/sql"
	"path/filepath"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

func TestSQLiteConnectionChecker_CanConnect_WhenDatabaseExists(t *testing.T) {
	// Arrange
	dbPath := filepath.Join(t.TempDir(), "existing.sqlite")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open sqlite database: %v", err)
	}
	if _, err := db.Exec(`CREATE TABLE users (id INTEGER PRIMARY KEY);`); err != nil {
		t.Fatalf("failed to initialize sqlite database: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("failed to close sqlite database: %v", err)
	}
	checker := NewSQLiteConnectionChecker()

	// Act
	err = checker.CanConnect(context.Background(), dbPath)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestSQLiteConnectionChecker_CanConnect_WhenDatabaseFileMissing(t *testing.T) {
	// Arrange
	dbPath := filepath.Join(t.TempDir(), "missing.sqlite")
	checker := NewSQLiteConnectionChecker()

	// Act
	err := checker.CanConnect(context.Background(), dbPath)

	// Assert
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("expected missing file error, got %v", err)
	}
}

func TestSQLiteConnectionChecker_CanConnect_WhenPathIsDirectory(t *testing.T) {
	// Arrange
	dirPath := t.TempDir()
	checker := NewSQLiteConnectionChecker()

	// Act
	err := checker.CanConnect(context.Background(), dirPath)

	// Assert
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "directory") {
		t.Fatalf("expected directory path error, got %v", err)
	}
}
