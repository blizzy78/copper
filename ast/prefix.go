package ast

// PrefixExpression is an expression that starts with an operator, such as "-15" or "!x" (where x is a bool.)
type PrefixExpression struct {
	StartLine int
	StartCol  int
	Operator  string
	Expression
}

func (p *PrefixExpression) Line() int {
	return p.StartLine
}

func (p *PrefixExpression) Col() int {
	return p.StartCol
}

func (p *PrefixExpression) expression() {}

var _ Node = (*PrefixExpression)(nil)
var _ Expression = (*PrefixExpression)(nil)
