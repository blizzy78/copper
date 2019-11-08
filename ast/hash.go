package ast

// HashExpression creates a map of expressions indexed by strings.
type HashExpression struct {
	StartLine int
	StartCol  int
	Values    map[string]Expression
}

func (h *HashExpression) Line() int {
	return h.StartLine
}

func (h *HashExpression) Col() int {
	return h.StartCol
}

func (h *HashExpression) String() string {
	return "<HASH>"
}

func (h *HashExpression) expression() {}

var _ Node = (*HashExpression)(nil)
var _ Expression = (*HashExpression)(nil)
