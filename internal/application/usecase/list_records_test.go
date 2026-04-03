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

func TestListRecords_MapsValues(t *testing.T) {
	t.Parallel()

	engine := &engineStub{
		records: model.RecordPage{
			Records: []model.Record{
				{Values: []model.Value{{Text: "1"}, {Text: "alice"}, {IsNull: true}}},
			},
			HasMore:    true,
			TotalCount: 37,
		},
	}
	uc := usecase.NewListRecords(engine)

	result, err := uc.Execute(context.Background(), "users", 0, 10, nil, nil)

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
	t.Parallel()

	engine := &engineStub{
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

	result, err := uc.Execute(context.Background(), "records", 0, 10, nil, nil)

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
	t.Parallel()

	engine := &engineStub{
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

	result, err := uc.Execute(context.Background(), "records", 0, 10, nil, nil)

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
	t.Parallel()

	engine := &engineStub{}
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

	_, err := uc.Execute(context.Background(), "users", 5, 20, filter, nil)

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
	t.Parallel()

	engine := &engineStub{}
	uc := usecase.NewListRecords(engine)
	sort := &dto.Sort{
		Column:    "created_at",
		Direction: dto.SortDirectionDesc,
	}

	_, err := uc.Execute(context.Background(), "users", 0, 50, nil, sort)

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

func TestListRecords_PropagatesEngineError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("list records failed")
	uc := usecase.NewListRecords(&engineStub{listRecordsErr: expectedErr})

	_, err := uc.Execute(context.Background(), "users", 0, 10, nil, nil)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
}
