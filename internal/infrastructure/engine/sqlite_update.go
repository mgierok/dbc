package engine

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/mgierok/dbc/internal/domain/model"
)

type txExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func (e *SQLiteEngine) ApplyRecordChanges(ctx context.Context, tableName string, changes model.TableChanges) (int, error) {
	if strings.TrimSpace(tableName) == "" {
		return 0, fmt.Errorf("table name is required")
	}
	if len(changes.Inserts) == 0 && len(changes.Updates) == 0 && len(changes.Deletes) == 0 {
		return 0, model.ErrMissingTableChanges
	}

	tx, err := e.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}

	inserted, err := applyRecordInserts(ctx, tx, tableName, changes.Inserts)
	if err != nil {
		return 0, withRollbackError(err, tx.Rollback)
	}
	updated, err := applyRecordUpdates(ctx, tx, tableName, changes.Updates, changes.Deletes)
	if err != nil {
		return 0, withRollbackError(err, tx.Rollback)
	}
	deleted, err := applyRecordDeletes(ctx, tx, tableName, changes.Deletes)
	if err != nil {
		return 0, withRollbackError(err, tx.Rollback)
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return inserted + updated + deleted, nil
}

func applyRecordInserts(ctx context.Context, tx txExecutor, tableName string, inserts []model.RecordInsert) (int, error) {
	total := 0
	for _, insert := range inserts {
		orderedColumns := make([]string, 0, len(insert.Values)+len(insert.ExplicitAutoValues))
		columnValues := make(map[string]model.Value, len(insert.Values)+len(insert.ExplicitAutoValues))

		for _, value := range insert.Values {
			column := strings.TrimSpace(value.Column)
			if column == "" {
				return 0, model.ErrMissingInsertValues
			}
			if _, exists := columnValues[column]; !exists {
				orderedColumns = append(orderedColumns, column)
			}
			columnValues[column] = value.Value
		}
		for _, value := range insert.ExplicitAutoValues {
			column := strings.TrimSpace(value.Column)
			if column == "" {
				return 0, model.ErrMissingInsertValues
			}
			if _, exists := columnValues[column]; !exists {
				orderedColumns = append(orderedColumns, column)
			}
			columnValues[column] = value.Value
		}
		if len(orderedColumns) == 0 {
			return 0, model.ErrMissingInsertValues
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
		affected, err := execAffectedRows(ctx, tx, query, args...)
		if err != nil {
			return 0, err
		}
		total += affected
	}
	return total, nil
}

func applyRecordUpdates(ctx context.Context, tx txExecutor, tableName string, updates []model.RecordUpdate, deletes []model.RecordDelete) (int, error) {
	skippedDeletes := make(map[string]struct{}, len(deletes))
	for _, deleteChange := range deletes {
		signature, err := recordIdentitySignature(deleteChange.Identity)
		if err != nil {
			return 0, err
		}
		skippedDeletes[signature] = struct{}{}
	}

	total := 0
	for _, update := range updates {
		if len(update.Changes) == 0 {
			return 0, model.ErrMissingRecordChanges
		}
		signature, err := recordIdentitySignature(update.Identity)
		if err != nil {
			return 0, err
		}
		if _, skip := skippedDeletes[signature]; skip {
			continue
		}
		whereClause, whereArgs, err := buildRecordIdentityClause(update.Identity)
		if err != nil {
			return 0, err
		}
		setParts := make([]string, 0, len(update.Changes))
		args := make([]any, 0, len(update.Changes)+len(whereArgs))
		for _, change := range update.Changes {
			if strings.TrimSpace(change.Column) == "" {
				return 0, model.ErrMissingRecordChanges
			}
			setParts = append(setParts, fmt.Sprintf("%s = ?", quoteIdentifier(change.Column)))
			args = append(args, bindValue(change.Value))
		}
		args = append(args, whereArgs...)
		query := fmt.Sprintf("UPDATE %s SET %s %s", quoteIdentifier(tableName), strings.Join(setParts, ", "), whereClause)
		affected, err := execAffectedRows(ctx, tx, query, args...)
		if err != nil {
			return 0, err
		}
		total += affected
	}
	return total, nil
}

func applyRecordDeletes(ctx context.Context, tx txExecutor, tableName string, deletes []model.RecordDelete) (int, error) {
	total := 0
	for _, deleteChange := range deletes {
		whereClause, whereArgs, err := buildRecordIdentityClause(deleteChange.Identity)
		if err != nil {
			return 0, err
		}
		query := fmt.Sprintf("DELETE FROM %s %s", quoteIdentifier(tableName), whereClause)
		affected, err := execAffectedRows(ctx, tx, query, whereArgs...)
		if err != nil {
			return 0, err
		}
		total += affected
	}
	return total, nil
}

func recordIdentitySignature(identity model.RecordIdentity) (string, error) {
	clause, args, err := buildRecordIdentityClause(identity)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s|%v", clause, args), nil
}

func buildRecordIdentityClause(identity model.RecordIdentity) (string, []any, error) {
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

func execAffectedRows(ctx context.Context, tx txExecutor, query string, args ...any) (int, error) {
	result, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return int(affected), nil
}

func withRollbackError(cause error, rollback func() error) error {
	if cause == nil {
		return nil
	}
	if rollbackErr := rollback(); rollbackErr != nil {
		return errors.Join(cause, rollbackErr)
	}
	return cause
}
