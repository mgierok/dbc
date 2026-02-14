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
	if !isAllowedOperator(filter.Operator.SQL) {
		return "", nil, ErrUnknownOperator
	}

	clause := fmt.Sprintf("WHERE %s %s", quoteIdentifier(filter.Column), filter.Operator.SQL)
	var args []any
	if filter.Operator.RequiresValue {
		clause += " ?"
		args = append(args, filter.Value)
	}
	return clause, args, nil
}

func isAllowedOperator(operatorSQL string) bool {
	for _, operator := range sqliteOperators {
		if strings.EqualFold(operator.SQL, operatorSQL) {
			return true
		}
	}
	return false
}

func quoteIdentifier(identifier string) string {
	escaped := strings.ReplaceAll(identifier, `"`, `""`)
	return `"` + escaped + `"`
}
