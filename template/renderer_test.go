package template

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/matryer/is"

	"github.com/blizzy78/copper/evaluator"
	"github.com/blizzy78/copper/scope"
)

type FooData string

func TestRenderer_Render(t *testing.T) {
	is := is.New(t)

	tmpl1 := `hello <% t("tmpl2", { "name": "world " + foo(" baz") + " " + scopeVar }) %>!`
	tmpl2 := `<% safe(name) %>`

	l := LoaderFunc(func(name string) (io.ReadCloser, error) {
		if name == "tmpl1" {
			return io.NopCloser(strings.NewReader(tmpl1)), nil
		}
		return io.NopCloser(strings.NewReader(tmpl2)), nil
	})

	type ctxKey string
	var valueFromCtx string

	r := NewRenderer(l,
		WithScopeData("safe", safe),
		WithScopeData("foo", func(s string, ctx context.Context, sc *scope.Scope) string {
			valueFromCtx = ctx.Value(ctxKey("key")).(string)
			sc.Set("scopeVar", "scopeValue")
			return "bar" + s
		}),
	)

	buf := bytes.Buffer{}

	ctx := context.Background()
	ctx = context.WithValue(ctx, ctxKey("key"), "value")

	err := r.Render(ctx, &buf, "tmpl1", nil)
	is.NoErr(err)
	is.Equal(buf.String(), "hello world bar baz scopeValue!")
	is.Equal(valueFromCtx, "value")
}

func TestRender(t *testing.T) {
	is := is.New(t)

	tmpl := ` aäöüÄÖÜß€ <% safe("b") %> c <% safe("d") %> e <% if 1 > 2 %> foo <% end %><% if 1 < 2 %> bar <% end %><% safe("hi") %> zzz `
	expected := ` aäöüÄÖÜß€ b c d e  bar hi zzz `

	w := strings.Builder{}

	s := scope.Scope{}

	s.Set("safe", safe)

	ls := evaluator.LiteralStringerFunc(func(s string) (interface{}, error) {
		return SafeString(s), nil
	})

	err := Render(strings.NewReader(tmpl), &w, nil, &s, evaluator.WithLiteralStringer(ls))

	is.NoErr(err)

	res := w.String()
	is.Equal(res, expected)
}

func TestRender_Unsafe(t *testing.T) {
	is := is.New(t)

	tmpl := `<% "foo" %>`
	expected := "!UNSAFE!"

	w := strings.Builder{}
	s := scope.Scope{}

	err := Render(strings.NewReader(tmpl), &w, nil, &s)

	is.NoErr(err)

	res := w.String()
	is.Equal(res, expected)
}

func safe(s string) SafeString {
	return SafeString(s)
}
