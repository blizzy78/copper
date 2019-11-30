package ast

// ContinueStatement starts the next iteration of the for loop it is used in.
type ContinueStatement struct {
	StartLine int
	StartCol  int
}

func (c *ContinueStatement) Line() int {
	return c.StartLine
}

func (c *ContinueStatement) Col() int {
	return c.StartCol
}

func (c *ContinueStatement) statement() {}

var _ Node = (*ContinueStatement)(nil)
var _ Statement = (*ContinueStatement)(nil)
