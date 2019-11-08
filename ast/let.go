package ast

import "fmt"

// LetStatement assigns a value to an identifier, usually in the current scope.
type LetStatement struct {
	StartLine int
	StartCol  int
	Ident
	Expression
}

func (l *LetStatement) Line() int {
	return l.StartLine
}

func (l *LetStatement) Col() int {
	return l.StartCol
}

func (l *LetStatement) String() string {
	return fmt.Sprintf("%s = %s", l.Ident.String(), l.Expression.String())
}

func (l *LetStatement) expression() {}

func (l *LetStatement) statement() {}

var _ Node = (*LetStatement)(nil)
var _ Statement = (*LetStatement)(nil)
var _ Expression = (*LetStatement)(nil)
