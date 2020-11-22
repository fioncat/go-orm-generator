package dbtypes

type Table struct {
	Name    string
	Comment string
	Fields  []Field
}

type Field struct {
	Name    string
	Comment string
	Type    string
}