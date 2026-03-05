package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/mgierok/dbc/internal/application/dto"
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

func (uc *SaveTableChanges) ExecuteDTO(ctx context.Context, tableName string, changes dto.TableChanges) error {
	return uc.Execute(ctx, tableName, toDomainTableChanges(changes))
}

func toDomainTableChanges(changes dto.TableChanges) model.TableChanges {
	mapped := model.TableChanges{
		Inserts: make([]model.RecordInsert, 0, len(changes.Inserts)),
		Updates: make([]model.RecordUpdate, 0, len(changes.Updates)),
		Deletes: make([]model.RecordDelete, 0, len(changes.Deletes)),
	}

	for _, insert := range changes.Inserts {
		mappedInsert := model.RecordInsert{
			Values:             toDomainColumnValues(insert.Values),
			ExplicitAutoValues: toDomainColumnValues(insert.ExplicitAutoValues),
		}
		mapped.Inserts = append(mapped.Inserts, mappedInsert)
	}

	for _, update := range changes.Updates {
		mappedUpdate := model.RecordUpdate{
			Identity: toDomainRecordIdentity(update.Identity),
			Changes:  toDomainColumnValues(update.Changes),
		}
		mapped.Updates = append(mapped.Updates, mappedUpdate)
	}

	for _, deleteChange := range changes.Deletes {
		mappedDelete := model.RecordDelete{
			Identity: toDomainRecordIdentity(deleteChange.Identity),
		}
		mapped.Deletes = append(mapped.Deletes, mappedDelete)
	}

	return mapped
}

func toDomainRecordIdentity(identity dto.RecordIdentity) model.RecordIdentity {
	return model.RecordIdentity{
		Keys: toDomainRecordIdentityKeys(identity.Keys),
	}
}

func toDomainRecordIdentityKeys(keys []dto.RecordIdentityKey) []model.RecordIdentityKey {
	mapped := make([]model.RecordIdentityKey, 0, len(keys))
	for _, key := range keys {
		mapped = append(mapped, model.RecordIdentityKey{
			Column: key.Column,
			Value: model.Value{
				IsNull: key.Value.IsNull,
				Text:   key.Value.Text,
				Raw:    key.Value.Raw,
			},
		})
	}
	return mapped
}

func toDomainColumnValues(values []dto.ColumnValue) []model.ColumnValue {
	mapped := make([]model.ColumnValue, 0, len(values))
	for _, value := range values {
		mapped = append(mapped, model.ColumnValue{
			Column: value.Column,
			Value: model.Value{
				IsNull: value.Value.IsNull,
				Text:   value.Value.Text,
				Raw:    value.Value.Raw,
			},
		})
	}
	return mapped
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
		if len(update.Identity.Keys) == 0 {
			return model.ErrMissingRecordIdentity
		}
		if len(update.Changes) == 0 {
			return model.ErrMissingRecordChanges
		}
	}
	for _, deleteChange := range changes.Deletes {
		if len(deleteChange.Identity.Keys) == 0 {
			return model.ErrMissingDeleteIdentity
		}
	}
	return nil
}
