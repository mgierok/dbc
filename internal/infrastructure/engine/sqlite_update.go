package engine

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/mgierok/dbc/internal/domain/model"
)

type txExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func (e *SQLiteEngine) ApplyRecordChanges(ctx context.Context, tableName string, changes model.TableChanges) error {
	if strings.TrimSpace(tableName) == "" {
		return fmt.Errorf("table name is required")
	}
	if len(changes.Inserts) == 0 && len(changes.Updates) == 0 && len(changes.Deletes) == 0 {
		return model.ErrMissingTableChanges
	}

	tx, err := e.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if err := applyRecordInserts(ctx, tx, tableName, changes.Inserts); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := applyRecordUpdates(ctx, tx, tableName, changes.Updates, changes.Deletes); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := applyRecordDeletes(ctx, tx, tableName, changes.Deletes); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

func applyRecordInserts(ctx context.Context, tx txExecutor, tableName string, inserts []model.RecordInsert) error {
	for _, insert := range inserts {
		orderedColumns := make([]string, 0, len(insert.Values)+len(insert.ExplicitAutoValues))
		columnValues := make(map[string]model.Value, len(insert.Values)+len(insert.ExplicitAutoValues))

		for _, value := range insert.Values {
			column := strings.TrimSpace(value.Column)
			if column == "" {
				return model.ErrMissingInsertValues
			}
			if _, exists := columnValues[column]; !exists {
				orderedColumns = append(orderedColumns, column)
			}
			columnValues[column] = value.Value
		}
		for _, value := range insert.ExplicitAutoValues {
			column := strings.TrimSpace(value.Column)
			if column == "" {
				return model.ErrMissingInsertValues
			}
			if _, exists := columnValues[column]; !exists {
				orderedColumns = append(orderedColumns, column)
			}
			columnValues[column] = value.Value
		}
		if len(orderedColumns) == 0 {
			return model.ErrMissingInsertValues
		}

		columns := make([]string, 0, len(orderedColumns))
		placeholders := make([]string, 0, len(orderedColumns))
		args := make([]any, 0, len(orderedColumns))
		for _, column := range orderedColumns {
			columns = append(columns, quoteIdentifier(column))
			placeholders = append(placeholders, "?")
			args = append(args, bindValue(columnValues[column]))
		}

		query := fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES (%s)",
			quoteIdentifier(tableName),
			strings.Join(columns, ", "),
			strings.Join(placeholders, ", "),
		)
		if _, err := tx.ExecContext(ctx, query, args...); err != nil {
			return err
		}
	}
	return nil
}

func applyRecordUpdates(ctx context.Context, tx txExecutor, tableName string, updates []model.RecordUpdate, deletes []model.RecordDelete) error {
	skippedDeletes := make(map[string]struct{}, len(deletes))
	for _, deleteChange := range deletes {
		signature, err := recordIdentitySignature(deleteChange.Identity)
		if err != nil {
			return err
		}
		skippedDeletes[signature] = struct{}{}
	}

	for _, update := range updates {
		if len(update.Changes) == 0 {
			return model.ErrMissingRecordChanges
		}
		signature, err := recordIdentitySignature(update.Identity)
		if err != nil {
			return err
		}
		if _, skip := skippedDeletes[signature]; skip {
			continue
		}
		whereClause, whereArgs, err := buildRecordIdentityClause(update.Identity)
		if err != nil {
			return err
		}
		setParts := make([]string, 0, len(update.Changes))
		args := make([]any, 0, len(update.Changes)+len(whereArgs))
		for _, change := range update.Changes {
			if strings.TrimSpace(change.Column) == "" {
				return model.ErrMissingRecordChanges
			}
			setParts = append(setParts, fmt.Sprintf("%s = ?", quoteIdentifier(change.Column)))
			args = append(args, bindValue(change.Value))
		}
		args = append(args, whereArgs...)
		query := fmt.Sprintf("UPDATE %s SET %s %s", quoteIdentifier(tableName), strings.Join(setParts, ", "), whereClause)
		if _, err := tx.ExecContext(ctx, query, args...); err != nil {
			return err
		}
	}
	return nil
}

func applyRecordDeletes(ctx context.Context, tx txExecutor, tableName string, deletes []model.RecordDelete) error {
	for _, deleteChange := range deletes {
		whereClause, whereArgs, err := buildRecordIdentityClause(deleteChange.Identity)
		if err != nil {
			return err
		}
		query := fmt.Sprintf("DELETE FROM %s %s", quoteIdentifier(tableName), whereClause)
		if _, err := tx.ExecContext(ctx, query, whereArgs...); err != nil {
			return err
		}
	}
	return nil
}

func recordIdentitySignature(identity model.RecordIdentity) (string, error) {
	clause, args, err := buildRecordIdentityClause(identity)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s|%v", clause, args), nil
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
