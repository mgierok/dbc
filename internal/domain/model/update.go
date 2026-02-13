package model

import "errors"

var (
	ErrMissingRecordIdentity = errors.New("record identity is required")
	ErrMissingRecordChanges  = errors.New("record changes are required")
	ErrMissingTableChanges   = errors.New("table changes are required")
	ErrMissingInsertValues   = errors.New("insert values are required")
	ErrMissingDeleteIdentity = errors.New("delete identity is required")
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

type RecordInsert struct {
	Values             []ColumnValue
	ExplicitAutoValues []ColumnValue
}

type RecordDelete struct {
	Identity RecordIdentity
}

type TableChanges struct {
	Inserts []RecordInsert
	Updates []RecordUpdate
	Deletes []RecordDelete
}
