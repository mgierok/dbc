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
	tableSQL, err := e.tableDefinitionSQL(ctx, tableName)
	if err != nil {
		return model.Schema{}, err
	}

	query := fmt.Sprintf("PRAGMA table_info(%s)", quoteIdentifier(tableName))
	rows, err := e.db.QueryContext(ctx, query)
	if err != nil {
		return model.Schema{}, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	var columns []model.Column
	for rows.Next() {
		var (
			cid     int
			name    string
			typ     string
			notnull int
			dflt    sql.NullString
			pk      int
		)
		if err := rows.Scan(&cid, &name, &typ, &notnull, &dflt, &pk); err != nil {
			return model.Schema{}, err
		}
		var defaultValue *string
		if dflt.Valid {
			defaultValue = &dflt.String
		}
		autoIncrement := false
		if pk > 0 {
			autoIncrement = columnHasAutoIncrement(tableSQL, name)
		}
		columns = append(columns, model.Column{
			Name:          name,
			Type:          typ,
			Nullable:      notnull == 0,
			PrimaryKey:    pk > 0,
			DefaultValue:  defaultValue,
			AutoIncrement: autoIncrement,
		})
	}
	if err := rows.Err(); err != nil {
		return model.Schema{}, err
	}

	return model.Schema{
		Table:   model.Table{Name: tableName},
		Columns: columns,
	}, nil
}

func (e *SQLiteEngine) ListRecords(ctx context.Context, tableName string, offset, limit int, filter *model.Filter) (page model.RecordPage, err error) {
	if limit <= 0 {
		return model.RecordPage{}, nil
	}
	if offset < 0 {
		offset = 0
	}

	// #nosec G202 -- table name is treated as an SQL identifier and escaped via quoteIdentifier.
	query := "SELECT * FROM " + quoteIdentifier(tableName)
	clause, args, err := buildFilterClause(filter)
	if err != nil {
		return model.RecordPage{}, err
	}
	if clause != "" {
		query = query + " " + clause
	}
	query = query + " LIMIT ? OFFSET ?"
	args = append(args, limit+1, offset)

	rows, err := e.db.QueryContext(ctx, query, args...)
	if err != nil {
		return model.RecordPage{}, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	columns, err := rows.Columns()
	if err != nil {
		return model.RecordPage{}, err
	}
	records := make([]model.Record, 0, limit)
	values := make([]any, len(columns))
	destinations := make([]any, len(columns))
	for i := range values {
		destinations[i] = &values[i]
	}

	for rows.Next() {
		if err := rows.Scan(destinations...); err != nil {
			return model.RecordPage{}, err
		}
		record := model.Record{Values: make([]model.Value, len(values))}
		for i, value := range values {
			if value == nil {
				record.Values[i] = model.Value{IsNull: true}
				continue
			}
			switch typed := value.(type) {
			case []byte:
				record.Values[i] = model.Value{Text: string(typed)}
			default:
				record.Values[i] = model.Value{Text: fmt.Sprint(typed)}
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

	return model.RecordPage{Records: records, HasMore: hasMore}, nil
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
