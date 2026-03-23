package engine

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	"github.com/mgierok/dbc/internal/application/port"
	"github.com/mgierok/dbc/internal/domain/model"
)

type SQLiteEngine struct {
	db *sql.DB
}

var _ port.Engine = (*SQLiteEngine)(nil)

func NewSQLiteEngine(db *sql.DB) *SQLiteEngine {
	return &SQLiteEngine{db: db}
}

func (e *SQLiteEngine) ListTables(ctx context.Context) (tables []model.Table, err error) {
	const query = `
		SELECT name
		FROM sqlite_master
		WHERE type = 'table'
		  AND name NOT LIKE 'sqlite_%'
	`
	rows, err := e.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		tables = append(tables, model.Table{Name: name})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tables, nil
}

func (e *SQLiteEngine) GetSchema(ctx context.Context, tableName string) (schema model.Schema, err error) {
	columnInfos, err := e.tableColumnInfos(ctx, tableName)
	if err != nil {
		return model.Schema{}, err
	}

	columns := make([]model.Column, len(columnInfos))
	for i, column := range columnInfos {
		columns[i] = column.toModelColumn()
	}

	return model.Schema{
		Table:   model.Table{Name: tableName},
		Columns: columns,
	}, nil
}

func (e *SQLiteEngine) ListRecords(ctx context.Context, tableName string, offset, limit int, filter *model.Filter, sort *model.Sort) (page model.RecordPage, err error) {
	if limit <= 0 {
		return model.RecordPage{}, nil
	}
	if offset < 0 {
		offset = 0
	}

	columnInfos, err := e.tableColumnInfos(ctx, tableName)
	if err != nil {
		return model.RecordPage{}, err
	}
	pkColumns := primaryKeyColumnsInOrder(columnInfos)
	selectParts := make([]string, 0, len(columnInfos)*2+len(pkColumns)+1)
	for index, column := range columnInfos {
		selectParts = append(selectParts, displayProjectionForColumn(column, index))
	}
	selectParts = appendEditableProjection(selectParts, columnInfos)
	selectParts = appendIdentityProjection(selectParts, pkColumns)

	var queryBuilder strings.Builder
	queryBuilder.WriteString("SELECT ")
	queryBuilder.WriteString(strings.Join(selectParts, ", "))
	queryBuilder.WriteString(" FROM ")
	queryBuilder.WriteString(quoteIdentifier(tableName))
	query := queryBuilder.String()
	clause, args, err := buildFilterClause(filter)
	if err != nil {
		return model.RecordPage{}, err
	}
	sortClause, err := e.buildSortClause(ctx, tableName, sort)
	if err != nil {
		return model.RecordPage{}, err
	}

	var countQueryBuilder strings.Builder
	countQueryBuilder.WriteString("SELECT COUNT(*) FROM ")
	countQueryBuilder.WriteString(quoteIdentifier(tableName))
	countQuery := countQueryBuilder.String()
	if clause != "" {
		query = query + " " + clause
		countQuery = countQuery + " " + clause
	}
	var totalCount int
	if err := e.db.QueryRowContext(ctx, countQuery, args...).Scan(&totalCount); err != nil {
		return model.RecordPage{}, err
	}

	if sortClause != "" {
		query = query + " " + sortClause
	}
	query += " LIMIT ? OFFSET ?"
	queryArgs := append(append([]any{}, args...), limit+1, offset)

	rows, err := e.db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return model.RecordPage{}, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	records := make([]model.Record, 0, limit)
	displayColumnCount := len(columnInfos)
	editableColumnCount := len(columnInfos)
	identityColumnOffset := displayColumnCount + editableColumnCount
	scanValues := make([]any, identityColumnOffset+len(pkColumns))
	if len(pkColumns) > 0 {
		scanValues = append(scanValues, nil)
	}
	destinations := make([]any, len(scanValues))
	for i := range scanValues {
		destinations[i] = &scanValues[i]
	}

	for rows.Next() {
		if err := rows.Scan(destinations...); err != nil {
			return model.RecordPage{}, err
		}
		record := model.Record{
			Values:             make([]model.Value, len(columnInfos)),
			EditableFromBrowse: make([]bool, len(columnInfos)),
		}
		for i := range columnInfos {
			record.Values[i] = materializeDisplayValue(scanValues[i])
			record.EditableFromBrowse[i] = projectedFlagEnabled(scanValues[displayColumnCount+i])
		}
		if len(pkColumns) > 0 {
			identityFlagIndex := identityColumnOffset + len(pkColumns)
			if identityAvailable(scanValues[identityFlagIndex]) {
				keys := make([]model.RecordIdentityKey, 0, len(pkColumns))
				for i, column := range pkColumns {
					value, err := materializeIdentityValue(column.typ, scanValues[identityColumnOffset+i])
					if err != nil {
						return model.RecordPage{}, err
					}
					keys = append(keys, model.RecordIdentityKey{
						Column: column.name,
						Value:  value,
					})
				}
				record.Identity = model.RecordIdentity{Keys: keys}
				record.RowKey = recordRowKey(record.Identity)
			} else {
				record.IdentityUnavailable = true
			}
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return model.RecordPage{}, err
	}

	hasMore := false
	if len(records) > limit {
		hasMore = true
		records = records[:limit]
	}

	return model.RecordPage{
		Records:    records,
		HasMore:    hasMore,
		TotalCount: totalCount,
	}, nil
}

func (e *SQLiteEngine) ListOperators(ctx context.Context, columnType string) ([]model.Operator, error) {
	return operatorsForType(columnType), nil
}

func (e *SQLiteEngine) tableDefinitionSQL(ctx context.Context, tableName string) (string, error) {
	var tableSQL sql.NullString
	const query = `SELECT sql FROM sqlite_master WHERE type = 'table' AND name = ?`
	if err := e.db.QueryRowContext(ctx, query, tableName).Scan(&tableSQL); err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}
	if !tableSQL.Valid {
		return "", nil
	}
	return tableSQL.String, nil
}

func columnHasAutoIncrement(tableSQL, columnName string) bool {
	if strings.TrimSpace(tableSQL) == "" {
		return false
	}
	if !strings.Contains(strings.ToUpper(tableSQL), "AUTOINCREMENT") {
		return false
	}
	quotedColumn := regexp.QuoteMeta(columnName)
	pattern := fmt.Sprintf("(?is)([\"`\\[]?%s[\"`\\]]?\\s+[^,]*PRIMARY\\s+KEY[^,]*AUTOINCREMENT)", quotedColumn)
	return regexp.MustCompile(pattern).MatchString(tableSQL)
}
