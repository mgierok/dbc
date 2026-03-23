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

func (uc *ListRecords) Execute(ctx context.Context, tableName string, offset, limit int, filter *dto.Filter, sort *dto.Sort) (dto.RecordPage, error) {
	var domainFilter *model.Filter
	if filter != nil {
		domainFilter = &model.Filter{
			Column: filter.Column,
			Operator: model.Operator{
				Name:          filter.Operator.Name,
				Kind:          model.OperatorKind(filter.Operator.Kind),
				RequiresValue: filter.Operator.RequiresValue,
			},
			Value: filter.Value,
		}
	}

	var domainSort *model.Sort
	if sort != nil {
		domainSort = &model.Sort{
			Column:    sort.Column,
			Direction: model.SortDirection(sort.Direction),
		}
	}

	page, err := uc.engine.ListRecords(ctx, tableName, offset, limit, domainFilter, domainSort)
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
		rows[i] = dto.RecordRow{
			Values:              values,
			EditableFromBrowse:  cloneEditableFromBrowse(record.EditableFromBrowse),
			RowKey:              record.RowKey,
			Identity:            mapRecordIdentityToDTO(record.Identity),
			IdentityUnavailable: record.IdentityUnavailable,
		}
	}

	return dto.RecordPage{
		Rows:       rows,
		HasMore:    page.HasMore,
		TotalCount: page.TotalCount,
	}, nil
}

func mapRecordIdentityToDTO(identity model.RecordIdentity) dto.RecordIdentity {
	if len(identity.Keys) == 0 {
		return dto.RecordIdentity{}
	}
	keys := make([]dto.RecordIdentityKey, len(identity.Keys))
	for i, key := range identity.Keys {
		keys[i] = dto.RecordIdentityKey{
			Column: key.Column,
			Value: dto.StagedValue{
				IsNull: key.Value.IsNull,
				Text:   key.Value.Text,
				Raw:    key.Value.Raw,
			},
		}
	}
	return dto.RecordIdentity{Keys: keys}
}

func cloneEditableFromBrowse(values []bool) []bool {
	if len(values) == 0 {
		return nil
	}
	cloned := make([]bool, len(values))
	copy(cloned, values)
	return cloned
}
