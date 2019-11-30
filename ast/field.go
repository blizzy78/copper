package ast

// FieldExpression looks up a "field" in a callee, returning it. The "field" may be a struct member or a method.
// In the case of a method, a CallExpression can be used to call it.
type FieldExpression struct {
	StartLine int
	StartCol  int
	Callee    Expression
	Index     Expression
}

func (f *FieldExpression) Line() int {
	return f.StartLine
}

func (f *FieldExpression) Col() int {
	return f.StartCol
}

func (f *FieldExpression) expression() {}

var _ Node = (*FieldExpression)(nil)
var _ Expression = (*FieldExpression)(nil)
