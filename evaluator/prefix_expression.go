package evaluator

import (
	"github.com/blizzy78/copper/ast"
)

func (ev *Evaluator) evalPrefixExpression(p ast.PrefixExpression) (o interface{}, err error) {
	var v interface{}
	if v, err = ev.eval(p.Expression); err != nil {
		return
	}

	switch p.Operator {
	case "-":
		o, err = evalMinusPrefix(v, p.StartLine, p.StartCol)

	case "!":
		o, err = evalBangPrefix(v, p.StartLine, p.StartCol)

	default:
		panic(newEvalErrorf(p.StartLine, p.StartCol, "unknown prefix expression operator: %s", p.Operator))
	}

	return
}

func evalMinusPrefix(right interface{}, line int, col int) (o interface{}, err error) {
	var r int64
	if r, err = toInt64(right); err != nil {
		err = newEvalErrorf(line, col, "incompatible expression type for '-' prefix expression: %T", right)
	} else {
		o = -r
	}
	return
}

func evalBangPrefix(right interface{}, line int, col int) (o interface{}, err error) {
	var r bool
	if r, err = toBool(right); err != nil {
		err = newEvalErrorf(line, col, "incompatible expression type for '!' prefix expression: %T", right)
	} else {
		o = !r
	}
	return
}
