package dto

type OperatorKind string

const (
	OperatorKindEq        OperatorKind = "eq"
	OperatorKindNeq       OperatorKind = "neq"
	OperatorKindLt        OperatorKind = "lt"
	OperatorKindLte       OperatorKind = "lte"
	OperatorKindGt        OperatorKind = "gt"
	OperatorKindGte       OperatorKind = "gte"
	OperatorKindLike      OperatorKind = "like"
	OperatorKindIsNull    OperatorKind = "is_null"
	OperatorKindIsNotNull OperatorKind = "is_not_null"
)

type Operator struct {
	Name          string
	Kind          OperatorKind
	RequiresValue bool
}

type Filter struct {
	Column   string
	Operator Operator
	Value    string
}
