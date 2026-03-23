package model

type Value struct {
	Text   string
	IsNull bool
	Raw    any
}

type Record struct {
	Values              []Value
	EditableFromBrowse  []bool
	RowKey              string
	Identity            RecordIdentity
	IdentityUnavailable bool
}
