package engine

import (
	"strings"

	"github.com/mgierok/dbc/internal/domain/model"
)

var sqliteOperators = []model.Operator{
	{Name: "Equals", SQL: "=", RequiresValue: true},
	{Name: "Not Equals", SQL: "!=", RequiresValue: true},
	{Name: "Less Than", SQL: "<", RequiresValue: true},
	{Name: "Less Or Equal", SQL: "<=", RequiresValue: true},
	{Name: "Greater Than", SQL: ">", RequiresValue: true},
	{Name: "Greater Or Equal", SQL: ">=", RequiresValue: true},
	{Name: "Like", SQL: "LIKE", RequiresValue: true},
	{Name: "Is Null", SQL: "IS NULL", RequiresValue: false},
	{Name: "Is Not Null", SQL: "IS NOT NULL", RequiresValue: false},
}

func operatorsForType(columnType string) []model.Operator {
	normalized := strings.ToUpper(strings.TrimSpace(columnType))
	if normalized == "" {
		return sqliteOperators
	}
	return sqliteOperators
}
