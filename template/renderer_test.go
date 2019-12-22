package template

import (
	"bytes"
	"context"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/matryer/is"

	"github.com/blizzy78/copper/evaluator"
	"github.com/blizzy78/copper/scope"
)

type FooData string

func TestRenderer_Render(t *testing.T) {
	is := is.New(t)

	tmpl1 := `hello <% t("tmpl2", { "name": "world " + foo() }) %>!`
	tmpl2 := `<% safe(name) %>`

	l := LoaderFunc(func(name string) (io.Reader, error) {
		if name == "tmpl1" {
			return strings.NewReader(tmpl1), nil
		}
		return strings.NewReader(tmpl2), nil
	})

	foo := func(f FooData) string {
		return string(f)
	}

	fooData := FooData("bar")
	ra := evaluator.ArgumentResolverFunc(func(t reflect.Type) (interface{}, error) {
		if reflect.ValueOf(fooData).Type().ConvertibleTo(t) {
			return fooData, nil
		}
		return nil, nil
	})

	r := NewRenderer(l, WithScopeData("safe", safe), WithScopeData("foo", foo), WithArgumentResolver(ra))

	buf := bytes.Buffer{}

	err := r.Render(context.Background(), &buf, "tmpl1", nil)
	is.NoErr(err)
	is.Equal(buf.String(), "hello world bar!")
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
