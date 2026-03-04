package engine

import (
	"strings"

	"github.com/mgierok/dbc/internal/domain/model"
)

type sqliteOperatorSpec struct {
	kind          model.OperatorKind
	name          string
	sql           string
	requiresValue bool
}

var sqliteOperators = []sqliteOperatorSpec{
	{kind: model.OperatorKindEq, name: "Equals", sql: "=", requiresValue: true},
	{kind: model.OperatorKindNeq, name: "Not Equals", sql: "!=", requiresValue: true},
	{kind: model.OperatorKindLt, name: "Less Than", sql: "<", requiresValue: true},
	{kind: model.OperatorKindLte, name: "Less Or Equal", sql: "<=", requiresValue: true},
	{kind: model.OperatorKindGt, name: "Greater Than", sql: ">", requiresValue: true},
	{kind: model.OperatorKindGte, name: "Greater Or Equal", sql: ">=", requiresValue: true},
	{kind: model.OperatorKindLike, name: "Like", sql: "LIKE", requiresValue: true},
	{kind: model.OperatorKindIsNull, name: "Is Null", sql: "IS NULL", requiresValue: false},
	{kind: model.OperatorKindIsNotNull, name: "Is Not Null", sql: "IS NOT NULL", requiresValue: false},
}

func operatorsForType(columnType string) []model.Operator {
	normalized := strings.ToUpper(strings.TrimSpace(columnType))

	operators := make([]model.Operator, 0, len(sqliteOperators))
	for _, operator := range sqliteOperators {
		operators = append(operators, model.Operator{
			Name:          operator.name,
			Kind:          operator.kind,
			RequiresValue: operator.requiresValue,
		})
	}

	if normalized == "" {
		return operators
	}
	return operators
}

func sqliteOperatorSQL(kind model.OperatorKind) (string, bool) {
	for _, operator := range sqliteOperators {
		if operator.kind == kind {
			return operator.sql, true
		}
	}
	return "", false
}
