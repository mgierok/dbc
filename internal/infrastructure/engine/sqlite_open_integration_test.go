package engine

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

type stubSQLiteDatabaseHandle struct {
	pingErr  error
	closeErr error
}

func (s stubSQLiteDatabaseHandle) PingContext(ctx context.Context) error {
	return s.pingErr
}

func (s stubSQLiteDatabaseHandle) Close() error {
	return s.closeErr
}

func TestOpenSQLiteDatabase_UsesBackgroundContextWhenContextIsNil(t *testing.T) {
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

	// Act
	var nilCtx context.Context
	db, err := OpenSQLiteDatabase(nilCtx, dbPath)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if db == nil {
		t.Fatal("expected database handle, got nil")
	}
	if err := db.Close(); err != nil {
		t.Fatalf("expected close without error, got %v", err)
	}
}

func TestOpenSQLiteDatabase_ReturnsPingErrorWhenCloseSucceeds(t *testing.T) {
	// Arrange
	dbPath := filepath.Join(t.TempDir(), "existing.sqlite")
	if err := os.WriteFile(dbPath, []byte{}, 0o600); err != nil {
		t.Fatalf("failed to prepare sqlite file: %v", err)
	}

	expectedPingErr := errors.New("ping failed")
	originalOpen := openSQLiteHandle
	t.Cleanup(func() {
		openSQLiteHandle = originalOpen
	})
	openSQLiteHandle = func(path string) (sqliteDatabaseHandle, error) {
		return stubSQLiteDatabaseHandle{
			pingErr:  expectedPingErr,
			closeErr: nil,
		}, nil
	}

	// Act
	db, err := OpenSQLiteDatabase(context.Background(), dbPath)

	// Assert
	if db != nil {
		t.Fatalf("expected nil database on ping failure, got %#v", db)
	}
	if !errors.Is(err, expectedPingErr) {
		t.Fatalf("expected error %v, got %v", expectedPingErr, err)
	}
}

func TestOpenSQLiteDatabase_ReturnsJoinedPingAndCloseErrorsWhenBothFail(t *testing.T) {
	// Arrange
	dbPath := filepath.Join(t.TempDir(), "existing.sqlite")
	if err := os.WriteFile(dbPath, []byte{}, 0o600); err != nil {
		t.Fatalf("failed to prepare sqlite file: %v", err)
	}

	expectedPingErr := errors.New("ping failed")
	expectedCloseErr := errors.New("close failed")
	originalOpen := openSQLiteHandle
	t.Cleanup(func() {
		openSQLiteHandle = originalOpen
	})
	openSQLiteHandle = func(path string) (sqliteDatabaseHandle, error) {
		return stubSQLiteDatabaseHandle{
			pingErr:  expectedPingErr,
			closeErr: expectedCloseErr,
		}, nil
	}

	// Act
	db, err := OpenSQLiteDatabase(context.Background(), dbPath)

	// Assert
	if db != nil {
		t.Fatalf("expected nil database on ping failure, got %#v", db)
	}
	if !errors.Is(err, expectedPingErr) {
		t.Fatalf("expected joined error to contain ping error %v, got %v", expectedPingErr, err)
	}
	if !errors.Is(err, expectedCloseErr) {
		t.Fatalf("expected joined error to contain close error %v, got %v", expectedCloseErr, err)
	}
}
