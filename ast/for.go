package ast

// ForExpression ranges over a range of values, executing a block of statements for each iteration.
type ForExpression struct {
	StartLine int
	StartCol  int
	Ident
	RangeExpr Expression
	Block
}

func (f *ForExpression) Line() int {
	return f.StartLine
}

func (f *ForExpression) Col() int {
	return f.StartCol
}

func (f *ForExpression) String() string {
	return "<FOR>"
}

func (f *ForExpression) expression() {}

var _ Node = (*ForExpression)(nil)
var _ Expression = (*ForExpression)(nil)
