package engine

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/mgierok/dbc/internal/domain/model"
)

type txExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type changePhase int

const (
	changePhaseInsert changePhase = iota
	changePhaseUpdate
	changePhaseDelete
)

type namedTableChangeBatch struct {
	tableName     string
	changes       model.TableChanges
	incomingIndex int
}

type plannedChangeOperation struct {
	batch namedTableChangeBatch
	phase changePhase
}

func (e *SQLiteEngine) ApplyRecordChanges(ctx context.Context, tableName string, changes model.TableChanges) error {
	return e.ApplyDatabaseChanges(ctx, []model.NamedTableChanges{{
		TableName: tableName,
		Changes:   changes,
	}})
}

func (e *SQLiteEngine) ApplyDatabaseChanges(ctx context.Context, changes []model.NamedTableChanges) error {
	if len(changes) == 0 {
		return model.ErrMissingTableChanges
	}

	batches, tableNames, err := buildNamedTableChangeBatches(changes)
	if err != nil {
		return err
	}

	tx, err := e.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	parentsByChild, err := loadDirtyTableForeignKeys(ctx, tx, tableNames)
	if err != nil {
		return withRollbackError(err, tx.Rollback)
	}

	plannedOperations := planBatchChangeOperations(batches, parentsByChild)
	if err := applyPlannedChangeOperations(ctx, tx, plannedOperations); err != nil {
		return withRollbackError(err, tx.Rollback)
	}

	return tx.Commit()
}

func buildNamedTableChangeBatches(changes []model.NamedTableChanges) ([]namedTableChangeBatch, []string, error) {
	batches := make([]namedTableChangeBatch, 0, len(changes))
	tableNames := make([]string, 0, len(changes))
	seenTables := make(map[string]struct{}, len(changes))

	for index, change := range changes {
		tableName := strings.TrimSpace(change.TableName)
		if tableName == "" {
			return nil, nil, fmt.Errorf("table name is required")
		}
		if len(change.Changes.Inserts) == 0 && len(change.Changes.Updates) == 0 && len(change.Changes.Deletes) == 0 {
			return nil, nil, model.ErrMissingTableChanges
		}
		batches = append(batches, namedTableChangeBatch{
			tableName:     tableName,
			changes:       change.Changes,
			incomingIndex: index,
		})
		if _, seen := seenTables[tableName]; seen {
			continue
		}
		seenTables[tableName] = struct{}{}
		tableNames = append(tableNames, tableName)
	}
	return batches, tableNames, nil
}

func loadDirtyTableForeignKeys(ctx context.Context, tx *sql.Tx, tableNames []string) (map[string]map[string]struct{}, error) {
	dirtyTables := make(map[string]struct{}, len(tableNames))
	for _, tableName := range tableNames {
		dirtyTables[tableName] = struct{}{}
	}

	parentsByChild := make(map[string]map[string]struct{}, len(tableNames))
	for _, tableName := range tableNames {
		parentTables, err := loadForeignKeyParentTables(ctx, tx, tableName)
		if err != nil {
			return nil, err
		}

		for _, parentTable := range parentTables {
			if parentTable == tableName {
				continue
			}
			if _, dirty := dirtyTables[parentTable]; !dirty {
				continue
			}
			if parentsByChild[tableName] == nil {
				parentsByChild[tableName] = make(map[string]struct{})
			}
			parentsByChild[tableName][parentTable] = struct{}{}
		}
	}
	return parentsByChild, nil
}

func loadForeignKeyParentTables(ctx context.Context, tx *sql.Tx, tableName string) (parentTables []string, err error) {
	query := fmt.Sprintf("PRAGMA foreign_key_list(%s)", quoteIdentifier(tableName))
	rows, err := tx.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			if err != nil {
				err = errors.Join(err, closeErr)
				return
			}
			err = closeErr
		}
	}()
	return scanForeignKeyParentTables(rows)
}

func scanForeignKeyParentTables(rows *sql.Rows) ([]string, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	if len(columns) < 3 {
		return nil, fmt.Errorf("foreign key metadata must include referenced table column")
	}

	rawValues := make([]sql.RawBytes, len(columns))
	scanTargets := make([]any, len(columns))
	for index := range rawValues {
		scanTargets[index] = &rawValues[index]
	}

	parentTables := make([]string, 0)
	seenParents := make(map[string]struct{})
	for rows.Next() {
		if err := rows.Scan(scanTargets...); err != nil {
			return nil, err
		}
		parentTable := strings.TrimSpace(string(rawValues[2]))
		if parentTable == "" {
			continue
		}
		if _, seen := seenParents[parentTable]; seen {
			continue
		}
		seenParents[parentTable] = struct{}{}
		parentTables = append(parentTables, parentTable)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return parentTables, nil
}

func planBatchChangeOperations(batches []namedTableChangeBatch, parentsByChild map[string]map[string]struct{}) []plannedChangeOperation {
	operations := make([]plannedChangeOperation, 0, len(batches)*3)
	operationIndexesByBatch := make(map[int][]int, len(batches))
	insertIndexesByTable := make(map[string][]int, len(batches))
	updateIndexesByTable := make(map[string][]int, len(batches))
	deleteIndexesByTable := make(map[string][]int, len(batches))
	batchIndexesByTable := make(map[string][]int, len(batches))
	firstOperationByBatch := make(map[int]int, len(batches))
	lastOperationByBatch := make(map[int]int, len(batches))

	for _, batch := range batches {
		batchIndexesByTable[batch.tableName] = append(batchIndexesByTable[batch.tableName], batch.incomingIndex)
		for _, phase := range nonEmptyChangePhases(batch.changes) {
			operationIndex := len(operations)
			operations = append(operations, plannedChangeOperation{
				batch: batch,
				phase: phase,
			})
			operationIndexesByBatch[batch.incomingIndex] = append(operationIndexesByBatch[batch.incomingIndex], operationIndex)
			switch phase {
			case changePhaseInsert:
				insertIndexesByTable[batch.tableName] = append(insertIndexesByTable[batch.tableName], operationIndex)
			case changePhaseUpdate:
				updateIndexesByTable[batch.tableName] = append(updateIndexesByTable[batch.tableName], operationIndex)
			case changePhaseDelete:
				deleteIndexesByTable[batch.tableName] = append(deleteIndexesByTable[batch.tableName], operationIndex)
			}
		}
		indexes := operationIndexesByBatch[batch.incomingIndex]
		firstOperationByBatch[batch.incomingIndex] = indexes[0]
		lastOperationByBatch[batch.incomingIndex] = indexes[len(indexes)-1]
	}

	edges := make([]map[int]struct{}, len(operations))
	indegree := make([]int, len(operations))
	for index := range operations {
		edges[index] = make(map[int]struct{})
	}
	addEdge := func(from, to int) {
		if from == to {
			return
		}
		if _, exists := edges[from][to]; exists {
			return
		}
		edges[from][to] = struct{}{}
		indegree[to]++
	}

	for _, operationIndexes := range operationIndexesByBatch {
		for index := 1; index < len(operationIndexes); index++ {
			addEdge(operationIndexes[index-1], operationIndexes[index])
		}
	}

	for _, batchIndexes := range batchIndexesByTable {
		for index := 1; index < len(batchIndexes); index++ {
			addEdge(lastOperationByBatch[batchIndexes[index-1]], firstOperationByBatch[batchIndexes[index]])
		}
	}

	for childTable, parentTables := range parentsByChild {
		for parentTable := range parentTables {
			for _, parentInsertIndex := range insertIndexesByTable[parentTable] {
				for _, childInsertIndex := range insertIndexesByTable[childTable] {
					addEdge(parentInsertIndex, childInsertIndex)
				}
				for _, childUpdateIndex := range updateIndexesByTable[childTable] {
					addEdge(parentInsertIndex, childUpdateIndex)
				}
			}
			for _, parentDeleteIndex := range deleteIndexesByTable[parentTable] {
				for _, childUpdateIndex := range updateIndexesByTable[childTable] {
					addEdge(childUpdateIndex, parentDeleteIndex)
				}
				for _, childDeleteIndex := range deleteIndexesByTable[childTable] {
					addEdge(childDeleteIndex, parentDeleteIndex)
				}
			}
		}
	}

	available := make([]int, 0, len(operations))
	for index := range operations {
		if indegree[index] == 0 {
			available = append(available, index)
		}
	}
	sortOperationIndexes(available)

	ordered := make([]plannedChangeOperation, 0, len(operations))
	processed := make([]bool, len(operations))
	for len(available) > 0 {
		operationIndex := available[0]
		available = available[1:]
		if processed[operationIndex] {
			continue
		}
		processed[operationIndex] = true
		ordered = append(ordered, operations[operationIndex])
		for dependentIndex := range edges[operationIndex] {
			indegree[dependentIndex]--
			if indegree[dependentIndex] == 0 {
				available = append(available, dependentIndex)
			}
		}
		sortOperationIndexes(available)
	}

	for index, operation := range operations {
		if processed[index] {
			continue
		}
		ordered = append(ordered, operation)
	}
	return ordered
}

func nonEmptyChangePhases(changes model.TableChanges) []changePhase {
	phases := make([]changePhase, 0, 3)
	if len(changes.Inserts) > 0 {
		phases = append(phases, changePhaseInsert)
	}
	if len(changes.Updates) > 0 {
		phases = append(phases, changePhaseUpdate)
	}
	if len(changes.Deletes) > 0 {
		phases = append(phases, changePhaseDelete)
	}
	return phases
}

func sortOperationIndexes(indexes []int) {
	sort.Ints(indexes)
}

func applyPlannedChangeOperations(ctx context.Context, tx *sql.Tx, operations []plannedChangeOperation) error {
	for _, operation := range operations {
		switch operation.phase {
		case changePhaseInsert:
			if err := applyRecordInserts(ctx, tx, operation.batch.tableName, operation.batch.changes.Inserts); err != nil {
				return err
			}
		case changePhaseUpdate:
			if err := applyRecordUpdates(ctx, tx, operation.batch.tableName, operation.batch.changes.Updates, operation.batch.changes.Deletes); err != nil {
				return err
			}
		case changePhaseDelete:
			if err := applyRecordDeletes(ctx, tx, operation.batch.tableName, operation.batch.changes.Deletes); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported change phase")
		}
	}
	return nil
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

func withRollbackError(cause error, rollback func() error) error {
	if cause == nil {
		return nil
	}
	if rollbackErr := rollback(); rollbackErr != nil {
		return errors.Join(cause, rollbackErr)
	}
	return cause
}
