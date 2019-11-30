package ast

// BreakStatement breaks the execution of the loop it is used in.
type BreakStatement struct {
	StartLine int
	StartCol  int
}

func (b *BreakStatement) Line() int {
	return b.StartLine
}

func (b *BreakStatement) Col() int {
	return b.StartCol
}

func (b *BreakStatement) statement() {}

var _ Node = (*BreakStatement)(nil)
var _ Statement = (*BreakStatement)(nil)
