package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/mgierok/dbc/internal/application/usecase"
	"github.com/mgierok/dbc/internal/domain/model"
)

type spyEngine struct {
	changes model.TableChanges
	err     error
}

func (s *spyEngine) ListTables(ctx context.Context) ([]model.Table, error) {
	return nil, nil
}

func (s *spyEngine) GetSchema(ctx context.Context, tableName string) (model.Schema, error) {
	return model.Schema{}, nil
}

func (s *spyEngine) ListRecords(ctx context.Context, tableName string, offset, limit int, filter *model.Filter) (model.RecordPage, error) {
	return model.RecordPage{}, nil
}

func (s *spyEngine) ListOperators(ctx context.Context, columnType string) ([]model.Operator, error) {
	return nil, nil
}

func (s *spyEngine) ApplyRecordChanges(ctx context.Context, tableName string, changes model.TableChanges) error {
	s.changes = changes
	return s.err
}

func TestSaveTableChanges_DelegatesChanges(t *testing.T) {
	// Arrange
	engine := &spyEngine{}
	uc := usecase.NewSaveTableChanges(engine)
	changes := model.TableChanges{
		Inserts: []model.RecordInsert{
			{
				Values: []model.ColumnValue{{Column: "name", Value: model.Value{Text: "new", Raw: "new"}}},
			},
		},
	}

	// Act
	err := uc.Execute(context.Background(), "users", changes)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
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
	err := uc.Execute(context.Background(), "users", model.TableChanges{
		Updates: []model.RecordUpdate{
			{
				Identity: model.RecordIdentity{
					Keys: []model.ColumnValue{{Column: "id", Value: model.Value{Text: "1", Raw: int64(1)}}},
				},
				Changes: []model.ColumnValue{{Column: "name", Value: model.Value{Text: "bob", Raw: "bob"}}},
			},
		},
	})

	// Assert
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
}

func TestSaveTableChanges_ValidatesMissingChanges(t *testing.T) {
	// Arrange
	engine := &spyEngine{}
	uc := usecase.NewSaveTableChanges(engine)

	// Act
	err := uc.Execute(context.Background(), "users", model.TableChanges{})

	// Assert
	if !errors.Is(err, model.ErrMissingTableChanges) {
		t.Fatalf("expected error %v, got %v", model.ErrMissingTableChanges, err)
	}
}

func TestSaveTableChanges_ValidatesInsertValues(t *testing.T) {
	// Arrange
	engine := &spyEngine{}
	uc := usecase.NewSaveTableChanges(engine)

	// Act
	err := uc.Execute(context.Background(), "users", model.TableChanges{
		Inserts: []model.RecordInsert{{}},
	})

	// Assert
	if !errors.Is(err, model.ErrMissingInsertValues) {
		t.Fatalf("expected error %v, got %v", model.ErrMissingInsertValues, err)
	}
}

func TestSaveTableChanges_ValidatesDeleteIdentity(t *testing.T) {
	// Arrange
	engine := &spyEngine{}
	uc := usecase.NewSaveTableChanges(engine)

	// Act
	err := uc.Execute(context.Background(), "users", model.TableChanges{
		Deletes: []model.RecordDelete{{}},
	})

	// Assert
	if !errors.Is(err, model.ErrMissingDeleteIdentity) {
		t.Fatalf("expected error %v, got %v", model.ErrMissingDeleteIdentity, err)
	}
}
