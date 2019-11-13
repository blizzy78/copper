package evaluator

import (
	"bytes"
	"testing"

	"github.com/blizzy78/copper/ast"
	"github.com/blizzy78/copper/lexer"
	"github.com/blizzy78/copper/parser"
	"github.com/blizzy78/copper/ranger"
	"github.com/blizzy78/copper/scope"
)

type MockObject struct {
	Field        int
	MockFieldPtr *MockObject
	MockField    MockObject2
}

type MockObject2 struct {
	Field int
}

func (m *MockObject) Five() int {
	return 5
}

func (m *MockObject) Double(x int) int {
	return x * 2
}

func (m *MockObject) Sum(x int, y int) int {
	return x + y
}

func (m *MockObject) SumWithMap(mp map[string]interface{}) int {
	x := mp["x"].(int64)
	y := mp["y"].(int64)
	return int(x + y)
}

func TestEvalIntExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"0", 0},
		{"5", 5},
		{"12", 12},
		{"1000", 1000},
		{"-5", -5},
		{"-12", -12},
		{"-1234", -1234},
		{"1 + 2 * 3", 7},
		{"1 + (2 * 3)", 7},
		{"(1 + 2) * 3", 9},
		{"29 % 5", 4},
		{"29 - 5", 24},
		{"29 / 5", 5},
	}

	for i, test := range tests {
		o := evalExpr(i, test.input, true, t)
		testObject(i, o, test.expected, t)
	}
}

func TestEvalBoolExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"false", false},
		{"!true", false},
		{"!false", true},
		{"!!true", true},
		{"!!false", false},
		{"true == true", true},
		{"true == false", false},
		{"!true == true", false},
		{"false == !false", false},
		{"true != true", false},
		{"true != false", true},
		{"true != !false", false},
		{"!!true != !!!false", false},
		{"5 == 5", true},
		{"5 == 6", false},
		{"5 != 5", false},
		{"5 != 6", true},
		{"5 < 6", true},
		{"5 > 6", false},
		{"6 < 5", false},
		{"6 > 5", true},
		{"5 <= 6", true},
		{"5 >= 6", false},
		{"6 <= 5", false},
		{"6 >= 5", true},
		{"5 <= 5", true},
		{"5 >= 5", true},
		{`"x" == "x"`, true},
		{`"x" == "y"`, false},
		{`"x" != "x"`, false},
		{`"x" != "y"`, true},
	}

	for i, test := range tests {
		o := evalExpr(i, test.input, true, t)
		testObject(i, o, test.expected, t)
	}
}

func TestEvalStringExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"x"`, `x`},
		{`"x" + "y"`, `xy`},
		{`"x" + "y" + "z"`, `xyz`},
	}

	for i, test := range tests {
		o := evalExpr(i, test.input, true, t)
		testObject(i, o, test.expected, t)
	}
}

func TestEvalIfExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{`if 2 < 3 5 end`, 5},
		{`if 2 < 3 5 else 7 end`, 5},
		{`if 2 > 3 5 else 7 end`, 7},
		{
			`if 1 > 2
				10
			elseif 1 > 3
				20
			elseif 1 > 4
				30
			elseif 1 > 5
				40
			else
				50
			end`,
			50,
		},
		{
			`if 1 > 2
				10
			elseif 1 > 3
				20
			elseif 1 < 4
				30
			elseif 1 > 5
				40
			else
				50
			end`,
			30,
		},
		{
			`if 1 > 2
				if 3 > 4
					10
				else
					20
				end
			elseif 5 < 6
				if 7 > 8
					30
				else
					40
				end
			end`,
			40,
		},
	}

	for i, test := range tests {
		o := evalExpr(i, test.input, true, t)
		testObject(i, o, test.expected, t)
	}
}

func TestLetStatement(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{`let x = 5`, 5},
		{`let x = true`, true},
		{`let x = "y"`, "y"},
		{`let x = if 1 < 2 10 else 20 end`, 10},
		{`let x = if 1 > 2 10 else 20 end`, 20},
		{`let x = m.Field`, 5},
		{`let x = m.Sum(3, 6)`, 9},
	}

	for i, test := range tests {
		s := scope.Scope{}

		s.Set("m", &MockObject{
			Field: 5,
		})

		o := evalWithScope(i, test.input, true, &s, t)

		if o != nil {
			t.Fatalf("[%d] did not return nil", i)
		}

		testScopeValue(i, &s, "x", test.expected, t)
	}
}

func TestIdentExpression(t *testing.T) {
	tests := []struct {
		input       string
		scopeValues map[string]interface{}
		expected    interface{}
	}{
		{
			`x`,
			map[string]interface{}{
				"x": 5,
			},
			5,
		},
		{
			`x + y`,
			map[string]interface{}{
				"x": 5,
				"y": 6,
			},
			11,
		},
		{
			`x + y`,
			map[string]interface{}{
				"x": "a",
				"y": "b",
			},
			"ab",
		},
		{
			`let x = 5
			let y = 6
			x + y`,
			map[string]interface{}{},
			11,
		},
	}

	for i, test := range tests {
		s := scope.Scope{}

		for k, v := range test.scopeValues {
			s.Set(k, v)
		}

		o := evalWithScope(i, test.input, true, &s, t)
		testObject(i, o, test.expected, t)
	}
}

func TestFieldExpression(t *testing.T) {
	tests := []struct {
		input       string
		scopeValues map[string]interface{}
		expected    interface{}
	}{
		{
			"x.y",
			map[string]interface{}{
				"x": map[string]interface{}{
					"y": 5,
				},
			},
			5,
		},
		{
			`x["y"]`,
			map[string]interface{}{
				"x": map[string]interface{}{
					"y": 5,
				},
			},
			5,
		},
		{
			"x.y.z",
			map[string]interface{}{
				"x": map[string]interface{}{
					"y": map[string]interface{}{
						"z": 5,
					},
				},
			},
			5,
		},
		{
			"x.Field",
			map[string]interface{}{
				"x": &MockObject{
					Field: 5,
				},
			},
			5,
		},
		{
			"x.MockFieldPtr.Field",
			map[string]interface{}{
				"x": &MockObject{
					MockFieldPtr: &MockObject{
						Field: 5,
					},
				},
			},
			5,
		},
		{
			"x.Field",
			map[string]interface{}{
				"x": MockObject{
					Field: 5,
				},
			},
			5,
		},
		{
			"x.MockFieldPtr.Field",
			map[string]interface{}{
				"x": MockObject{
					MockFieldPtr: &MockObject{
						Field: 5,
					},
				},
			},
			5,
		},
		{
			"x.MockField.Field",
			map[string]interface{}{
				"x": MockObject{
					MockField: MockObject2{
						Field: 5,
					},
				},
			},
			5,
		},
	}

	for i, test := range tests {
		s := scope.Scope{}

		for k, v := range test.scopeValues {
			s.Set(k, v)
		}

		o := evalWithScope(i, test.input, true, &s, t)
		testObject(i, o, test.expected, t)
	}
}

func TestCallExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{
			"x.Five()",
			5,
		},
		{
			"x.Double(21)",
			42,
		},
		{
			"x.Sum(6, 7)",
			13,
		},
		{
			`x.SumWithMap({
				"x": 10,
				"y": 20
			})`,
			30,
		},
		{
			"foo(3, 4)",
			12,
		},
		{
			`let x = foo
			x(3, 4)`,
			12,
		},
	}

	for i, test := range tests {
		s := scope.Scope{}

		s.Set("x", &MockObject{})

		s.Set("foo", func(a int, b int) int {
			return a * b
		})

		o := evalWithScope(i, test.input, true, &s, t)
		testObject(i, o, test.expected, t)
	}
}

func TestForStatement(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{
			`let x = 10
			for i in range(1, 6)
				let x = x + 1
			end
			for i in range(11, 16)
				let x = x + 2
			end`,
			25,
		},
		{
			`let x = 10
			for i in range(1, 3)
				for j in range(1, 11)
					let x = x + 1
				end
			end`,
			30,
		},
		{
			`let x = 10
			for i in range(1, 3)
				for j in range(1, 11)
					let x = x + 1
					if j == 5
						break
					end
					let x = x + 1
				end
			end`,
			28,
		},
		{
			`let x = 10
			for i in range(1, 3)
				for j in range(1, 11)
					let x = x + 1
					if j >= 5
						continue
					end
					let x = x + 1
				end
			end`,
			38,
		},
	}

	for i, test := range tests {
		s := scope.Scope{}

		s.Set("range", ranger.NewInt)

		evalWithScope(i, test.input, true, &s, t)
		v, _ := s.Value("x")
		testObject(i, v, test.expected, t)
	}
}

func TestCaptureExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected []interface{}
	}{
		{
			`let x = capture
				"a"
				"b"
				"c"
				1 + 2
				true
			end`,
			[]interface{}{
				"a",
				"b",
				"c",
				3,
				true,
			},
		},
	}

	for i, test := range tests {
		s := scope.Scope{}

		evalWithScope(i, test.input, true, &s, t)
		v, _ := s.Value("x")
		testObject(i, v, test.expected, t)
	}
}

func TestStartInLiteral(t *testing.T) {
	tests := []struct {
		input    string
		expected []interface{}
	}{
		{
			`<% let x = capture
			%>foo <% "bar" %> baz<%
			end
			%>`,
			[]interface{}{
				"foo ",
				"bar",
				" baz",
			},
		},
	}

	for i, test := range tests {
		s := scope.Scope{}

		evalWithScope(i, test.input, false, &s, t)
		v, _ := s.Value("x")
		testObject(i, v, test.expected, t)
	}
}

func testObject(i int, actual interface{}, expected interface{}, t *testing.T) {
	t.Helper()

	switch e := expected.(type) {
	case int:
		testIntObject(i, actual, int64(e), t)
	case int64:
		testIntObject(i, actual, e, t)
	case bool:
		testBoolObject(i, actual, e, t)
	case string:
		testStringObject(i, actual, e, t)
	case []interface{}:
		testSliceObject(i, actual, e, t)
	default:
		t.Fatalf("unexpected type for 'expected': %T", expected)
	}
}

func mustToInt64(v interface{}) (i int64) {
	i, _ = toInt64(v)
	return
}

func mustToBool(v interface{}) (b bool) {
	b, _ = toBool(v)
	return
}

func mustToString(v interface{}) (s string) {
	s, _ = toString(v)
	return
}

func mustToSlice(v interface{}) (s []interface{}) {
	s, _ = toSlice(v)
	return
}

func testIntObject(i int, actual interface{}, expected int64, t *testing.T) {
	t.Helper()

	if mustToInt64(actual) != expected {
		t.Fatalf("[%d] wrong int value, expected=%d, got=%s (%T)", i, expected, actual, actual)
	}
}

func testBoolObject(i int, actual interface{}, expected bool, t *testing.T) {
	t.Helper()

	if mustToBool(actual) != expected {
		t.Fatalf("[%d] wrong bool value, expected=%t, got=%s", i, expected, actual)
	}
}

func testStringObject(i int, actual interface{}, expected string, t *testing.T) {
	t.Helper()

	if mustToString(actual) != expected {
		t.Fatalf("[%d] wrong string value, expected=%s, got=%s", i, expected, actual)
	}
}

func testSliceObject(i int, actual interface{}, expected []interface{}, t *testing.T) {
	t.Helper()

	a := mustToSlice(actual)

	if len(a) != len(expected) {
		t.Fatalf("[%d] wrong number of slice elements, expected=%d, got=%d",
			i, len(expected), len(a))
	}

	for j := range a {
		testObject(i, a[j], expected[j], t)
	}
}

func testScopeValue(i int, s *scope.Scope, name string, expected interface{}, t *testing.T) {
	t.Helper()

	if v, ok := s.Value(name); ok {
		testObject(i, v, expected, t)
	} else {
		t.Fatalf("[%d] name not found in scope: %s", i, name)
	}
}

func evalWithScope(i int, input string, startInCode bool, s *scope.Scope, t *testing.T) (o interface{}) {
	t.Helper()

	prog := parse(i, input, startInCode, t)

	ev := Evaluator{}

	var err error
	if o, err = ev.Eval(prog, s); err != nil {
		t.Fatalf("[%d] error evaluating expression: %v", i, err)
	}

	return
}

func evalExpr(i int, input string, startInCode bool, t *testing.T) (o interface{}) {
	t.Helper()

	prog := parse(i, input, startInCode, t)
	expr := prog.Statements[0].(*ast.ExpressionStatement).Expression

	ev := Evaluator{}

	var err error
	if o, err = ev.Eval(expr, &scope.Scope{}); err != nil {
		t.Fatalf("[%d] error evaluating expression: %v", i, err)
	}

	return
}

func parse(i int, input string, startInCode bool, t *testing.T) (prog *ast.Program) {
	t.Helper()

	l := newLexerString(input, startInCode, t)

	tCh, doneCh := l.Tokens()

	p := parser.New(tCh, doneCh)
	var err error
	if prog, err = p.Parse(); err != nil {
		t.Fatalf("[%d] error parsing program: %v", i, err)
	}

	return

}

func newLexerString(s string, startInCode bool, t *testing.T) (l *lexer.Lexer) {
	t.Helper()

	r := bytes.NewReader([]byte(s))
	return lexer.New(r, startInCode)
}
