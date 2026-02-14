package engine

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/mgierok/dbc/internal/domain/model"
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

func TestSQLiteEngine_ListRecords_AppliesFilterAndPagination(t *testing.T) {
	// Arrange
	db := setupSQLiteSchemaDB(t, `
		CREATE TABLE users (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL
		);
		INSERT INTO users (id, name)
		VALUES (1, 'alice'),
		       (2, 'bob'),
		       (3, 'alice');
	`)
	engine := NewSQLiteEngine(db)
	filter := &model.Filter{
		Column: "name",
		Operator: model.Operator{
			SQL:           "=",
			RequiresValue: true,
		},
		Value: "alice",
	}

	// Act
	page, err := engine.ListRecords(context.Background(), "users", 0, 1, filter)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(page.Records) != 1 {
		t.Fatalf("expected exactly one record in page, got %d", len(page.Records))
	}
	if !page.HasMore {
		t.Fatal("expected hasMore to be true when filtered result exceeds page size")
	}
	if got := page.Records[0].Values[1].Text; got != "alice" {
		t.Fatalf("expected name alice, got %q", got)
	}
}

func TestSQLiteEngine_ListRecords_SupportsQuotedTableName(t *testing.T) {
	// Arrange
	db := setupSQLiteSchemaDB(t, `
		CREATE TABLE "audit""log" (
			id INTEGER PRIMARY KEY,
			note TEXT NOT NULL
		);
		INSERT INTO "audit""log" (id, note)
		VALUES (1, 'entry');
	`)
	engine := NewSQLiteEngine(db)

	// Act
	page, err := engine.ListRecords(context.Background(), `audit"log`, 0, 10, nil)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(page.Records) != 1 {
		t.Fatalf("expected one record, got %d", len(page.Records))
	}
	if page.HasMore {
		t.Fatal("expected hasMore to be false for single row result")
	}
	if got := page.Records[0].Values[1].Text; got != "entry" {
		t.Fatalf("expected note entry, got %q", got)
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
