package parser

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/blizzy78/copper/ast"
	"github.com/blizzy78/copper/lexer"
)

func TestParserStartInLiteral(t *testing.T) {
	tests := []struct {
		input   string
		program *ast.Program
	}{
		{
			`foo <% 5 + 6 * 7 %> bar`,
			&ast.Program{
				Statements: []ast.Statement{
					&ast.ExpressionStatement{
						Expression: newLiteral("foo "),
					},
					&ast.ExpressionStatement{
						Expression: &ast.InfixExpression{
							Left:     newIntLiteral(5),
							Operator: "+",
							Right: &ast.InfixExpression{
								Left:     newIntLiteral(6),
								Operator: "*",
								Right:    newIntLiteral(7),
							},
						},
					},
					&ast.ExpressionStatement{
						Expression: newLiteral(" bar"),
					},
				},
			},
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			testParser(test.input, test.program, t)
		})
	}
}

func TestParseExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected ast.Expression
	}{
		{
			"true",
			&ast.BoolLiteral{
				Value: true,
			},
		},
		{
			"false",
			&ast.BoolLiteral{
				Value: false,
			},
		},
		{
			"-a * b",
			&ast.InfixExpression{
				Left: &ast.PrefixExpression{
					Operator: "-",
					Expression: &ast.Ident{
						Name: "a",
					},
				},
				Operator: "*",
				Right: &ast.Ident{
					Name: "b",
				},
			},
		},
		{
			"!-a",
			&ast.PrefixExpression{
				Operator: "!",
				Expression: &ast.PrefixExpression{
					Operator: "-",
					Expression: &ast.Ident{
						Name: "a",
					},
				},
			},
		},
		{
			"a + b + c",
			&ast.InfixExpression{
				Left: &ast.InfixExpression{
					Left: &ast.Ident{
						Name: "a",
					},
					Operator: "+",
					Right: &ast.Ident{
						Name: "b",
					},
				},
				Operator: "+",
				Right: &ast.Ident{
					Name: "c",
				},
			},
		},
		{
			"a + b - c",
			&ast.InfixExpression{
				Left: &ast.InfixExpression{
					Left: &ast.Ident{
						Name: "a",
					},
					Operator: "+",
					Right: &ast.Ident{
						Name: "b",
					},
				},
				Operator: "-",
				Right: &ast.Ident{
					Name: "c",
				},
			},
		},
		{
			"a * b * c",
			&ast.InfixExpression{
				Left: &ast.InfixExpression{
					Left: &ast.Ident{
						Name: "a",
					},
					Operator: "*",
					Right: &ast.Ident{
						Name: "b",
					},
				},
				Operator: "*",
				Right: &ast.Ident{
					Name: "c",
				},
			},
		},
		{
			"a * b / c",
			&ast.InfixExpression{
				Left: &ast.InfixExpression{
					Left: &ast.Ident{
						Name: "a",
					},
					Operator: "*",
					Right: &ast.Ident{
						Name: "b",
					},
				},
				Operator: "/",
				Right: &ast.Ident{
					Name: "c",
				},
			},
		},
		{
			"a + b / c",
			&ast.InfixExpression{
				Left: &ast.Ident{
					Name: "a",
				},
				Operator: "+",
				Right: &ast.InfixExpression{
					Left: &ast.Ident{
						Name: "b",
					},
					Operator: "/",
					Right: &ast.Ident{
						Name: "c",
					},
				},
			},
		},
		{
			"a || b && c",
			&ast.InfixExpression{
				Left: &ast.Ident{
					Name: "a",
				},
				Operator: "||",
				Right: &ast.InfixExpression{
					Left: &ast.Ident{
						Name: "b",
					},
					Operator: "&&",
					Right: &ast.Ident{
						Name: "c",
					},
				},
			},
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			l := newLexerString(test.input, t, lexer.WithStartInCodeMode())
			prog := parse(l, t)
			testExpression(prog.Statements[0].(*ast.ExpressionStatement).Expression, test.expected, t)
		})
	}
}

func TestParseExpressionBool(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{
			"true",
			true,
		},
		{
			"false",
			false,
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			l := newLexerString(test.input, t, lexer.WithStartInCodeMode())
			prog := parse(l, t)
			b := prog.Statements[0].(*ast.ExpressionStatement).Expression.(*ast.BoolLiteral)
			if b.Value != test.expected {
				t.Fatalf("wrong boolean literal, expected=%t, got=%t", test.expected, b.Value)
			}
		})
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		input    string
		expected []ast.Statement
	}{
		{
			`x`,
			[]ast.Statement{
				&ast.ExpressionStatement{
					Expression: newIdent("x"),
				},
			},
		},
		{
			`5`,
			[]ast.Statement{
				&ast.ExpressionStatement{
					Expression: newIntLiteral(5),
				},
			},
		},
		{
			`!x`,
			[]ast.Statement{
				&ast.ExpressionStatement{
					Expression: &ast.PrefixExpression{
						Operator:   "!",
						Expression: newIdent("x"),
					},
				},
			},
		},
		{
			`-5`,
			[]ast.Statement{
				&ast.ExpressionStatement{
					Expression: &ast.PrefixExpression{
						Operator:   "-",
						Expression: newIntLiteral(5),
					},
				},
			},
		},
		{
			`5 + 6 * 7`,
			[]ast.Statement{
				&ast.ExpressionStatement{
					Expression: &ast.InfixExpression{
						Left:     newIntLiteral(5),
						Operator: "+",
						Right: &ast.InfixExpression{
							Left:     newIntLiteral(6),
							Operator: "*",
							Right:    newIntLiteral(7),
						},
					},
				},
			},
		},
		{
			`let x = 5`,
			[]ast.Statement{
				&ast.LetStatement{
					Ident: ast.Ident{
						Name: "x",
					},
					Expression: newIntLiteral(5),
				},
			},
		},
		{
			`true`,
			[]ast.Statement{
				&ast.ExpressionStatement{
					Expression: newBoolLiteral(true),
				},
			},
		},
		{
			`false`,
			[]ast.Statement{
				&ast.ExpressionStatement{
					Expression: newBoolLiteral(false),
				},
			},
		},
		{
			`nil`,
			[]ast.Statement{
				&ast.ExpressionStatement{
					Expression: newNilLiteral(),
				},
			},
		},
		{
			"let x = 5",
			[]ast.Statement{
				&ast.LetStatement{
					Ident: ast.Ident{
						Name: "x",
					},
					Expression: newIntLiteral(5),
				},
			},
		},
		{
			`if x == 5
			   y
			 end
			 `,
			[]ast.Statement{
				&ast.ExpressionStatement{
					Expression: &ast.IfExpression{
						Conditionals: []ast.ConditionalBlock{
							{
								Condition: &ast.InfixExpression{
									Left:     newIdent("x"),
									Operator: "==",
									Right:    newIntLiteral(5),
								},
								Block: ast.Block{
									Statements: []ast.Statement{
										&ast.ExpressionStatement{
											Expression: newIdent("y"),
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			`if a == 5
			   b
			 elseif c == 6
			   d
			 end`,
			[]ast.Statement{
				&ast.ExpressionStatement{
					Expression: &ast.IfExpression{
						Conditionals: []ast.ConditionalBlock{
							{
								Condition: &ast.InfixExpression{
									Left:     newIdent("a"),
									Operator: "==",
									Right:    newIntLiteral(5),
								},
								Block: ast.Block{
									Statements: []ast.Statement{
										&ast.ExpressionStatement{
											Expression: newIdent("b"),
										},
									},
								},
							},
							{
								Condition: &ast.InfixExpression{
									Left:     newIdent("c"),
									Operator: "==",
									Right:    newIntLiteral(6),
								},
								Block: ast.Block{
									Statements: []ast.Statement{
										&ast.ExpressionStatement{
											Expression: newIdent("d"),
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			`if a == 5
			   b
			 elseif c == 6
			   d
			 elseif e == 7
			   f
			 end`,
			[]ast.Statement{
				&ast.ExpressionStatement{
					Expression: &ast.IfExpression{
						Conditionals: []ast.ConditionalBlock{
							{
								Condition: &ast.InfixExpression{
									Left:     newIdent("a"),
									Operator: "==",
									Right:    newIntLiteral(5),
								},
								Block: ast.Block{
									Statements: []ast.Statement{
										&ast.ExpressionStatement{
											Expression: newIdent("b"),
										},
									},
								},
							},
							{
								Condition: &ast.InfixExpression{
									Left:     newIdent("c"),
									Operator: "==",
									Right:    newIntLiteral(6),
								},
								Block: ast.Block{
									Statements: []ast.Statement{
										&ast.ExpressionStatement{
											Expression: newIdent("d"),
										},
									},
								},
							},
							{
								Condition: &ast.InfixExpression{
									Left:     newIdent("e"),
									Operator: "==",
									Right:    newIntLiteral(7),
								},
								Block: ast.Block{
									Statements: []ast.Statement{
										&ast.ExpressionStatement{
											Expression: newIdent("f"),
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			`if a == 5
			   b
			 elseif c == 6
			   d
			 elseif e == 7
			   f
			 else
			   g
			 end`,
			[]ast.Statement{
				&ast.ExpressionStatement{
					Expression: &ast.IfExpression{
						Conditionals: []ast.ConditionalBlock{
							{
								Condition: &ast.InfixExpression{
									Left:     newIdent("a"),
									Operator: "==",
									Right:    newIntLiteral(5),
								},
								Block: ast.Block{
									Statements: []ast.Statement{
										&ast.ExpressionStatement{
											Expression: newIdent("b"),
										},
									},
								},
							},
							{
								Condition: &ast.InfixExpression{
									Left:     newIdent("c"),
									Operator: "==",
									Right:    newIntLiteral(6),
								},
								Block: ast.Block{
									Statements: []ast.Statement{
										&ast.ExpressionStatement{
											Expression: newIdent("d"),
										},
									},
								},
							},
							{
								Condition: &ast.InfixExpression{
									Left:     newIdent("e"),
									Operator: "==",
									Right:    newIntLiteral(7),
								},
								Block: ast.Block{
									Statements: []ast.Statement{
										&ast.ExpressionStatement{
											Expression: newIdent("f"),
										},
									},
								},
							},
							{
								Block: ast.Block{
									Statements: []ast.Statement{
										&ast.ExpressionStatement{
											Expression: newIdent("g"),
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			`if a == 5
			   b
			 else
			   c
			 end`,
			[]ast.Statement{
				&ast.ExpressionStatement{
					Expression: &ast.IfExpression{
						Conditionals: []ast.ConditionalBlock{
							{
								Condition: &ast.InfixExpression{
									Left:     newIdent("a"),
									Operator: "==",
									Right:    newIntLiteral(5),
								},
								Block: ast.Block{
									Statements: []ast.Statement{
										&ast.ExpressionStatement{
											Expression: newIdent("b"),
										},
									},
								},
							},
							{
								Block: ast.Block{
									Statements: []ast.Statement{
										&ast.ExpressionStatement{
											Expression: newIdent("c"),
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			"x()",
			[]ast.Statement{
				&ast.ExpressionStatement{
					Expression: &ast.CallExpression{
						Callee: newIdent("x"),
					},
				},
			},
		},
		{
			"x(y)",
			[]ast.Statement{
				&ast.ExpressionStatement{
					Expression: &ast.CallExpression{
						Callee: newIdent("x"),
						Params: []ast.Expression{
							newIdent("y"),
						},
					},
				},
			},
		},
		{
			"x(y, z)",
			[]ast.Statement{
				&ast.ExpressionStatement{
					Expression: &ast.CallExpression{
						Callee: newIdent("x"),
						Params: []ast.Expression{
							newIdent("y"),
							newIdent("z"),
						},
					},
				},
			},
		},
		{
			"x(1 * 2, 3 + 4, 5 / y)",
			[]ast.Statement{
				&ast.ExpressionStatement{
					Expression: &ast.CallExpression{
						Callee: newIdent("x"),
						Params: []ast.Expression{
							&ast.InfixExpression{
								Left:     newIntLiteral(1),
								Operator: "*",
								Right:    newIntLiteral(2),
							},
							&ast.InfixExpression{
								Left:     newIntLiteral(3),
								Operator: "+",
								Right:    newIntLiteral(4),
							},
							&ast.InfixExpression{
								Left:     newIntLiteral(5),
								Operator: "/",
								Right:    newIdent("y"),
							},
						},
					},
				},
			},
		},
		{
			`"abc" == 'def'`,
			[]ast.Statement{
				&ast.ExpressionStatement{
					Expression: &ast.InfixExpression{
						Left:     newStringLiteral("abc"),
						Operator: "==",
						Right:    newStringLiteral("def"),
					},
				},
			},
		},
		{
			`a.b == x["y"]`,
			[]ast.Statement{
				&ast.ExpressionStatement{
					Expression: &ast.InfixExpression{
						Left: &ast.FieldExpression{
							Callee: newIdent("a"),
							Index:  newStringLiteral("b"),
						},
						Operator: "==",
						Right: &ast.FieldExpression{
							Callee: newIdent("x"),
							Index:  newStringLiteral("y"),
						},
					},
				},
			},
		},
		{
			`a["b"] != x.y`,
			[]ast.Statement{
				&ast.ExpressionStatement{
					Expression: &ast.InfixExpression{
						Left: &ast.FieldExpression{
							Callee: newIdent("a"),
							Index:  newStringLiteral("b"),
						},
						Operator: "!=",
						Right: &ast.FieldExpression{
							Callee: newIdent("x"),
							Index:  newStringLiteral("y"),
						},
					},
				},
			},
		},
		{
			`a.b.c.d`,
			[]ast.Statement{
				&ast.ExpressionStatement{
					Expression: &ast.FieldExpression{
						Callee: &ast.FieldExpression{
							Callee: &ast.FieldExpression{
								Callee: newIdent("a"),
								Index:  newStringLiteral("b"),
							},
							Index: newStringLiteral("c"),
						},
						Index: newStringLiteral("d"),
					},
				},
			},
		},
		{
			`a["b"]["c"]["d"]`,
			[]ast.Statement{
				&ast.ExpressionStatement{
					Expression: &ast.FieldExpression{
						Callee: &ast.FieldExpression{
							Callee: &ast.FieldExpression{
								Callee: newIdent("a"),
								Index:  newStringLiteral("b"),
							},
							Index: newStringLiteral("c"),
						},
						Index: newStringLiteral("d"),
					},
				},
			},
		},
		{
			`a.b["c"].d`,
			[]ast.Statement{
				&ast.ExpressionStatement{
					Expression: &ast.FieldExpression{
						Callee: &ast.FieldExpression{
							Callee: &ast.FieldExpression{
								Callee: newIdent("a"),
								Index:  newStringLiteral("b"),
							},
							Index: newStringLiteral("c"),
						},
						Index: newStringLiteral("d"),
					},
				},
			},
		},
		{
			`a["b"].c["d"]`,
			[]ast.Statement{
				&ast.ExpressionStatement{
					Expression: &ast.FieldExpression{
						Callee: &ast.FieldExpression{
							Callee: &ast.FieldExpression{
								Callee: newIdent("a"),
								Index:  newStringLiteral("b"),
							},
							Index: newStringLiteral("c"),
						},
						Index: newStringLiteral("d"),
					},
				},
			},
		},
		{
			`a.b(x)["c"].d`,
			[]ast.Statement{
				&ast.ExpressionStatement{
					Expression: &ast.FieldExpression{
						Callee: &ast.FieldExpression{
							Callee: &ast.CallExpression{
								Callee: &ast.FieldExpression{
									Callee: newIdent("a"),
									Index:  newStringLiteral("b"),
								},
								Params: []ast.Expression{
									newIdent("x"),
								},
							},
							Index: newStringLiteral("c"),
						},
						Index: newStringLiteral("d"),
					},
				},
			},
		},
		{
			`{ "x": 42, "y": "foo" }`,
			[]ast.Statement{
				&ast.ExpressionStatement{
					Expression: &ast.HashExpression{
						Values: map[string]ast.Expression{
							"x": newIntLiteral(42),
							"y": newStringLiteral("foo"),
						},
					},
				},
			},
		},
		{
			`break
			continue`,
			[]ast.Statement{
				&ast.BreakStatement{},
				&ast.ContinueStatement{},
			},
		},
		{
			`for i in range(x)
			  "foo"
			end`,
			[]ast.Statement{
				&ast.ExpressionStatement{
					Expression: &ast.ForExpression{
						Ident: ast.Ident{
							Name: "i",
						},
						RangeExpr: &ast.CallExpression{
							Callee: newIdent("range"),
							Params: []ast.Expression{
								newIdent("x"),
							},
						},
						Block: ast.Block{
							Statements: []ast.Statement{
								&ast.ExpressionStatement{
									Expression: newStringLiteral("foo"),
								},
							},
						},
					},
				},
			},
		},
		{
			`for i, st in range(x)
			  "foo"
			end`,
			[]ast.Statement{
				&ast.ExpressionStatement{
					Expression: &ast.ForExpression{
						Ident: ast.Ident{
							Name: "i",
						},
						StatusIdent: &ast.Ident{
							Name: "st",
						},
						RangeExpr: &ast.CallExpression{
							Callee: newIdent("range"),
							Params: []ast.Expression{
								newIdent("x"),
							},
						},
						Block: ast.Block{
							Statements: []ast.Statement{
								&ast.ExpressionStatement{
									Expression: newStringLiteral("foo"),
								},
							},
						},
					},
				},
			},
		},
		{
			`capture
			  "foo"
			end`,
			[]ast.Statement{
				&ast.ExpressionStatement{
					Expression: &ast.CaptureExpression{
						Block: ast.Block{
							Statements: []ast.Statement{
								&ast.ExpressionStatement{
									Expression: newStringLiteral("foo"),
								},
							},
						},
					},
				},
			},
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			testParser(test.input, &ast.Program{Statements: test.expected}, t, lexer.WithStartInCodeMode())
		})
	}
}

func testStatement(actual ast.Statement, expected ast.Statement, t *testing.T) {
	t.Helper()

	if reflect.TypeOf(actual) != reflect.TypeOf(expected) {
		t.Fatalf("wrong statement type, expected=%T, got=%T", expected, actual)
	}

	switch ex := expected.(type) {
	case *ast.LetStatement:
		testLetStatement(actual.(*ast.LetStatement), ex, t)
	case *ast.ExpressionStatement:
		testExpressionStatement(actual.(*ast.ExpressionStatement), ex, t)
	case *ast.BreakStatement:
		// okay
	case *ast.ContinueStatement:
		// okay
	default:
		t.Fatalf("unknown statement type: %T", expected)
	}
}

func testLetStatement(actual *ast.LetStatement, expected *ast.LetStatement, t *testing.T) {
	t.Helper()

	testIdentifier(&actual.Ident, &expected.Ident, t)
	testExpression(actual.Expression, expected.Expression, t)
}

func testExpressionStatement(actual *ast.ExpressionStatement, expected *ast.ExpressionStatement, t *testing.T) {
	t.Helper()

	testExpression(actual.Expression, expected.Expression, t)
}

func testExpression(actual ast.Expression, expected ast.Expression, t *testing.T) { //nolint:gocyclo
	t.Helper()

	if reflect.TypeOf(actual) != reflect.TypeOf(expected) {
		t.Fatalf("wrong expression type, expected=%T, got=%T", expected, actual)
	}

	switch ex := expected.(type) {
	case *ast.Literal:
		testLiteral(actual.(*ast.Literal), ex, t)
	case *ast.Ident:
		testIdentifier(actual.(*ast.Ident), ex, t)
	case *ast.IntLiteral:
		testIntLiteral(actual.(*ast.IntLiteral), ex, t)
	case *ast.StringLiteral:
		testStringLiteral(actual.(*ast.StringLiteral), ex, t)
	case *ast.BoolLiteral:
		testBoolLiteral(actual.(*ast.BoolLiteral), ex, t)
	case *ast.NilLiteral:
		// okay
	case *ast.IfExpression:
		testIfExpression(actual.(*ast.IfExpression), ex, t)
	case *ast.InfixExpression:
		testInfixExpression(actual.(*ast.InfixExpression), ex, t)
	case *ast.CallExpression:
		testCallExpression(actual.(*ast.CallExpression), ex, t)
	case *ast.FieldExpression:
		testFieldExpression(actual.(*ast.FieldExpression), ex, t)
	case *ast.PrefixExpression:
		testPrefixExpression(actual.(*ast.PrefixExpression), ex, t)
	case *ast.ForExpression:
		testForExpression(actual.(*ast.ForExpression), ex, t)
	case *ast.CaptureExpression:
		testCaptureExpression(actual.(*ast.CaptureExpression), ex, t)
	case *ast.HashExpression:
		testHashExpression(actual.(*ast.HashExpression), ex, t)
	default:
		t.Fatalf("unknown expression type: %T", expected)
	}
}

func testLiteral(actual *ast.Literal, expected *ast.Literal, t *testing.T) {
	t.Helper()

	if actual.Text != expected.Text {
		t.Fatalf("wrong literal, expected=%s, got=%s", expected.Text, actual.Text)
	}
}

func testIntLiteral(actual *ast.IntLiteral, expected *ast.IntLiteral, t *testing.T) {
	t.Helper()

	if actual.Value != expected.Value {
		t.Fatalf("wrong integer literal, expected=%d, got=%d", expected.Value, actual.Value)
	}
}

func testStringLiteral(actual *ast.StringLiteral, expected *ast.StringLiteral, t *testing.T) {
	t.Helper()

	if actual.Value != expected.Value {
		t.Fatalf("wrong string literal, expected=%s, got=%s", expected.Value, actual.Value)
	}
}

func testBoolLiteral(actual *ast.BoolLiteral, expected *ast.BoolLiteral, t *testing.T) {
	t.Helper()

	if actual.Value != expected.Value {
		t.Fatalf("wrong bool literal, expected=%t, got=%t", expected.Value, actual.Value)
	}
}

func testIfExpression(actual *ast.IfExpression, expected *ast.IfExpression, t *testing.T) {
	t.Helper()

	if len(actual.Conditionals) != len(expected.Conditionals) {
		t.Fatalf("wrong number of conditionals in if-expression, expected=%d, got=%d",
			len(expected.Conditionals), len(actual.Conditionals))
	}

	for i := range expected.Conditionals {
		t.Run(fmt.Sprintf("conditional %d", i), func(t *testing.T) {
			testConditionalBlock(&actual.Conditionals[i], &expected.Conditionals[i], t)
		})
	}
}

func testConditionalBlock(actual *ast.ConditionalBlock, expected *ast.ConditionalBlock, t *testing.T) {
	t.Helper()

	if actual.Condition != nil && expected.Condition != nil {
		testExpression(actual.Condition, expected.Condition, t)
	} else if (actual.Condition == nil && expected.Condition != nil) || (actual.Condition != nil && expected.Condition == nil) {
		t.Fatal("cannot compare condition of conditional-block, one of them is nil")
	}

	testBlock(&actual.Block, &expected.Block, t)
}

func testBlock(actual *ast.Block, expected *ast.Block, t *testing.T) {
	t.Helper()

	if len(actual.Statements) != len(expected.Statements) {
		t.Fatalf("wrong number of statements in block, expected=%d, got=%d",
			len(expected.Statements), len(actual.Statements))
	}

	for i := range expected.Statements {
		t.Run(fmt.Sprintf("statement %d", i), func(t *testing.T) {
			testStatement(actual.Statements[i], expected.Statements[i], t)
		})
	}
}

func testInfixExpression(actual *ast.InfixExpression, expected *ast.InfixExpression, t *testing.T) {
	t.Helper()

	if actual.Operator != expected.Operator {
		t.Fatalf("wrong operator in infix-expression, expected=%s, got=%s",
			expected.Operator, actual.Operator)
	}

	testExpression(actual.Left, expected.Left, t)
	testExpression(actual.Right, expected.Right, t)
}

func testIdentifier(actual *ast.Ident, expected *ast.Ident, t *testing.T) {
	t.Helper()

	if actual.Name != expected.Name {
		t.Fatalf("wrong identifier name, expected=%s, got=%s", expected.Name, actual.Name)
	}
}

func testCallExpression(actual *ast.CallExpression, expected *ast.CallExpression, t *testing.T) {
	t.Helper()

	if len(actual.Params) != len(expected.Params) {
		t.Fatalf("wrong number of arguments in call-expression, expected=%d, got=%d",
			len(expected.Params), len(actual.Params))
	}

	testExpression(actual.Callee, expected.Callee, t)

	for i := range expected.Params {
		t.Run(fmt.Sprintf("param %d", i), func(t *testing.T) {
			testExpression(actual.Params[i], expected.Params[i], t)
		})
	}
}

func testFieldExpression(actual *ast.FieldExpression, expected *ast.FieldExpression, t *testing.T) {
	t.Helper()

	testExpression(actual.Callee, expected.Callee, t)
	testExpression(actual.Index, expected.Index, t)
}

func testPrefixExpression(actual *ast.PrefixExpression, expected *ast.PrefixExpression, t *testing.T) {
	t.Helper()

	if actual.Operator != expected.Operator {
		t.Fatalf("wrong prefix operator, expected=%s, got=%s", expected.Operator, actual.Operator)
	}

	testExpression(actual.Expression, expected.Expression, t)
}

func testForExpression(actual *ast.ForExpression, expected *ast.ForExpression, t *testing.T) {
	t.Helper()

	testIdentifier(&actual.Ident, &expected.Ident, t)
	if actual.StatusIdent != nil || expected.StatusIdent != nil {
		testIdentifier(actual.StatusIdent, expected.StatusIdent, t)
	}
	testExpression(actual.RangeExpr, expected.RangeExpr, t)
	testBlock(&actual.Block, &expected.Block, t)
}

func testCaptureExpression(actual *ast.CaptureExpression, expected *ast.CaptureExpression, t *testing.T) {
	t.Helper()

	testBlock(&actual.Block, &expected.Block, t)
}

func testHashExpression(actual *ast.HashExpression, expected *ast.HashExpression, t *testing.T) {
	t.Helper()

	if len(actual.Values) != len(expected.Values) {
		t.Fatalf("wrong number of elements in hash expression, expected=%d, got=%d",
			len(expected.Values), len(actual.Values))
	}

	for k := range expected.Values {
		testExpression(actual.Values[k], expected.Values[k], t)
	}
}

func testParser(input string, expected *ast.Program, t *testing.T, lexerOpts ...lexer.Opt) {
	t.Helper()

	l := newLexerString(input, t, lexerOpts...)
	prog := parse(l, t)

	if len(prog.Statements) != len(expected.Statements) {
		t.Fatalf("program does not have expected number of statements, expected=%d, got=%d",
			len(expected.Statements), len(prog.Statements))
	}

	for i := 0; i < len(expected.Statements); i++ {
		t.Run(fmt.Sprintf("statement %d", i), func(t *testing.T) {
			testStatement(prog.Statements[i], expected.Statements[i], t)
		})
	}
}

func parse(l *lexer.Lexer, t *testing.T) *ast.Program {
	tCh, doneCh := l.Tokens()

	p := New(tCh, doneCh)

	prog, err := p.Parse()
	if err != nil {
		t.Fatalf("error parsing program: %v", err)
	}

	return prog
}

func newLexerString(s string, tb testing.TB, opts ...lexer.Opt) *lexer.Lexer {
	tb.Helper()

	r := bytes.NewReader([]byte(s))
	return lexer.New(r, opts...)
}

func newIdent(n string) *ast.Ident {
	return &ast.Ident{
		Name: n,
	}
}

func newStringLiteral(s string) *ast.StringLiteral {
	return &ast.StringLiteral{
		Value: s,
	}
}

func newBoolLiteral(b bool) *ast.BoolLiteral {
	return &ast.BoolLiteral{
		Value: b,
	}
}

func newIntLiteral(i int64) *ast.IntLiteral {
	return &ast.IntLiteral{
		Value: i,
	}
}

func newLiteral(t string) *ast.Literal {
	return &ast.Literal{
		Text: t,
	}
}

func newNilLiteral() *ast.NilLiteral {
	return &ast.NilLiteral{}
}
