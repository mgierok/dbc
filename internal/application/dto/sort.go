package dto

type SortDirection string

const (
	SortDirectionAsc  SortDirection = "ASC"
	SortDirectionDesc SortDirection = "DESC"
)

type Sort struct {
	Column    string
	Direction SortDirection
}
