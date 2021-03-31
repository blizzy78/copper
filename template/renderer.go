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
	Load(name string) (io.ReadCloser, error)
}

// A LoaderFunc is an adapter type that allows ordinary functions to be used as Loaders.
// If f is a function with the appropriate signature, LoaderFunc(f) is a loader that calls f.
type LoaderFunc func(name string) (io.ReadCloser, error)

// Opt is the type of a function that configures r.
type Opt func(*Renderer)

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
func (r *Renderer) Render(ctx context.Context, w io.Writer, name string, data map[string]interface{}) error {
	userScope := scope.Scope{}

	if r.scopeData != nil {
		for k, v := range r.scopeData {
			userScope.Set(k, v)
		}
	}

	if userScope.HasValue(r.templateFuncName) {
		return fmt.Errorf("cannot use template function name, identifer already in use: %s", r.templateFuncName)
	}

	userScope.Lock()

	rendererScope := scope.Scope{
		Parent: &userScope,
	}

	renderTemplateFunc := func(name string, data map[string]interface{}, ctx context.Context) (SafeString, error) {
		buf := bytes.Buffer{}
		if err := r.Render(ctx, &buf, name, data); err != nil {
			return "", err
		}
		return SafeString(buf.String()), nil
	}

	rendererScope.Set(r.templateFuncName, renderTemplateFunc)

	rendererScope.Lock()

	rd, err := r.loader.Load(name)
	if err != nil {
		return err
	}
	defer rd.Close()

	ls := evaluator.LiteralStringerFunc(func(s string) (interface{}, error) {
		return SafeString(s), nil
	})

	ra := evaluator.ArgumentResolverFunc(func(t reflect.Type) (interface{}, error) {
		return resolveContext(t, ctx)
	})

	evaluatorOpts := make([]evaluator.Opt, len(r.argumentResolvers)+2)
	evaluatorOpts[0] = evaluator.WithLiteralStringer(ls)
	evaluatorOpts[1] = evaluator.WithArgumentResolver(ra)
	for i := 0; i < len(r.argumentResolvers); i++ {
		evaluatorOpts[2+i] = evaluator.WithArgumentResolver(r.argumentResolvers[i])
	}

	if err = Render(rd, w, data, &rendererScope, evaluatorOpts...); err != nil {
		return fmt.Errorf("error rendering template %s: %w", name, err)
	}

	return nil
}

// Render loads a template from r, evaluates it using scope s, optionally passing additional data,
// and writes the output to w.
func Render(r io.Reader, w io.Writer, data map[string]interface{}, s *scope.Scope, evaluatorOpts ...evaluator.Opt) error {
	templateScope := newTemplateScope(data, s)

	ra := evaluator.ArgumentResolverFunc(func(t reflect.Type) (interface{}, error) {
		return resolveScope(t, templateScope)
	})

	newEvaluatorOpts := make([]evaluator.Opt, len(evaluatorOpts)+1)
	newEvaluatorOpts[0] = evaluator.WithArgumentResolver(ra)
	for i := 0; i < len(evaluatorOpts); i++ {
		newEvaluatorOpts[1+i] = evaluatorOpts[i]
	}

	o, err := render(r, templateScope, newEvaluatorOpts...)
	if err != nil {
		return err
	}

	return write(w, o)
}

func (s SafeString) String() string {
	return string(s)
}

func newTemplateScope(data map[string]interface{}, parent *scope.Scope) *scope.Scope {
	s := scope.Scope{
		Parent: parent,
	}

	for k, v := range data {
		if v != nil {
			s.Set(k, v)
		}
	}

	return &s
}

func render(r io.Reader, s *scope.Scope, evaluatorOpts ...evaluator.Opt) (interface{}, error) {
	l := lexer.New(r)
	tCh, doneCh := l.Tokens()

	p := parser.New(tCh, doneCh)
	prog, err := p.Parse()
	if err != nil {
		return nil, err
	}

	// wrap capture around the original statements to capture all output
	prog = &ast.Program{
		Statements: []ast.Statement{
			capture(prog.Statements),
		},
	}

	return renderProgram(prog, s, evaluatorOpts...)
}

func renderProgram(p *ast.Program, s *scope.Scope, evaluatorOpts ...evaluator.Opt) (interface{}, error) {
	ev := evaluator.New(evaluatorOpts...)
	return ev.Eval(p, s)
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

func resolveContext(t reflect.Type, ctx context.Context) (interface{}, error) {
	if !reflect.ValueOf(ctx).Type().ConvertibleTo(t) {
		return nil, nil
	}
	return ctx, nil
}

func resolveScope(t reflect.Type, s *scope.Scope) (interface{}, error) {
	if !reflect.ValueOf(s).Type().ConvertibleTo(t) {
		return nil, nil
	}
	return s, nil
}

func write(w io.Writer, o interface{}) error {
	if sl, ok := o.([]interface{}); ok {
		for _, el := range sl {
			if err := writeSingle(w, el); err != nil {
				return err
			}
		}
		return nil
	}
	return writeSingle(w, o)
}

func writeSingle(w io.Writer, o interface{}) error {
	s := expectSafe(o)
	_, err := w.Write([]byte(s))
	return err
}

func expectSafe(v interface{}) string {
	switch value := v.(type) {
	case nil:
		return ""
	case SafeString:
		return value.String()
	case []interface{}:
		buf := strings.Builder{}
		for _, el := range value {
			buf.WriteString(expectSafe(el))
		}
		return buf.String()
	case string:
		if value != "" {
			return "!UNSAFE!"
		}
	default:
		return "!UNSAFE!"
	}
	return ""
}

func (l LoaderFunc) Load(name string) (io.ReadCloser, error) {
	return l(name)
}
