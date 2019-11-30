package lexer

import "fmt"

type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Col     int
	Err     error
}

type TokenType string

const (
	// Illegal is the token type used for unknown tokens.
	Illegal = "ILLEGAL"

	// EOF is the token type returned when a lexer has reached the end of its input.
	EOF = "EOF"

	// True is the token type used for a literal true bool value.
	True = "TRUE"

	// True is the token type used for a literal false bool value.
	False = "FALSE"

	// Nil is the token type used for a literal nil value.
	Nil = "NIL"

	// Ident is the token type used for an identifier.
	Ident = "IDENT"

	// Int is the token type used for a signed integer value.
	Int = "INT"

	// String is the token type used for a literal string value.
	String = "STRING"

	// Assign is the token type used for the assignment character '='.
	// If the character is part of the sequences "==", "!=", "<=", or ">=",
	// the token types Equal, NotEqual, LessOrEqual, or GreaterOrEqual are used for
	// the whole sequence instead, respectively.
	Assign = "="

	// Bang is the token type used for the bang character '!'.
	Bang = "!"

	// Plus is the token type used for the plus character '+'.
	Plus = "+"

	// Minus is the token type used for the minus character '-'.
	Minus = "-"

	// Asterisk is the token type used for the asterisk character '*'.
	Asterisk = "*"

	// Slash is the token type used for the slash character '/'.
	Slash = "/"

	// Mod is the token type used for the modulo character '%'.
	Mod = "%"

	// Equal is the token type used for the equals comparison character sequence "==".
	Equal = "=="

	// NotEqual is the token type used for the not equals comparison character sequence "!=".
	NotEqual = "!="

	// LessThan is the token type used for the less than character '<'. If the character is followed by
	// the equals character '=', the token type LessOrEqual is used for the whole sequence instead.
	LessThan = "<"

	// GreaterThan is the token type used for the greater than character '>'. If the character is followed by
	// the equals character '=', the token type GreaterOrEqual is used for the whole sequence instead.
	GreaterThan = ">"

	// LessOrEqual is the token type used for the less or equals character sequence "<=".
	LessOrEqual = "<="

	// GreaterOrEqual is the token type used for the greater or equals character sequence ">=".
	GreaterOrEqual = ">="

	// Dot is the token type used for the dot character '.'.
	Dot = "."

	// Comma is the token type used for the modulo character '%'.
	Comma = ","

	// Colon is the token type used for the colon character ':'.
	Colon = ":"

	// LeftParen is the token type used for the left parenthesis character '('.
	LeftParen = "("

	// RightParen is the token type used for the right parenthesis character ')'.
	RightParen = ")"

	// LeftBracket is the token type used for the left square bracket character '['.
	LeftBracket = "["

	// RightBracket is the token type used for the right square bracket character ']'.
	RightBracket = "]"

	// LeftBrace is the token type used for the left curly brace character '{'.
	LeftBrace = "{"

	// RightBrace is the token type used for the right curly brace character '}'.
	RightBrace = "}"

	// Let is the token type used for the let keyword.
	Let = "LET"

	// If is the token type used for the if keyword.
	If = "IF"

	// Else is the token type used for the else keyword.
	Else = "ELSE"

	// ElseIf is the token type used for the elseif keyword.
	ElseIf = "ELSE_IF"

	// End is the token type used for the end keyword.
	End = "END"

	// For is the token type used for the for keyword.
	For = "FOR"

	// Break is the token type used for the break keyword.
	Break = "BREAK"

	// Continue is the token type used for the continue keyword.
	Continue = "CONTINUE"

	// In is the token type used for the in keyword.
	In = "IN"

	// Capture is the token type used for the capture keyword.
	Capture = "CAPTURE"

	// Literal is the token type used for literal strings in the template, outside of code blocks.
	Literal = "LITERAL"

	Error = "ERROR"
)

func (t *Token) String() string {
	return fmt.Sprintf("'%s' (%s)", t.Literal, t.Type)
}
