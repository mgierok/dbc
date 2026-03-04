package engine

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"

	_ "modernc.org/sqlite"
)

type sqliteDatabaseHandle interface {
	PingContext(ctx context.Context) error
	Close() error
}

var openSQLiteHandle = func(dbPath string) (sqliteDatabaseHandle, error) {
	return sql.Open("sqlite", dbPath)
}

// OpenSQLiteDatabase validates the sqlite path and returns an open, reachable DB handle.
func OpenSQLiteDatabase(ctx context.Context, dbPath string) (*sql.DB, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	info, err := os.Stat(dbPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("database file does not exist: %s", dbPath)
		}
		return nil, fmt.Errorf("check database path: %w", err)
	}
	if info.IsDir() {
		return nil, fmt.Errorf("database path points to a directory: %s", dbPath)
	}

	db, err := openSQLiteHandle(dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite database: %w", err)
	}

	pingErr := db.PingContext(ctx)
	if pingErr == nil {
		sqlDB, ok := db.(*sql.DB)
		if !ok {
			return nil, fmt.Errorf("open sqlite database: unexpected database handle type %T", db)
		}
		return sqlDB, nil
	}

	closeErr := db.Close()
	if closeErr != nil {
		return nil, errors.Join(
			fmt.Errorf("ping sqlite database: %w", pingErr),
			fmt.Errorf("close sqlite database after ping failure: %w", closeErr),
		)
	}
	return nil, fmt.Errorf("ping sqlite database: %w", pingErr)
}
