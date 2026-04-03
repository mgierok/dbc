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

func TestListOperators_MapsOperators(t *testing.T) {
	t.Parallel()

	engine := &engineStub{
		operators: []model.Operator{
			{Name: "Equals", Kind: model.OperatorKindEq, RequiresValue: true},
		},
	}
	uc := usecase.NewListOperators(engine)

	result, err := uc.Execute(context.Background(), "TEXT")

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

func TestListOperators_PropagatesEngineError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("list operators failed")
	uc := usecase.NewListOperators(&engineStub{listOperatorsErr: expectedErr})

	_, err := uc.Execute(context.Background(), "TEXT")

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
}
