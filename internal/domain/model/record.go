package model

type Value struct {
	Text   string
	IsNull bool
}

type Record struct {
	Values []Value
}
