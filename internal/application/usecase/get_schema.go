package usecase

import (
	"context"

	"github.com/mgierok/dbc/internal/application/dto"
	"github.com/mgierok/dbc/internal/application/port"
	"github.com/mgierok/dbc/internal/domain/service"
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
		inputSpec := service.InputSpecForType(column.Type)
		inputKind := dto.ColumnInputText
		if inputSpec.Kind == service.InputSelect {
			inputKind = dto.ColumnInputSelect
		}
		columns[i] = dto.SchemaColumn{
			Name:          column.Name,
			Type:          column.Type,
			Nullable:      column.Nullable,
			PrimaryKey:    column.PrimaryKey,
			DefaultValue:  column.DefaultValue,
			AutoIncrement: column.AutoIncrement,
			Input: dto.ColumnInput{
				Kind:    inputKind,
				Options: inputSpec.Options,
			},
		}
	}

	return dto.Schema{
		TableName: schema.Table.Name,
		Columns:   columns,
	}, nil
}
