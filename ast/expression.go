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

func (e *ExpressionStatement) String() string {
	return e.Expression.String()
}

func (e *ExpressionStatement) statement() {}

func stringParens(e Expression) string {
	switch e.(type) {
	case *IntLiteral:
		return e.String()
	case *BoolLiteral:
		return e.String()
	case *Ident:
		return e.String()
	case *PrefixExpression:
		return e.String()
	default:
		return "(" + e.String() + ")"
	}
}

var _ Node = (*ExpressionStatement)(nil)
var _ Statement = (*ExpressionStatement)(nil)
