package ast

// NilLiteral represents a literal nil value.
type NilLiteral struct {
	StartLine int
	StartCol  int
}

func (n *NilLiteral) Line() int {
	return n.StartLine
}

func (n *NilLiteral) Col() int {
	return n.StartCol
}

func (n *NilLiteral) String() string {
	return "nil"
}

func (n *NilLiteral) expression() {}

var _ Node = (*NilLiteral)(nil)
var _ Expression = (*NilLiteral)(nil)
