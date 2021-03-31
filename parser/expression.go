package parser

import (
	"strconv"

	"github.com/blizzy78/copper/ast"
	"github.com/blizzy78/copper/lexer"
)

func (p *Parser) parseExpression(precedence int) (ast.Expression, error) {
	if p.currTokenIs(lexer.EOF) {
		return nil, newParseErrorf(p.currToken.Line, p.currToken.Col, "expression expected")
	}

	parsePrefixFunc, ok := p.prefixParseFuncs[p.currToken.Type]
	if !ok {
		return nil, newParseErrorf(p.currToken.Line, p.currToken.Col, "no prefix parse function found for %s", p.currToken)
	}

	e, err := parsePrefixFunc()
	if err != nil {
		return nil, err
	}

	for !p.currTokenIs(lexer.EOF) {
		currPrec, ok := p.currPrecedence()
		if !ok {
			// next token is not an operator, so let's stop here
			break
		}

		if precedence >= currPrec {
			break
		}

		parseInfixFunc, ok := p.infixParseFuncs[p.currToken.Type]
		if !ok {
			panic(newParseErrorf(p.currToken.Line, p.currToken.Col, "no infix parse function found for %s", p.currToken))
		}

		e, ok, err = parseInfixFunc(e, currPrec)
		if err != nil {
			return nil, err
		}
		if !ok {
			break
		}
	}

	return e, nil
}

func (p *Parser) parseLiteralExpression() (ast.Expression, error) {
	e := ast.Literal{
		StartLine: p.currToken.Line,
		StartCol:  p.currToken.Col,
		Text:      p.currToken.Literal,
	}
	return &e, p.readNextToken()
}

func (p *Parser) parsePrefixExpression() (ast.Expression, error) {
	op := p.currToken.Literal
	line := p.currToken.Line
	col := p.currToken.Col

	if err := p.readNextToken(); err != nil {
		return nil, err
	}

	expr, err := p.parseExpression(precedencePrefix)
	if err != nil {
		return nil, err
	}

	return &ast.PrefixExpression{
		StartLine:  line,
		StartCol:   col,
		Operator:   op,
		Expression: expr,
	}, nil
}

func (p *Parser) parseInfixExpression(left ast.Expression, currPrecedence int) (ast.Expression, bool, error) {
	op := p.currToken.Literal

	if err := p.readNextToken(); err != nil {
		return nil, false, err
	}

	right, err := p.parseExpression(currPrecedence)
	if err != nil {
		return nil, false, err
	}

	return &ast.InfixExpression{
		StartLine: left.Line(),
		StartCol:  left.Col(),
		Left:      left,
		Operator:  op,
		Right:     right,
	}, true, nil
}

func (p *Parser) parseGroupedExpression() (ast.Expression, error) {
	if err := p.readNextToken(); err != nil {
		return nil, err
	}

	e, err := p.parseExpression(precedenceLowest)
	if err != nil {
		return nil, err
	}

	if !p.currTokenIs(lexer.RightParen) {
		return nil, newParseErrorf(p.currToken.Line, p.currToken.Col, "right paren expected")
	}

	return e, p.readNextToken()
}

func (p *Parser) parseIdentExpression() (ast.Expression, error) {
	return p.parseIdentExpr()
}

func (p *Parser) parseIdentExpr() (*ast.Ident, error) {
	e := ast.Ident{
		StartLine: p.currToken.Line,
		StartCol:  p.currToken.Col,
		Name:      p.currToken.Literal,
	}
	return &e, p.readNextToken()
}

func (p *Parser) parseIntLiteral() (ast.Expression, error) {
	value, err := strconv.ParseInt(p.currToken.Literal, 10, 64)
	if err != nil {
		return nil, newParseErrorf(p.currToken.Line, p.currToken.Col, "error parsing int literal: %v", err)
	}

	e := ast.IntLiteral{
		StartLine: p.currToken.Line,
		StartCol:  p.currToken.Col,
		Value:     value,
	}
	return &e, p.readNextToken()
}

func (p *Parser) parseStringLiteral() (ast.Expression, error) {
	e := ast.StringLiteral{
		StartLine: p.currToken.Line,
		StartCol:  p.currToken.Col,
		Value:     p.currToken.Literal,
	}
	return &e, p.readNextToken()
}

func (p *Parser) parseBoolLiteral() (ast.Expression, error) {
	e := ast.BoolLiteral{
		StartLine: p.currToken.Line,
		StartCol:  p.currToken.Col,
		Value:     p.currTokenIs(lexer.True),
	}
	return &e, p.readNextToken()
}

func (p *Parser) parseNilLiteral() (ast.Expression, error) {
	if err := p.readNextToken(); err != nil {
		return nil, err
	}

	return &ast.NilLiteral{
		StartLine: p.currToken.Line,
		StartCol:  p.currToken.Col,
	}, nil
}

func (p *Parser) parseIfExpression() (ast.Expression, error) {
	ifLine := p.currToken.Line
	ifCol := p.currToken.Col

	blockStartTokenType := p.currToken.Type
	blockStartLine := p.currToken.Line
	blockStartCol := p.currToken.Col

	if err := p.readNextToken(); err != nil {
		return nil, err
	}

	conditionals := []ast.ConditionalBlock{}

	haveElse := false
	haveEnd := false

	for !p.currTokenIs(lexer.EOF) {
		switch {
		case blockStartTokenType == lexer.Else && !haveElse:
			haveElse = true
		case blockStartTokenType == lexer.Else:
			return nil, newParseErrorf(blockStartLine, blockStartCol, "if expression can only have a single else block")
		case blockStartTokenType == lexer.ElseIf && haveElse:
			return nil, newParseErrorf(blockStartLine, blockStartCol, "else block must be last in if expression")
		}

		var expr ast.Expression
		if blockStartTokenType == lexer.If || blockStartTokenType == lexer.ElseIf {
			var err error
			expr, err = p.parseExpression(precedenceLowest)
			if err != nil {
				return nil, err
			}
		}

		b, endToken, err := p.parseBlock([]lexer.TokenType{
			lexer.ElseIf,
			lexer.Else,
			lexer.End,
		})
		if err != nil {
			return nil, err
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

	// might not be needed
	if !haveEnd {
		return nil, newParseErrorf(p.currToken.Line, p.currToken.Col, "premature end of file")
	}

	// might not be needed
	if len(conditionals) == 0 {
		panic(newParseErrorf(p.currToken.Line, p.currToken.Col, "no conditionals in if block"))
	}

	return &ast.IfExpression{
		StartLine:    ifLine,
		StartCol:     ifCol,
		Conditionals: conditionals,
	}, nil
}

func (p *Parser) parseForExpression() (ast.Expression, error) {
	line := p.currToken.Line
	col := p.currToken.Col

	if err := p.readNextToken(); err != nil {
		return nil, err
	}

	ident, err := p.parseIdentExpr()
	if err != nil {
		return nil, err
	}

	var statusIdent *ast.Ident
	if p.currTokenIs(lexer.Comma) {
		if err = p.readNextToken(); err != nil {
			return nil, err
		}

		statusIdent, err = p.parseIdentExpr()
		if err != nil {
			return nil, err
		}
	}

	if !p.currTokenIs(lexer.In) {
		return nil, newParseErrorf(p.currToken.Line, p.currToken.Col, "in keyword expected")
	}

	if err = p.readNextToken(); err != nil {
		return nil, err
	}

	expr, err := p.parseExpression(precedenceLowest)
	if err != nil {
		return nil, err
	}

	// TODO: replace all this by call to parseBlock

	blockLine := p.currToken.Line
	blockCol := p.currToken.Col

	stmts := []ast.Statement{}

	for !p.currTokenIs(lexer.EOF) {
		if p.currTokenIs(lexer.End) {
			break
		}

		st, err := p.parseStatement()
		if err != nil {
			return nil, err
		}

		stmts = append(stmts, st)
	}

	if !p.currTokenIs(lexer.End) {
		return nil, newParseErrorf(p.currToken.Line, p.currToken.Col, "end of for expression not found")
	}

	if err = p.readNextToken(); err != nil {
		return nil, err
	}

	return &ast.ForExpression{
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
	}, nil
}

func (p *Parser) parseBlock(endTokenTypes []lexer.TokenType) (*ast.Block, *lexer.Token, error) {
	line := p.currToken.Line
	col := p.currToken.Col

	statements := []ast.Statement{}

	for !p.currTokenIs(lexer.EOF) && !p.currTokenIsOneOf(endTokenTypes) {
		s, err := p.parseStatement()
		if err != nil {
			return nil, nil, err
		}

		statements = append(statements, s)
	}

	endToken := p.currToken

	if p.currTokenIs(lexer.EOF) {
		return nil, nil, newParseErrorf(p.currToken.Line, p.currToken.Col, "end of block not found")
	}

	if err := p.readNextToken(); err != nil {
		return nil, nil, err
	}

	return &ast.Block{
		StartLine:  line,
		StartCol:   col,
		Statements: statements,
	}, endToken, nil
}

func (p *Parser) parseCallExpression(left ast.Expression, currPrecedence int) (ast.Expression, bool, error) {
	if err := p.readNextToken(); err != nil {
		return nil, false, err
	}

	params := []ast.Expression{}

	for !p.currTokenIs(lexer.EOF) {
		if p.currTokenIs(lexer.RightParen) {
			break
		}

		param, err := p.parseExpression(precedenceLowest)
		if err != nil {
			return nil, false, err
		}

		params = append(params, param)

		if p.currTokenIs(lexer.RightParen) {
			break
		}

		if !p.currTokenIs(lexer.Comma) {
			return nil, false, newParseErrorf(p.currToken.Line, p.currToken.Col, "comma expected")
		}

		if err = p.readNextToken(); err != nil {
			return nil, false, err
		}
	}

	if !p.currTokenIs(lexer.RightParen) {
		return nil, false, newParseErrorf(p.currToken.Line, p.currToken.Col, "right paren expected")
	}

	if err := p.readNextToken(); err != nil {
		return nil, false, err
	}

	return &ast.CallExpression{
		StartLine: left.Line(),
		StartCol:  left.Col(),
		Callee:    left,
		Params:    params,
	}, true, nil
}

func (p *Parser) parseFieldExpression(left ast.Expression, currPrecedence int) (ast.Expression, bool, error) {
	dot := p.currTokenIs(lexer.Dot)

	if err := p.readNextToken(); err != nil {
		return nil, false, err
	}

	// x.y -- which is syntactic sugar for: x["y"]
	if dot {
		if !p.currTokenIs(lexer.Ident) {
			return nil, false, newParseErrorf(p.currToken.Line, p.currToken.Col, "expected identifier as field index")
		}

		return &ast.FieldExpression{
			StartLine: left.Line(),
			StartCol:  left.Col(),
			Callee:    left,
			Index: &ast.StringLiteral{
				StartLine: p.currToken.Line,
				StartCol:  p.currToken.Col,
				Value:     p.currToken.Literal,
			},
		}, true, p.readNextToken()
	}

	expr, err := p.parseExpression(precedenceLowest)
	if err != nil {
		return nil, false, err
	}

	if !p.currTokenIs(lexer.RightBracket) {
		return nil, false, newParseErrorf(p.currToken.Line, p.currToken.Col, "expected right bracket")
	}

	e := ast.FieldExpression{
		StartLine: left.Line(),
		StartCol:  left.Col(),
		Callee:    left,
		Index:     expr,
	}
	return &e, true, p.readNextToken()
}

func (p *Parser) parseCaptureExpression() (ast.Expression, error) {
	line := p.currToken.Line
	col := p.currToken.Col

	if err := p.readNextToken(); err != nil {
		return nil, err
	}

	b, _, err := p.parseBlock([]lexer.TokenType{lexer.End})
	if err != nil {
		return nil, err
	}
	return &ast.CaptureExpression{
		StartLine: line,
		StartCol:  col,
		Block:     *b,
	}, nil
}

func (p *Parser) parseHashExpression() (ast.Expression, error) {
	line := p.currToken.Line
	col := p.currToken.Col

	if err := p.readNextToken(); err != nil {
		return nil, err
	}

	values := map[string]ast.Expression{}

	first := true
	for !p.currTokenIs(lexer.EOF) {
		if p.currTokenIs(lexer.RightBrace) {
			break
		}

		if !first {
			if !p.currTokenIs(lexer.Comma) {
				return nil, newParseErrorf(p.currToken.Line, p.currToken.Col, "expected comma before next hash element")
			}

			if err := p.readNextToken(); err != nil {
				return nil, err
			}
		}

		keyLine := p.currToken.Line
		keyCol := p.currToken.Col

		keyExpr, err := p.parseExpression(precedenceLowest)
		if err != nil {
			return nil, err
		}

		if _, ok := keyExpr.(*ast.StringLiteral); !ok {
			return nil, newParseErrorf(keyLine, keyCol, "key in hash expression is not a string: %T", keyExpr)
		}

		if !p.currTokenIs(lexer.Colon) {
			return nil, newParseErrorf(p.currToken.Line, p.currToken.Col, "expected colon after key in hash expression")
		}

		if err = p.readNextToken(); err != nil {
			return nil, err
		}

		key := keyExpr.(*ast.StringLiteral).Value
		if key == "" {
			return nil, newParseErrorf(keyLine, keyCol, "empty key in hash expression")
		}

		if _, ok := values[key]; ok {
			return nil, newParseErrorf(keyLine, keyCol, "duplicate key in hash expression: %s", key)
		}

		value, err := p.parseExpression(precedenceLowest)
		if err != nil {
			return nil, err
		}

		values[key] = value

		first = false
	}

	if !p.currTokenIs(lexer.RightBrace) {
		return nil, newParseErrorf(p.currToken.Line, p.currToken.Col, "expected right brace to end hash expression")
	}

	if err := p.readNextToken(); err != nil {
		return nil, err
	}

	return &ast.HashExpression{
		StartLine: line,
		StartCol:  col,
		Values:    values,
	}, nil
}
