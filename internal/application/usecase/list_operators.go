package usecase

import (
	"context"

	"github.com/mgierok/dbc/internal/application/dto"
	"github.com/mgierok/dbc/internal/application/port"
)

type ListOperators struct {
	engine port.Engine
}

func NewListOperators(engine port.Engine) *ListOperators {
	return &ListOperators{engine: engine}
}

func (uc *ListOperators) Execute(ctx context.Context, columnType string) ([]dto.Operator, error) {
	operators, err := uc.engine.ListOperators(ctx, columnType)
	if err != nil {
		return nil, err
	}
	result := make([]dto.Operator, len(operators))
	for i, operator := range operators {
		result[i] = dto.Operator{
			Name:          operator.Name,
			SQL:           operator.SQL,
			RequiresValue: operator.RequiresValue,
		}
	}
	return result, nil
}
