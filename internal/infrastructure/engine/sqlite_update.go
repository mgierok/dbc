package engine

import (
	"context"
	"fmt"
	"strings"

	"github.com/mgierok/dbc/internal/domain/model"
)

func (e *SQLiteEngine) ApplyRecordUpdates(ctx context.Context, tableName string, updates []model.RecordUpdate) error {
	if len(updates) == 0 {
		return nil
	}
	if strings.TrimSpace(tableName) == "" {
		return fmt.Errorf("table name is required")
	}

	tx, err := e.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	for _, update := range updates {
		if len(update.Changes) == 0 {
			_ = tx.Rollback()
			return model.ErrMissingRecordChanges
		}
		whereClause, whereArgs, err := buildRecordIdentityClause(update.Identity)
		if err != nil {
			_ = tx.Rollback()
			return err
		}
		setParts := make([]string, 0, len(update.Changes))
		args := make([]any, 0, len(update.Changes)+len(whereArgs))
		for _, change := range update.Changes {
			if strings.TrimSpace(change.Column) == "" {
				_ = tx.Rollback()
				return model.ErrMissingRecordChanges
			}
			setParts = append(setParts, fmt.Sprintf("%s = ?", quoteIdentifier(change.Column)))
			args = append(args, bindValue(change.Value))
		}
		args = append(args, whereArgs...)
		query := fmt.Sprintf("UPDATE %s SET %s %s", quoteIdentifier(tableName), strings.Join(setParts, ", "), whereClause)
		if _, err := tx.ExecContext(ctx, query, args...); err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func buildRecordIdentityClause(identity model.RecordIdentity) (string, []any, error) {
	if identity.RowID != nil {
		return "WHERE rowid = ?", []any{*identity.RowID}, nil
	}
	if len(identity.Keys) == 0 {
		return "", nil, model.ErrMissingRecordIdentity
	}
	parts := make([]string, 0, len(identity.Keys))
	args := make([]any, 0, len(identity.Keys))
	for _, key := range identity.Keys {
		if strings.TrimSpace(key.Column) == "" {
			return "", nil, model.ErrMissingRecordIdentity
		}
		if key.Value.IsNull {
			parts = append(parts, fmt.Sprintf("%s IS NULL", quoteIdentifier(key.Column)))
			continue
		}
		parts = append(parts, fmt.Sprintf("%s = ?", quoteIdentifier(key.Column)))
		args = append(args, bindValue(key.Value))
	}
	return "WHERE " + strings.Join(parts, " AND "), args, nil
}

func bindValue(value model.Value) any {
	if value.IsNull {
		return nil
	}
	if value.Raw != nil {
		return value.Raw
	}
	return value.Text
}
