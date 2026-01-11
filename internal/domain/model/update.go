package model

import "errors"

var (
	ErrMissingRecordIdentity = errors.New("record identity is required")
	ErrMissingRecordChanges  = errors.New("record changes are required")
)

type ColumnValue struct {
	Column string
	Value  Value
}

type RecordIdentity struct {
	RowID *int64
	Keys  []ColumnValue
}

type RecordUpdate struct {
	Identity RecordIdentity
	Changes  []ColumnValue
}
