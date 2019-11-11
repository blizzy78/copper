package ast

// Block is a list of statements to execute.
type Block struct {
	StartLine  int
	StartCol   int
	Statements []Statement
}

func (b *Block) Line() int {
	return b.StartLine
}

func (b *Block) Col() int {
	return b.StartCol
}

func (b *Block) String() string {
	return "<BLOCK>"
}

var _ Node = (*Block)(nil)