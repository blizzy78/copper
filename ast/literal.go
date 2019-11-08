package ast

// Literal represents literal text outside of code blocks.
type Literal struct {
	StartLine int
	StartCol  int
	Text      string
}

func (l *Literal) Line() int {
	return l.StartLine
}

func (l *Literal) Col() int {
	return l.StartCol
}

func (l *Literal) String() string {
	return l.Text
}

func (l *Literal) expression() {}

var _ Node = (*Literal)(nil)
var _ Expression = (*Literal)(nil)
