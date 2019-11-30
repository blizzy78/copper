package ast

// BoolLiteral represents a literal bool value.
type BoolLiteral struct {
	StartLine int
	StartCol  int
	Value     bool
}

func (b *BoolLiteral) Line() int {
	return b.StartLine
}

func (b *BoolLiteral) Col() int {
	return b.StartCol
}

func (b *BoolLiteral) expression() {}

var _ Node = (*BoolLiteral)(nil)
var _ Expression = (*BoolLiteral)(nil)
