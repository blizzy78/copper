package lexer

import (
	"bufio"
	"io"
	"strings"
	"sync"
)

// Lexer parses a series of statements or expressions, a template, from a reader and returns them
// as a sequence of lexical tokens.
type Lexer struct {
	r              io.RuneReader
	optStartInCode bool
	line           int
	col            int
	startedInCode  bool
	inCode         bool
	initOnce       sync.Once
	currChar       rune
	nextChar       rune
	currEOF        bool
	nextEOF        bool
}

// Opt is the type of a function that configures an option of l.
type Opt func(l *Lexer)

var (
	keywords = map[string]TokenType{
		// TODO
		// "func":     Function,
		// "return":   Return,
		"func":   Illegal,
		"return": Illegal,

		"let":      Let,
		"if":       If,
		"else":     Else,
		"elseif":   ElseIf,
		"end":      End,
		"for":      For,
		"break":    Break,
		"continue": Continue,
		"in":       In,
		"true":     True,
		"false":    False,
		"nil":      Nil,
		"capture":  Capture,
	}
)

// New returns a new lexer, configured with opts, that reads a template from r.
func New(r io.Reader, opts ...Opt) *Lexer {
	l := &Lexer{
		r: bufio.NewReader(r),
	}

	for _, opt := range opts {
		opt(l)
	}

	l.startedInCode = l.optStartInCode
	l.inCode = l.optStartInCode

	return l
}

// WithStartInCodeMode configures a lexer to start in code mode. The default is to start in literal mode.
// If the lexer starts in literal mode, code blocks (<% %>) must be used to switch to code mode.
func WithStartInCodeMode() Opt {
	return func(l *Lexer) {
		l.optStartInCode = true
	}
}

// Tokens reads from the lexer's input and writes a sequence of tokens into tCh. If an error occurs
// when producing tokens, the error is associated with the next token in the channel. Token production
// stops when there was an error, or when the done channel is closed.
func (l *Lexer) Tokens() (tCh <-chan *Token, done chan<- struct{}) {
	tokenCh := make(chan *Token)
	tCh = tokenCh

	doneCh := make(chan struct{})
	done = doneCh

	go func() {
		defer close(tokenCh)

	loop:
		for {
			t, err := l.next()
			if err != nil {
				t.Err = err
			}

			select {
			case <-doneCh:
				break loop
			case tokenCh <- &t:
				// okay
			}

			if t.Type == EOF || t.Err != nil {
				break
			}
		}
	}()

	return
}

func (l *Lexer) next() (t Token, err error) {
	l.initOnce.Do(func() {
		err = l.initialize()
	})

	if err != nil {
		return
	}

	if l.inCode {
		if err = l.skipWhitespace(); err != nil {
			return
		}
	}

	if l.currEOF {
		// if the lexer started in literal mode, it is illegal to end inside code mode
		if !l.startedInCode && l.inCode {
			err = newParseErrorf(l.line, l.col, "end of code mode block expected")
			return
		}

		t = newToken(EOF, "", l.line, l.col)

		return
	}

	if l.inCode {
		if isIntChar(l.currChar) {
			t, err = l.readInt()
		} else if isIdentFirstChar(l.currChar) {
			if t, err = l.readIdent(); err == nil {
				t.Type = lookupIdentTokenType(t.Literal)
			}
		} else if l.currChar == '"' || l.currChar == '\'' {
			t, err = l.readString()
		} else if l.currChar == '/' && l.nextCharIs('/') {
			t, err = l.readLineComment()
		} else if l.currChar == '/' && l.nextCharIs('*') {
			t, err = l.readBlockComment()
		} else {
			t, err = l.readToken()
		}
	} else {
		t, err = l.readLiteralOrCodeStart()
	}

	if err != nil {
		return
	}

	// if the lexer started in code mode, it is illegal to try to break out into literal mode
	if l.startedInCode && (t.Type == codeStart || t.Type == codeEnd) {
		t = newToken(Illegal, t.Literal, l.line, l.col)
		return
	}

	if l.inCode && t.Type == codeEnd {
		l.inCode = false
	} else if !l.inCode && t.Type == codeStart {
		l.inCode = true
	}

	// skip internal tokens
	if t.Type == codeStart || t.Type == codeEnd || t.Type == lineComment || t.Type == blockComment {
		t, err = l.next()
	}

	return
}

func (l *Lexer) initialize() (err error) {
	if err = l.readNextChar(); err != nil {
		return
	}
	err = l.readNextChar()

	l.line = 1
	l.col = 1

	return
}

func (l *Lexer) skipWhitespace() (err error) {
	for !l.currEOF && isWhitespaceChar(l.currChar) {
		if err = l.readNextChar(); err != nil {
			break
		}
	}
	return
}

func (l *Lexer) readInt() (t Token, err error) {
	b := strings.Builder{}
	line := l.line
	col := l.col

	for !l.currEOF && isIntChar(l.currChar) {
		if _, err = b.WriteRune(l.currChar); err != nil {
			break
		}

		if err = l.readNextChar(); err != nil {
			break
		}
	}

	if err != nil {
		return
	}

	t = newToken(Int, b.String(), line, col)

	return
}

func (l *Lexer) readLiteralOrCodeStart() (t Token, err error) {
	b := strings.Builder{}
	line := l.line
	col := l.col

	if !l.currEOF && l.isAtCodeStart() {
		if err = l.readNextChar(); err != nil {
			return
		}
		if err = l.readNextChar(); err != nil {
			return
		}

		t = newToken(codeStart, "<%", line, col)

		return
	}

	for !l.currEOF && !l.isAtCodeStart() {
		if _, err = b.WriteRune(l.currChar); err != nil {
			break
		}

		if err = l.readNextChar(); err != nil {
			break
		}
	}

	if err != nil {
		return
	}

	t = newToken(Literal, b.String(), line, col)

	return
}

func (l *Lexer) isAtCodeStart() bool {
	return l.currChar == '<' && l.nextCharIs('%')
}

func (l *Lexer) isAtCodeEnd() bool {
	return l.currChar == '%' && l.nextCharIs('>')
}

func (l *Lexer) isAtBlockCommentEnd() bool {
	return l.currChar == '*' && l.nextCharIs('/')
}

func (l *Lexer) readIdent() (t Token, err error) {
	b := strings.Builder{}
	line := l.line
	col := l.col

	for !l.currEOF && isIdentChar(l.currChar) {
		if _, err = b.WriteRune(l.currChar); err != nil {
			break
		}

		if err = l.readNextChar(); err != nil {
			break
		}
	}

	if err != nil {
		return
	}

	t = newToken(Ident, b.String(), line, col)

	return
}

func (l *Lexer) readString() (t Token, err error) {
	startChar := l.currChar
	line := l.line
	col := l.col

	if err = l.readNextChar(); err != nil {
		return
	}

	buf := strings.Builder{}

	prevBackslash := false

	for !l.currEOF && (prevBackslash || l.currChar != startChar) {
		if _, err = buf.WriteRune(l.currChar); err != nil {
			break
		}

		prevBackslash = l.currChar == '\\'

		if err = l.readNextChar(); err != nil {
			return
		}
	}

	if err != nil {
		return
	}

	if l.currEOF {
		err = newParseErrorf(line, col, "end of string not found")
		return
	}

	if err = l.readNextChar(); err != nil {
		return
	}

	s := buf.String()
	s = strings.ReplaceAll(s, `\r`, "\r")
	s = strings.ReplaceAll(s, `\n`, "\n")
	s = strings.ReplaceAll(s, `\t`, "\t")
	s = strings.ReplaceAll(s, `\"`, `"`)
	s = strings.ReplaceAll(s, `\'`, "'")
	s = strings.ReplaceAll(s, `\\`, "\\")

	t = newToken(String, s, line, col)

	return
}

func (l *Lexer) readToken() (t Token, err error) {
	line := l.line
	col := l.col

	switch l.currChar {
	case '=':
		if l.nextCharIs('=') {
			t = newToken(Equal, "==", line, col)
			err = l.readNextChar()
		} else {
			t = newToken(Assign, "=", line, col)
		}
	case '!':
		if l.nextCharIs('=') {
			t = newToken(NotEqual, "!=", line, col)
			err = l.readNextChar()
		} else {
			t = newToken(Bang, "!", line, col)
		}
	case '+':
		t = newToken(Plus, "+", line, col)
	case '-':
		t = newToken(Minus, "-", line, col)
	case '*':
		t = newToken(Asterisk, "*", line, col)
	case '/':
		t = newToken(Slash, "/", line, col)
	case '%':
		if l.nextCharIs('>') {
			t = newToken(codeEnd, "%>", line, col)
			err = l.readNextChar()
		} else {
			t = newToken(Mod, "%", line, col)
		}
	case '(':
		t = newToken(LeftParen, "(", line, col)
	case ')':
		t = newToken(RightParen, ")", line, col)
	case ',':
		t = newToken(Comma, ",", line, col)
	case '.':
		t = newToken(Dot, ".", line, col)
	case ':':
		t = newToken(Colon, ":", line, col)
	case '[':
		t = newToken(LeftBracket, "[", line, col)
	case ']':
		t = newToken(RightBracket, "]", line, col)
	case '{':
		t = newToken(LeftBrace, "{", line, col)
	case '}':
		t = newToken(RightBrace, "}", line, col)
	case '<':
		if l.nextCharIs('=') {
			t = newToken(LessOrEqual, "<=", line, col)
			err = l.readNextChar()
		} else {
			t = newToken(LessThan, "<", line, col)
		}
	case '>':
		if l.nextCharIs('=') {
			t = newToken(GreaterOrEqual, ">=", line, col)
			err = l.readNextChar()
		} else {
			t = newToken(GreaterThan, ">", line, col)
		}
	default:
		t = newToken(Illegal, string(l.currChar), line, col)
	}

	if err != nil {
		return
	}

	err = l.readNextChar()

	return
}

func (l *Lexer) readBlockComment() (t Token, err error) {
	line := l.line
	col := l.col

	if err = l.readNextChar(); err != nil {
		return
	}
	if err = l.readNextChar(); err != nil {
		return
	}

	buf := strings.Builder{}

	endOk := false
	for !l.currEOF {
		if l.isAtBlockCommentEnd() {
			endOk = true
			break
		}

		if _, err = buf.WriteRune(l.currChar); err != nil {
			break
		}

		if err = l.readNextChar(); err != nil {
			break
		}
	}

	if err != nil {
		return
	}

	if !endOk {
		err = newParseErrorf(line, col, "end of block comment not found")
		return
	}

	if err = l.readNextChar(); err != nil {
		return
	}
	if err = l.readNextChar(); err != nil {
		return
	}

	t = newToken(blockComment, buf.String(), line, col)

	return
}

func (l *Lexer) readLineComment() (t Token, err error) {
	line := l.line
	col := l.col

	if err = l.readNextChar(); err != nil {
		return
	}
	if err = l.readNextChar(); err != nil {
		return
	}

	buf := strings.Builder{}

	for !l.currEOF && l.currChar != '\n' {
		// if the lexer started in literal mode, stop line comments at end of code blocks
		if !l.startedInCode && l.isAtCodeEnd() {
			break
		}

		if _, err = buf.WriteRune(l.currChar); err != nil {
			break
		}

		if err = l.readNextChar(); err != nil {
			break
		}
	}

	if err != nil {
		return
	}

	t = newToken(lineComment, buf.String(), line, col)

	return
}

func (l *Lexer) readNextChar() (err error) {
	if l.currEOF {
		return
	}

	if l.nextEOF {
		l.currEOF = true
		l.col++
		return
	}

	switch l.currChar {
	case '\n':
		l.line++
		l.col = 1
	default:
		l.col++
	}

	var r rune
	var i int

	r, i, err = l.r.ReadRune()

	if i > 0 {
		l.currChar = l.nextChar
		l.nextChar = r
	}

	if err == io.EOF {
		l.nextEOF = true
		l.currChar = l.nextChar
		err = nil
	}

	return
}

func (l *Lexer) nextCharIs(c rune) bool {
	return !l.nextEOF && (l.nextChar == c)
}

func newToken(t TokenType, literal string, line int, col int) Token {
	return Token{
		Type:    t,
		Literal: literal,
		Line:    line,
		Col:     col,
	}
}

func isWhitespaceChar(c rune) bool {
	return c == ' ' || c == '\t' || c == '\r' || c == '\n'
}

func isIntChar(c rune) bool {
	return c >= '0' && c <= '9'
}

func isIdentFirstChar(c rune) bool {
	return isIdentChar(c) && !isIntChar(c)
}

func isIdentChar(c rune) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_' || isIntChar(c)
}

func lookupIdentTokenType(ident string) (t TokenType) {
	var ok bool
	if t, ok = keywords[ident]; !ok {
		t = Ident
	}
	return
}
