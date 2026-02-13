package model

type Column struct {
	Name          string
	Type          string
	Nullable      bool
	PrimaryKey    bool
	DefaultValue  *string
	AutoIncrement bool
}
