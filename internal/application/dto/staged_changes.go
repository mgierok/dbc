package dto

type StagedValue struct {
	IsNull bool
	Text   string
	Raw    any
}

type ColumnValue struct {
	Column string
	Value  StagedValue
}

type RecordIdentityKey struct {
	Column string
	Value  StagedValue
}

type RecordIdentity struct {
	Keys []RecordIdentityKey
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

type NamedTableChanges struct {
	TableName string
	Changes   TableChanges
}

type StagedEdit struct {
	Value StagedValue
}

type PendingInsertRow struct {
	Values       map[int]StagedEdit
	ExplicitAuto map[int]bool
}

type PendingRecordEdits struct {
	Identity RecordIdentity
	Changes  map[int]StagedEdit
}

type PendingRecordDelete struct {
	Identity RecordIdentity
}
