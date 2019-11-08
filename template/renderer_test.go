package template

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/matryer/is"

	"github.com/blizzy78/copper/scope"
)

func TestRenderer_Render(t *testing.T) {
	is := is.New(t)

	tmpl1 := `hello <% t("tmpl2", { "name": "world" }) %>!`
	tmpl2 := `<% safe(name) %>`

	load := func(name string) (io.Reader, error) {
		if name == "tmpl1" {
			return strings.NewReader(tmpl1), nil
		}
		return strings.NewReader(tmpl2), nil
	}

	scopeData := map[string]interface{}{
		"safe": safe,
	}

	r := NewRenderer(load, nil, scopeData, "t")

	buf := bytes.Buffer{}

	err := r.Render(context.Background(), &buf, "tmpl1", nil)
	is.NoErr(err)
	is.Equal(buf.String(), "hello world!")
}

func TestRender(t *testing.T) {
	tmpl := ` aäöüÄÖÜß€ <% safe("b") %> c <% safe("d") %> e <% if 1 > 2 %> foo <% end %><% if 1 < 2 %> bar <% end %><% safe("hi") %> zzz `
	expected := ` aäöüÄÖÜß€ b c d e  bar hi zzz `

	w := strings.Builder{}

	s := scope.Scope{}

	s.Set("safe", safe)

	ls := func(s string) (interface{}, error) {
		return SafeString(s), nil
	}

	if err := Render(bytes.NewReader([]byte(tmpl)), &w, nil, &s, ls, nil); err != nil {
		t.Fatalf("error while rendering: %v", err)
	}

	res := w.String()

	if res != expected {
		t.Fatalf("wrong string, expected=<%s>, got=<%s>", expected, res)
	}
}

func safe(s string) SafeString {
	return SafeString(s)
}
