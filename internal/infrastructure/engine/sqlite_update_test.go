package engine

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/mgierok/dbc/internal/domain/model"
)

func TestSQLiteEngine_ApplyRecordChanges_AppliesInsertUpdateDeleteInOneTransaction(t *testing.T) {
	// Arrange
	db := setupSQLiteUpdateDB(t, `
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			active BOOLEAN NOT NULL,
			score REAL NOT NULL,
			note TEXT
		);
		INSERT INTO users (id, name, active, score, note)
		VALUES (1, 'alice', 1, 1.25, 'hello'),
		       (2, 'carol', 1, 2.10, 'keep');
	`)
	engine := NewSQLiteEngine(db)
	changes := model.TableChanges{
		Inserts: []model.RecordInsert{
			{
				Values: []model.ColumnValue{
					{Column: "name", Value: model.Value{Text: "dan", Raw: "dan"}},
					{Column: "active", Value: model.Value{Text: "true", Raw: int64(1)}},
					{Column: "score", Value: model.Value{Text: "4.5", Raw: 4.5}},
					{Column: "note", Value: model.Value{IsNull: true}},
				},
			},
		},
		Updates: []model.RecordUpdate{
			{
				Identity: model.RecordIdentity{
					Keys: []model.ColumnValue{{Column: "id", Value: model.Value{Text: "1", Raw: int64(1)}}},
				},
				Changes: []model.ColumnValue{{Column: "name", Value: model.Value{Text: "bob", Raw: "bob"}}},
			},
		},
		Deletes: []model.RecordDelete{
			{
				Identity: model.RecordIdentity{
					Keys: []model.ColumnValue{{Column: "id", Value: model.Value{Text: "2", Raw: int64(2)}}},
				},
			},
		},
	}

	// Act
	err := engine.ApplyRecordChanges(context.Background(), "users", changes)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	var (
		countAll   int
		updated    string
		insertedID int64
		inserted   string
	)
	if err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&countAll); err != nil {
		t.Fatalf("failed to count rows: %v", err)
	}
	if countAll != 2 {
		t.Fatalf("expected 2 rows after insert+delete, got %d", countAll)
	}
	if err := db.QueryRow("SELECT name FROM users WHERE id = 1").Scan(&updated); err != nil {
		t.Fatalf("failed to fetch updated row: %v", err)
	}
	if updated != "bob" {
		t.Fatalf("expected updated name bob, got %q", updated)
	}
	if err := db.QueryRow("SELECT id, name FROM users WHERE name = 'dan'").Scan(&insertedID, &inserted); err != nil {
		t.Fatalf("failed to fetch inserted row: %v", err)
	}
	if insertedID <= 2 {
		t.Fatalf("expected inserted id > 2, got %d", insertedID)
	}
	if inserted != "dan" {
		t.Fatalf("expected inserted name dan, got %q", inserted)
	}
}

func TestSQLiteEngine_ApplyRecordChanges_RollsBackOnError(t *testing.T) {
	// Arrange
	db := setupSQLiteUpdateDB(t, `
		CREATE TABLE users (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL
		);
		INSERT INTO users (id, name) VALUES (1, 'alice');
	`)
	engine := NewSQLiteEngine(db)
	changes := model.TableChanges{
		Inserts: []model.RecordInsert{
			{
				Values: []model.ColumnValue{
					{Column: "id", Value: model.Value{Text: "2", Raw: int64(2)}},
					{Column: "name", Value: model.Value{Text: "bob", Raw: "bob"}},
				},
			},
		},
		Updates: []model.RecordUpdate{
			{
				Identity: model.RecordIdentity{
					Keys: []model.ColumnValue{{Column: "id", Value: model.Value{Text: "1", Raw: int64(1)}}},
				},
				Changes: []model.ColumnValue{{Column: "missing_column", Value: model.Value{Text: "oops", Raw: "oops"}}},
			},
		},
	}

	// Act
	err := engine.ApplyRecordChanges(context.Background(), "users", changes)

	// Assert
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var countAll int
	if err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&countAll); err != nil {
		t.Fatalf("failed to count rows: %v", err)
	}
	if countAll != 1 {
		t.Fatalf("expected rollback to keep 1 row, got %d", countAll)
	}
	var name string
	if err := db.QueryRow("SELECT name FROM users WHERE id = 1").Scan(&name); err != nil {
		t.Fatalf("failed to scan row: %v", err)
	}
	if name != "alice" {
		t.Fatalf("expected rollback to keep name %q, got %q", "alice", name)
	}
}

func TestSQLiteEngine_ApplyRecordChanges_DeletesByCompositePrimaryKey(t *testing.T) {
	// Arrange
	db := setupSQLiteUpdateDB(t, `
		CREATE TABLE memberships (
			user_id INTEGER NOT NULL,
			group_id INTEGER NOT NULL,
			role TEXT NOT NULL,
			PRIMARY KEY (user_id, group_id)
		);
		INSERT INTO memberships (user_id, group_id, role)
		VALUES (1, 1, 'owner'), (1, 2, 'viewer');
	`)
	engine := NewSQLiteEngine(db)
	changes := model.TableChanges{
		Deletes: []model.RecordDelete{
			{
				Identity: model.RecordIdentity{
					Keys: []model.ColumnValue{
						{Column: "user_id", Value: model.Value{Text: "1", Raw: int64(1)}},
						{Column: "group_id", Value: model.Value{Text: "2", Raw: int64(2)}},
					},
				},
			},
		},
	}

	// Act
	err := engine.ApplyRecordChanges(context.Background(), "memberships", changes)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	var countAll int
	if err := db.QueryRow("SELECT COUNT(*) FROM memberships").Scan(&countAll); err != nil {
		t.Fatalf("failed to count rows: %v", err)
	}
	if countAll != 1 {
		t.Fatalf("expected one row after delete, got %d", countAll)
	}
}

func TestSQLiteEngine_ApplyRecordChanges_InsertsWithDefaultsAndExplicitAutoIncrement(t *testing.T) {
	// Arrange
	db := setupSQLiteUpdateDB(t, `
		CREATE TABLE events (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL DEFAULT 'anonymous',
			level INTEGER NOT NULL DEFAULT 5
		);
	`)
	engine := NewSQLiteEngine(db)
	changes := model.TableChanges{
		Inserts: []model.RecordInsert{
			{
				Values: []model.ColumnValue{
					{Column: "name", Value: model.Value{Text: "'anonymous'", Raw: "'anonymous'"}},
					{Column: "level", Value: model.Value{Text: "5", Raw: int64(5)}},
				},
			},
			{
				Values: []model.ColumnValue{
					{Column: "name", Value: model.Value{Text: "manual", Raw: "manual"}},
					{Column: "level", Value: model.Value{Text: "7", Raw: int64(7)}},
				},
				ExplicitAutoValues: []model.ColumnValue{
					{Column: "id", Value: model.Value{Text: "10", Raw: int64(10)}},
				},
			},
		},
	}

	// Act
	err := engine.ApplyRecordChanges(context.Background(), "events", changes)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	var (
		firstName  string
		firstLevel int64
		secondName string
		secondID   int64
	)
	if err := db.QueryRow("SELECT name, level FROM events WHERE id = 1").Scan(&firstName, &firstLevel); err != nil {
		t.Fatalf("failed to fetch first row: %v", err)
	}
	if firstName != "'anonymous'" || firstLevel != 5 {
		t.Fatalf("expected first row defaults to be persisted, got name=%q level=%d", firstName, firstLevel)
	}
	if err := db.QueryRow("SELECT id, name FROM events WHERE id = 10").Scan(&secondID, &secondName); err != nil {
		t.Fatalf("failed to fetch explicit-id row: %v", err)
	}
	if secondID != 10 || secondName != "manual" {
		t.Fatalf("expected explicit auto value insert, got id=%d name=%q", secondID, secondName)
	}
}

func TestSQLiteEngine_ApplyRecordChanges_SkipsUpdatesForRowsMarkedDelete(t *testing.T) {
	// Arrange
	db := setupSQLiteUpdateDB(t, `
		CREATE TABLE users (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL
		);
		INSERT INTO users (id, name) VALUES (1, 'alice');
	`)
	engine := NewSQLiteEngine(db)
	changes := model.TableChanges{
		Updates: []model.RecordUpdate{
			{
				Identity: model.RecordIdentity{
					Keys: []model.ColumnValue{{Column: "id", Value: model.Value{Text: "1", Raw: int64(1)}}},
				},
				Changes: []model.ColumnValue{{Column: "missing_column", Value: model.Value{Text: "x", Raw: "x"}}},
			},
		},
		Deletes: []model.RecordDelete{
			{
				Identity: model.RecordIdentity{
					Keys: []model.ColumnValue{{Column: "id", Value: model.Value{Text: "1", Raw: int64(1)}}},
				},
			},
		},
	}

	// Act
	err := engine.ApplyRecordChanges(context.Background(), "users", changes)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	var countAll int
	if err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&countAll); err != nil {
		t.Fatalf("failed to count rows: %v", err)
	}
	if countAll != 0 {
		t.Fatalf("expected deleted row to be removed, got %d rows", countAll)
	}
}

func setupSQLiteUpdateDB(t *testing.T, schema string) *sql.DB {
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
