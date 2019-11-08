package ast

import "fmt"

// InfixExpression is an expression with an operator in the middle, such as "1 + 2" or "x % 5".
type InfixExpression struct {
	StartLine int
	StartCol  int
	Left      Expression
	Operator  string
	Right     Expression
}

func (i *InfixExpression) Line() int {
	return i.StartLine
}

func (i *InfixExpression) Col() int {
	return i.StartCol
}

func (i *InfixExpression) String() string {
	return fmt.Sprintf("%s %s %s", stringParens(i.Left), i.Operator, stringParens(i.Right))
}

func (i *InfixExpression) expression() {}

var _ Node = (*InfixExpression)(nil)
var _ Expression = (*InfixExpression)(nil)
