package dto

type RecordRow struct {
	Values []string
}

type RecordPage struct {
	Rows    []RecordRow
	HasMore bool
}
