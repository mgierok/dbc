package usecase_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/mgierok/dbc/internal/application/dto"
	"github.com/mgierok/dbc/internal/application/usecase"
	"github.com/mgierok/dbc/internal/domain/model"
)

type spyDatabaseEngine struct {
	databaseChanges []model.NamedTableChanges
	err             error
}

func (s *spyDatabaseEngine) ListTables(ctx context.Context) ([]model.Table, error) {
	return nil, nil
}

func (s *spyDatabaseEngine) GetSchema(ctx context.Context, tableName string) (model.Schema, error) {
	return model.Schema{}, nil
}

func (s *spyDatabaseEngine) ListRecords(ctx context.Context, tableName string, offset, limit int, filter *model.Filter, sort *model.Sort) (model.RecordPage, error) {
	return model.RecordPage{}, nil
}

func (s *spyDatabaseEngine) ListOperators(ctx context.Context, columnType string) ([]model.Operator, error) {
	return nil, nil
}

func (s *spyDatabaseEngine) ApplyRecordChanges(ctx context.Context, tableName string, changes model.TableChanges) error {
	return nil
}

func (s *spyDatabaseEngine) ApplyDatabaseChanges(ctx context.Context, changes []model.NamedTableChanges) error {
	s.databaseChanges = append([]model.NamedTableChanges(nil), changes...)
	return s.err
}

func TestSaveDatabaseChanges_DelegatesAllNamedTableBatches(t *testing.T) {
	engine := &spyDatabaseEngine{}
	uc := usecase.NewSaveDatabaseChanges(engine)
	changes := []model.NamedTableChanges{
		{
			TableName: "users",
			Changes: model.TableChanges{
				Inserts: []model.RecordInsert{
					{
						Values: []model.ColumnValue{{Column: "name", Value: model.Value{Text: "alice", Raw: "alice"}}},
					},
				},
			},
		},
		{
			TableName: "orders",
			Changes: model.TableChanges{
				Deletes: []model.RecordDelete{
					{
						Identity: model.RecordIdentity{
							Keys: []model.RecordIdentityKey{{Column: "id", Value: model.Value{Text: "7", Raw: int64(7)}}},
						},
					},
				},
			},
		},
	}

	err := uc.Execute(context.Background(), changes)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !reflect.DeepEqual(engine.databaseChanges, changes) {
		t.Fatalf("expected delegated changes %+v, got %+v", changes, engine.databaseChanges)
	}
}

func TestSaveDatabaseChanges_ReturnsEngineError(t *testing.T) {
	expectedErr := errors.New("boom")
	engine := &spyDatabaseEngine{err: expectedErr}
	uc := usecase.NewSaveDatabaseChanges(engine)

	err := uc.Execute(context.Background(), []model.NamedTableChanges{
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
	})

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
}

func TestSaveDatabaseChanges_ValidatesMissingChanges(t *testing.T) {
	engine := &spyDatabaseEngine{}
	uc := usecase.NewSaveDatabaseChanges(engine)

	err := uc.Execute(context.Background(), nil)

	if !errors.Is(err, model.ErrMissingTableChanges) {
		t.Fatalf("expected error %v, got %v", model.ErrMissingTableChanges, err)
	}
}

func TestSaveDatabaseChanges_ValidatesMissingTableName(t *testing.T) {
	engine := &spyDatabaseEngine{}
	uc := usecase.NewSaveDatabaseChanges(engine)

	err := uc.Execute(context.Background(), []model.NamedTableChanges{
		{
			Changes: model.TableChanges{
				Inserts: []model.RecordInsert{
					{
						Values: []model.ColumnValue{{Column: "name", Value: model.Value{Text: "alice", Raw: "alice"}}},
					},
				},
			},
		},
	})

	if err == nil || err.Error() != "table name is required" {
		t.Fatalf("expected missing table name error, got %v", err)
	}
}

func TestSaveDatabaseChanges_ExecuteDTO_MapsNamedTableBatches(t *testing.T) {
	engine := &spyDatabaseEngine{}
	uc := usecase.NewSaveDatabaseChanges(engine)
	changes := []dto.NamedTableChanges{
		{
			TableName: "users",
			Changes: dto.TableChanges{
				Updates: []dto.RecordUpdate{
					{
						Identity: dto.RecordIdentity{
							Keys: []dto.RecordIdentityKey{
								{Column: "id", Value: dto.StagedValue{Text: "1", Raw: int64(1)}},
							},
						},
						Changes: []dto.ColumnValue{
							{Column: "name", Value: dto.StagedValue{Text: "alice", Raw: "alice"}},
						},
					},
				},
			},
		},
		{
			TableName: "orders",
			Changes: dto.TableChanges{
				Deletes: []dto.RecordDelete{
					{
						Identity: dto.RecordIdentity{
							Keys: []dto.RecordIdentityKey{
								{Column: "id", Value: dto.StagedValue{Text: "9", Raw: int64(9)}},
							},
						},
					},
				},
			},
		},
	}

	err := uc.ExecuteDTO(context.Background(), changes)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(engine.databaseChanges) != 2 {
		t.Fatalf("expected two mapped table changes, got %d", len(engine.databaseChanges))
	}
	if engine.databaseChanges[0].TableName != "users" {
		t.Fatalf("expected first table users, got %q", engine.databaseChanges[0].TableName)
	}
	if len(engine.databaseChanges[0].Changes.Updates) != 1 {
		t.Fatalf("expected one users update, got %d", len(engine.databaseChanges[0].Changes.Updates))
	}
	if engine.databaseChanges[0].Changes.Updates[0].Changes[0].Column != "name" {
		t.Fatalf("expected users update column name, got %q", engine.databaseChanges[0].Changes.Updates[0].Changes[0].Column)
	}
	if engine.databaseChanges[1].TableName != "orders" {
		t.Fatalf("expected second table orders, got %q", engine.databaseChanges[1].TableName)
	}
	if len(engine.databaseChanges[1].Changes.Deletes) != 1 {
		t.Fatalf("expected one orders delete, got %d", len(engine.databaseChanges[1].Changes.Deletes))
	}
}
