package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/mgierok/dbc/internal/application/dto"
	"github.com/mgierok/dbc/internal/application/port"
	"github.com/mgierok/dbc/internal/domain/model"
)

type SaveDatabaseChanges struct {
	engine port.Engine
}

func NewSaveDatabaseChanges(engine port.Engine) *SaveDatabaseChanges {
	return &SaveDatabaseChanges{engine: engine}
}

func (uc *SaveDatabaseChanges) Execute(ctx context.Context, changes []model.NamedTableChanges) error {
	if len(changes) == 0 {
		return model.ErrMissingTableChanges
	}
	for _, change := range changes {
		if strings.TrimSpace(change.TableName) == "" {
			return fmt.Errorf("table name is required")
		}
		if err := validateTableChanges(change.Changes); err != nil {
			return err
		}
	}
	return uc.engine.ApplyDatabaseChanges(ctx, changes)
}

func (uc *SaveDatabaseChanges) ExecuteDTO(ctx context.Context, changes []dto.NamedTableChanges) error {
	return uc.Execute(ctx, toDomainNamedTableChanges(changes))
}

func toDomainNamedTableChanges(changes []dto.NamedTableChanges) []model.NamedTableChanges {
	mapped := make([]model.NamedTableChanges, 0, len(changes))
	for _, change := range changes {
		mapped = append(mapped, model.NamedTableChanges{
			TableName: change.TableName,
			Changes:   toDomainTableChanges(change.Changes),
		})
	}
	return mapped
}
