package template

import (
	"strconv"
	"strings"
	"testing"

	"github.com/blizzy78/copper/lexer"
	"github.com/blizzy78/copper/parser"
	"github.com/blizzy78/copper/ranger"
	"github.com/blizzy78/copper/scope"
)

var Result string

func BenchmarkRender_PureEval(b *testing.B) {
	b.StopTimer()

	tmpl := `<%
let a = 0
let b = 1
for i in fromTo(1, 20)
	let result = result + " " + safe(b)
	let newB = a + b
	let a = b
	let b = newB
end
	%>`

	l := lexer.New(strings.NewReader(tmpl))
	tCh, doneCh := l.Tokens()
	p := parser.New(tCh, doneCh)
	prog, err := p.Parse()
	if err != nil {
		b.Fatalf("error while parsing: %v", err)
	}

	safe := func(i int64) SafeString {
		return SafeString(strconv.FormatInt(i, 10))
	}

	s := scope.Scope{}
	s.Set("safe", safe)
	s.Set("fromTo", ranger.NewFromTo)
	s.Set("result", "")

	for i := 0; i < b.N; i++ {
		b.StartTimer()
		_, err = renderProgram(prog, &s)
		b.StopTimer()

		if err != nil {
			b.Fatalf("error while rendering: %v", err)
		}

		res, _ := s.Value("result")
		Result = res.(string)
	}
}
