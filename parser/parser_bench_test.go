package parser

import (
	"testing"

	"github.com/blizzy78/copper/ast"
)

var Prog *ast.Program

const tmpl = `let x = 123 <% let y = 234 %> let z = 345 <% foo() %> test`

func BenchmarkParser(b *testing.B) {
	b.StopTimer()

	for i := 0; i < b.N; i++ {
		l := newLexerString(tmpl, b)
		tCh, doneCh := l.Tokens()

		p := New(tCh, doneCh)

		b.StartTimer()

		pr, err := p.Parse()

		b.StopTimer()

		Prog = pr

		if err != nil {
			b.Fatalf("error while parsing: %v", err)
		}
	}
}
