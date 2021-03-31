package parser

import (
	"github.com/blizzy78/copper/ast"
	"github.com/blizzy78/copper/lexer"
)

func (p *Parser) parseStatement() (ast.Statement, error) {
	switch p.currToken.Type {
	case lexer.Let:
		return p.parseLetStatement()
	case lexer.Break:
		return p.parseBreakStatement()
	case lexer.Continue:
		return p.parseContinueStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseLetStatement() (*ast.LetStatement, error) {
	line := p.currToken.Line
	col := p.currToken.Col

	if err := p.expectNext(lexer.Ident); err != nil {
		return nil, err
	}

	name := p.currToken.Literal

	if err := p.expectNext(lexer.Assign); err != nil {
		return nil, err
	}

	if err := p.readNextToken(); err != nil {
		return nil, err
	}

	expr, err := p.parseExpression(precedenceLowest)
	if err != nil {
		return nil, err
	}

	return &ast.LetStatement{
		Ident: ast.Ident{
			StartLine: line,
			StartCol:  col,
			Name:      name,
		},
		Expression: expr,
	}, nil
}

func (p *Parser) parseBreakStatement() (*ast.BreakStatement, error) {
	line := p.currToken.Line
	col := p.currToken.Col

	if err := p.readNextToken(); err != nil {
		return nil, err
	}

	return &ast.BreakStatement{
		StartLine: line,
		StartCol:  col,
	}, nil
}

func (p *Parser) parseContinueStatement() (*ast.ContinueStatement, error) {
	line := p.currToken.Line
	col := p.currToken.Col

	if err := p.readNextToken(); err != nil {
		return nil, err
	}

	return &ast.ContinueStatement{
		StartLine: line,
		StartCol:  col,
	}, nil
}

func (p *Parser) parseExpressionStatement() (*ast.ExpressionStatement, error) {
	line := p.currToken.Line
	col := p.currToken.Col

	expr, err := p.parseExpression(precedenceLowest)
	if err != nil {
		return nil, err
	}

	return &ast.ExpressionStatement{
		StartLine:  line,
		StartCol:   col,
		Expression: expr,
	}, nil
}
