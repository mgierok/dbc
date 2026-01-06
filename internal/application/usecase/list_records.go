package usecase

import (
	"context"

	"github.com/mgierok/dbc/internal/application/dto"
	"github.com/mgierok/dbc/internal/application/port"
	"github.com/mgierok/dbc/internal/domain/model"
)

type ListRecords struct {
	engine port.Engine
}

func NewListRecords(engine port.Engine) *ListRecords {
	return &ListRecords{engine: engine}
}

func (uc *ListRecords) Execute(ctx context.Context, tableName string, offset, limit int, filter *dto.Filter) (dto.RecordPage, error) {
	var domainFilter *model.Filter
	if filter != nil {
		domainFilter = &model.Filter{
			Column: filter.Column,
			Operator: model.Operator{
				Name:          filter.Operator.Name,
				SQL:           filter.Operator.SQL,
				RequiresValue: filter.Operator.RequiresValue,
			},
			Value: filter.Value,
		}
	}

	page, err := uc.engine.ListRecords(ctx, tableName, offset, limit, domainFilter)
	if err != nil {
		return dto.RecordPage{}, err
	}

	rows := make([]dto.RecordRow, len(page.Records))
	for i, record := range page.Records {
		values := make([]string, len(record.Values))
		for j, value := range record.Values {
			if value.IsNull {
				values[j] = "NULL"
			} else {
				values[j] = value.Text
			}
		}
		rows[i] = dto.RecordRow{Values: values}
	}

	return dto.RecordPage{Rows: rows, HasMore: page.HasMore}, nil
}
