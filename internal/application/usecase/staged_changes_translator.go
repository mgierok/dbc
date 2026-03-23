package usecase

import (
	"errors"
	"fmt"
	"strings"

	"github.com/mgierok/dbc/internal/application/dto"
	"github.com/mgierok/dbc/internal/domain/model"
	"github.com/mgierok/dbc/internal/domain/service"
)

type StagedChangesTranslator struct{}

var ErrSelectedRecordIdentityExceedsSafeBrowseLimit = errors.New("selected record identity exceeds safe browse limit")

func NewStagedChangesTranslator() *StagedChangesTranslator {
	return &StagedChangesTranslator{}
}

func (uc *StagedChangesTranslator) ParseStagedValue(column dto.SchemaColumn, input string, isNull bool) (dto.StagedValue, error) {
	parsed, err := service.ParseValue(column.Type, input, isNull, column.Nullable)
	if err != nil {
		return dto.StagedValue{}, err
	}
	return dto.StagedValue{
		IsNull: parsed.IsNull,
		Text:   parsed.Text,
		Raw:    parsed.Raw,
	}, nil
}

func (uc *StagedChangesTranslator) BuildRecordIdentity(schema dto.Schema, row dto.RecordRow) (string, dto.RecordIdentity, error) {
	if row.IdentityUnavailable {
		return "", dto.RecordIdentity{}, ErrSelectedRecordIdentityExceedsSafeBrowseLimit
	}
	if row.RowKey != "" || len(row.Identity.Keys) > 0 {
		if row.RowKey == "" || len(row.Identity.Keys) == 0 {
			return "", dto.RecordIdentity{}, fmt.Errorf("record identity missing")
		}
		return row.RowKey, row.Identity, nil
	}

	pkColumns := primaryKeyColumns(schema.Columns)
	if len(pkColumns) == 0 {
		return "", dto.RecordIdentity{}, fmt.Errorf("table has no primary key")
	}
	values := row.Values
	keys := make([]dto.RecordIdentityKey, 0, len(pkColumns))
	parts := make([]string, 0, len(pkColumns))
	for _, pk := range pkColumns {
		if pk.index < 0 || pk.index >= len(values) {
			return "", dto.RecordIdentity{}, fmt.Errorf("primary key index out of range")
		}
		rawValue := values[pk.index]
		isNull := strings.EqualFold(rawValue, "NULL")
		nullable := pk.column.Nullable && !pk.column.PrimaryKey
		parsed, err := service.ParseValue(pk.column.Type, rawValue, isNull, nullable)
		if err != nil {
			return "", dto.RecordIdentity{}, err
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
	return strings.Join(parts, "|"), dto.RecordIdentity{Keys: keys}, nil
}

func (uc *StagedChangesTranslator) BuildTableChanges(
	schema dto.Schema,
	pendingInserts []dto.PendingInsertRow,
	pendingUpdates map[string]dto.PendingRecordEdits,
	pendingDeletes map[string]dto.PendingRecordDelete,
) (dto.TableChanges, error) {
	changes := dto.TableChanges{}

	for _, row := range pendingInserts {
		insert, err := uc.buildInsertChange(schema.Columns, row)
		if err != nil {
			return dto.TableChanges{}, err
		}
		changes.Inserts = append(changes.Inserts, insert)
	}

	deleteKeys := make(map[string]struct{}, len(pendingDeletes))
	for key, deleteChange := range pendingDeletes {
		deleteKeys[key] = struct{}{}
		changes.Deletes = append(changes.Deletes, dto.RecordDelete(deleteChange))
	}

	for key, edits := range pendingUpdates {
		if _, deleted := deleteKeys[key]; deleted {
			continue
		}
		if len(edits.Changes) == 0 {
			continue
		}
		if len(edits.Identity.Keys) == 0 {
			return dto.TableChanges{}, fmt.Errorf("record identity missing")
		}
		updateChanges := make([]dto.ColumnValue, 0, len(edits.Changes))
		for colIndex, change := range edits.Changes {
			if colIndex < 0 || colIndex >= len(schema.Columns) {
				return dto.TableChanges{}, fmt.Errorf("column index out of range")
			}
			column := schema.Columns[colIndex]
			updateChanges = append(updateChanges, dto.ColumnValue{Column: column.Name, Value: change.Value})
		}
		changes.Updates = append(changes.Updates, dto.RecordUpdate{
			Identity: edits.Identity,
			Changes:  updateChanges,
		})
	}

	return changes, nil
}

type schemaPKColumn struct {
	index  int
	column dto.SchemaColumn
}

func primaryKeyColumns(columns []dto.SchemaColumn) []schemaPKColumn {
	if len(columns) == 0 {
		return nil
	}
	var pkColumns []schemaPKColumn
	for i, column := range columns {
		if column.PrimaryKey {
			pkColumns = append(pkColumns, schemaPKColumn{index: i, column: column})
		}
	}
	return pkColumns
}

func (uc *StagedChangesTranslator) buildInsertChange(columns []dto.SchemaColumn, row dto.PendingInsertRow) (dto.RecordInsert, error) {
	insert := dto.RecordInsert{}
	for colIndex, column := range columns {
		value, ok := row.Values[colIndex]
		if !ok {
			continue
		}
		textValue := stagedDisplayValue(value.Value)
		if !column.Nullable && column.DefaultValue == nil && !column.AutoIncrement && strings.TrimSpace(textValue) == "" {
			return dto.RecordInsert{}, fmt.Errorf("value for column %q is required", column.Name)
		}
		columnValue := dto.ColumnValue{
			Column: column.Name,
			Value:  value.Value,
		}
		if column.AutoIncrement {
			if row.ExplicitAuto[colIndex] {
				insert.ExplicitAutoValues = append(insert.ExplicitAutoValues, columnValue)
			}
			continue
		}
		insert.Values = append(insert.Values, columnValue)
	}
	if len(insert.Values) == 0 && len(insert.ExplicitAutoValues) == 0 {
		return dto.RecordInsert{}, model.ErrMissingInsertValues
	}
	return insert, nil
}

func stagedDisplayValue(value dto.StagedValue) string {
	if value.IsNull {
		return "NULL"
	}
	if strings.TrimSpace(value.Text) != "" {
		return value.Text
	}
	if value.Raw != nil {
		return fmt.Sprint(value.Raw)
	}
	return ""
}
