package usecase

import (
	"context"

	"github.com/mgierok/dbc/internal/application/dto"
	"github.com/mgierok/dbc/internal/application/port"
	"github.com/mgierok/dbc/internal/domain/service"
)

type ListTables struct {
	engine port.Engine
}

func NewListTables(engine port.Engine) *ListTables {
	return &ListTables{engine: engine}
}

func (uc *ListTables) Execute(ctx context.Context) ([]dto.Table, error) {
	tables, err := uc.engine.ListTables(ctx)
	if err != nil {
		return nil, err
	}
	sorted := service.SortedTablesByName(tables)
	result := make([]dto.Table, len(sorted))
	for i, table := range sorted {
		result[i] = dto.Table{Name: table.Name}
	}
	return result, nil
}
