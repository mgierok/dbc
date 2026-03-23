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

type fakeEngine struct {
	tables    []model.Table
	schema    model.Schema
	records   model.RecordPage
	operators []model.Operator

	listTablesErr    error
	getSchemaErr     error
	listRecordsErr   error
	listOperatorsErr error

	lastRecordsTable  string
	lastRecordsOffset int
	lastRecordsLimit  int
	lastRecordsFilter *model.Filter
	lastRecordsSort   *model.Sort
}

func (f *fakeEngine) ListTables(ctx context.Context) ([]model.Table, error) {
	if f.listTablesErr != nil {
		return nil, f.listTablesErr
	}
	return f.tables, nil
}

func (f *fakeEngine) GetSchema(ctx context.Context, tableName string) (model.Schema, error) {
	if f.getSchemaErr != nil {
		return model.Schema{}, f.getSchemaErr
	}
	f.schema.Table = model.Table{Name: tableName}
	return f.schema, nil
}

func (f *fakeEngine) ListRecords(ctx context.Context, tableName string, offset, limit int, filter *model.Filter, sort *model.Sort) (model.RecordPage, error) {
	if f.listRecordsErr != nil {
		return model.RecordPage{}, f.listRecordsErr
	}
	f.lastRecordsTable = tableName
	f.lastRecordsOffset = offset
	f.lastRecordsLimit = limit
	f.lastRecordsFilter = filter
	f.lastRecordsSort = sort
	return f.records, nil
}

func (f *fakeEngine) ListOperators(ctx context.Context, columnType string) ([]model.Operator, error) {
	if f.listOperatorsErr != nil {
		return nil, f.listOperatorsErr
	}
	return f.operators, nil
}

func (f *fakeEngine) ApplyRecordChanges(ctx context.Context, tableName string, changes model.TableChanges) (int, error) {
	return 0, nil
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
	defaultName := "'guest'"
	engine := &fakeEngine{
		schema: model.Schema{
			Columns: []model.Column{
				{Name: "id", Type: "INTEGER", Nullable: false, PrimaryKey: true, AutoIncrement: true},
				{Name: "name", Type: "TEXT", Nullable: true},
				{Name: "display_name", Type: "TEXT", Nullable: false, DefaultValue: &defaultName},
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
		{
			Name:          "id",
			Type:          "INTEGER",
			Nullable:      false,
			PrimaryKey:    true,
			AutoIncrement: true,
			Input:         dto.ColumnInput{Kind: dto.ColumnInputText},
		},
		{
			Name:          "name",
			Type:          "TEXT",
			Nullable:      true,
			AutoIncrement: false,
			Input:         dto.ColumnInput{Kind: dto.ColumnInputText},
		},
		{
			Name:          "display_name",
			Type:          "TEXT",
			Nullable:      false,
			DefaultValue:  &defaultName,
			AutoIncrement: false,
			Input:         dto.ColumnInput{Kind: dto.ColumnInputText},
		},
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
			HasMore:    true,
			TotalCount: 37,
		},
	}
	uc := usecase.NewListRecords(engine)

	// Act
	result, err := uc.Execute(context.Background(), "users", 0, 10, nil, nil)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	expected := dto.RecordPage{
		Rows: []dto.RecordRow{
			{Values: []string{"1", "alice", "NULL"}},
		},
		HasMore:    true,
		TotalCount: 37,
	}
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("expected %v, got %v", expected, result)
	}
}

func TestListRecords_MapsPrecomputedIdentity(t *testing.T) {
	// Arrange
	engine := &fakeEngine{
		records: model.RecordPage{
			Records: []model.Record{
				{
					Values: []model.Value{{Text: "visible"}, {IsNull: true}},
					RowKey: "id=0x0102",
					Identity: model.RecordIdentity{
						Keys: []model.RecordIdentityKey{
							{
								Column: "id",
								Value:  model.Value{Text: "0x0102", Raw: []byte{0x01, 0x02}},
							},
						},
					},
				},
				{
					Values:              []model.Value{{Text: "<truncated 262145 bytes>"}},
					IdentityUnavailable: true,
				},
			},
		},
	}
	uc := usecase.NewListRecords(engine)

	// Act
	result, err := uc.Execute(context.Background(), "records", 0, 10, nil, nil)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	expected := dto.RecordPage{
		Rows: []dto.RecordRow{
			{
				Values: []string{"visible", "NULL"},
				RowKey: "id=0x0102",
				Identity: dto.RecordIdentity{
					Keys: []dto.RecordIdentityKey{
						{
							Column: "id",
							Value:  dto.StagedValue{Text: "0x0102", Raw: []byte{0x01, 0x02}},
						},
					},
				},
			},
			{
				Values:              []string{"<truncated 262145 bytes>"},
				IdentityUnavailable: true,
			},
		},
	}
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("expected %v, got %v", expected, result)
	}
}

func TestListRecords_MapsEditableFromBrowseMetadata(t *testing.T) {
	// Arrange
	engine := &fakeEngine{
		records: model.RecordPage{
			Records: []model.Record{
				{
					Values:             []model.Value{{Text: "alice"}, {Text: "<truncated 262145 bytes>"}},
					EditableFromBrowse: []bool{true, false},
				},
			},
		},
	}
	uc := usecase.NewListRecords(engine)

	// Act
	result, err := uc.Execute(context.Background(), "records", 0, 10, nil, nil)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	expected := dto.RecordPage{
		Rows: []dto.RecordRow{
			{
				Values:             []string{"alice", "<truncated 262145 bytes>"},
				EditableFromBrowse: []bool{true, false},
			},
		},
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
			Kind:          dto.OperatorKindEq,
			RequiresValue: true,
		},
		Value: "alice",
	}

	// Act
	_, err := uc.Execute(context.Background(), "users", 5, 20, filter, nil)

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
			Kind:          model.OperatorKindEq,
			RequiresValue: true,
		},
		Value: "alice",
	}
	if !reflect.DeepEqual(engine.lastRecordsFilter, expectedFilter) {
		t.Fatalf("expected filter %v, got %v", expectedFilter, engine.lastRecordsFilter)
	}
}

func TestListRecords_MapsSort(t *testing.T) {
	// Arrange
	engine := &fakeEngine{}
	uc := usecase.NewListRecords(engine)
	sort := &dto.Sort{
		Column:    "created_at",
		Direction: dto.SortDirectionDesc,
	}

	// Act
	_, err := uc.Execute(context.Background(), "users", 0, 50, nil, sort)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	expectedSort := &model.Sort{
		Column:    "created_at",
		Direction: model.SortDirectionDesc,
	}
	if !reflect.DeepEqual(engine.lastRecordsSort, expectedSort) {
		t.Fatalf("expected sort %v, got %v", expectedSort, engine.lastRecordsSort)
	}
}

func TestListOperators_MapsOperators(t *testing.T) {
	// Arrange
	engine := &fakeEngine{
		operators: []model.Operator{
			{Name: "Equals", Kind: model.OperatorKindEq, RequiresValue: true},
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
		{Name: "Equals", Kind: dto.OperatorKindEq, RequiresValue: true},
	}
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("expected %v, got %v", expected, result)
	}
}

func TestListTables_PropagatesEngineError(t *testing.T) {
	// Arrange
	expectedErr := errors.New("list tables failed")
	engine := &fakeEngine{listTablesErr: expectedErr}
	uc := usecase.NewListTables(engine)

	// Act
	_, err := uc.Execute(context.Background())

	// Assert
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
}

func TestGetSchema_PropagatesEngineError(t *testing.T) {
	// Arrange
	expectedErr := errors.New("get schema failed")
	engine := &fakeEngine{getSchemaErr: expectedErr}
	uc := usecase.NewGetSchema(engine)

	// Act
	_, err := uc.Execute(context.Background(), "users")

	// Assert
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
}

func TestListRecords_PropagatesEngineError(t *testing.T) {
	// Arrange
	expectedErr := errors.New("list records failed")
	engine := &fakeEngine{listRecordsErr: expectedErr}
	uc := usecase.NewListRecords(engine)

	// Act
	_, err := uc.Execute(context.Background(), "users", 0, 10, nil, nil)

	// Assert
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
}

func TestListOperators_PropagatesEngineError(t *testing.T) {
	// Arrange
	expectedErr := errors.New("list operators failed")
	engine := &fakeEngine{listOperatorsErr: expectedErr}
	uc := usecase.NewListOperators(engine)

	// Act
	_, err := uc.Execute(context.Background(), "TEXT")

	// Assert
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
}
