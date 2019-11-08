package lexer

import (
	"testing"
)

var token *Token

func BenchmarkX(b *testing.B) {
	s := `let x = 123 <% let y = 234 %> let z = 345 <% foo() %> test`

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		l := newLexerString(s, false, b)
		b.StartTimer()

		for {
			var err error
			if token, err = l.Next(); err != nil {
				b.Fatalf("error while getting next token: %v", err)
			}

			if token.Type == EOF {
				break
			}
		}
	}
}
