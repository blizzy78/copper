package lexer

import (
	"bytes"
	"strconv"
	"testing"
)

type expectedToken struct {
	typ     TokenType
	literal string
}

func TestLexer(t *testing.T) {
	tests := []struct {
		input    string
		expected []expectedToken
	}{
		{
			``,
			[]expectedToken{
				{EOF, ""},
			},
		},
		{
			`"x"`,
			[]expectedToken{
				{String, "x"},
				{EOF, ""},
			},
		},
		{
			`"x\ny"`,
			[]expectedToken{
				{String, "x\ny"},
				{EOF, ""},
			},
		},
		{
			`"x\"y"`,
			[]expectedToken{
				{String, `x"y`},
				{EOF, ""},
			},
		},
		{
			`+`,
			[]expectedToken{
				{Plus, "+"},
				{EOF, ""},
			},
		},
		{
			`!`,
			[]expectedToken{
				{Bang, "!"},
				{EOF, ""},
			},
		},
		{
			`=`,
			[]expectedToken{
				{Assign, "="},
				{EOF, ""},
			},
		},
		{
			` `,
			[]expectedToken{
				{EOF, ""},
			},
		},
		{
			`   `,
			[]expectedToken{
				{EOF, ""},
			},
		},
		{
			`=+-*/%(),!<>.:[]{}`,
			[]expectedToken{
				{Assign, "="},
				{Plus, "+"},
				{Minus, "-"},
				{Asterisk, "*"},
				{Slash, "/"},
				{Mod, "%"},
				{LeftParen, "("},
				{RightParen, ")"},
				{Comma, ","},
				{Bang, "!"},
				{LessThan, "<"},
				{GreaterThan, ">"},
				{Dot, "."},
				{Colon, ":"},
				{LeftBracket, "["},
				{RightBracket, "]"},
				{LeftBrace, "{"},
				{RightBrace, "}"},
				{EOF, ""},
			},
		},
		{
			`=+(@),`,
			[]expectedToken{
				{Assign, "="},
				{Plus, "+"},
				{LeftParen, "("},
				{Illegal, "@"},
				{RightParen, ")"},
				{Comma, ","},
				{EOF, ""},
			},
		},
		{
			`  = + (   ) , ! == != <= >= "ab  cd" '  ef
gh  ' `,
			[]expectedToken{
				{Assign, "="},
				{Plus, "+"},
				{LeftParen, "("},
				{RightParen, ")"},
				{Comma, ","},
				{Bang, "!"},
				{Equal, "=="},
				{NotEqual, "!="},
				{LessOrEqual, "<="},
				{GreaterOrEqual, ">="},
				{String, "ab  cd"},
				{String, `  ef
gh  `},
				{EOF, ""},
			},
		},
		{
			`5`,
			[]expectedToken{
				{Int, "5"},
				{EOF, ""},
			},
		},
		{
			`123`,
			[]expectedToken{
				{Int, "123"},
				{EOF, ""},
			},
		},
		{
			`123 234 345`,
			[]expectedToken{
				{Int, "123"},
				{Int, "234"},
				{Int, "345"},
				{EOF, ""},
			},
		},
		{
			`123 + 234 * 345`,
			[]expectedToken{
				{Int, "123"},
				{Plus, "+"},
				{Int, "234"},
				{Asterisk, "*"},
				{Int, "345"},
				{EOF, ""},
			},
		},
		{
			`123+234*345`,
			[]expectedToken{
				{Int, "123"},
				{Plus, "+"},
				{Int, "234"},
				{Asterisk, "*"},
				{Int, "345"},
				{EOF, ""},
			},
		},
		{
			`x`,
			[]expectedToken{
				{Ident, "x"},
				{EOF, ""},
			},
		},
		{
			`xyz`,
			[]expectedToken{
				{Ident, "xyz"},
				{EOF, ""},
			},
		},
		{
			`foo bar baz`,
			[]expectedToken{
				{Ident, "foo"},
				{Ident, "bar"},
				{Ident, "baz"},
				{EOF, ""},
			},
		},
		{
			`foo2 + bar * baz`,
			[]expectedToken{
				{Ident, "foo2"},
				{Plus, "+"},
				{Ident, "bar"},
				{Asterisk, "*"},
				{Ident, "baz"},
				{EOF, ""},
			},
		},
		{
			`foo+bar*baz`,
			[]expectedToken{
				{Ident, "foo"},
				{Plus, "+"},
				{Ident, "bar"},
				{Asterisk, "*"},
				{Ident, "baz"},
				{EOF, ""},
			},
		},
		{
			` a*2 + x%3 `,
			[]expectedToken{
				{Ident, "a"},
				{Asterisk, "*"},
				{Int, "2"},
				{Plus, "+"},
				{Ident, "x"},
				{Mod, "%"},
				{Int, "3"},
				{EOF, ""},
			},
		},
		{
			`let x = y`,
			[]expectedToken{
				{Let, "let"},
				{Ident, "x"},
				{Assign, "="},
				{Ident, "y"},
				{EOF, ""},
			},
		},
		{
			`if else elseif end for let func return break continue in nil`,
			[]expectedToken{
				{If, "if"},
				{Else, "else"},
				{ElseIf, "elseif"},
				{End, "end"},
				{For, "for"},
				{Let, "let"},
				// {Function, "func"},
				{Illegal, "func"},
				// {Return, "return"},
				{Illegal, "return"},
				{Break, "break"},
				{Continue, "continue"},
				{In, "in"},
				{Nil, "nil"},
				{EOF, ""},
			},
		},
		{
			`func foo(a, b)
			  return a + b
			end
			for i in bar
			  let x = foo(i, i + 1)
			  if x >= 10
			    break
			  end
			end`,
			[]expectedToken{
				// {Function, "func"},
				{Illegal, "func"},
				{Ident, "foo"},
				{LeftParen, "("},
				{Ident, "a"},
				{Comma, ","},
				{Ident, "b"},
				{RightParen, ")"},
				// {Return, "return"},
				{Illegal, "return"},
				{Ident, "a"},
				{Plus, "+"},
				{Ident, "b"},
				{End, "end"},
				{For, "for"},
				{Ident, "i"},
				{In, "in"},
				{Ident, "bar"},
				{Let, "let"},
				{Ident, "x"},
				{Assign, "="},
				{Ident, "foo"},
				{LeftParen, "("},
				{Ident, "i"},
				{Comma, ","},
				{Ident, "i"},
				{Plus, "+"},
				{Int, "1"},
				{RightParen, ")"},
				{If, "if"},
				{Ident, "x"},
				{GreaterOrEqual, ">="},
				{Int, "10"},
				{Break, "break"},
				{End, "end"},
				{End, "end"},
				{EOF, ""},
			},
		},
		{
			`// comment
			"foo"
			// comment 2
			"bar" // "comment 3"
			"baz"`,
			[]expectedToken{
				{String, "foo"},
				{String, "bar"},
				{String, "baz"},
				{EOF, ""},
			},
		},
	}

	for i, test := range tests {
		test := test
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			testTokenString(test.input, test.expected, t, WithStartInCodeMode())
		})
	}
}

func TestLexerStartInLiteral(t *testing.T) {
	tests := []struct {
		input    string
		expected []expectedToken
	}{
		{
			``,
			[]expectedToken{
				{EOF, ""},
			},
		},
		{
			`let x = 123 <% let y = 234 %> let z = 345 <% foo() %> test`,
			[]expectedToken{
				{Literal, "let x = 123 "},
				{Let, "let"},
				{Ident, "y"},
				{Assign, "="},
				{Int, "234"},
				{Literal, " let z = 345 "},
				{Ident, "foo"},
				{LeftParen, "("},
				{RightParen, ")"},
				{Literal, " test"},
				{EOF, ""},
			},
		},
		{
			`a <% // b %> c <% "d" %> e <%// f
			"g" %> h`,
			[]expectedToken{
				{Literal, "a "},
				{Literal, " c "},
				{String, "d"},
				{Literal, " e "},
				{String, "g"},
				{Literal, " h"},
				{EOF, ""},
			},
		},
		{
			`a <% /* b */ "c" /* d */ %> e <% /* f %> g <%
			"h" */ %> i`,
			[]expectedToken{
				{Literal, "a "},
				{String, "c"},
				{Literal, " e "},
				{Literal, " i"},
				{EOF, ""},
			},
		},
	}

	for i, test := range tests {
		test := test
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			testTokenString(test.input, test.expected, t)
		})
	}
}

func testTokenString(input string, expectedTokens []expectedToken, t *testing.T, opts ...Opt) {
	t.Helper()

	l := newLexerString(input, t, opts...)
	tCh, doneCh := l.Tokens()

	defer func() {
		close(doneCh)
	}()

	expectedIdx := 0

loop:
	for tok := range tCh {
		if tok.Err != nil {
			t.Fatalf("error reading next token: %v", tok.Err)
		}

		expected := expectedTokens[expectedIdx]
		expectedIdx++

		if tok.Type != expected.typ || tok.Literal != expected.literal {
			t.Fatalf("wrong token, expected=%s, got=%s", expected.String(), tok.String())
		}

		if tok.Type == EOF {
			break loop
		}
	}
}

func newLexerString(s string, tb testing.TB, opts ...Opt) (l *Lexer) {
	tb.Helper()

	r := bytes.NewReader([]byte(s))
	return New(r, opts...)
}

func (e *expectedToken) String() string {
	return string(e.typ) + "{" + e.literal + "}"
}
