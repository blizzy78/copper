package parser

import (
	"github.com/blizzy78/copper/ast"
	"github.com/blizzy78/copper/lexer"
)

// Parser parses a sequence of lexical tokens produced by a lexer, transforming them to an
// abstract syntax tree. The tree can then be evaluated (executed) by an evaluator.Evaluator.
type Parser struct {
	ch               <-chan *lexer.Token
	doneCh           chan<- struct{}
	currToken        *lexer.Token
	nextToken        *lexer.Token
	prefixParseFuncs map[lexer.TokenType]prefixParseFunc
	infixParseFuncs  map[lexer.TokenType]infixParseFunc
}

type prefixParseFunc func() (expr ast.Expression, err error)

type infixParseFunc func(left ast.Expression, currPrecedence int) (e ast.Expression, ok bool, err error)

const (
	precedenceLowest = iota + 1
	precedenceEquals
	precedenceLessGreater
	precedenceSum
	precedenceProduct
	precedencePrefix
	precedenceField
)

var (
	startToken = lexer.Token{
		Type: "START",
	}

	precedences = map[lexer.TokenType]int{
		lexer.Equal:          precedenceEquals,
		lexer.NotEqual:       precedenceEquals,
		lexer.LessThan:       precedenceLessGreater,
		lexer.LessOrEqual:    precedenceLessGreater,
		lexer.GreaterThan:    precedenceLessGreater,
		lexer.GreaterOrEqual: precedenceLessGreater,
		lexer.Plus:           precedenceSum,
		lexer.Minus:          precedenceSum,
		lexer.Slash:          precedenceProduct,
		lexer.Asterisk:       precedenceProduct,
		lexer.Mod:            precedenceProduct,
		lexer.LeftParen:      precedenceField,
		lexer.Dot:            precedenceField,
		lexer.LeftBracket:    precedenceField,
	}
)

// New returns a new parser that reads a sequence of tokens from tCh. When the parser is done parsing,
// or when an error occurred, it closes doneCh.
func New(tCh <-chan *lexer.Token, doneCh chan<- struct{}) *Parser {
	return &Parser{
		ch:     tCh,
		doneCh: doneCh,
	}
}

// Parse reads the sequence of tokens and transforms it into an abstract syntax tree, a program.
// The tree can be evaluated (executed) by an evaluator.Evaluator.
func (p *Parser) Parse() (prog *ast.Program, err error) {
	defer close(p.doneCh)

	if err = p.initialize(); err != nil {
		return
	}

	line := p.currToken.Line
	col := p.currToken.Col

	var statements []ast.Statement

	for !p.currTokenIs(lexer.EOF) {
		var s ast.Statement
		if s, err = p.parseStatement(); err != nil {
			break
		}

		statements = append(statements, s)
	}

	if err != nil {
		return
	}

	prog = &ast.Program{
		StartLine:  line,
		StartCol:   col,
		Statements: statements,
	}

	return
}

func (p *Parser) initialize() (err error) {
	p.prefixParseFuncs = map[lexer.TokenType]prefixParseFunc{}
	p.registerPrefixParseFunc(lexer.Ident, p.parseIdentExpression)
	p.registerPrefixParseFunc(lexer.Int, p.parseIntLiteral)
	p.registerPrefixParseFunc(lexer.String, p.parseStringLiteral)
	p.registerPrefixParseFunc(lexer.Bang, p.parsePrefixExpression)
	p.registerPrefixParseFunc(lexer.Minus, p.parsePrefixExpression)
	p.registerPrefixParseFunc(lexer.True, p.parseBoolLiteral)
	p.registerPrefixParseFunc(lexer.False, p.parseBoolLiteral)
	p.registerPrefixParseFunc(lexer.LeftParen, p.parseGroupedExpression)
	p.registerPrefixParseFunc(lexer.If, p.parseIfExpression)
	p.registerPrefixParseFunc(lexer.Nil, p.parseNilLiteral)
	p.registerPrefixParseFunc(lexer.Capture, p.parseCaptureExpression)
	p.registerPrefixParseFunc(lexer.For, p.parseForExpression)
	p.registerPrefixParseFunc(lexer.LeftBrace, p.parseHashExpression)
	p.registerPrefixParseFunc(lexer.Literal, p.parseLiteralExpression)

	p.infixParseFuncs = map[lexer.TokenType]infixParseFunc{}
	p.registerInfixParseFunc(lexer.Equal, p.parseInfixExpression)
	p.registerInfixParseFunc(lexer.NotEqual, p.parseInfixExpression)
	p.registerInfixParseFunc(lexer.LessThan, p.parseInfixExpression)
	p.registerInfixParseFunc(lexer.GreaterThan, p.parseInfixExpression)
	p.registerInfixParseFunc(lexer.LessOrEqual, p.parseInfixExpression)
	p.registerInfixParseFunc(lexer.GreaterOrEqual, p.parseInfixExpression)
	p.registerInfixParseFunc(lexer.Plus, p.parseInfixExpression)
	p.registerInfixParseFunc(lexer.Minus, p.parseInfixExpression)
	p.registerInfixParseFunc(lexer.Slash, p.parseInfixExpression)
	p.registerInfixParseFunc(lexer.Asterisk, p.parseInfixExpression)
	p.registerInfixParseFunc(lexer.Mod, p.parseInfixExpression)
	p.registerInfixParseFunc(lexer.LeftParen, p.parseCallExpression)
	p.registerInfixParseFunc(lexer.Dot, p.parseFieldExpression)
	p.registerInfixParseFunc(lexer.LeftBracket, p.parseFieldExpression)

	// prevent nil pointers
	p.currToken = &startToken
	p.nextToken = &startToken

	if err = p.readNextToken(); err != nil {
		return
	}

	if err = p.readNextToken(); err != nil {
		return
	}

	return
}

func (p *Parser) expectNext(t lexer.TokenType) (err error) {
	if !p.nextTokenIs(t) {
		err = newParseErrorf(p.nextToken.Line, p.nextToken.Col, "expected token %s, got %s instead", t, p.nextToken)
		return
	}

	err = p.readNextToken()

	return
}

func (p *Parser) currTokenIs(t lexer.TokenType) bool {
	return p.currToken.Type == t
}

func (p *Parser) currTokenIsOneOf(t []lexer.TokenType) bool {
	for _, tt := range t {
		if p.currTokenIs(tt) {
			return true
		}
	}
	return false
}

func (p *Parser) nextTokenIs(t lexer.TokenType) bool {
	return p.nextToken.Type == t
}

func (p *Parser) readNextToken() (err error) {
	if p.currTokenIs(lexer.EOF) {
		return
	}

	p.currToken = p.nextToken

	if p.currTokenIs(lexer.EOF) {
		return
	}

	p.nextToken = <-p.ch

	if p.nextToken.Err != nil {
		return p.nextToken.Err
	}

	return
}

func (p *Parser) registerPrefixParseFunc(t lexer.TokenType, f prefixParseFunc) {
	p.prefixParseFuncs[t] = f
}

func (p *Parser) registerInfixParseFunc(t lexer.TokenType, f infixParseFunc) {
	p.infixParseFuncs[t] = f
}

func (p *Parser) currPrecedence() (pr int, ok bool) {
	pr, ok = precedences[p.currToken.Type]
	return
}
