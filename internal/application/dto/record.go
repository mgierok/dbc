package dto

type RecordRow struct {
	Values              []string
	EditableFromBrowse  []bool
	RowKey              string
	Identity            RecordIdentity
	IdentityUnavailable bool
}

type RecordPage struct {
	Rows       []RecordRow
	HasMore    bool
	TotalCount int
}
