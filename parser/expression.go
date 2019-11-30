package parser

import (
	"strconv"

	"github.com/blizzy78/copper/ast"
	"github.com/blizzy78/copper/lexer"
)

func (p *Parser) parseExpression(precedence int) (e ast.Expression, err error) {
	if p.currTokenIs(lexer.EOF) {
		err = newParseErrorf(p.currToken.Line, p.currToken.Col, "expression expected")
		return
	}

	parsePrefixFunc, ok := p.prefixParseFuncs[p.currToken.Type]
	if !ok {
		err = newParseErrorf(p.currToken.Line, p.currToken.Col, "no prefix parse function found for %s", p.currToken)
		return
	}

	if e, err = parsePrefixFunc(); err != nil {
		return
	}

	for !p.currTokenIs(lexer.EOF) {
		var currPrec int
		if currPrec, ok = p.currPrecedence(); !ok {
			// next token is not an operator, so let's stop here
			break
		}

		if precedence >= currPrec {
			break
		}

		var parseInfixFunc infixParseFunc
		if parseInfixFunc, ok = p.infixParseFuncs[p.currToken.Type]; !ok {
			panic(newParseErrorf(p.currToken.Line, p.currToken.Col, "no infix parse function found for %s", p.currToken))
		}

		var ok bool
		if e, ok, err = parseInfixFunc(e, currPrec); !ok || err != nil {
			break
		}
	}

	return
}

func (p *Parser) parseLiteralExpression() (e ast.Expression, err error) {
	e = &ast.Literal{
		StartLine: p.currToken.Line,
		StartCol:  p.currToken.Col,
		Text:      p.currToken.Literal,
	}

	err = p.readNextToken()

	return
}

func (p *Parser) parsePrefixExpression() (e ast.Expression, err error) {
	op := p.currToken.Literal
	line := p.currToken.Line
	col := p.currToken.Col

	if err = p.readNextToken(); err != nil {
		return
	}

	var expr ast.Expression
	if expr, err = p.parseExpression(precedencePrefix); err != nil {
		return
	}

	e = &ast.PrefixExpression{
		StartLine:  line,
		StartCol:   col,
		Operator:   op,
		Expression: expr,
	}

	return
}

func (p *Parser) parseInfixExpression(left ast.Expression, currPrecedence int) (e ast.Expression, ok bool, err error) {
	op := p.currToken.Literal

	if err = p.readNextToken(); err != nil {
		return
	}

	var right ast.Expression
	if right, err = p.parseExpression(currPrecedence); err != nil {
		return
	}

	e = &ast.InfixExpression{
		StartLine: left.Line(),
		StartCol:  left.Col(),
		Left:      left,
		Operator:  op,
		Right:     right,
	}

	ok = true

	return
}

func (p *Parser) parseGroupedExpression() (e ast.Expression, err error) {
	if err = p.readNextToken(); err != nil {
		return
	}

	if e, err = p.parseExpression(precedenceLowest); err != nil {
		return
	}

	if !p.currTokenIs(lexer.RightParen) {
		err = newParseErrorf(p.currToken.Line, p.currToken.Col, "right paren expected")
		return
	}

	err = p.readNextToken()

	return
}

func (p *Parser) parseIdentExpression() (e ast.Expression, err error) {
	return p.parseIdentExpr()
}

func (p *Parser) parseIdentExpr() (e *ast.Ident, err error) {
	e = &ast.Ident{
		StartLine: p.currToken.Line,
		StartCol:  p.currToken.Col,
		Name:      p.currToken.Literal,
	}

	err = p.readNextToken()

	return
}

func (p *Parser) parseIntLiteral() (e ast.Expression, err error) {
	var value int64
	if value, err = strconv.ParseInt(p.currToken.Literal, 10, 64); err != nil {
		err = newParseErrorf(p.currToken.Line, p.currToken.Col, "error parsing int literal: %v", err)
		return
	}

	e = &ast.IntLiteral{
		StartLine: p.currToken.Line,
		StartCol:  p.currToken.Col,
		Value:     value,
	}

	err = p.readNextToken()

	return
}

func (p *Parser) parseStringLiteral() (e ast.Expression, err error) {
	e = &ast.StringLiteral{
		StartLine: p.currToken.Line,
		StartCol:  p.currToken.Col,
		Value:     p.currToken.Literal,
	}

	err = p.readNextToken()

	return
}

func (p *Parser) parseBoolLiteral() (e ast.Expression, err error) {
	e = &ast.BoolLiteral{
		StartLine: p.currToken.Line,
		StartCol:  p.currToken.Col,
		Value:     p.currTokenIs(lexer.True),
	}

	err = p.readNextToken()

	return
}

func (p *Parser) parseNilLiteral() (e ast.Expression, err error) {
	e = &ast.NilLiteral{
		StartLine: p.currToken.Line,
		StartCol:  p.currToken.Col,
	}

	err = p.readNextToken()

	return
}

func (p *Parser) parseIfExpression() (e ast.Expression, err error) {
	ifLine := p.currToken.Line
	ifCol := p.currToken.Col

	blockStartTokenType := p.currToken.Type
	blockStartLine := p.currToken.Line
	blockStartCol := p.currToken.Col

	if err = p.readNextToken(); err != nil {
		return
	}

	conditionals := []ast.ConditionalBlock{}

	haveElse := false
	haveEnd := false

	for !p.currTokenIs(lexer.EOF) {
		if blockStartTokenType == lexer.Else && !haveElse {
			haveElse = true
		} else if blockStartTokenType == lexer.Else {
			err = newParseErrorf(blockStartLine, blockStartCol, "if expression can only have a single else block")
			break
		} else if blockStartTokenType == lexer.ElseIf && haveElse {
			err = newParseErrorf(blockStartLine, blockStartCol, "else block must be last in if expression")
			break
		}

		var expr ast.Expression
		if blockStartTokenType == lexer.If || blockStartTokenType == lexer.ElseIf {
			if expr, err = p.parseExpression(precedenceLowest); err != nil {
				return
			}
		}

		var b *ast.Block
		var endToken *lexer.Token
		b, endToken, err = p.parseBlock([]lexer.TokenType{
			lexer.ElseIf,
			lexer.Else,
			lexer.End,
		})
		if err != nil {
			return
		}

		c := ast.ConditionalBlock{
			StartLine: blockStartLine,
			StartCol:  blockStartCol,
			Condition: expr,
			Block:     *b,
		}

		conditionals = append(conditionals, c)

		if endToken.Type == lexer.End {
			haveEnd = true
			break
		}

		blockStartTokenType = endToken.Type
		blockStartLine = endToken.Line
		blockStartCol = endToken.Col
	}

	if err != nil {
		return
	}

	// might not be needed
	if !haveEnd {
		err = newParseErrorf(p.currToken.Line, p.currToken.Col, "premature end of file")
		return
	}

	// might not be needed
	if len(conditionals) == 0 {
		panic(newParseErrorf(p.currToken.Line, p.currToken.Col, "no conditionals in if block"))
	}

	e = &ast.IfExpression{
		StartLine:    ifLine,
		StartCol:     ifCol,
		Conditionals: conditionals,
	}

	return
}

func (p *Parser) parseForExpression() (e ast.Expression, err error) {
	line := p.currToken.Line
	col := p.currToken.Col

	if err = p.readNextToken(); err != nil {
		return
	}

	var ident *ast.Ident
	if ident, err = p.parseIdentExpr(); err != nil {
		return
	}

	var statusIdent *ast.Ident
	if p.currTokenIs(lexer.Comma) {
		if err = p.readNextToken(); err != nil {
			return
		}

		if statusIdent, err = p.parseIdentExpr(); err != nil {
			return
		}
	}

	if !p.currTokenIs(lexer.In) {
		err = newParseErrorf(p.currToken.Line, p.currToken.Col, "in keyword expected")
		return
	}

	if err = p.readNextToken(); err != nil {
		return
	}

	var expr ast.Expression
	if expr, err = p.parseExpression(precedenceLowest); err != nil {
		return
	}

	// TODO: replace all this by call to parseBlock

	blockLine := p.currToken.Line
	blockCol := p.currToken.Col

	stmts := []ast.Statement{}

	for !p.currTokenIs(lexer.EOF) {
		if p.currTokenIs(lexer.End) {
			break
		}

		var st ast.Statement
		if st, err = p.parseStatement(); err != nil {
			break
		}

		stmts = append(stmts, st)
	}

	if err != nil {
		return
	}

	if !p.currTokenIs(lexer.End) {
		err = newParseErrorf(p.currToken.Line, p.currToken.Col, "end of for expression not found")
		return
	}

	if err = p.readNextToken(); err != nil {
		return
	}

	e = &ast.ForExpression{
		StartLine:   line,
		StartCol:    col,
		Ident:       *ident,
		StatusIdent: statusIdent,
		RangeExpr:   expr,
		Block: ast.Block{
			StartLine:  blockLine,
			StartCol:   blockCol,
			Statements: stmts,
		},
	}

	return
}

func (p *Parser) parseBlock(endTokenTypes []lexer.TokenType) (b *ast.Block, endToken *lexer.Token, err error) {
	line := p.currToken.Line
	col := p.currToken.Col

	statements := []ast.Statement{}

	for !p.currTokenIs(lexer.EOF) && !p.currTokenIsOneOf(endTokenTypes) {
		var s ast.Statement
		if s, err = p.parseStatement(); err != nil {
			break
		}

		statements = append(statements, s)
	}

	if err != nil {
		return
	}

	endToken = p.currToken

	if p.currTokenIs(lexer.EOF) {
		err = newParseErrorf(p.currToken.Line, p.currToken.Col, "end of block not found")
		return
	}

	if err = p.readNextToken(); err != nil {
		return
	}

	b = &ast.Block{
		StartLine:  line,
		StartCol:   col,
		Statements: statements,
	}

	return
}

func (p *Parser) parseCallExpression(left ast.Expression, currPrecedence int) (e ast.Expression, ok bool, err error) {
	if err = p.readNextToken(); err != nil {
		return
	}

	params := []ast.Expression{}

	for !p.currTokenIs(lexer.EOF) {
		if p.currTokenIs(lexer.RightParen) {
			break
		}

		var param ast.Expression
		if param, err = p.parseExpression(precedenceLowest); err != nil {
			break
		}

		params = append(params, param)

		if p.currTokenIs(lexer.RightParen) {
			break
		}

		if !p.currTokenIs(lexer.Comma) {
			err = newParseErrorf(p.currToken.Line, p.currToken.Col, "comma expected")
			break
		}

		if err = p.readNextToken(); err != nil {
			break
		}
	}

	if err != nil {
		return
	}

	if !p.currTokenIs(lexer.RightParen) {
		err = newParseErrorf(p.currToken.Line, p.currToken.Col, "right paren expected")
		return
	}

	if err = p.readNextToken(); err != nil {
		return
	}

	e = &ast.CallExpression{
		StartLine: left.Line(),
		StartCol:  left.Col(),
		Callee:    left,
		Params:    params,
	}

	ok = true

	return
}

func (p *Parser) parseFieldExpression(left ast.Expression, currPrecedence int) (e ast.Expression, ok bool, err error) {
	dot := p.currTokenIs(lexer.Dot)

	if err = p.readNextToken(); err != nil {
		return
	}

	// x.y -- which is syntactic sugar for: x["y"]
	if dot {
		if p.currTokenIs(lexer.Ident) {
			s := ast.StringLiteral{
				StartLine: p.currToken.Line,
				StartCol:  p.currToken.Col,
				Value:     p.currToken.Literal,
			}
			e = &ast.FieldExpression{
				StartLine: left.Line(),
				StartCol:  left.Col(),
				Callee:    left,
				Index:     &s,
			}

			ok = true
		} else {
			err = newParseErrorf(p.currToken.Line, p.currToken.Col, "expected identifier as field index")
		}
	} else if !dot {
		var expr ast.Expression
		if expr, err = p.parseExpression(precedenceLowest); err != nil {
			return
		}

		if !p.currTokenIs(lexer.RightBracket) {
			err = newParseErrorf(p.currToken.Line, p.currToken.Col, "expected right bracket")
			return
		}

		e = &ast.FieldExpression{
			StartLine: left.Line(),
			StartCol:  left.Col(),
			Callee:    left,
			Index:     expr,
		}

		ok = true
	}

	if err == nil {
		err = p.readNextToken()
	}

	return
}

func (p *Parser) parseCaptureExpression() (e ast.Expression, err error) {
	line := p.currToken.Line
	col := p.currToken.Col

	if err = p.readNextToken(); err != nil {
		return
	}

	var b *ast.Block
	if b, _, err = p.parseBlock([]lexer.TokenType{lexer.End}); err != nil {
		return
	}

	e = &ast.CaptureExpression{
		StartLine: line,
		StartCol:  col,
		Block:     *b,
	}

	return
}

func (p *Parser) parseHashExpression() (e ast.Expression, err error) {
	line := p.currToken.Line
	col := p.currToken.Col

	if err = p.readNextToken(); err != nil {
		return
	}

	values := map[string]ast.Expression{}

	first := true
	for !p.currTokenIs(lexer.EOF) {
		if p.currTokenIs(lexer.RightBrace) {
			break
		}

		if !first {
			if !p.currTokenIs(lexer.Comma) {
				err = newParseErrorf(p.currToken.Line, p.currToken.Col, "expected comma before next hash element")
				break
			}

			if err = p.readNextToken(); err != nil {
				return
			}
		}

		keyLine := p.currToken.Line
		keyCol := p.currToken.Col

		var keyExpr ast.Expression
		if keyExpr, err = p.parseExpression(precedenceLowest); err != nil {
			break
		}

		if _, ok := keyExpr.(*ast.StringLiteral); !ok {
			err = newParseErrorf(keyLine, keyCol, "key in hash expression is not a string: %T", keyExpr)
			break
		}

		if !p.currTokenIs(lexer.Colon) {
			err = newParseErrorf(p.currToken.Line, p.currToken.Col, "expected colon after key in hash expression")
			break
		}

		if err = p.readNextToken(); err != nil {
			return
		}

		key := keyExpr.(*ast.StringLiteral).Value
		if key == "" {
			err = newParseErrorf(keyLine, keyCol, "empty key in hash expression")
			break
		}

		if _, ok := values[key]; ok {
			err = newParseErrorf(keyLine, keyCol, "duplicate key in hash expression: %s", key)
			break
		}

		var value ast.Expression
		if value, err = p.parseExpression(precedenceLowest); err != nil {
			break
		}

		values[key] = value

		first = false
	}

	if err != nil {
		return
	}

	if !p.currTokenIs(lexer.RightBrace) {
		err = newParseErrorf(p.currToken.Line, p.currToken.Col, "expected right brace to end hash expression")
		return
	}

	if err = p.readNextToken(); err != nil {
		return
	}

	e = &ast.HashExpression{
		StartLine: line,
		StartCol:  col,
		Values:    values,
	}

	return
}
