package template

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/blizzy78/copper/ast"
	"github.com/blizzy78/copper/evaluator"
	"github.com/blizzy78/copper/lexer"
	"github.com/blizzy78/copper/parser"
	"github.com/blizzy78/copper/scope"
)

// Renderer parses templates, evaluates their code, and writes out the output.
type Renderer struct {
	load             LoadFunc
	resolveArg       evaluator.ResolveArgumentFunc
	scopeData        map[string]interface{}
	templateFuncName string
}

// LoadFunc is the type of a function that loads a template with a specific name and returns it as a rune reader.
type LoadFunc func(name string) (r io.Reader, err error)

// SafeString encapsulates a regular string to mark it as safe for output.
// If template code tries to output a regular string, it will be rendered only as "!UNSAFE!".
// Instead, regular strings must be wrapped in SafeString to render them as expected.
// Before wrapping in SafeString, strings should be HTML-escaped etc., depending on the output's language.
type SafeString string

// NewRenderer returns a new template renderer that loads templates using a load function.
//
// resolveArg may be nil, in which case no additional method or function call arguments can be resolved automatically.
//
// scopeData is a map of values that is provided to all templates being rendered. The values are provided as a scope
// to the evaluator.
//
// templateFuncName is the name of a function that can be called in a template to render another template.
// The signature of that function is as follows:
//
// 		func(name string, data map[string]interface{}) SafeString
//
// The data map is again provided as the scopeData argument to the new renderer.
func NewRenderer(load LoadFunc, resolveArg evaluator.ResolveArgumentFunc, scopeData map[string]interface{}, templateFuncName string) *Renderer {
	return &Renderer{
		load:             load,
		resolveArg:       resolveArg,
		scopeData:        scopeData,
		templateFuncName: templateFuncName,
	}
}

// Render loads a template with a specific name, evaluates it (optionally passing additional data), and writes the output to w.
//
// If the template calls the renderer's templateFuncName to render other templates, the data map passed to Render will not be passed
// to those templates.
//
// Literal output is wrapped in SafeString without further escaping.
//
// The context is passed to an internal evaluator.ResolveArgumentFunc and can therefore be resolved automatically
// as an argument to method or function calls in template code.
func (r *Renderer) Render(ctx context.Context, w io.Writer, name string, data map[string]interface{}) (err error) {
	userScope := scope.Scope{}

	if r.scopeData != nil {
		for k, v := range r.scopeData {
			userScope.Set(k, v)
		}
	}

	if userScope.HasValue(r.templateFuncName) {
		err = fmt.Errorf("cannot use template function name, identifer already in use: %s", r.templateFuncName)
		return
	}

	userScope.Lock()

	rendererScope := scope.Scope{
		Parent: &userScope,
	}

	renderTemplateFunc := func(name string, data map[string]interface{}, ctx context.Context) (s SafeString, err error) {
		buf := bytes.Buffer{}
		if err = r.Render(ctx, &buf, name, data); err != nil {
			return
		}
		s = SafeString(buf.String())
		return
	}

	rendererScope.Set(r.templateFuncName, renderTemplateFunc)

	rendererScope.Lock()

	var rd io.Reader
	if rd, err = r.load(name); err != nil {
		return
	}

	ls := func(s string) (interface{}, error) {
		return SafeString(s), nil
	}

	ra := func(t reflect.Type) (v interface{}, err error) {
		return r.resolve(t, ctx)
	}

	if err = Render(rd, w, data, &rendererScope, ls, ra); err != nil {
		err = fmt.Errorf("error rendering template %s: %w", name, err)
	}

	return
}

func (r *Renderer) resolve(t reflect.Type, ctx context.Context) (v interface{}, err error) {
	if reflect.ValueOf(ctx).Type().ConvertibleTo(t) {
		v = ctx
	} else if r.resolveArg != nil {
		v, err = r.resolveArg(t)
	} else {
		err = fmt.Errorf("cannot resolve argument of type: %s", t.String())
	}
	return
}

// Render loads a template from r, evaluates it using scope s (optionally passing additional data),
// and writes the output to w. Literal output is passed to ls so that it may be escaped and wrapped in
// SafeString. If ra is nil, no additional arguments can be resolved in method or function calls in
// template code.
func Render(r io.Reader, w io.Writer, data map[string]interface{}, s *scope.Scope,
	ls evaluator.LiteralStringFunc, ra evaluator.ResolveArgumentFunc) (err error) {

	templateScope := newTemplateScope(data, s)

	templateRA := func(t reflect.Type) (v interface{}, err error) {
		return resolve(t, templateScope, ra)
	}

	var o interface{}
	if o, err = render(r, templateScope, ls, evaluator.ResolveArgumentFunc(templateRA)); err != nil {
		return
	}

	err = write(w, o)

	return
}

func (s SafeString) String() string {
	return string(s)
}

func newTemplateScope(data map[string]interface{}, parent *scope.Scope) (s *scope.Scope) {
	s = &scope.Scope{
		Parent: parent,
	}

	for k, v := range data {
		if v != nil {
			s.Set(k, v)
		}
	}

	return
}

func render(r io.Reader, s *scope.Scope, ls evaluator.LiteralStringFunc, ra evaluator.ResolveArgumentFunc) (o interface{}, err error) {
	l := lexer.New(r, false)

	tCh, errCh, doneCh := l.Tokens()

	p := parser.New(tCh, doneCh)

	var prog *ast.Program
	if prog, err = p.Parse(); err != nil {
		return
	}

	if err = <-errCh; err != nil {
		return
	}

	// wrap capture around the original statements to capture all output
	prog.Statements = []ast.Statement{
		capture(prog.Statements),
	}

	ev := evaluator.Evaluator{
		LiteralStringFunc:   ls,
		ResolveArgumentFunc: ra,
	}
	o, err = ev.Eval(prog, s)

	return
}

func capture(statements []ast.Statement) ast.Statement {
	return &ast.ExpressionStatement{
		Expression: &ast.CaptureExpression{
			Block: ast.Block{
				Statements: statements,
			},
		},
	}
}

func resolve(t reflect.Type, s *scope.Scope, parent evaluator.ResolveArgumentFunc) (v interface{}, err error) {
	if reflect.ValueOf(s).Type().ConvertibleTo(t) {
		v = s
	} else if parent != nil {
		v, err = parent(t)
	}
	return
}

func write(w io.Writer, o interface{}) (err error) {
	if sl, ok := o.([]interface{}); ok {
		for _, el := range sl {
			if err = writeSingle(w, el); err != nil {
				break
			}
		}
	} else {
		err = writeSingle(w, o)
	}
	return
}

func writeSingle(w io.Writer, o interface{}) (err error) {
	s := expectSafe(o)
	_, err = w.Write([]byte(s))
	return
}

func expectSafe(v interface{}) (s string) {
	switch value := v.(type) {
	case nil:
		s = ""
	case SafeString:
		s = value.String()
	case []interface{}:
		buf := strings.Builder{}
		for _, el := range value {
			buf.WriteString(expectSafe(el))
		}
		s = buf.String()
	default:
		s = "!UNSAFE!"
	}
	return
}
