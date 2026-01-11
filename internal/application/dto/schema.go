package dto

type Schema struct {
	TableName string
	Columns   []SchemaColumn
}

type ColumnInputKind string

const (
	ColumnInputText   ColumnInputKind = "text"
	ColumnInputSelect ColumnInputKind = "select"
)

type ColumnInput struct {
	Kind    ColumnInputKind
	Options []string
}

type SchemaColumn struct {
	Name       string
	Type       string
	Nullable   bool
	PrimaryKey bool
	Input      ColumnInput
}
