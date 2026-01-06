package usecase

import (
	"context"

	"github.com/mgierok/dbc/internal/application/dto"
	"github.com/mgierok/dbc/internal/application/port"
)

type GetSchema struct {
	engine port.Engine
}

func NewGetSchema(engine port.Engine) *GetSchema {
	return &GetSchema{engine: engine}
}

func (uc *GetSchema) Execute(ctx context.Context, tableName string) (dto.Schema, error) {
	schema, err := uc.engine.GetSchema(ctx, tableName)
	if err != nil {
		return dto.Schema{}, err
	}

	columns := make([]dto.SchemaColumn, len(schema.Columns))
	for i, column := range schema.Columns {
		columns[i] = dto.SchemaColumn{Name: column.Name, Type: column.Type}
	}

	return dto.Schema{
		TableName: schema.Table.Name,
		Columns:   columns,
	}, nil
}
