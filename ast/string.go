package ast

import (
	"fmt"
)

// StringLiteral represents a literal string, such as "foo" (including quotes.)
type StringLiteral struct {
	StartLine int
	StartCol  int
	Value     string
}

func (s *StringLiteral) Line() int {
	return s.StartLine
}

func (s *StringLiteral) Col() int {
	return s.StartCol
}

func (s *StringLiteral) String() string {
	return fmt.Sprintf(`"%s"`, s.Value)
}

func (s *StringLiteral) expression() {}

var _ Node = (*StringLiteral)(nil)
var _ Expression = (*StringLiteral)(nil)
