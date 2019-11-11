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
		tCh, errCh := l.Tokens()

		b.StartTimer()

		for {
			select {
			case tok := <-tCh:
				token = tok

			case err := <-errCh:
				b.Fatalf("error while getting next token: %v", err)
			}
		}
	}
}
