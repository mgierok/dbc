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

func TestListTables_SortsAlphabetically(t *testing.T) {
	t.Parallel()

	engine := &engineStub{
		tables: []model.Table{
			{Name: "users"},
			{Name: "accounts"},
		},
	}
	uc := usecase.NewListTables(engine)

	result, err := uc.Execute(context.Background())

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

func TestListTables_PropagatesEngineError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("list tables failed")
	uc := usecase.NewListTables(&engineStub{listTablesErr: expectedErr})

	_, err := uc.Execute(context.Background())

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
}
