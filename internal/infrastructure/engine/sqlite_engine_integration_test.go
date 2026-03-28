package engine

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
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

func TestSQLiteEngine_GetSchema_MapsSingleColumnUniqueAndForeignKeys(t *testing.T) {
	// Arrange
	db := setupSQLiteSchemaDB(t, `
		PRAGMA foreign_keys = ON;

		CREATE TABLE accounts (
			id INTEGER PRIMARY KEY
		);

		CREATE TABLE users (
			id INTEGER PRIMARY KEY,
			email TEXT NOT NULL UNIQUE,
			account_id INTEGER,
			FOREIGN KEY (account_id) REFERENCES accounts(id)
		);
	`)
	engine := NewSQLiteEngine(db)

	// Act
	schema, err := engine.GetSchema(context.Background(), "users")

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !schema.Columns[1].Unique {
		t.Fatalf("expected email column to be marked unique")
	}
	expectedRefs := []model.ForeignKeyRef{{Table: "accounts", Column: "id"}}
	if !reflect.DeepEqual(schema.Columns[2].ForeignKeys, expectedRefs) {
		t.Fatalf("expected account_id foreign keys %v, got %v", expectedRefs, schema.Columns[2].ForeignKeys)
	}
}

func TestSQLiteEngine_GetSchema_DoesNotMarkCompositeUniqueMembershipAsPerColumnUnique(t *testing.T) {
	// Arrange
	db := setupSQLiteSchemaDB(t, `
		CREATE TABLE memberships (
			user_id INTEGER NOT NULL,
			role_id INTEGER NOT NULL,
			UNIQUE (user_id, role_id)
		);
	`)
	engine := NewSQLiteEngine(db)

	// Act
	schema, err := engine.GetSchema(context.Background(), "memberships")

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if schema.Columns[0].Unique {
		t.Fatalf("expected user_id not to be marked unique from composite unique index")
	}
	if schema.Columns[1].Unique {
		t.Fatalf("expected role_id not to be marked unique from composite unique index")
	}
}

func TestSQLiteEngine_GetSchema_DoesNotMarkPartialUniqueIndexAsPerColumnUnique(t *testing.T) {
	// Arrange
	db := setupSQLiteSchemaDB(t, `
		CREATE TABLE users (
			id INTEGER PRIMARY KEY,
			email TEXT NOT NULL,
			deleted_at TEXT
		);
		CREATE UNIQUE INDEX users_email_active_idx
			ON users(email)
			WHERE deleted_at IS NULL;
	`)
	engine := NewSQLiteEngine(db)

	// Act
	schema, err := engine.GetSchema(context.Background(), "users")

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if schema.Columns[1].Unique {
		t.Fatalf("expected email not to be marked unique from partial unique index")
	}
}

func TestSQLiteEngine_GetSchema_MapsCompositeForeignKeysPerSourceColumn(t *testing.T) {
	// Arrange
	db := setupSQLiteSchemaDB(t, `
		PRAGMA foreign_keys = ON;

		CREATE TABLE parents (
			parent_id INTEGER NOT NULL,
			parent_code TEXT NOT NULL,
			PRIMARY KEY (parent_id, parent_code)
		);

		CREATE TABLE children (
			child_parent_id INTEGER NOT NULL,
			child_parent_code TEXT NOT NULL,
			FOREIGN KEY (child_parent_id, child_parent_code)
				REFERENCES parents(parent_id, parent_code)
		);
	`)
	engine := NewSQLiteEngine(db)

	// Act
	schema, err := engine.GetSchema(context.Background(), "children")

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got := schema.Columns[0].ForeignKeys; !reflect.DeepEqual(got, []model.ForeignKeyRef{{Table: "parents", Column: "parent_id"}}) {
		t.Fatalf("expected child_parent_id foreign key mapping, got %v", got)
	}
	if got := schema.Columns[1].ForeignKeys; !reflect.DeepEqual(got, []model.ForeignKeyRef{{Table: "parents", Column: "parent_code"}}) {
		t.Fatalf("expected child_parent_code foreign key mapping, got %v", got)
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

func TestSQLiteEngine_ListRecords_SortsIntegersByStoredValue(t *testing.T) {
	// Arrange
	db := setupSQLiteSchemaDB(t, `
		CREATE TABLE scores (
			id INTEGER PRIMARY KEY,
			score INTEGER NOT NULL
		);
		INSERT INTO scores (id, score)
		VALUES (1, 10),
		       (2, 2);
	`)
	engine := NewSQLiteEngine(db)

	// Act
	page, err := engine.ListRecords(context.Background(), "scores", 0, 10, nil, &model.Sort{
		Column:    "score",
		Direction: model.SortDirectionAsc,
	})

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got := recordPageKeys(page); !reflect.DeepEqual(got, []string{"id=2", "id=1"}) {
		t.Fatalf("expected integer sort by stored value, got %v", got)
	}
}

func TestSQLiteEngine_ListRecords_SortsRealsByStoredValue(t *testing.T) {
	// Arrange
	db := setupSQLiteSchemaDB(t, `
		CREATE TABLE metrics (
			id INTEGER PRIMARY KEY,
			score REAL NOT NULL
		);
		INSERT INTO metrics (id, score)
		VALUES (1, 10.5),
		       (2, 2.25);
	`)
	engine := NewSQLiteEngine(db)

	// Act
	page, err := engine.ListRecords(context.Background(), "metrics", 0, 10, nil, &model.Sort{
		Column:    "score",
		Direction: model.SortDirectionAsc,
	})

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got := recordPageKeys(page); !reflect.DeepEqual(got, []string{"id=2", "id=1"}) {
		t.Fatalf("expected real sort by stored value, got %v", got)
	}
}

func TestSQLiteEngine_ListRecords_SortsBlobsByStoredValue(t *testing.T) {
	// Arrange
	db := setupSQLiteSchemaDB(t, `
		CREATE TABLE files (
			id INTEGER PRIMARY KEY,
			payload BLOB NOT NULL
		);
		INSERT INTO files (id, payload)
		VALUES (1, X'02'),
		       (2, X'01');
	`)
	engine := NewSQLiteEngine(db)

	// Act
	page, err := engine.ListRecords(context.Background(), "files", 0, 10, nil, &model.Sort{
		Column:    "payload",
		Direction: model.SortDirectionAsc,
	})

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got := recordPageKeys(page); !reflect.DeepEqual(got, []string{"id=2", "id=1"}) {
		t.Fatalf("expected blob sort by stored value, got %v", got)
	}
}

func TestSQLiteEngine_ListRecords_SortsOversizedTextByStoredValue(t *testing.T) {
	// Arrange
	db := setupSQLiteSchemaDB(t, `
		CREATE TABLE notes (
			id INTEGER PRIMARY KEY,
			note TEXT NOT NULL
		);
	`)
	oversizedA := strings.Repeat("a", 262145)
	oversizedB := strings.Repeat("b", 262145)
	if _, err := db.Exec(`INSERT INTO notes (id, note) VALUES (?, ?)`, 1, oversizedB); err != nil {
		t.Fatalf("failed to insert oversized B note: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO notes (id, note) VALUES (?, ?)`, 2, oversizedA); err != nil {
		t.Fatalf("failed to insert oversized A note: %v", err)
	}
	engine := NewSQLiteEngine(db)

	// Act
	page, err := engine.ListRecords(context.Background(), "notes", 0, 10, nil, &model.Sort{
		Column:    "note",
		Direction: model.SortDirectionAsc,
	})

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got := recordPageKeys(page); !reflect.DeepEqual(got, []string{"id=2", "id=1"}) {
		t.Fatalf("expected oversized text sort by stored value, got %v", got)
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

func TestSQLiteEngine_ListRecords_UsesSafeDisplayMaterialization(t *testing.T) {
	// Arrange
	db := setupSQLiteSchemaDB(t, `
		CREATE TABLE records (
			id INTEGER PRIMARY KEY,
			note TEXT,
			payload BLOB,
			optional_note TEXT
		);
	`)
	oversized := strings.Repeat("a", 262145)
	if _, err := db.Exec(`INSERT INTO records (id, note, payload, optional_note) VALUES (?, ?, X'0102', NULL)`, 1, "short"); err != nil {
		t.Fatalf("failed to insert short record: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO records (id, note, payload, optional_note) VALUES (?, ?, zeroblob(?), NULL)`, 2, oversized, 262145); err != nil {
		t.Fatalf("failed to insert oversized record: %v", err)
	}
	engine := NewSQLiteEngine(db)

	// Act
	page, err := engine.ListRecords(context.Background(), "records", 0, 10, nil, nil)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(page.Records) != 2 {
		t.Fatalf("expected two records, got %d", len(page.Records))
	}
	if got := page.Records[0].Values[1].Text; got != "short" {
		t.Fatalf("expected unchanged short text, got %q", got)
	}
	if got := page.Records[0].Values[2].Text; got != "<blob 2 bytes>" {
		t.Fatalf("expected blob placeholder, got %q", got)
	}
	if !page.Records[0].Values[3].IsNull {
		t.Fatal("expected NULL column to stay NULL")
	}
	if got := page.Records[1].Values[1].Text; got != "<truncated 262145 bytes>" {
		t.Fatalf("expected oversized text placeholder, got %q", got)
	}
	if got := page.Records[1].Values[2].Text; got != "<blob truncated 262145 bytes>" {
		t.Fatalf("expected oversized blob placeholder, got %q", got)
	}
}

func TestSQLiteEngine_ListRecords_PreservesWithinLimitPrimaryKeyIdentity(t *testing.T) {
	// Arrange
	db := setupSQLiteSchemaDB(t, `
		CREATE TABLE files (
			id BLOB PRIMARY KEY,
			name TEXT NOT NULL
		);
		INSERT INTO files (id, name)
		VALUES (X'0102', 'report');
	`)
	engine := NewSQLiteEngine(db)

	// Act
	page, err := engine.ListRecords(context.Background(), "files", 0, 10, nil, nil)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(page.Records) != 1 {
		t.Fatalf("expected one record, got %d", len(page.Records))
	}
	record := page.Records[0]
	if record.IdentityUnavailable {
		t.Fatal("expected identity to stay available within the safe limit")
	}
	if record.RowKey != "id=0x0102" {
		t.Fatalf("expected blob row key, got %q", record.RowKey)
	}
	if len(record.Identity.Keys) != 1 {
		t.Fatalf("expected one identity key, got %d", len(record.Identity.Keys))
	}
	if got := record.Identity.Keys[0].Value.Text; got != "0x0102" {
		t.Fatalf("expected blob identity text, got %q", got)
	}
	typed, ok := record.Identity.Keys[0].Value.Raw.([]byte)
	if !ok || string(typed) != string([]byte{0x01, 0x02}) {
		t.Fatalf("expected blob raw bytes, got %#v", record.Identity.Keys[0].Value.Raw)
	}
}

func TestSQLiteEngine_ListRecords_MarksIdentityUnavailableWhenPrimaryKeyExceedsSafeLimit(t *testing.T) {
	// Arrange
	db := setupSQLiteSchemaDB(t, `
		CREATE TABLE records (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL
		);
	`)
	oversizedID := strings.Repeat("k", 262145)
	if _, err := db.Exec(`INSERT INTO records (id, name) VALUES (?, 'oversized')`, oversizedID); err != nil {
		t.Fatalf("failed to insert oversized identity row: %v", err)
	}
	engine := NewSQLiteEngine(db)

	// Act
	page, err := engine.ListRecords(context.Background(), "records", 0, 10, nil, nil)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(page.Records) != 1 {
		t.Fatalf("expected one record, got %d", len(page.Records))
	}
	record := page.Records[0]
	if got := record.Values[0].Text; got != "<truncated 262145 bytes>" {
		t.Fatalf("expected truncated PK display placeholder, got %q", got)
	}
	if !record.IdentityUnavailable {
		t.Fatal("expected oversized primary key to disable record identity")
	}
	if record.RowKey != "" {
		t.Fatalf("expected empty row key for unavailable identity, got %q", record.RowKey)
	}
	if len(record.Identity.Keys) != 0 {
		t.Fatalf("expected empty identity for unavailable row, got %+v", record.Identity)
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

func recordPageKeys(page model.RecordPage) []string {
	keys := make([]string, len(page.Records))
	for i, record := range page.Records {
		keys[i] = record.RowKey
	}
	return keys
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
