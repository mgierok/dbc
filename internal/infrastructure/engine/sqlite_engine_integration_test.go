package engine

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/mgierok/dbc/internal/domain/model"
)

func TestSQLiteEngine_ListTables_ReturnsUserTablesWithoutSQLiteInternals(t *testing.T) {
	// Arrange
	db := setupSQLiteSchemaDB(t, `
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT
		);
		CREATE TABLE orders (
			id INTEGER PRIMARY KEY
		);
		INSERT INTO users (id) VALUES (NULL);
	`)
	engine := NewSQLiteEngine(db)

	// Act
	tables, err := engine.ListTables(context.Background())

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(tables) != 2 {
		t.Fatalf("expected 2 user tables, got %d", len(tables))
	}
	names := make(map[string]struct{}, len(tables))
	for _, table := range tables {
		names[table.Name] = struct{}{}
	}
	if _, ok := names["users"]; !ok {
		t.Fatalf("expected users table in result, got %+v", tables)
	}
	if _, ok := names["orders"]; !ok {
		t.Fatalf("expected orders table in result, got %+v", tables)
	}
	if _, ok := names["sqlite_sequence"]; ok {
		t.Fatalf("expected sqlite internal tables to be excluded, got %+v", tables)
	}
}

func TestSQLiteEngine_ListOperators_ReturnsSQLiteOperatorContract(t *testing.T) {
	// Arrange
	db := setupSQLiteSchemaDB(t, `
		CREATE TABLE users (
			id INTEGER PRIMARY KEY
		);
	`)
	engine := NewSQLiteEngine(db)

	// Act
	operators, err := engine.ListOperators(context.Background(), "INTEGER")

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	expected := operatorsForType("INTEGER")
	if len(operators) != len(expected) {
		t.Fatalf("expected %d operators, got %d", len(expected), len(operators))
	}
	for i := range expected {
		if operators[i] != expected[i] {
			t.Fatalf("expected operator %+v at index %d, got %+v", expected[i], i, operators[i])
		}
	}
}

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
			Kind:          model.OperatorKindEq,
			RequiresValue: true,
		},
		Value: "alice",
	}

	// Act
	page, err := engine.ListRecords(context.Background(), "users", 0, 1, filter, nil)

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
	if page.TotalCount != 2 {
		t.Fatalf("expected total count 2 for filtered result, got %d", page.TotalCount)
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
	page, err := engine.ListRecords(context.Background(), `audit"log`, 0, 10, nil, nil)

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
	if page.TotalCount != 1 {
		t.Fatalf("expected total count 1, got %d", page.TotalCount)
	}
	if got := page.Records[0].Values[1].Text; got != "entry" {
		t.Fatalf("expected note entry, got %q", got)
	}
}

func TestSQLiteEngine_ListRecords_AppliesSortAscAndDesc(t *testing.T) {
	// Arrange
	db := setupSQLiteSchemaDB(t, `
		CREATE TABLE users (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL
		);
		INSERT INTO users (id, name)
		VALUES (1, 'charlie'),
		       (2, 'alice'),
		       (3, 'bob');
	`)
	engine := NewSQLiteEngine(db)

	// Act
	ascPage, ascErr := engine.ListRecords(context.Background(), "users", 0, 10, nil, &model.Sort{
		Column:    "name",
		Direction: model.SortDirectionAsc,
	})
	descPage, descErr := engine.ListRecords(context.Background(), "users", 0, 10, nil, &model.Sort{
		Column:    "name",
		Direction: model.SortDirectionDesc,
	})

	// Assert
	if ascErr != nil {
		t.Fatalf("expected no asc error, got %v", ascErr)
	}
	if descErr != nil {
		t.Fatalf("expected no desc error, got %v", descErr)
	}
	if got := ascPage.Records[0].Values[1].Text; got != "alice" {
		t.Fatalf("expected asc first value alice, got %q", got)
	}
	if got := descPage.Records[0].Values[1].Text; got != "charlie" {
		t.Fatalf("expected desc first value charlie, got %q", got)
	}
}

func TestSQLiteEngine_ListRecords_RejectsUnknownSortColumn(t *testing.T) {
	// Arrange
	db := setupSQLiteSchemaDB(t, `
		CREATE TABLE users (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL
		);
	`)
	engine := NewSQLiteEngine(db)

	// Act
	_, err := engine.ListRecords(context.Background(), "users", 0, 10, nil, &model.Sort{
		Column:    "missing",
		Direction: model.SortDirectionAsc,
	})

	// Assert
	if !errors.Is(err, ErrUnknownSortColumn) {
		t.Fatalf("expected error %v, got %v", ErrUnknownSortColumn, err)
	}
}

func TestSQLiteEngine_ListRecords_RejectsUnknownSortDirection(t *testing.T) {
	// Arrange
	db := setupSQLiteSchemaDB(t, `
		CREATE TABLE users (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL
		);
	`)
	engine := NewSQLiteEngine(db)

	// Act
	_, err := engine.ListRecords(context.Background(), "users", 0, 10, nil, &model.Sort{
		Column:    "name",
		Direction: model.SortDirection("SIDEWAYS"),
	})

	// Assert
	if !errors.Is(err, ErrUnknownSortDirection) {
		t.Fatalf("expected error %v, got %v", ErrUnknownSortDirection, err)
	}
}

func TestSQLiteEngine_ListRecords_FilterSortAndPagination(t *testing.T) {
	// Arrange
	db := setupSQLiteSchemaDB(t, `
		CREATE TABLE users (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			score INTEGER NOT NULL
		);
		INSERT INTO users (id, name, score)
		VALUES (1, 'a', 30),
		       (2, 'a', 10),
		       (3, 'a', 20),
		       (4, 'b', 5);
	`)
	engine := NewSQLiteEngine(db)
	filter := &model.Filter{
		Column: "name",
		Operator: model.Operator{
			Kind:          model.OperatorKindEq,
			RequiresValue: true,
		},
		Value: "a",
	}
	sort := &model.Sort{
		Column:    "score",
		Direction: model.SortDirectionAsc,
	}

	// Act
	page, err := engine.ListRecords(context.Background(), "users", 1, 1, filter, sort)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(page.Records) != 1 {
		t.Fatalf("expected one record, got %d", len(page.Records))
	}
	if got := page.Records[0].Values[2].Text; got != "20" {
		t.Fatalf("expected sorted paged score 20, got %q", got)
	}
	if !page.HasMore {
		t.Fatal("expected hasMore for remaining filtered sorted rows")
	}
	if page.TotalCount != 3 {
		t.Fatalf("expected total count 3 for filtered result, got %d", page.TotalCount)
	}
}

func TestSQLiteEngine_ListRecords_OffsetBeyondFilteredRange_ReturnsEmptyPageWithTotalCount(t *testing.T) {
	// Arrange
	db := setupSQLiteSchemaDB(t, `
		CREATE TABLE users (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL
		);
		INSERT INTO users (id, name)
		VALUES (1, 'alice'),
		       (2, 'alice'),
		       (3, 'bob');
	`)
	engine := NewSQLiteEngine(db)
	filter := &model.Filter{
		Column: "name",
		Operator: model.Operator{
			Kind:          model.OperatorKindEq,
			RequiresValue: true,
		},
		Value: "alice",
	}

	// Act
	page, err := engine.ListRecords(context.Background(), "users", 20, 10, filter, nil)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(page.Records) != 0 {
		t.Fatalf("expected no records, got %d", len(page.Records))
	}
	if page.HasMore {
		t.Fatal("expected hasMore to be false for empty out-of-range page")
	}
	if page.TotalCount != 2 {
		t.Fatalf("expected total count 2 for filtered result, got %d", page.TotalCount)
	}
}

func BenchmarkSQLiteEngine_ListRecords_PaginatedFilteredSorted(b *testing.B) {
	const rowCount = 10000
	db := setupSQLiteBenchmarkDB(b, rowCount)
	engine := NewSQLiteEngine(db)
	filter := &model.Filter{
		Column: "group_name",
		Operator: model.Operator{
			Kind:          model.OperatorKindEq,
			RequiresValue: true,
		},
		Value: "group-07",
	}
	sort := &model.Sort{
		Column:    "score",
		Direction: model.SortDirectionDesc,
	}
	const (
		offset = 40
		limit  = 20
	)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		page, err := engine.ListRecords(context.Background(), "users", offset, limit, filter, sort)
		if err != nil {
			b.Fatalf("expected no error, got %v", err)
		}
		if page.TotalCount <= 0 {
			b.Fatalf("expected positive total count, got %d", page.TotalCount)
		}
	}
}

func setupSQLiteBenchmarkDB(b *testing.B, rowCount int) *sql.DB {
	b.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", b.Name())
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		b.Fatalf("failed to open sqlite db: %v", err)
	}
	b.Cleanup(func() {
		if err := db.Close(); err != nil {
			b.Fatalf("failed to close db: %v", err)
		}
	})
	schema := `
		CREATE TABLE users (
			id INTEGER PRIMARY KEY,
			group_name TEXT NOT NULL,
			score INTEGER NOT NULL
		);
	`
	if _, err := db.Exec(schema); err != nil {
		b.Fatalf("failed to setup schema: %v", err)
	}
	tx, err := db.Begin()
	if err != nil {
		b.Fatalf("failed to begin transaction: %v", err)
	}
	stmt, err := tx.Prepare("INSERT INTO users (id, group_name, score) VALUES (?, ?, ?)")
	if err != nil {
		b.Fatalf("failed to prepare insert statement: %v", err)
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			b.Fatalf("failed to close insert statement: %v", err)
		}
	}()
	for i := 1; i <= rowCount; i++ {
		group := fmt.Sprintf("group-%02d", i%50)
		score := rowCount - i
		if _, err := stmt.Exec(i, group, score); err != nil {
			b.Fatalf("failed to insert benchmark row: %v", err)
		}
	}
	if err := tx.Commit(); err != nil {
		b.Fatalf("failed to commit benchmark data: %v", err)
	}
	if _, err := db.Exec("CREATE INDEX idx_users_group_name_score ON users(group_name, score DESC)"); err != nil {
		b.Fatalf("failed to create benchmark index: %v", err)
	}
	return db
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
