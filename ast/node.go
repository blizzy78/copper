package ast

// Node represents a node in the abstract syntax tree.
type Node interface {
	Line() int
	Col() int
}
