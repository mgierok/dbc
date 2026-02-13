package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/mgierok/dbc/internal/application/port"
	"github.com/mgierok/dbc/internal/domain/model"
)

type SaveTableChanges struct {
	engine port.Engine
}

func NewSaveTableChanges(engine port.Engine) *SaveTableChanges {
	return &SaveTableChanges{engine: engine}
}

func (uc *SaveTableChanges) Execute(ctx context.Context, tableName string, changes model.TableChanges) error {
	if strings.TrimSpace(tableName) == "" {
		return fmt.Errorf("table name is required")
	}
	if err := validateTableChanges(changes); err != nil {
		return err
	}
	return uc.engine.ApplyRecordChanges(ctx, tableName, changes)
}

func validateTableChanges(changes model.TableChanges) error {
	if len(changes.Inserts) == 0 && len(changes.Updates) == 0 && len(changes.Deletes) == 0 {
		return model.ErrMissingTableChanges
	}
	for _, insert := range changes.Inserts {
		if len(insert.Values) == 0 && len(insert.ExplicitAutoValues) == 0 {
			return model.ErrMissingInsertValues
		}
	}
	for _, update := range changes.Updates {
		if update.Identity.RowID == nil && len(update.Identity.Keys) == 0 {
			return model.ErrMissingRecordIdentity
		}
		if len(update.Changes) == 0 {
			return model.ErrMissingRecordChanges
		}
	}
	for _, deleteChange := range changes.Deletes {
		if deleteChange.Identity.RowID == nil && len(deleteChange.Identity.Keys) == 0 {
			return model.ErrMissingDeleteIdentity
		}
	}
	return nil
}
