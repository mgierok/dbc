package engine

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"

	_ "modernc.org/sqlite"

	"github.com/mgierok/dbc/internal/application/port"
)

var _ port.DatabaseConnectionChecker = (*SQLiteConnectionChecker)(nil)

type SQLiteConnectionChecker struct{}

func NewSQLiteConnectionChecker() *SQLiteConnectionChecker {
	return &SQLiteConnectionChecker{}
}

func (c *SQLiteConnectionChecker) CanConnect(ctx context.Context, dbPath string) error {
	info, err := os.Stat(dbPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("database file does not exist: %s", dbPath)
		}
		return fmt.Errorf("check database path: %w", err)
	}
	if info.IsDir() {
		return fmt.Errorf("database path points to a directory: %s", dbPath)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("open sqlite database: %w", err)
	}

	pingErr := db.PingContext(ctx)
	closeErr := db.Close()
	if pingErr != nil {
		if closeErr != nil {
			return errors.Join(
				fmt.Errorf("ping sqlite database: %w", pingErr),
				fmt.Errorf("close sqlite database after ping failure: %w", closeErr),
			)
		}
		return fmt.Errorf("ping sqlite database: %w", pingErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close sqlite database: %w", closeErr)
	}

	return nil
}
