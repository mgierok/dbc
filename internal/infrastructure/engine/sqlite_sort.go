package engine

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/mgierok/dbc/internal/domain/model"
)

var (
	ErrMissingSortColumn    = errors.New("sort column is required")
	ErrUnknownSortDirection = errors.New("unknown sort direction")
	ErrUnknownSortColumn    = errors.New("unknown sort column")
)

func normalizeSortDirection(direction model.SortDirection) (string, error) {
	value := strings.ToUpper(strings.TrimSpace(string(direction)))
	switch value {
	case "":
		return "", ErrUnknownSortDirection
	case string(model.SortDirectionAsc):
		return value, nil
	case string(model.SortDirectionDesc):
		return value, nil
	default:
		return "", ErrUnknownSortDirection
	}
}

func (e *SQLiteEngine) buildSortClause(ctx context.Context, tableName string, sort *model.Sort) (string, error) {
	if sort == nil {
		return "", nil
	}
	column := strings.TrimSpace(sort.Column)
	if column == "" {
		return "", ErrMissingSortColumn
	}

	columns, err := e.tableColumns(ctx, tableName)
	if err != nil {
		return "", err
	}
	normalizedColumn, ok := columns[strings.ToLower(column)]
	if !ok {
		return "", ErrUnknownSortColumn
	}

	direction, err := normalizeSortDirection(sort.Direction)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("ORDER BY %s %s", quoteIdentifier(normalizedColumn), direction), nil
}

func (e *SQLiteEngine) tableColumns(ctx context.Context, tableName string) (columns map[string]string, err error) {
	query := fmt.Sprintf("PRAGMA table_info(%s)", quoteIdentifier(tableName))
	rows, err := e.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	columns = make(map[string]string)
	for rows.Next() {
		var (
			cid     int
			name    string
			typ     string
			notnull int
			dflt    any
			pk      int
		)
		if err := rows.Scan(&cid, &name, &typ, &notnull, &dflt, &pk); err != nil {
			return nil, err
		}
		columns[strings.ToLower(name)] = name
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return columns, nil
}
