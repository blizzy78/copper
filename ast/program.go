package ast

// Program contains a list of statements to be executed.
type Program struct {
	StartLine  int
	StartCol   int
	Statements []Statement
}

// Statement is a single statement to be executed, such as a "let" or "if" statement.
type Statement interface {
	Node

	// statement is a no-op method to differentiate this interface from other interfaces.
	statement()
}

func (p *Program) Line() int {
	return p.StartLine
}

func (p *Program) Col() int {
	return p.StartCol
}

func (p *Program) String() string {
	return "<PROGRAM>"
}

var _ Node = (*Program)(nil)
