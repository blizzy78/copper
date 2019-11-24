package lexer

import (
	"testing"
)

var token *Token

func BenchmarkX(b *testing.B) {
	s := `let x = 123 <% let y = 234 %> let z = 345 <% foo() %> test`

	for i := 0; i < b.N; i++ {
		b.StopTimer()

		l := newLexerString(s, b)
		tCh, doneCh := l.Tokens()

		defer func() {
			close(doneCh)
		}()

		b.StartTimer()

		for tok := range tCh {
			if tok.Err != nil {
				b.Fatalf("error while getting next token: %v", tok.Err)
			}

			token = tok

			if tok.Type == EOF {
				break
			}
		}
	}
}
