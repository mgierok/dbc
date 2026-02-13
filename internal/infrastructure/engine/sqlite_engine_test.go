package engine

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	_ "modernc.org/sqlite"
)

func TestSQLiteEngine_GetSchema_MapsDefaultValuesAndAutoIncrement(t *testing.T) {
	// Arrange
	db := setupSQLiteSchemaDB(t, `
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL DEFAULT 'guest',
			created_at TEXT DEFAULT CURRENT_TIMESTAMP
		);
	`)
	engine := NewSQLiteEngine(db)

	// Act
	schema, err := engine.GetSchema(context.Background(), "users")

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(schema.Columns) != 3 {
		t.Fatalf("expected 3 columns, got %d", len(schema.Columns))
	}
	if !schema.Columns[0].AutoIncrement {
		t.Fatalf("expected id column to be marked auto increment")
	}
	if schema.Columns[0].DefaultValue != nil {
		t.Fatalf("expected id default to be nil")
	}
	if schema.Columns[1].DefaultValue == nil || *schema.Columns[1].DefaultValue != "'guest'" {
		t.Fatalf("expected name default to be 'guest', got %v", schema.Columns[1].DefaultValue)
	}
	if schema.Columns[2].DefaultValue == nil || *schema.Columns[2].DefaultValue != "CURRENT_TIMESTAMP" {
		t.Fatalf("expected created_at default CURRENT_TIMESTAMP, got %v", schema.Columns[2].DefaultValue)
	}
}

func setupSQLiteSchemaDB(t *testing.T, schema string) *sql.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatalf("failed to close db: %v", err)
		}
	})
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to setup schema: %v", err)
	}
	return db
}
