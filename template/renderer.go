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
	loader            Loader
	argumentResolvers []evaluator.ArgumentResolver
	scopeData         map[string]interface{}
	templateFuncName  string
}

// A Loader loads a template with a specific name and returns it as a reader.
type Loader interface {
	Load(name string) (r io.Reader, err error)
}

// A LoaderFunc is an adapter type that allows ordinary functions to be used as Loaders.
// If f is a function with the appropriate signature, LoaderFunc(f) is a loader that calls f.
type LoaderFunc func(name string) (r io.Reader, err error)

// Opt is the type of a function that configures r.
type Opt func(r *Renderer)

// SafeString encapsulates a regular string to mark it as safe for output.
// If template code tries to output a regular string, it will be rendered only as "!UNSAFE!".
// Instead, regular strings must be wrapped in SafeString to render them as expected.
// Before wrapping in SafeString, strings should be HTML-escaped etc., depending on the output's language.
type SafeString string

// NewRenderer returns a new renderer, configured with opts, that loads templates via load.
func NewRenderer(loader Loader, opts ...Opt) *Renderer {
	r := &Renderer{
		loader:           loader,
		templateFuncName: "t",
		scopeData:        map[string]interface{}{},
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

// WithArgumentResolver configures a renderer to use r to automatically resolve additional arguments of
// method or function calls in a template. The default is to only resolve the current scope.Scope.
//
// WithArgumentResolver may be used multiple times to configure additional resolvers.
func WithArgumentResolver(r evaluator.ArgumentResolver) Opt {
	return func(rend *Renderer) {
		rend.argumentResolvers = append(rend.argumentResolvers, r)
	}
}

// WithScopeData configures a renderer to provide additional data to all templates being rendered.
// WithScopeData may be used multiple times, also in combination with WithScopeDataMap.
func WithScopeData(k string, v interface{}) Opt {
	return func(r *Renderer) {
		r.scopeData[k] = v
	}
}

// WithScopeDataMap configures a renderer to provide additional data to all templates being rendered.
// WithScopeDataMap may be used multiple times, also in combination with WithScopeData.
func WithScopeDataMap(data map[string]interface{}) Opt {
	return func(r *Renderer) {
		for k, v := range data {
			r.scopeData[k] = v
		}
	}
}

// WithTemplateFuncName configures a renderer to use n as the name of the function that may be called in
// templates to render other templates. The default name of this function is "t".
//
// The signature of that function is as follows:
//
// 		func(name string, data map[string]interface{}) SafeString
//
// where name is the name of the template to render.
//
// The data map is in turn provided to the new renderer using WithScopeData.
func WithTemplateFuncName(n string) Opt {
	return func(r *Renderer) {
		r.templateFuncName = n
	}
}

// Render loads a template with a specific name, evaluates it (optionally passing additional data), and writes the output to w.
//
// If the template calls the renderer's function to render other templates (see WithTemplateFuncName), the data map passed to
// Render will not be passed to those templates.
//
// Literal output is wrapped in SafeString without further escaping.
//
// The context is passed to an internal evaluator.ArgumentResolver and can therefore be resolved automatically
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
	if rd, err = r.loader.Load(name); err != nil {
		return
	}

	ls := evaluator.LiteralStringerFunc(func(s string) (interface{}, error) {
		return SafeString(s), nil
	})

	ra := evaluator.ArgumentResolverFunc(func(t reflect.Type) (v interface{}, err error) {
		return resolveContext(t, ctx)
	})

	evaluatorOpts := make([]evaluator.Opt, len(r.argumentResolvers)+2)
	evaluatorOpts[0] = evaluator.WithLiteralStringer(ls)
	evaluatorOpts[1] = evaluator.WithArgumentResolver(ra)
	for i := 0; i < len(r.argumentResolvers); i++ {
		evaluatorOpts[2+i] = evaluator.WithArgumentResolver(r.argumentResolvers[i])
	}

	if err = Render(rd, w, data, &rendererScope, evaluatorOpts...); err != nil {
		err = fmt.Errorf("error rendering template %s: %w", name, err)
	}

	return
}

// Render loads a template from r, evaluates it using scope s, optionally passing additional data,
// and writes the output to w.
func Render(r io.Reader, w io.Writer, data map[string]interface{}, s *scope.Scope, evaluatorOpts ...evaluator.Opt) (err error) {
	templateScope := newTemplateScope(data, s)

	ra := evaluator.ArgumentResolverFunc(func(t reflect.Type) (v interface{}, err error) {
		return resolveScope(t, templateScope)
	})

	newEvaluatorOpts := make([]evaluator.Opt, len(evaluatorOpts)+1)
	newEvaluatorOpts[0] = evaluator.WithArgumentResolver(ra)
	for i := 0; i < len(evaluatorOpts); i++ {
		newEvaluatorOpts[1+i] = evaluatorOpts[i]
	}

	var o interface{}
	if o, err = render(r, templateScope, newEvaluatorOpts...); err != nil {
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

func render(r io.Reader, s *scope.Scope, evaluatorOpts ...evaluator.Opt) (o interface{}, err error) {
	l := lexer.New(r)

	tCh, doneCh := l.Tokens()

	p := parser.New(tCh, doneCh)

	var prog *ast.Program
	if prog, err = p.Parse(); err != nil {
		return
	}

	// wrap capture around the original statements to capture all output
	prog.Statements = []ast.Statement{
		capture(prog.Statements),
	}

	ev := evaluator.New(evaluatorOpts...)
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

func resolveContext(t reflect.Type, ctx context.Context) (v interface{}, err error) {
	if reflect.ValueOf(ctx).Type().ConvertibleTo(t) {
		v = ctx
	}
	return
}

func resolveScope(t reflect.Type, s *scope.Scope) (v interface{}, err error) {
	if reflect.ValueOf(s).Type().ConvertibleTo(t) {
		v = s
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

func (l LoaderFunc) Load(name string) (r io.Reader, err error) {
	return l(name)
}
