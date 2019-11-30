package ast

// CaptureExpression captures the return values of all statements in its block, returning them as
// elements of a slice. If there is only one value, it is returned directly rather than inside a slice.
type CaptureExpression struct {
	StartLine int
	StartCol  int
	Block
}

func (c *CaptureExpression) Line() int {
	return c.StartLine
}

func (c *CaptureExpression) Col() int {
	return c.StartCol
}

func (c *CaptureExpression) expression() {
}

var _ Node = (*CallExpression)(nil)
var _ Expression = (*CallExpression)(nil)
