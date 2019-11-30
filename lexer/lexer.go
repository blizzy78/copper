package lexer

import (
	"bufio"
	"io"
	"strings"
)

// Lexer parses a series of statements or expressions, a template, from a reader and returns them
// as a sequence of lexical tokens.
type Lexer struct {
	r              io.RuneReader
	optStartInCode bool
	line           int
	col            int
	currChar       rune
	nextChar       rune
	currEOF        bool
	nextEOF        bool
}

// Opt is the type of a function that configures an option of l.
type Opt func(l *Lexer)

type stateFunc func(tCh chan<- *Token) stateFunc

var (
	keywords = map[string]TokenType{
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

	startState := l.parseLiteral
	if l.optStartInCode {
		startState = l.parseCode
	}

	if err := l.initialize(); err != nil {
		startState = l.parseError(err, l.line, l.col)
	}

	go func(state stateFunc) {
		defer close(tokenCh)

		for state != nil {
			if l.currEOF {
				state = l.parseEOF
			}

			state = state(tokenCh)
		}
	}(startState)

	return
}

func (l *Lexer) parseLiteral(tCh chan<- *Token) stateFunc {
	buf := strings.Builder{}

	defer emitTokenBuffer(tCh, Literal, &buf, l.line, l.col)

	for {
		if l.currEOF {
			return l.parseEOF
		}

		if l.isAtCodeStart() {
			return l.parseCodeStart
		}

		if _, err := buf.WriteRune(l.currChar); err != nil {
			return l.parseError(err, l.line, l.col)
		}

		if err := l.readNextChar(); err != nil {
			return l.parseError(err, l.line, l.col)
		}
	}
}

func (l *Lexer) parseEOF(tCh chan<- *Token) stateFunc {
	tCh <- newToken(EOF, "", l.line, l.col)
	return nil
}

func (l *Lexer) parseCodeStart(tCh chan<- *Token) stateFunc {
	if err := l.readNextChar(); err != nil {
		return l.parseError(err, l.line, l.col)
	}
	if err := l.readNextChar(); err != nil {
		return l.parseError(err, l.line, l.col)
	}

	return l.parseCode
}

func (l *Lexer) parseCodeEnd(tCh chan<- *Token) stateFunc {
	if err := l.readNextChar(); err != nil {
		return l.parseError(err, l.line, l.col)
	}
	if err := l.readNextChar(); err != nil {
		return l.parseError(err, l.line, l.col)
	}

	return l.parseLiteral
}

func (l *Lexer) parseCode(tCh chan<- *Token) stateFunc {
	if err := l.skipWhitespace(); err != nil {
		return l.parseError(err, l.line, l.col)
	}

	if isIntChar(l.currChar) {
		return l.parseInt
	}

	if isIdentFirstChar(l.currChar) {
		return l.parseIdent
	}

	switch l.currChar {
	case '"':
		fallthrough
	case '\'':
		return l.parseString
	case '=':
		return l.parseAssignOrEqual
	case '+':
		return l.parseToken(Plus, "+")
	case '-':
		return l.parseToken(Minus, "-")
	case '*':
		return l.parseToken(Asterisk, "*")
	case '/':
		return l.parseSlashOrComment
	case '!':
		return l.parseBangOrNotEqual
	case '%':
		return l.parseModOrCodeEnd
	case '(':
		return l.parseToken(LeftParen, "(")
	case ')':
		return l.parseToken(RightParen, ")")
	case '[':
		return l.parseToken(LeftBracket, "[")
	case ']':
		return l.parseToken(RightBracket, "]")
	case '{':
		return l.parseToken(LeftBrace, "{")
	case '}':
		return l.parseToken(RightBrace, "}")
	case '.':
		return l.parseToken(Dot, ".")
	case ',':
		return l.parseToken(Comma, ",")
	case ':':
		return l.parseToken(Colon, ":")
	case '<':
		return l.parseLessThanOrLessEqual
	case '>':
		return l.parseGreaterThanOrGreaterEqual
	}

	return l.parseIllegal
}

func (l *Lexer) parseInt(tCh chan<- *Token) stateFunc {
	buf := strings.Builder{}

	defer emitTokenBuffer(tCh, Int, &buf, l.line, l.col)

	for {
		if l.currEOF {
			return l.parseEOF
		}

		if !isIntChar(l.currChar) {
			return l.parseCode
		}

		if _, err := buf.WriteRune(l.currChar); err != nil {
			return l.parseError(err, l.line, l.col)
		}

		if err := l.readNextChar(); err != nil {
			return l.parseError(err, l.line, l.col)
		}
	}
}

func (l *Lexer) parseIdent(tCh chan<- *Token) stateFunc {
	b := strings.Builder{}

	defer func(line int, col int) {
		literal := b.String()

		var t TokenType
		var ok bool
		if t, ok = keywords[literal]; !ok {
			t = Ident
		}

		tCh <- newToken(t, literal, line, col)
	}(l.line, l.col)

	for {
		if l.currEOF {
			return l.parseEOF
		}

		if !isIdentChar(l.currChar) {
			return l.parseCode
		}

		if _, err := b.WriteRune(l.currChar); err != nil {
			return l.parseError(err, l.line, l.col)
		}

		if err := l.readNextChar(); err != nil {
			return l.parseError(err, l.line, l.col)
		}
	}
}

func (l *Lexer) parseString(tCh chan<- *Token) stateFunc {
	startChar := l.currChar

	buf := strings.Builder{}

	defer emitTokenBuffer(tCh, String, &buf, l.line, l.col)

	if err := l.readNextChar(); err != nil {
		return l.parseError(err, l.line, l.col)
	}

	prevBackslash := false

	for {
		if l.currEOF {
			return l.parseEOF
		}

		if l.currChar == startChar && !prevBackslash {
			break
		}

		if _, err := buf.WriteRune(l.currChar); err != nil {
			return l.parseError(err, l.line, l.col)
		}

		prevBackslash = l.currChar == '\\'

		if err := l.readNextChar(); err != nil {
			return l.parseError(err, l.line, l.col)
		}
	}

	if err := l.readNextChar(); err != nil {
		return l.parseError(err, l.line, l.col)
	}

	s := buf.String()
	s = strings.ReplaceAll(s, `\r`, "\r")
	s = strings.ReplaceAll(s, `\n`, "\n")
	s = strings.ReplaceAll(s, `\t`, "\t")
	s = strings.ReplaceAll(s, `\"`, `"`)
	s = strings.ReplaceAll(s, `\'`, "'")
	s = strings.ReplaceAll(s, `\\`, "\\")
	buf.Reset()
	buf.WriteString(s)

	return l.parseCode
}

func (l *Lexer) parseLineComment(tCh chan<- *Token) stateFunc {
	if err := l.readNextChar(); err != nil {
		return l.parseError(err, l.line, l.col)
	}
	if err := l.readNextChar(); err != nil {
		return l.parseError(err, l.line, l.col)
	}

	for {
		if l.currEOF {
			return l.parseEOF
		}

		if l.currChar == '\n' {
			return l.parseCode
		}

		if !l.optStartInCode && l.isAtCodeEnd() {
			return l.parseCodeEnd
		}

		if err := l.readNextChar(); err != nil {
			return l.parseError(err, l.line, l.col)
		}
	}
}

func (l *Lexer) parseBlockComment(tCh chan<- *Token) stateFunc {
	if err := l.readNextChar(); err != nil {
		return l.parseError(err, l.line, l.col)
	}
	if err := l.readNextChar(); err != nil {
		return l.parseError(err, l.line, l.col)
	}

	for {
		if l.currEOF {
			return l.parseEOF
		}

		if l.currChar == '*' && l.nextCharIs('/') {
			if err := l.readNextChar(); err != nil {
				return l.parseError(err, l.line, l.col)
			}
			if err := l.readNextChar(); err != nil {
				return l.parseError(err, l.line, l.col)
			}

			return l.parseCode
		}

		if err := l.readNextChar(); err != nil {
			return l.parseError(err, l.line, l.col)
		}
	}
}

func (l *Lexer) parseAssignOrEqual(tCh chan<- *Token) stateFunc {
	if l.nextCharIs('=') {
		return l.parseToken(Equal, "==")
	}

	return l.parseToken(Assign, "=")
}

func (l *Lexer) parseBangOrNotEqual(tCh chan<- *Token) stateFunc {
	if l.nextCharIs('=') {
		return l.parseToken(NotEqual, "!=")
	}

	return l.parseToken(Bang, "!")
}

func (l *Lexer) parseModOrCodeEnd(tCh chan<- *Token) stateFunc {
	if l.nextCharIs('>') {
		return l.parseCodeEnd
	}

	return l.parseToken(Mod, "%")
}

func (l *Lexer) parseLessThanOrLessEqual(tCh chan<- *Token) stateFunc {
	if l.nextCharIs('=') {
		return l.parseToken(LessOrEqual, "<=")
	}

	return l.parseToken(LessThan, "<")
}

func (l *Lexer) parseGreaterThanOrGreaterEqual(tCh chan<- *Token) stateFunc {
	if l.nextCharIs('=') {
		return l.parseToken(GreaterOrEqual, ">=")
	}

	return l.parseToken(GreaterThan, ">")
}

func (l *Lexer) parseSlashOrComment(tCh chan<- *Token) stateFunc {
	if l.nextCharIs('/') {
		return l.parseLineComment
	}

	if l.nextCharIs('*') {
		return l.parseBlockComment
	}

	return l.parseToken(Slash, "/")
}

func (l *Lexer) parseToken(t TokenType, literal string) stateFunc {
	return func(tCh chan<- *Token) stateFunc {
		tCh <- newToken(t, literal, l.line, l.col)

		for range literal {
			if err := l.readNextChar(); err != nil {
				return l.parseError(err, l.line, l.col)
			}
		}

		return l.parseCode
	}
}

func (l *Lexer) parseIllegal(tCh chan<- *Token) stateFunc {
	buf := strings.Builder{}

	defer emitTokenBuffer(tCh, Illegal, &buf, l.line, l.col)

	if _, err := buf.WriteRune(l.currChar); err != nil {
		return l.parseError(err, l.line, l.col)
	}

	return nil
}

func (l *Lexer) parseError(err error, line int, col int) stateFunc {
	return func(tCh chan<- *Token) stateFunc {
		tCh <- newErrorToken(err, line, col)
		return nil
	}
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

func (l *Lexer) isAtCodeStart() bool {
	return l.currChar == '<' && l.nextCharIs('%')
}

func (l *Lexer) isAtCodeEnd() bool {
	return l.currChar == '%' && l.nextCharIs('>')
}

func (l *Lexer) isAtBlockCommentEnd() bool {
	return l.currChar == '*' && l.nextCharIs('/')
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

func emitTokenBuffer(tCh chan<- *Token, t TokenType, buf *strings.Builder, line int, col int) {
	tCh <- newToken(t, buf.String(), line, col)
}

func newToken(t TokenType, literal string, line int, col int) *Token {
	return &Token{
		Type:    t,
		Literal: literal,
		Line:    line,
		Col:     col,
	}
}

func newErrorToken(err error, line int, col int) *Token {
	return &Token{
		Type: Error,
		Line: line,
		Col:  col,
		Err:  newParseError(err, line, col),
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
