package evaluator

import (
	"github.com/blizzy78/copper/ast"
)

func (ev *Evaluator) evalPrefixExpression(p ast.PrefixExpression) (interface{}, error) {
	v, err := ev.eval(p.Expression)
	if err != nil {
		return nil, err
	}

	switch p.Operator {
	case "-":
		return evalMinusPrefix(v, p.StartLine, p.StartCol)

	case "!":
		return evalBangPrefix(v, p.StartLine, p.StartCol)

	default:
		panic(newEvalErrorf(p.StartLine, p.StartCol, "unknown prefix expression operator: %s", p.Operator))
	}
}

func evalMinusPrefix(right interface{}, line int, col int) (interface{}, error) {
	r, err := toInt64(right)
	if err != nil {
		return nil, newEvalErrorf(line, col, "incompatible expression type for '-' prefix expression: %T", right)
	}
	return -r, nil
}

func evalBangPrefix(right interface{}, line int, col int) (interface{}, error) {
	r, err := toBool(right)
	if err != nil {
		return nil, newEvalErrorf(line, col, "incompatible expression type for '!' prefix expression: %T", right)
	}
	return !r, nil
}
