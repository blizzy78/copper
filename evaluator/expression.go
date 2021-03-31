package evaluator

import (
	"reflect"

	"github.com/blizzy78/copper/ast"
	"github.com/blizzy78/copper/ranger"
	"github.com/blizzy78/copper/scope"
)

func (ev *Evaluator) evalExpression(e ast.Expression) (interface{}, error) { //nolint:gocyclo
	switch ex := e.(type) {
	case *ast.NilLiteral:
		return evalNilLiteral(), nil
	case *ast.Literal:
		return ev.evalLiteral(*ex)
	case *ast.IntLiteral:
		return evalIntLiteral(*ex), nil
	case *ast.BoolLiteral:
		return evalBoolLiteral(*ex), nil
	case *ast.StringLiteral:
		return evalStringLiteral(*ex), nil
	case *ast.Ident:
		return ev.evalIdentExpression(*ex)
	case *ast.PrefixExpression:
		return ev.evalPrefixExpression(*ex)
	case *ast.InfixExpression:
		return ev.evalInfixExpression(*ex)
	case *ast.IfExpression:
		return ev.evalIfExpression(*ex)
	case *ast.FieldExpression:
		return ev.evalFieldExpression(*ex)
	case *ast.CallExpression:
		return ev.evalCallExpression(*ex)
	case *ast.CaptureExpression:
		return ev.evalCaptureExpression(*ex)
	case *ast.ForExpression:
		return ev.evalForExpression(*ex)
	case *ast.HashExpression:
		return ev.evalHashExpression(*ex)
	default:
		panic(newEvalErrorf(e.Line(), e.Col(), "unknown expression type: %T", e))
	}
}

func (ev *Evaluator) evalLiteral(l ast.Literal) (interface{}, error) {
	return ev.literalStringer.String(l.Text)
}

func (ev *Evaluator) evalIdentExpression(i ast.Ident) (interface{}, error) {
	name := i.Name
	o, ok := ev.scope.Value(name)
	if !ok {
		return nil, newEvalErrorf(i.StartLine, i.StartCol, "identifier not found in scope: %s", name)
	}
	return o, nil
}

func (ev *Evaluator) evalIfExpression(i ast.IfExpression) (interface{}, error) {
	for _, c := range i.Conditionals {
		cond := true

		if c.Condition != nil {
			v, err := ev.eval(c.Condition)
			if err != nil {
				return nil, err
			}

			cond, err = toBool(v)
			if err != nil {
				return nil, newEvalErrorf(c.Condition.Line(), c.Condition.Col(), "condition expression type in if expression is not bool: %s", v)
			}
		}

		if cond {
			os, err := ev.evalBlockCaptureAll(c.Block)
			if err != nil {
				return nil, err
			}

			return toSingleOrSliceObject(os), nil
		}
	}

	return nil, nil
}

func (ev *Evaluator) evalForExpression(f ast.ForExpression) (interface{}, error) {
	name := f.Ident.Name
	if ev.scope.HasValue(name) {
		return nil, newEvalErrorf(f.Ident.StartLine, f.Ident.StartCol, "identifier in for statement already in use: %s", name)
	}

	var statusName *string
	if f.StatusIdent != nil {
		statusName = &f.StatusIdent.Name
	}

	if statusName != nil && ev.scope.HasValue(*statusName) {
		return nil, newEvalErrorf(f.Ident.StartLine, f.Ident.StartCol, "status identifier in for statement already in use: %s", *statusName)
	}

	r, err := ev.eval(f.RangeExpr)
	if err != nil {
		return nil, err
	}

	rg, ok := r.(ranger.Ranger)
	if !ok {
		return nil, newEvalErrorf(f.RangeExpr.Line(), f.RangeExpr.Col(), "range expression in for statement did not produce a ranger.Ranger: %T", r)
	}

	defer func(oldScope *scope.Scope) {
		ev.scope = oldScope
		ev.loopLevel--
	}(ev.scope)

	loopScope := scope.Scope{
		Parent: ev.scope,
	}
	ev.scope = &loopScope

	ev.loopLevel++

	os := []interface{}{}

	for rg.Next() {
		v := rg.Value()

		loopScope.ClearSelf()
		loopScope.Set(name, v)
		if statusName != nil {
			loopScope.Set(*statusName, rg.Status())
		}

		loopOs, err := ev.evalBlockCaptureAll(f.Block)
		if err != nil {
			return nil, err
		}

		os = append(os, loopOs...)

		if ev.breakRequested {
			ev.breakRequested = false
			break
		}

		ev.continueRequested = false
	}

	return toSingleOrSliceObject(os), nil
}

func (ev *Evaluator) evalCallExpression(c ast.CallExpression) (interface{}, error) {
	f, err := ev.eval(c.Callee)
	if err != nil {
		return nil, err
	}

	fValue := reflect.ValueOf(f)
	if fValue.Kind() != reflect.Func {
		return nil, newEvalErrorf(c.Callee.Line(), c.Callee.Col(), "callee expression in call expression is not a function: %T", f)
	}

	fValueType := fValue.Type()
	numExpectedParams := fValueType.NumIn()

	if len(c.Params) > numExpectedParams {
		return nil, newEvalErrorf(c.StartLine, c.StartCol, "too many arguments for function call")
	}

	params := []reflect.Value{}

	for i, e := range c.Params {
		po, err := ev.eval(e)
		if err != nil {
			return nil, err
		}

		pType := fValueType.In(i)
		if po != nil {
			pValue := reflect.ValueOf(po)
			if !pValue.Type().ConvertibleTo(pType) {
				return nil, newEvalErrorf(e.Line(), e.Col(), "cannot convert argument of type %T to required type %s", po, pType)
			}

			pValue = pValue.Convert(pType)
			params = append(params, pValue)
		} else {
			nilValue := reflect.New(pType).Elem()
			params = append(params, nilValue)
		}
	}

	for i := len(c.Params); i < numExpectedParams; i++ {
		pType := fValueType.In(i)
		ok := false
		for _, ra := range ev.argumentResolvers {
			v, err := ra.Resolve(pType)
			if err != nil {
				return nil, err
			}
			if v == nil {
				continue
			}

			vValue := reflect.ValueOf(v)
			pValue := vValue.Convert(pType)
			params = append(params, pValue)
			ok = true
			break
		}

		if !ok {
			return nil, newEvalErrorf(c.StartLine, c.StartCol, "cannot resolve argument #%d for function call: %s", i+1, pType.Name())
		}
	}

	if len(params) != numExpectedParams {
		return nil, newEvalErrorf(c.StartLine, c.StartCol, "not enough arguments for function call")
	}

	rs := fValue.Call(params)
	if len(rs) == 0 {
		return nil, nil
	}

	lastR := rs[len(rs)-1].Interface()
	if lastR != nil {
		if lastErr, ok := lastR.(error); ok {
			return nil, lastErr
		}
	}

	return rs[0].Interface(), nil
}

func (ev *Evaluator) evalCaptureExpression(c ast.CaptureExpression) (interface{}, error) {
	os, err := ev.evalBlockCaptureAll(c.Block)
	if err != nil {
		return nil, err
	}
	return toSingleOrSliceObject(os), nil
}

func (ev *Evaluator) evalHashExpression(h ast.HashExpression) (interface{}, error) {
	values := map[string]interface{}{}

	for key, expr := range h.Values {
		if _, ok := values[key]; ok {
			return nil, newEvalErrorf(h.StartLine, h.StartCol, "duplicate key in hash expression: %s", key)
		}

		v, err := ev.eval(expr)
		if err != nil {
			return nil, err
		}

		values[key] = v
	}

	return values, nil
}

func defaultLiteral(s string) (interface{}, error) {
	return s, nil
}

func defaultResolve(t reflect.Type) (interface{}, error) {
	return nil, nil
}

func toSingleOrSliceObject(os []interface{}) interface{} {
	if len(os) > 1 {
		return os
	}
	if len(os) == 1 {
		return os[0]
	}
	return nil
}
