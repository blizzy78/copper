package ast

// Expression represents any expression, such as "1 + 2" or "foo.bar()".
type Expression interface {
	Node

	// expression is a no-op method to differentiate this interface from other interfaces.
	expression()
}

type ExpressionStatement struct {
	StartLine int
	StartCol  int
	Expression
}

func (e *ExpressionStatement) Line() int {
	return e.StartLine
}

func (e *ExpressionStatement) Col() int {
	return e.StartCol
}

func (e *ExpressionStatement) statement() {}

var _ Node = (*ExpressionStatement)(nil)
var _ Statement = (*ExpressionStatement)(nil)
