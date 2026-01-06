package dto

type Operator struct {
	Name          string
	SQL           string
	RequiresValue bool
}

type Filter struct {
	Column   string
	Operator Operator
	Value    string
}
