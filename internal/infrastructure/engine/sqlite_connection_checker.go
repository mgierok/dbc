package engine

import (
	"context"
	"fmt"

	"github.com/mgierok/dbc/internal/application/port"
)

var _ port.DatabaseConnectionChecker = (*SQLiteConnectionChecker)(nil)

type SQLiteConnectionChecker struct{}

func NewSQLiteConnectionChecker() *SQLiteConnectionChecker {
	return &SQLiteConnectionChecker{}
}

func (c *SQLiteConnectionChecker) CanConnect(ctx context.Context, dbPath string) error {
	db, err := OpenSQLiteDatabase(ctx, dbPath)
	if err != nil {
		return err
	}

	if closeErr := db.Close(); closeErr != nil {
		return fmt.Errorf("close sqlite database: %w", closeErr)
	}
	return nil
}
