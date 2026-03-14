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

type spyEngine struct {
	tableName string
	changes   model.TableChanges
	count     int
	err       error
}

func (s *spyEngine) ListTables(ctx context.Context) ([]model.Table, error) {
	return nil, nil
}

func (s *spyEngine) GetSchema(ctx context.Context, tableName string) (model.Schema, error) {
	return model.Schema{}, nil
}

func (s *spyEngine) ListRecords(ctx context.Context, tableName string, offset, limit int, filter *model.Filter, sort *model.Sort) (model.RecordPage, error) {
	return model.RecordPage{}, nil
}

func (s *spyEngine) ListOperators(ctx context.Context, columnType string) ([]model.Operator, error) {
	return nil, nil
}

func (s *spyEngine) ApplyRecordChanges(ctx context.Context, tableName string, changes model.TableChanges) (int, error) {
	s.tableName = tableName
	s.changes = changes
	return s.count, s.err
}

func TestSaveTableChanges_DelegatesChangesAndReturnsAppliedRowCount(t *testing.T) {
	// Arrange
	engine := &spyEngine{count: 1}
	uc := usecase.NewSaveTableChanges(engine)
	changes := model.TableChanges{
		Inserts: []model.RecordInsert{
			{
				Values: []model.ColumnValue{{Column: "name", Value: model.Value{Text: "new", Raw: "new"}}},
			},
		},
	}

	// Act
	count, err := uc.Execute(context.Background(), "users", changes)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if count != 1 {
		t.Fatalf("expected applied row count 1, got %d", count)
	}
	if len(engine.changes.Inserts) != 1 {
		t.Fatalf("expected 1 insert, got %d", len(engine.changes.Inserts))
	}
}

func TestSaveTableChanges_ReturnsError(t *testing.T) {
	// Arrange
	expectedErr := errors.New("boom")
	engine := &spyEngine{err: expectedErr}
	uc := usecase.NewSaveTableChanges(engine)

	// Act
	count, err := uc.Execute(context.Background(), "users", model.TableChanges{
		Updates: []model.RecordUpdate{
			{
				Identity: model.RecordIdentity{
					Keys: []model.RecordIdentityKey{{Column: "id", Value: model.Value{Text: "1", Raw: int64(1)}}},
				},
				Changes: []model.ColumnValue{{Column: "name", Value: model.Value{Text: "bob", Raw: "bob"}}},
			},
		},
	})

	// Assert
	if count != 0 {
		t.Fatalf("expected zero count on error, got %d", count)
	}
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
}

func TestSaveTableChanges_ValidatesMissingChanges(t *testing.T) {
	// Arrange
	engine := &spyEngine{}
	uc := usecase.NewSaveTableChanges(engine)

	// Act
	count, err := uc.Execute(context.Background(), "users", model.TableChanges{})

	// Assert
	if count != 0 {
		t.Fatalf("expected zero count for validation failure, got %d", count)
	}
	if !errors.Is(err, model.ErrMissingTableChanges) {
		t.Fatalf("expected error %v, got %v", model.ErrMissingTableChanges, err)
	}
}

func TestSaveTableChanges_ValidatesInsertValues(t *testing.T) {
	// Arrange
	engine := &spyEngine{}
	uc := usecase.NewSaveTableChanges(engine)

	// Act
	count, err := uc.Execute(context.Background(), "users", model.TableChanges{
		Inserts: []model.RecordInsert{{}},
	})

	// Assert
	if count != 0 {
		t.Fatalf("expected zero count for validation failure, got %d", count)
	}
	if !errors.Is(err, model.ErrMissingInsertValues) {
		t.Fatalf("expected error %v, got %v", model.ErrMissingInsertValues, err)
	}
}

func TestSaveTableChanges_ValidatesDeleteIdentity(t *testing.T) {
	// Arrange
	engine := &spyEngine{}
	uc := usecase.NewSaveTableChanges(engine)

	// Act
	count, err := uc.Execute(context.Background(), "users", model.TableChanges{
		Deletes: []model.RecordDelete{{}},
	})

	// Assert
	if count != 0 {
		t.Fatalf("expected zero count for validation failure, got %d", count)
	}
	if !errors.Is(err, model.ErrMissingDeleteIdentity) {
		t.Fatalf("expected error %v, got %v", model.ErrMissingDeleteIdentity, err)
	}
}

func TestSaveTableChanges_ExecuteDTO_DelegatesChangesAndReturnsAppliedRowCount(t *testing.T) {
	// Arrange
	engine := &spyEngine{count: 1}
	uc := usecase.NewSaveTableChanges(engine)
	changes := dto.TableChanges{
		Inserts: []dto.RecordInsert{
			{
				Values: []dto.ColumnValue{{Column: "name", Value: dto.StagedValue{Text: "new", Raw: "new"}}},
			},
		},
	}

	// Act
	count, err := uc.ExecuteDTO(context.Background(), "users", changes)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if count != 1 {
		t.Fatalf("expected applied row count 1, got %d", count)
	}
	if len(engine.changes.Inserts) != 1 {
		t.Fatalf("expected 1 insert, got %d", len(engine.changes.Inserts))
	}
}

func TestSaveTableChanges_ExecuteDTO_MapsUpdateIdentityAndChanges(t *testing.T) {
	// Arrange
	engine := &spyEngine{}
	uc := usecase.NewSaveTableChanges(engine)
	changes := dto.TableChanges{
		Updates: []dto.RecordUpdate{
			{
				Identity: dto.RecordIdentity{
					Keys: []dto.RecordIdentityKey{
						{Column: "id", Value: dto.StagedValue{Text: "1", Raw: int64(1)}},
						{Column: "tenant_id", Value: dto.StagedValue{Text: "NULL", IsNull: true}},
					},
				},
				Changes: []dto.ColumnValue{
					{Column: "name", Value: dto.StagedValue{Text: "alice", Raw: "alice"}},
				},
			},
		},
	}

	// Act
	count, err := uc.ExecuteDTO(context.Background(), "users", changes)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if count != 0 {
		t.Fatalf("expected applied row count 0, got %d", count)
	}

	expected := []model.RecordUpdate{
		{
			Identity: model.RecordIdentity{
				Keys: []model.RecordIdentityKey{
					{Column: "id", Value: model.Value{Text: "1", Raw: int64(1)}},
					{Column: "tenant_id", Value: model.Value{Text: "NULL", IsNull: true}},
				},
			},
			Changes: []model.ColumnValue{
				{Column: "name", Value: model.Value{Text: "alice", Raw: "alice"}},
			},
		},
	}
	if len(engine.changes.Updates) != len(expected) {
		t.Fatalf("expected %d updates, got %d", len(expected), len(engine.changes.Updates))
	}
	for i := range expected {
		if !reflect.DeepEqual(engine.changes.Updates[i], expected[i]) {
			t.Fatalf("expected update %+v, got %+v", expected[i], engine.changes.Updates[i])
		}
	}
}

func TestSaveTableChanges_ExecuteDTO_MapsDeleteIdentity(t *testing.T) {
	// Arrange
	engine := &spyEngine{}
	uc := usecase.NewSaveTableChanges(engine)
	changes := dto.TableChanges{
		Deletes: []dto.RecordDelete{
			{
				Identity: dto.RecordIdentity{
					Keys: []dto.RecordIdentityKey{
						{Column: "id", Value: dto.StagedValue{Text: "7", Raw: int64(7)}},
						{Column: "tenant_id", Value: dto.StagedValue{Text: "NULL", IsNull: true}},
					},
				},
			},
		},
	}

	// Act
	count, err := uc.ExecuteDTO(context.Background(), "users", changes)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if count != 0 {
		t.Fatalf("expected applied row count 0, got %d", count)
	}

	expected := []model.RecordDelete{
		{
			Identity: model.RecordIdentity{
				Keys: []model.RecordIdentityKey{
					{Column: "id", Value: model.Value{Text: "7", Raw: int64(7)}},
					{Column: "tenant_id", Value: model.Value{Text: "NULL", IsNull: true}},
				},
			},
		},
	}
	if len(engine.changes.Deletes) != len(expected) {
		t.Fatalf("expected %d deletes, got %d", len(expected), len(engine.changes.Deletes))
	}
	for i := range expected {
		if !reflect.DeepEqual(engine.changes.Deletes[i], expected[i]) {
			t.Fatalf("expected delete %+v, got %+v", expected[i], engine.changes.Deletes[i])
		}
	}
}

func TestSaveTableChanges_ExecuteDTO_PreservesCombinedPayloadShape(t *testing.T) {
	// Arrange
	engine := &spyEngine{}
	uc := usecase.NewSaveTableChanges(engine)
	changes := dto.TableChanges{
		Updates: []dto.RecordUpdate{
			{
				Identity: dto.RecordIdentity{
					Keys: []dto.RecordIdentityKey{
						{Column: "id", Value: dto.StagedValue{Text: "1", Raw: int64(1)}},
						{Column: "tenant_id", Value: dto.StagedValue{Text: "NULL", IsNull: true}},
					},
				},
				Changes: []dto.ColumnValue{
					{Column: "name", Value: dto.StagedValue{Text: "alice", Raw: "alice"}},
				},
			},
		},
		Deletes: []dto.RecordDelete{
			{
				Identity: dto.RecordIdentity{
					Keys: []dto.RecordIdentityKey{
						{Column: "id", Value: dto.StagedValue{Text: "2", Raw: int64(2)}},
						{Column: "tenant_id", Value: dto.StagedValue{Text: "tenant-a", Raw: "tenant-a"}},
					},
				},
			},
		},
	}

	// Act
	count, err := uc.ExecuteDTO(context.Background(), "users", changes)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if count != 0 {
		t.Fatalf("expected applied row count 0, got %d", count)
	}
	if engine.tableName != "users" {
		t.Fatalf("expected table name users, got %q", engine.tableName)
	}
	if len(engine.changes.Inserts) != 0 {
		t.Fatalf("expected 0 inserts, got %d", len(engine.changes.Inserts))
	}
	if len(engine.changes.Updates) != 1 {
		t.Fatalf("expected 1 update, got %d", len(engine.changes.Updates))
	}
	if len(engine.changes.Deletes) != 1 {
		t.Fatalf("expected 1 delete, got %d", len(engine.changes.Deletes))
	}
	if engine.changes.Updates[0].Identity.Keys[1].Value.Text != "NULL" {
		t.Fatalf("expected null update identity value to be preserved, got %+v", engine.changes.Updates[0].Identity.Keys[1].Value)
	}
	if engine.changes.Updates[0].Changes[0].Value.Raw != "alice" {
		t.Fatalf("expected update raw value alice, got %#v", engine.changes.Updates[0].Changes[0].Value.Raw)
	}
	if engine.changes.Deletes[0].Identity.Keys[1].Value.Raw != "tenant-a" {
		t.Fatalf("expected delete raw value tenant-a, got %#v", engine.changes.Deletes[0].Identity.Keys[1].Value.Raw)
	}
}

func TestSaveTableChanges_ValidatesMissingTableName(t *testing.T) {
	// Arrange
	engine := &spyEngine{}
	uc := usecase.NewSaveTableChanges(engine)

	// Act
	count, err := uc.Execute(context.Background(), "   ", model.TableChanges{
		Inserts: []model.RecordInsert{
			{
				Values: []model.ColumnValue{{Column: "name", Value: model.Value{Text: "alice", Raw: "alice"}}},
			},
		},
	})

	// Assert
	if count != 0 {
		t.Fatalf("expected zero count for missing table name, got %d", count)
	}
	if err == nil {
		t.Fatal("expected error for missing table name")
	}
	if err.Error() != "table name is required" {
		t.Fatalf("expected missing-table-name error, got %v", err)
	}
}

func TestSaveTableChanges_ValidatesUpdateIdentity(t *testing.T) {
	// Arrange
	engine := &spyEngine{}
	uc := usecase.NewSaveTableChanges(engine)

	// Act
	count, err := uc.Execute(context.Background(), "users", model.TableChanges{
		Updates: []model.RecordUpdate{
			{
				Changes: []model.ColumnValue{{Column: "name", Value: model.Value{Text: "bob", Raw: "bob"}}},
			},
		},
	})

	// Assert
	if count != 0 {
		t.Fatalf("expected zero count for validation failure, got %d", count)
	}
	if !errors.Is(err, model.ErrMissingRecordIdentity) {
		t.Fatalf("expected error %v, got %v", model.ErrMissingRecordIdentity, err)
	}
}

func TestSaveTableChanges_ValidatesUpdateChanges(t *testing.T) {
	// Arrange
	engine := &spyEngine{}
	uc := usecase.NewSaveTableChanges(engine)

	// Act
	count, err := uc.Execute(context.Background(), "users", model.TableChanges{
		Updates: []model.RecordUpdate{
			{
				Identity: model.RecordIdentity{
					Keys: []model.RecordIdentityKey{{Column: "id", Value: model.Value{Text: "1", Raw: int64(1)}}},
				},
			},
		},
	})

	// Assert
	if count != 0 {
		t.Fatalf("expected zero count for validation failure, got %d", count)
	}
	if !errors.Is(err, model.ErrMissingRecordChanges) {
		t.Fatalf("expected error %v, got %v", model.ErrMissingRecordChanges, err)
	}
}
