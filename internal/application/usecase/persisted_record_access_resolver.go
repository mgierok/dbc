package usecase

import (
	"errors"
	"fmt"
	"strings"

	"github.com/mgierok/dbc/internal/application/dto"
	"github.com/mgierok/dbc/internal/domain/service"
)

var ErrSelectedRecordIdentityExceedsSafeBrowseLimit = errors.New("selected record identity exceeds safe browse limit")
var ErrSelectedCellHasNoSafeEditableSource = errors.New("selected cell has no safe editable source")

type PersistedRecordAccessResolver struct{}

func NewPersistedRecordAccessResolver() *PersistedRecordAccessResolver {
	return &PersistedRecordAccessResolver{}
}

func (r *PersistedRecordAccessResolver) ResolveForDelete(schema dto.Schema, row dto.RecordRow) (dto.PersistedRecordRef, error) {
	if row.IdentityUnavailable {
		return dto.PersistedRecordRef{}, ErrSelectedRecordIdentityExceedsSafeBrowseLimit
	}
	if row.RowKey != "" || len(row.Identity.Keys) > 0 {
		if row.RowKey == "" || len(row.Identity.Keys) == 0 {
			return dto.PersistedRecordRef{}, fmt.Errorf("record identity missing")
		}
		return dto.PersistedRecordRef{
			RowKey:   row.RowKey,
			Identity: row.Identity,
		}, nil
	}

	pkColumns := primaryKeyColumns(schema.Columns)
	if len(pkColumns) == 0 {
		return dto.PersistedRecordRef{}, fmt.Errorf("table has no primary key")
	}

	values := row.Values
	keys := make([]dto.RecordIdentityKey, 0, len(pkColumns))
	parts := make([]string, 0, len(pkColumns))
	for _, pk := range pkColumns {
		if pk.index < 0 || pk.index >= len(values) {
			return dto.PersistedRecordRef{}, fmt.Errorf("primary key index out of range")
		}
		rawValue := values[pk.index]
		isNull := strings.EqualFold(rawValue, "NULL")
		nullable := pk.column.Nullable && !pk.column.PrimaryKey
		parsed, err := service.ParseValue(pk.column.Type, rawValue, isNull, nullable)
		if err != nil {
			return dto.PersistedRecordRef{}, err
		}
		keys = append(keys, dto.RecordIdentityKey{
			Column: pk.column.Name,
			Value: dto.StagedValue{
				IsNull: parsed.IsNull,
				Text:   parsed.Text,
				Raw:    parsed.Raw,
			},
		})
		parts = append(parts, fmt.Sprintf("%s=%s", pk.column.Name, rawValue))
	}

	return dto.PersistedRecordRef{
		RowKey:   strings.Join(parts, "|"),
		Identity: dto.RecordIdentity{Keys: keys},
	}, nil
}

func (r *PersistedRecordAccessResolver) ResolveForEdit(schema dto.Schema, row dto.RecordRow, columnIndex int) (dto.PersistedRecordRef, error) {
	if columnIndex < 0 || columnIndex >= len(schema.Columns) {
		return dto.PersistedRecordRef{}, fmt.Errorf("column index out of range")
	}
	recordRef, err := r.ResolveForDelete(schema, row)
	if err != nil {
		return dto.PersistedRecordRef{}, err
	}
	if len(row.EditableFromBrowse) > 0 {
		if columnIndex >= len(row.EditableFromBrowse) || !row.EditableFromBrowse[columnIndex] {
			return dto.PersistedRecordRef{}, ErrSelectedCellHasNoSafeEditableSource
		}
	}
	return recordRef, nil
}
