package model

type Column struct {
	Name          string
	Type          string
	Nullable      bool
	PrimaryKey    bool
	Unique        bool
	DefaultValue  *string
	AutoIncrement bool
	ForeignKeys   []ForeignKeyRef
}
