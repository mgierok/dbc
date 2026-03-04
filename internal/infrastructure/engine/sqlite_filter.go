package engine

import (
	"errors"
	"fmt"
	"strings"

	"github.com/mgierok/dbc/internal/domain/model"
)

var (
	ErrMissingFilterColumn = errors.New("filter column is required")
	ErrUnknownOperator     = errors.New("unknown operator")
)

func buildFilterClause(filter *model.Filter) (string, []any, error) {
	if filter == nil {
		return "", nil, nil
	}
	if strings.TrimSpace(filter.Column) == "" {
		return "", nil, ErrMissingFilterColumn
	}
	operatorSQL, ok := sqliteOperatorSQL(filter.Operator.Kind)
	if !ok {
		return "", nil, ErrUnknownOperator
	}

	clause := fmt.Sprintf("WHERE %s %s", quoteIdentifier(filter.Column), operatorSQL)
	var args []any
	if filter.Operator.RequiresValue {
		clause += " ?"
		args = append(args, filter.Value)
	}
	return clause, args, nil
}

func quoteIdentifier(identifier string) string {
	escaped := strings.ReplaceAll(identifier, `"`, `""`)
	return `"` + escaped + `"`
}
