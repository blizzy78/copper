package ast

// IfExpression executes the statements in one of its conditional blocks if the condition of that block is met.
type IfExpression struct {
	StartLine    int
	StartCol     int
	Conditionals []ConditionalBlock
}

// ConditionalBlock contains a block of statements to be executed if the condition is met (if any.)
type ConditionalBlock struct {
	StartLine int
	StartCol  int
	Condition Expression
	Block
}

func (i *IfExpression) Line() int {
	return i.StartLine
}

func (i *IfExpression) Col() int {
	return i.StartCol
}

func (i *IfExpression) String() string {
	return "<IF>"
}

func (i *IfExpression) expression() {}

func (c *ConditionalBlock) Line() int {
	return c.StartLine
}

func (c *ConditionalBlock) Col() int {
	return c.StartCol
}

var _ Node = (*IfExpression)(nil)
var _ Expression = (*IfExpression)(nil)

var _ Node = (*ConditionalBlock)(nil)
