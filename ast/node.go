package ast

import "fmt"

// Node represents a node in the abstract syntax tree.
type Node interface {
	Line() int
	Col() int
	fmt.Stringer
}
