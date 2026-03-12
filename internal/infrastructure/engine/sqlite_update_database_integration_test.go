package engine

import (
	"context"
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/mgierok/dbc/internal/domain/model"
)

func TestSQLiteEngine_ApplyDatabaseChanges_AppliesMultipleTablesInOneTransaction(t *testing.T) {
	db := setupSQLiteUpdateDB(t, `
		CREATE TABLE users (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL
		);
		CREATE TABLE orders (
			id INTEGER PRIMARY KEY,
			status TEXT NOT NULL
		);
		INSERT INTO users (id, name) VALUES (1, 'alice');
		INSERT INTO orders (id, status) VALUES (5, 'pending');
	`)
	engine := NewSQLiteEngine(db)

	err := engine.ApplyDatabaseChanges(context.Background(), []model.NamedTableChanges{
		{
			TableName: "users",
			Changes: model.TableChanges{
				Updates: []model.RecordUpdate{
					{
						Identity: model.RecordIdentity{
							Keys: []model.RecordIdentityKey{{Column: "id", Value: model.Value{Text: "1", Raw: int64(1)}}},
						},
						Changes: []model.ColumnValue{{Column: "name", Value: model.Value{Text: "bob", Raw: "bob"}}},
					},
				},
			},
		},
		{
			TableName: "orders",
			Changes: model.TableChanges{
				Updates: []model.RecordUpdate{
					{
						Identity: model.RecordIdentity{
							Keys: []model.RecordIdentityKey{{Column: "id", Value: model.Value{Text: "5", Raw: int64(5)}}},
						},
						Changes: []model.ColumnValue{{Column: "status", Value: model.Value{Text: "paid", Raw: "paid"}}},
					},
				},
			},
		},
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var userName string
	if err := db.QueryRow("SELECT name FROM users WHERE id = 1").Scan(&userName); err != nil {
		t.Fatalf("failed to load updated user: %v", err)
	}
	if userName != "bob" {
		t.Fatalf("expected updated user name bob, got %q", userName)
	}

	var orderStatus string
	if err := db.QueryRow("SELECT status FROM orders WHERE id = 5").Scan(&orderStatus); err != nil {
		t.Fatalf("failed to load updated order: %v", err)
	}
	if orderStatus != "paid" {
		t.Fatalf("expected updated order status paid, got %q", orderStatus)
	}
}

func TestSQLiteEngine_ApplyDatabaseChanges_RollsBackAllTablesOnFailure(t *testing.T) {
	db := setupSQLiteUpdateDB(t, `
		CREATE TABLE users (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL
		);
		CREATE TABLE orders (
			id INTEGER PRIMARY KEY,
			status TEXT NOT NULL
		);
		INSERT INTO users (id, name) VALUES (1, 'alice');
		INSERT INTO orders (id, status) VALUES (5, 'pending');
	`)
	engine := NewSQLiteEngine(db)

	err := engine.ApplyDatabaseChanges(context.Background(), []model.NamedTableChanges{
		{
			TableName: "users",
			Changes: model.TableChanges{
				Updates: []model.RecordUpdate{
					{
						Identity: model.RecordIdentity{
							Keys: []model.RecordIdentityKey{{Column: "id", Value: model.Value{Text: "1", Raw: int64(1)}}},
						},
						Changes: []model.ColumnValue{{Column: "name", Value: model.Value{Text: "bob", Raw: "bob"}}},
					},
				},
			},
		},
		{
			TableName: "orders",
			Changes: model.TableChanges{
				Updates: []model.RecordUpdate{
					{
						Identity: model.RecordIdentity{
							Keys: []model.RecordIdentityKey{{Column: "id", Value: model.Value{Text: "5", Raw: int64(5)}}},
						},
						Changes: []model.ColumnValue{{Column: "missing_column", Value: model.Value{Text: "paid", Raw: "paid"}}},
					},
				},
			},
		},
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var userName string
	if err := db.QueryRow("SELECT name FROM users WHERE id = 1").Scan(&userName); err != nil {
		t.Fatalf("failed to load user after rollback: %v", err)
	}
	if userName != "alice" {
		t.Fatalf("expected user rollback to preserve alice, got %q", userName)
	}

	var orderStatus string
	if err := db.QueryRow("SELECT status FROM orders WHERE id = 5").Scan(&orderStatus); err != nil {
		t.Fatalf("failed to load order after rollback: %v", err)
	}
	if orderStatus != "pending" {
		t.Fatalf("expected order rollback to preserve pending, got %q", orderStatus)
	}
}

func TestSQLiteEngine_ApplyDatabaseChanges_ReordersChildFirstInsertBatchForForeignKeys(t *testing.T) {
	db := setupSQLiteUpdateDB(t, `
		CREATE TABLE authors (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL
		);
		CREATE TABLE books (
			id INTEGER PRIMARY KEY,
			author_id INTEGER NOT NULL,
			title TEXT NOT NULL,
			FOREIGN KEY (author_id) REFERENCES authors(id)
		);
	`)
	enableForeignKeys(t, db)
	engine := NewSQLiteEngine(db)

	err := engine.ApplyDatabaseChanges(context.Background(), []model.NamedTableChanges{
		{
			TableName: "books",
			Changes: model.TableChanges{
				Inserts: []model.RecordInsert{
					{
						Values: []model.ColumnValue{
							{Column: "id", Value: model.Value{Text: "1", Raw: int64(1)}},
							{Column: "author_id", Value: model.Value{Text: "10", Raw: int64(10)}},
							{Column: "title", Value: model.Value{Text: "FK-safe save", Raw: "FK-safe save"}},
						},
					},
				},
			},
		},
		{
			TableName: "authors",
			Changes: model.TableChanges{
				Inserts: []model.RecordInsert{
					{
						Values: []model.ColumnValue{
							{Column: "id", Value: model.Value{Text: "10", Raw: int64(10)}},
							{Column: "name", Value: model.Value{Text: "Ada", Raw: "Ada"}},
						},
					},
				},
			},
		},
	})

	if err != nil {
		t.Fatalf("expected FK-safe insert ordering, got %v", err)
	}

	assertCount(t, db, "SELECT COUNT(*) FROM authors", 1)
	assertCount(t, db, "SELECT COUNT(*) FROM books", 1)
}

func TestSQLiteEngine_ApplyDatabaseChanges_ReordersParentFirstDeleteBatchForForeignKeys(t *testing.T) {
	db := setupSQLiteUpdateDB(t, `
		CREATE TABLE authors (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL
		);
		CREATE TABLE books (
			id INTEGER PRIMARY KEY,
			author_id INTEGER NOT NULL,
			title TEXT NOT NULL,
			FOREIGN KEY (author_id) REFERENCES authors(id)
		);
		INSERT INTO authors (id, name) VALUES (10, 'Ada');
		INSERT INTO books (id, author_id, title) VALUES (1, 10, 'FK-safe save');
	`)
	enableForeignKeys(t, db)
	engine := NewSQLiteEngine(db)

	err := engine.ApplyDatabaseChanges(context.Background(), []model.NamedTableChanges{
		{
			TableName: "authors",
			Changes: model.TableChanges{
				Deletes: []model.RecordDelete{
					{
						Identity: model.RecordIdentity{
							Keys: []model.RecordIdentityKey{{Column: "id", Value: model.Value{Text: "10", Raw: int64(10)}}},
						},
					},
				},
			},
		},
		{
			TableName: "books",
			Changes: model.TableChanges{
				Deletes: []model.RecordDelete{
					{
						Identity: model.RecordIdentity{
							Keys: []model.RecordIdentityKey{{Column: "id", Value: model.Value{Text: "1", Raw: int64(1)}}},
						},
					},
				},
			},
		},
	})

	if err != nil {
		t.Fatalf("expected FK-safe delete ordering, got %v", err)
	}

	assertCount(t, db, "SELECT COUNT(*) FROM authors", 0)
	assertCount(t, db, "SELECT COUNT(*) FROM books", 0)
}

func TestSQLiteEngine_ApplyDatabaseChanges_PreservesCallerOrderForRepeatedTableBatches(t *testing.T) {
	db := setupSQLiteUpdateDB(t, `
		CREATE TABLE users (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL
		);
		INSERT INTO users (id, name) VALUES (1, 'alice');
	`)
	engine := NewSQLiteEngine(db)

	err := engine.ApplyDatabaseChanges(context.Background(), []model.NamedTableChanges{
		{
			TableName: "users",
			Changes: model.TableChanges{
				Deletes: []model.RecordDelete{
					{
						Identity: model.RecordIdentity{
							Keys: []model.RecordIdentityKey{{Column: "id", Value: model.Value{Text: "1", Raw: int64(1)}}},
						},
					},
				},
			},
		},
		{
			TableName: "users",
			Changes: model.TableChanges{
				Inserts: []model.RecordInsert{
					{
						Values: []model.ColumnValue{
							{Column: "id", Value: model.Value{Text: "1", Raw: int64(1)}},
							{Column: "name", Value: model.Value{Text: "bob", Raw: "bob"}},
						},
					},
				},
			},
		},
	})

	if err != nil {
		t.Fatalf("expected repeated-table batches to preserve caller order, got %v", err)
	}

	var userName string
	if err := db.QueryRow("SELECT name FROM users WHERE id = 1").Scan(&userName); err != nil {
		t.Fatalf("failed to load replacement user: %v", err)
	}
	if userName != "bob" {
		t.Fatalf("expected replacement user name bob, got %q", userName)
	}
}

func TestSQLiteEngine_ApplyDatabaseChanges_UsesInsertUpdateDeletePhasesAcrossRelatedTables(t *testing.T) {
	db := setupSQLiteUpdateDB(t, `
		CREATE TABLE authors (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL
		);
		CREATE TABLE books (
			id INTEGER PRIMARY KEY,
			author_id INTEGER NOT NULL,
			title TEXT NOT NULL,
			FOREIGN KEY (author_id) REFERENCES authors(id)
		);
		INSERT INTO authors (id, name) VALUES (1, 'Existing');
		INSERT INTO books (id, author_id, title) VALUES (1, 1, 'Old title');
	`)
	enableForeignKeys(t, db)
	engine := NewSQLiteEngine(db)

	err := engine.ApplyDatabaseChanges(context.Background(), []model.NamedTableChanges{
		{
			TableName: "books",
			Changes: model.TableChanges{
				Inserts: []model.RecordInsert{
					{
						Values: []model.ColumnValue{
							{Column: "id", Value: model.Value{Text: "2", Raw: int64(2)}},
							{Column: "author_id", Value: model.Value{Text: "2", Raw: int64(2)}},
							{Column: "title", Value: model.Value{Text: "New title", Raw: "New title"}},
						},
					},
				},
				Updates: []model.RecordUpdate{
					{
						Identity: model.RecordIdentity{
							Keys: []model.RecordIdentityKey{{Column: "id", Value: model.Value{Text: "1", Raw: int64(1)}}},
						},
						Changes: []model.ColumnValue{
							{Column: "author_id", Value: model.Value{Text: "2", Raw: int64(2)}},
						},
					},
				},
			},
		},
		{
			TableName: "authors",
			Changes: model.TableChanges{
				Inserts: []model.RecordInsert{
					{
						Values: []model.ColumnValue{
							{Column: "id", Value: model.Value{Text: "2", Raw: int64(2)}},
							{Column: "name", Value: model.Value{Text: "Replacement", Raw: "Replacement"}},
						},
					},
				},
				Deletes: []model.RecordDelete{
					{
						Identity: model.RecordIdentity{
							Keys: []model.RecordIdentityKey{{Column: "id", Value: model.Value{Text: "1", Raw: int64(1)}}},
						},
					},
				},
			},
		},
	})

	if err != nil {
		t.Fatalf("expected multi-phase save to succeed, got %v", err)
	}

	assertCount(t, db, "SELECT COUNT(*) FROM authors WHERE id = 1", 0)
	assertCount(t, db, "SELECT COUNT(*) FROM authors WHERE id = 2", 1)
	assertCount(t, db, "SELECT COUNT(*) FROM books WHERE author_id = 2", 2)
}

func TestSQLiteEngine_ApplyDatabaseChanges_RollsBackForeignKeyOrderedWorkWhenLaterStatementFails(t *testing.T) {
	db := setupSQLiteUpdateDB(t, `
		CREATE TABLE authors (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL
		);
		CREATE TABLE books (
			id INTEGER PRIMARY KEY,
			author_id INTEGER NOT NULL,
			title TEXT NOT NULL,
			FOREIGN KEY (author_id) REFERENCES authors(id)
		);
		INSERT INTO authors (id, name) VALUES (1, 'Existing');
		INSERT INTO books (id, author_id, title) VALUES (1, 1, 'Old title');
	`)
	enableForeignKeys(t, db)
	engine := NewSQLiteEngine(db)

	err := engine.ApplyDatabaseChanges(context.Background(), []model.NamedTableChanges{
		{
			TableName: "books",
			Changes: model.TableChanges{
				Inserts: []model.RecordInsert{
					{
						Values: []model.ColumnValue{
							{Column: "id", Value: model.Value{Text: "2", Raw: int64(2)}},
							{Column: "author_id", Value: model.Value{Text: "2", Raw: int64(2)}},
							{Column: "title", Value: model.Value{Text: "New title", Raw: "New title"}},
						},
					},
				},
				Updates: []model.RecordUpdate{
					{
						Identity: model.RecordIdentity{
							Keys: []model.RecordIdentityKey{{Column: "id", Value: model.Value{Text: "1", Raw: int64(1)}}},
						},
						Changes: []model.ColumnValue{
							{Column: "missing_column", Value: model.Value{Text: "boom", Raw: "boom"}},
						},
					},
				},
			},
		},
		{
			TableName: "authors",
			Changes: model.TableChanges{
				Inserts: []model.RecordInsert{
					{
						Values: []model.ColumnValue{
							{Column: "id", Value: model.Value{Text: "2", Raw: int64(2)}},
							{Column: "name", Value: model.Value{Text: "Replacement", Raw: "Replacement"}},
						},
					},
				},
			},
		},
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	assertCount(t, db, "SELECT COUNT(*) FROM authors WHERE id = 2", 0)
	assertCount(t, db, "SELECT COUNT(*) FROM books WHERE id = 2", 0)
	assertCount(t, db, "SELECT COUNT(*) FROM authors WHERE id = 1", 1)
	assertCount(t, db, "SELECT COUNT(*) FROM books WHERE id = 1", 1)
}

func enableForeignKeys(t *testing.T, db *sql.DB) {
	t.Helper()
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		t.Fatalf("failed to enable foreign keys: %v", err)
	}
}

func assertCount(t *testing.T, db *sql.DB, query string, expected int) {
	t.Helper()
	var count int
	if err := db.QueryRow(query).Scan(&count); err != nil {
		t.Fatalf("failed to count rows for %q: %v", query, err)
	}
	if count != expected {
		t.Fatalf("expected count %d for %q, got %d", expected, query, count)
	}
}
