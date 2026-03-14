package engine

import (
	"context"
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
