package ast

import "strings"

// CallExpression calls a method or function. The called method or function may return zero or more values.
// If no values are returned, the CallExpression returns nil, otherwise only the first value is returned.
// If the method or function returns an error as the last value, execution stops with that error.
//
// Not all method or function arguments need to be supplied. Any remaining arguments not supplied can be
// supplied automatically by a evaluator.ResolveArgumentFunc.
type CallExpression struct {
	StartLine int
	StartCol  int
	Callee    Expression
	Params    []Expression
}

func (c *CallExpression) Line() int {
	return c.StartLine
}

func (c *CallExpression) Col() int {
	return c.StartCol
}

func (c *CallExpression) String() string {
	buf := strings.Builder{}

	buf.WriteString(stringParens(c.Callee))

	buf.WriteString("(")

	first := true
	for _, p := range c.Params {
		if first {
			first = false
		} else {
			buf.WriteString(", ")
		}

		buf.WriteString(p.String())
	}

	buf.WriteString(")")

	return buf.String()
}

func (c *CallExpression) expression() {}

var _ Node = (*CallExpression)(nil)
var _ Expression = (*CallExpression)(nil)
