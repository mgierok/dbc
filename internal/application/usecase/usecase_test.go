package usecase_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/mgierok/dbc/internal/application/dto"
	"github.com/mgierok/dbc/internal/application/usecase"
	"github.com/mgierok/dbc/internal/domain/model"
)

type fakeEngine struct {
	tables    []model.Table
	schema    model.Schema
	records   model.RecordPage
	operators []model.Operator

	lastRecordsTable  string
	lastRecordsOffset int
	lastRecordsLimit  int
	lastRecordsFilter *model.Filter
}

func (f *fakeEngine) ListTables(ctx context.Context) ([]model.Table, error) {
	return f.tables, nil
}

func (f *fakeEngine) GetSchema(ctx context.Context, tableName string) (model.Schema, error) {
	f.schema.Table = model.Table{Name: tableName}
	return f.schema, nil
}

func (f *fakeEngine) ListRecords(ctx context.Context, tableName string, offset, limit int, filter *model.Filter) (model.RecordPage, error) {
	f.lastRecordsTable = tableName
	f.lastRecordsOffset = offset
	f.lastRecordsLimit = limit
	f.lastRecordsFilter = filter
	return f.records, nil
}

func (f *fakeEngine) ListOperators(ctx context.Context, columnType string) ([]model.Operator, error) {
	return f.operators, nil
}

func TestListTables_SortsAlphabetically(t *testing.T) {
	// Arrange
	engine := &fakeEngine{
		tables: []model.Table{
			{Name: "users"},
			{Name: "accounts"},
		},
	}
	uc := usecase.NewListTables(engine)

	// Act
	result, err := uc.Execute(context.Background())

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	expected := []dto.Table{
		{Name: "accounts"},
		{Name: "users"},
	}
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("expected %v, got %v", expected, result)
	}
}

func TestGetSchema_MapsColumns(t *testing.T) {
	// Arrange
	engine := &fakeEngine{
		schema: model.Schema{
			Columns: []model.Column{
				{Name: "id", Type: "INTEGER"},
				{Name: "name", Type: "TEXT"},
			},
		},
	}
	uc := usecase.NewGetSchema(engine)

	// Act
	result, err := uc.Execute(context.Background(), "users")

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.TableName != "users" {
		t.Fatalf("expected table name %q, got %q", "users", result.TableName)
	}
	expectedColumns := []dto.SchemaColumn{
		{Name: "id", Type: "INTEGER"},
		{Name: "name", Type: "TEXT"},
	}
	if !reflect.DeepEqual(result.Columns, expectedColumns) {
		t.Fatalf("expected %v, got %v", expectedColumns, result.Columns)
	}
}

func TestListRecords_MapsValues(t *testing.T) {
	// Arrange
	engine := &fakeEngine{
		records: model.RecordPage{
			Records: []model.Record{
				{Values: []model.Value{{Text: "1"}, {Text: "alice"}, {IsNull: true}}},
			},
			HasMore: true,
		},
	}
	uc := usecase.NewListRecords(engine)

	// Act
	result, err := uc.Execute(context.Background(), "users", 0, 10, nil)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	expected := dto.RecordPage{
		Rows: []dto.RecordRow{
			{Values: []string{"1", "alice", "NULL"}},
		},
		HasMore: true,
	}
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("expected %v, got %v", expected, result)
	}
}

func TestListRecords_MapsFilter(t *testing.T) {
	// Arrange
	engine := &fakeEngine{}
	uc := usecase.NewListRecords(engine)
	filter := &dto.Filter{
		Column: "name",
		Operator: dto.Operator{
			Name:          "Equals",
			SQL:           "=",
			RequiresValue: true,
		},
		Value: "alice",
	}

	// Act
	_, err := uc.Execute(context.Background(), "users", 5, 20, filter)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if engine.lastRecordsTable != "users" {
		t.Fatalf("expected table %q, got %q", "users", engine.lastRecordsTable)
	}
	if engine.lastRecordsOffset != 5 || engine.lastRecordsLimit != 20 {
		t.Fatalf("expected offset 5 and limit 20, got %d and %d", engine.lastRecordsOffset, engine.lastRecordsLimit)
	}
	expectedFilter := &model.Filter{
		Column: "name",
		Operator: model.Operator{
			Name:          "Equals",
			SQL:           "=",
			RequiresValue: true,
		},
		Value: "alice",
	}
	if !reflect.DeepEqual(engine.lastRecordsFilter, expectedFilter) {
		t.Fatalf("expected filter %v, got %v", expectedFilter, engine.lastRecordsFilter)
	}
}

func TestListOperators_MapsOperators(t *testing.T) {
	// Arrange
	engine := &fakeEngine{
		operators: []model.Operator{
			{Name: "Equals", SQL: "=", RequiresValue: true},
		},
	}
	uc := usecase.NewListOperators(engine)

	// Act
	result, err := uc.Execute(context.Background(), "TEXT")

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	expected := []dto.Operator{
		{Name: "Equals", SQL: "=", RequiresValue: true},
	}
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("expected %v, got %v", expected, result)
	}
}
