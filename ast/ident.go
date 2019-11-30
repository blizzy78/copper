package ast

// Ident represents a literal identifier, usually looked up automatically in the current scope.
type Ident struct {
	StartLine int
	StartCol  int
	Name      string
}

func (i *Ident) Line() int {
	return i.StartLine
}

func (i *Ident) Col() int {
	return i.StartCol
}

func (i *Ident) expression() {}

var _ Node = (*Ident)(nil)
var _ Expression = (*Ident)(nil)
