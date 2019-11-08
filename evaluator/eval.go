package evaluator

import (
	"fmt"
	"reflect"

	"github.com/blizzy78/copper/ast"
	"github.com/blizzy78/copper/scope"
)

// Evaluator evaluates an abstract syntax tree node and returns its result.
type Evaluator struct {
	// LiteralStringFunc is used to convert literal strings in the template to values suitable for output.
	// If LiteralStringFunc is nil, strings will be wrapped in renderer.SafeString and will therefore be output as-is.
	LiteralStringFunc

	// ResolveArgumentFunc is used to automatically resolve additional arguments of methods or functions that should be called.
	// If ResolveArgumentFunc is nil, additional arguments cannot be resolved and must be supplied with the function calls
	// in the template.
	ResolveArgumentFunc

	scope             *scope.Scope
	loopLevel         int
	breakRequested    bool
	continueRequested bool
}

// LiteralStringFunc is the type of a function that converts a literal string in a template to a value suitable for output.
// For example, it can wrap the string in a renderer.SafeString so that all literal strings are output as-is.
type LiteralStringFunc func(s string) (interface{}, error)

// ResolveArgumentFunc is the type of a function that resolves additional arguments of methods or functions that should be called.
// For example, a method could expect the arguments "x int, y *FooBar", but the method call only specifies the first argument:
// "a.b(123)". In that case, the second *FooBar argument can be automatically resolved by the resolve function.
//
// The resolve function inspects the type t and returns a value for it. If no actual value can be produced, nil may be returned
// as the value. The returned value must be convertible to the type t.
type ResolveArgumentFunc func(t reflect.Type) (v interface{}, err error)

// Eval evaluates the abstract syntax tree node n and returns its result. The scope s is used to look up and store
// variable state using identifiers. The scope may be pre-filled with identifiers which can be used during evaluation
// of expressions.
func (ev *Evaluator) Eval(n ast.Node, s *scope.Scope) (o interface{}, err error) {
	ev.scope = s
	return ev.eval(n)
}

func (ev *Evaluator) eval(n ast.Node) (o interface{}, err error) {
	switch node := n.(type) {
	case *ast.Program:
		o, err = ev.evalProgram(*node)
	case *ast.Block:
		o, err = ev.evalBlock(*node)
	case ast.Statement:
		o, err = ev.evalStatement(node)
	case ast.Expression:
		o, err = ev.evalExpression(node)
	default:
		panic(fmt.Errorf("unknown node type: %T", n))
	}

	o = normalize(o)

	return
}

func normalize(v interface{}) (o interface{}) {
	switch value := v.(type) {
	case int:
		return int64(value)
	case int8:
		return int64(value)
	case int16:
		return int64(value)
	case int32:
		return int64(value)
	case int64:
		return value

	case uint:
		return int64(value)
	case uint8:
		return int64(value)
	case uint16:
		return int64(value)
	case uint32:
		return int64(value)
	case uint64:
		return int64(value)

	default:
		return v
	}
}
