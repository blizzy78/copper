package evaluator

import (
	"fmt"
	"reflect"

	"github.com/blizzy78/copper/ast"
	"github.com/blizzy78/copper/scope"
)

// Evaluator evaluates an abstract syntax tree node and returns its result.
type Evaluator struct {
	literalStringer   LiteralStringer
	argumentResolvers []ArgumentResolver
	scope             *scope.Scope
	loopLevel         int
	breakRequested    bool
	continueRequested bool
}

// Opt is the type of a function that configures an option of ev.
type Opt func(ev *Evaluator)

// A LiteralStringer converts a literal string in a template to a value suitable for output.
// For example, it can wrap the string in a renderer.SafeString so that all literal strings are output as-is.
// It may also escape strings before wrapping them.
type LiteralStringer interface {
	String(s string) (interface{}, error)
}

// A LiteralStringerFunc is an adapter type that allows ordinary functions to be used as literal stringers.
// If f is a function with the appropriate signature, LiteralStringerFunc(f) is a literal stringer that calls f.
type LiteralStringerFunc func(s string) (interface{}, error)

// An ArgumentResolver resolves additional arguments of methods or functions that should be called.
// For example, a method could expect the arguments "x int, y *FooBar", but the method call only specifies the first argument:
// "a.b(123)". In that case, the second *FooBar argument can be automatically resolved by the argument resolver.
type ArgumentResolver interface {
	// Resolve inspects the type t and returns a value for it. If no actual value can be produced, nil may be returned
	// as the value. The returned value must be convertible to the type t.
	Resolve(t reflect.Type) (interface{}, error)
}

// An ArgumentResolverFunc is an adapter type that allows ordinary functions to be used as argument resolvers.
// If f is a function with the appropriate signature, ArgumentResolverFunc(f) is an argument resolver that calls f.
type ArgumentResolverFunc func(t reflect.Type) (interface{}, error)

// New returns a new evaluator, configured with opts.
func New(opts ...Opt) *Evaluator {
	ev := &Evaluator{
		literalStringer:   LiteralStringerFunc(defaultLiteral),
		argumentResolvers: []ArgumentResolver{ArgumentResolverFunc(defaultResolve)},
	}

	for _, opt := range opts {
		opt(ev)
	}

	return ev
}

// WithLiteralStringer configures an evaluator to use l to convert literal strings in a template to values
// suitable for output. The default is to return strings as-is, without escaping.
func WithLiteralStringer(l LiteralStringer) Opt {
	return func(ev *Evaluator) {
		ev.literalStringer = l
	}
}

// WithArgumentResolver configures an evaluator to use r to automatically resolve additional arguments of
// method or function calls in a template. The default is to not resolve any arguments.
//
// WithArgumentResolver may be used multiple times to configure additional resolvers. The first resolver
// to return a value other than nil wins.
func WithArgumentResolver(r ArgumentResolver) Opt {
	return func(ev *Evaluator) {
		ev.argumentResolvers = append(ev.argumentResolvers, r)
	}
}

// Eval evaluates the abstract syntax tree node n and returns its result. The scope s is used to look up and store
// variable state using identifiers. The scope may be pre-filled with identifiers which can be used during evaluation
// of expressions.
func (ev *Evaluator) Eval(n ast.Node, s *scope.Scope) (interface{}, error) {
	ev.scope = s
	return ev.eval(n)
}

func (ev *Evaluator) eval(n ast.Node) (interface{}, error) {
	switch node := n.(type) {
	case *ast.Program:
		o, err := ev.evalProgram(*node)
		o = normalize(o)
		return o, err
	case *ast.Block:
		o, err := ev.evalBlock(*node)
		o = normalize(o)
		return o, err
	case ast.Statement:
		o, err := ev.evalStatement(node)
		o = normalize(o)
		return o, err
	case ast.Expression:
		o, err := ev.evalExpression(node)
		o = normalize(o)
		return o, err
	default:
		panic(fmt.Errorf("unknown node type: %T", n))
	}
}

func normalize(v interface{}) interface{} { //nolint:gocyclo
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

func (f LiteralStringerFunc) String(s string) (interface{}, error) {
	return f(s)
}

func (r ArgumentResolverFunc) Resolve(t reflect.Type) (interface{}, error) {
	return r(t)
}
