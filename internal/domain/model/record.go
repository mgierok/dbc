package model

type Value struct {
	Text   string
	IsNull bool
	Raw    any
}

type Record struct {
	Values []Value
}
