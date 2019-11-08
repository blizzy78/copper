package parser

import (
	"github.com/blizzy78/copper/ast"
	"github.com/blizzy78/copper/lexer"
)

func (p *Parser) parseStatement() (s ast.Statement, err error) {
	switch p.currToken.Type {
	case lexer.Let:
		s, err = p.parseLetStatement()
	case lexer.Break:
		s, err = p.parseBreakStatement()
	case lexer.Continue:
		s, err = p.parseContinueStatement()
	default:
		s, err = p.parseExpressionStatement()
	}

	return
}

func (p *Parser) parseLetStatement() (s *ast.LetStatement, err error) {
	line := p.currToken.Line
	col := p.currToken.Col

	if err = p.expectNext(lexer.Ident); err != nil {
		return
	}

	name := p.currToken.Literal

	if err = p.expectNext(lexer.Assign); err != nil {
		return
	}

	if err = p.readNextToken(); err != nil {
		return
	}

	var expr ast.Expression
	if expr, err = p.parseExpression(precedenceLowest); err != nil {
		return
	}

	s = &ast.LetStatement{
		Ident: ast.Ident{
			StartLine: line,
			StartCol:  col,
			Name:      name,
		},
		Expression: expr,
	}

	return
}

func (p *Parser) parseBreakStatement() (s *ast.BreakStatement, err error) {
	line := p.currToken.Line
	col := p.currToken.Col

	if err = p.readNextToken(); err != nil {
		return
	}

	s = &ast.BreakStatement{
		StartLine: line,
		StartCol:  col,
	}

	return
}

func (p *Parser) parseContinueStatement() (s *ast.ContinueStatement, err error) {
	line := p.currToken.Line
	col := p.currToken.Col

	if err = p.readNextToken(); err != nil {
		return
	}

	s = &ast.ContinueStatement{
		StartLine: line,
		StartCol:  col,
	}

	return
}

func (p *Parser) parseExpressionStatement() (s *ast.ExpressionStatement, err error) {
	line := p.currToken.Line
	col := p.currToken.Col

	var expr ast.Expression
	if expr, err = p.parseExpression(precedenceLowest); err != nil {
		return
	}

	s = &ast.ExpressionStatement{
		StartLine:  line,
		StartCol:   col,
		Expression: expr,
	}

	return
}
