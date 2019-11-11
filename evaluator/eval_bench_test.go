package evaluator

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/blizzy78/copper/lexer"
	"github.com/blizzy78/copper/parser"
	"github.com/blizzy78/copper/ranger"
	"github.com/blizzy78/copper/scope"
)

func benchmarkEvaluator(tmpl string, b *testing.B) {
	b.Helper()
	b.StopTimer()

	s := scope.Scope{}

	s.Set("fromTo", ranger.NewFromTo)
	s.Set("intToString", func(i int64) string {
		return strconv.FormatInt(i, 10)
	})

	l := lexer.New(strings.NewReader(tmpl), true)
	tCh, errCh, doneCh := l.Tokens()

	p := parser.New(tCh, doneCh)

	prog, err := p.Parse()
	if err != nil {
		b.Fatalf("error parsing program: %v", err)
	}

	if err := <-errCh; err != nil {
		b.Fatalf("error parsing program (lexer): %v", err)
	}

	e := &Evaluator{}

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		_, err = e.Eval(prog, &s)
	}

	b.StopTimer()

	if err != nil {
		b.Fatalf("error evaluating program: %v", err)
	}
}

func benchmarkEvaluatorIncrement(c int, b *testing.B) {
	b.Helper()
	b.StopTimer()

	tmpl := `
let x = 0
let s = ""
for i in fromTo(1, %d)
	let x = x + 1
	let s = s + " " + intToString(x)
end
s
`
	tmpl = fmt.Sprintf(tmpl, c)
	benchmarkEvaluator(tmpl, b)
}

func BenchmarkEvaluatorIncrement5(b *testing.B) {
	benchmarkEvaluatorIncrement(5, b)
}

func BenchmarkEvaluatorIncrement10(b *testing.B) {
	benchmarkEvaluatorIncrement(10, b)
}

func BenchmarkEvaluatorIncrement100(b *testing.B) {
	benchmarkEvaluatorIncrement(100, b)
}
