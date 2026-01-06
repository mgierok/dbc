package dto

type Schema struct {
	TableName string
	Columns   []SchemaColumn
}

type SchemaColumn struct {
	Name string
	Type string
}
