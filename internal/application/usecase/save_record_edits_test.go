package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/mgierok/dbc/internal/application/usecase"
	"github.com/mgierok/dbc/internal/domain/model"
)

type spyEngine struct {
	updates []model.RecordUpdate
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

func (s *spyEngine) ApplyRecordUpdates(ctx context.Context, tableName string, updates []model.RecordUpdate) error {
	s.updates = updates
	return s.err
}

func TestSaveRecordEdits_DelegatesUpdates(t *testing.T) {
	// Arrange
	engine := &spyEngine{}
	uc := usecase.NewSaveRecordEdits(engine)
	updates := []model.RecordUpdate{
		{
			Identity: model.RecordIdentity{
				Keys: []model.ColumnValue{{Column: "id", Value: model.Value{Text: "1", Raw: int64(1)}}},
			},
			Changes: []model.ColumnValue{{Column: "name", Value: model.Value{Text: "bob", Raw: "bob"}}},
		},
	}

	// Act
	err := uc.Execute(context.Background(), "users", updates)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(engine.updates) != 1 {
		t.Fatalf("expected 1 update, got %d", len(engine.updates))
	}
}

func TestSaveRecordEdits_ReturnsError(t *testing.T) {
	// Arrange
	expectedErr := errors.New("boom")
	engine := &spyEngine{err: expectedErr}
	uc := usecase.NewSaveRecordEdits(engine)

	// Act
	err := uc.Execute(context.Background(), "users", []model.RecordUpdate{
		{
			Identity: model.RecordIdentity{
				Keys: []model.ColumnValue{{Column: "id", Value: model.Value{Text: "1", Raw: int64(1)}}},
			},
			Changes: []model.ColumnValue{{Column: "name", Value: model.Value{Text: "bob", Raw: "bob"}}},
		},
	})

	// Assert
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
}
