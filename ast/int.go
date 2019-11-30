package ast

// IntLiteral represents a literal signed integer value.
type IntLiteral struct {
	StartLine int
	StartCol  int
	Value     int64
}

func (i *IntLiteral) Line() int {
	return i.StartLine
}

func (i *IntLiteral) Col() int {
	return i.StartCol
}

func (i *IntLiteral) expression() {}

var _ Node = (*IntLiteral)(nil)
var _ Expression = (*IntLiteral)(nil)
