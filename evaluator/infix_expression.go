package evaluator

import (
	"reflect"

	"github.com/blizzy78/copper/ast"
)

func (ev *Evaluator) evalInfixExpression(i ast.InfixExpression) (o interface{}, err error) {
	var left interface{}
	if left, err = ev.eval(i.Left); err != nil {
		return
	}
	leftKind := reflect.ValueOf(left).Kind()

	// short-circuit expressions like "falsy && truthy"
	if left != nil && leftKind == reflect.Bool && i.Operator == "&&" {
		var l bool
		l, err = toBool(left)
		if err != nil {
			return
		}
		if !l {
			o = false
			return
		}
	}

	var right interface{}
	if right, err = ev.eval(i.Right); err != nil {
		return
	}
	rightKind := reflect.ValueOf(right).Kind()

	switch {
	case left != nil && right != nil && leftKind == reflect.String && rightKind == reflect.String:
		var l string
		if l, err = toString(left); err != nil {
			return
		}

		var r string
		if r, err = toString(right); err != nil {
			return
		}

		o, err = evalStringInfixExpression(l, r, i.Operator, i.StartLine, i.StartCol)

	case left != nil && right != nil && leftKind == reflect.Int64 && rightKind == reflect.Int64:
		var l int64
		if l, err = toInt64(left); err != nil {
			return
		}

		var r int64
		if r, err = toInt64(right); err != nil {
			return
		}

		o, err = evalIntInfixExpression(l, r, i.Operator, i.StartLine, i.StartCol)

	case left != nil && right != nil && leftKind == reflect.Bool && rightKind == reflect.Bool:
		var l bool
		if l, err = toBool(left); err != nil {
			return
		}

		var r bool
		if r, err = toBool(right); err != nil {
			return
		}

		o, err = evalBoolInfixExpression(l, r, i.Operator, i.StartLine, i.StartCol)

	default:
		err = newEvalErrorf(i.StartLine, i.StartCol, "cannot handle expression types in '%s' infix expression: %T vs %T", i.Operator, left, right)
	}

	return
}

func evalBoolInfixExpression(l bool, r bool, op string, line int, col int) (o interface{}, err error) {
	switch op {
	case "==":
		o = l == r
	case "!=":
		o = l != r
	case "||":
		o = l || r
	case "&&":
		o = l && r
	default:
		err = newEvalErrorf(line, col, "unexpected operator in bool infix expression: %s", op)
	}
	return
}

func evalIntInfixExpression(l int64, r int64, op string, line int, col int) (o interface{}, err error) {
	switch op {
	case "==":
		o = l == r
	case "!=":
		o = l != r
	case "<":
		o = l < r
	case "<=":
		o = l <= r
	case ">":
		o = l > r
	case ">=":
		o = l >= r
	case "+":
		o = l + r
	case "-":
		o = l - r
	case "*":
		o = l * r
	case "/":
		if r != 0 {
			o = l / r
		} else {
			err = newEvalErrorf(line, col, "division by zero")
		}
	case "%":
		if r != 0 {
			o = l % r
		} else {
			err = newEvalErrorf(line, col, "division by zero")
		}
	default:
		err = newEvalErrorf(line, col, "unexpected operator in int infix expression: %s", op)
	}
	return
}

func evalStringInfixExpression(l string, r string, op string, line int, col int) (o interface{}, err error) {
	switch op {
	case "==":
		o = l == r
	case "!=":
		o = l != r
	case "+":
		if l == "" {
			o = r
		} else if r == "" {
			o = l
		} else {
			o = l + r
		}
	default:
		err = newEvalErrorf(line, col, "unexpected operator in string infix expression: %s", op)
	}
	return
}
