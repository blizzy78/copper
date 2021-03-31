package evaluator

import (
	"reflect"

	"github.com/blizzy78/copper/ast"
)

func (ev *Evaluator) evalInfixExpression(i ast.InfixExpression) (interface{}, error) {
	left, err := ev.eval(i.Left)
	if err != nil {
		return nil, err
	}
	leftKind := reflect.ValueOf(left).Kind()

	// short-circuit expressions like "falsy && ..."
	if left != nil && leftKind == reflect.Bool && i.Operator == "&&" {
		l, err := toBool(left)
		if err != nil {
			return nil, err
		}
		if !l {
			return false, nil
		}
	}

	// short-circuit expressions like "truthy || ..."
	if left != nil && leftKind == reflect.Bool && i.Operator == "||" {
		l, err := toBool(left)
		if err != nil {
			return nil, err
		}
		if l {
			return true, nil
		}
	}

	right, err := ev.eval(i.Right)
	if err != nil {
		return nil, err
	}
	rightKind := reflect.ValueOf(right).Kind()

	switch {
	case left != nil && right != nil && leftKind == reflect.String && rightKind == reflect.String:
		l, err := toString(left)
		if err != nil {
			return nil, err
		}

		r, err := toString(right)
		if err != nil {
			return nil, err
		}

		return evalStringInfixExpression(l, r, i.Operator, i.StartLine, i.StartCol)

	case left != nil && right != nil && leftKind == reflect.Int64 && rightKind == reflect.Int64:
		l, err := toInt64(left)
		if err != nil {
			return nil, err
		}

		r, err := toInt64(right)
		if err != nil {
			return nil, err
		}

		return evalIntInfixExpression(l, r, i.Operator, i.StartLine, i.StartCol)

	case left != nil && right != nil && leftKind == reflect.Bool && rightKind == reflect.Bool:
		l, err := toBool(left)
		if err != nil {
			return nil, err
		}

		r, err := toBool(right)
		if err != nil {
			return nil, err
		}

		return evalBoolInfixExpression(l, r, i.Operator, i.StartLine, i.StartCol)

	default:
		return nil, newEvalErrorf(i.StartLine, i.StartCol, "cannot handle expression types in '%s' infix expression: %T vs %T", i.Operator, left, right)
	}
}

func evalBoolInfixExpression(l bool, r bool, op string, line int, col int) (interface{}, error) {
	switch op {
	case "==":
		return l == r, nil
	case "!=":
		return l != r, nil
	case "||":
		return l || r, nil
	case "&&":
		return l && r, nil
	default:
		return nil, newEvalErrorf(line, col, "unexpected operator in bool infix expression: %s", op)
	}
}

func evalIntInfixExpression(l int64, r int64, op string, line int, col int) (interface{}, error) { //nolint:gocyclo
	switch op {
	case "==":
		return l == r, nil
	case "!=":
		return l != r, nil
	case "<":
		return l < r, nil
	case "<=":
		return l <= r, nil
	case ">":
		return l > r, nil
	case ">=":
		return l >= r, nil
	case "+":
		return l + r, nil
	case "-":
		return l - r, nil
	case "*":
		return l * r, nil
	case "/":
		if r == 0 {
			return nil, newEvalErrorf(line, col, "division by zero")
		}
		return l / r, nil
	case "%":
		if r == 0 {
			return nil, newEvalErrorf(line, col, "division by zero")
		}
		return l % r, nil
	default:
		return nil, newEvalErrorf(line, col, "unexpected operator in int infix expression: %s", op)
	}
}

func evalStringInfixExpression(l string, r string, op string, line int, col int) (interface{}, error) {
	switch op {
	case "==":
		return l == r, nil
	case "!=":
		return l != r, nil
	case "+":
		if l == "" {
			return r, nil
		}
		if r == "" {
			return l, nil
		}
		return l + r, nil
	default:
		return nil, newEvalErrorf(line, col, "unexpected operator in string infix expression: %s", op)
	}
}
