package evaluator

import (
	"reflect"

	"github.com/blizzy78/copper/ast"
	"github.com/blizzy78/copper/ranger"
	"github.com/blizzy78/copper/scope"
)

func (ev *Evaluator) evalExpression(e ast.Expression) (o interface{}, err error) {
	switch ex := e.(type) {
	case *ast.NilLiteral:
		o = evalNilLiteral()
	case *ast.Literal:
		o, err = ev.evalLiteral(*ex)
	case *ast.IntLiteral:
		o = evalIntLiteral(*ex)
	case *ast.BoolLiteral:
		o = evalBoolLiteral(*ex)
	case *ast.StringLiteral:
		o = evalStringLiteral(*ex)
	case *ast.Ident:
		o, err = ev.evalIdentExpression(*ex)
	case *ast.PrefixExpression:
		o, err = ev.evalPrefixExpression(*ex)
	case *ast.InfixExpression:
		o, err = ev.evalInfixExpression(*ex)
	case *ast.IfExpression:
		o, err = ev.evalIfExpression(*ex)
	case *ast.FieldExpression:
		o, err = ev.evalFieldExpression(*ex)
	case *ast.CallExpression:
		o, err = ev.evalCallExpression(*ex)
	case *ast.CaptureExpression:
		o, err = ev.evalCaptureExpression(*ex)
	case *ast.ForExpression:
		o, err = ev.evalForExpression(*ex)
	case *ast.HashExpression:
		o, err = ev.evalHashExpression(*ex)
	default:
		panic(newEvalErrorf(e.Line(), e.Col(), "unknown expression type: %T", e))
	}

	return
}

func (ev *Evaluator) evalLiteral(l ast.Literal) (o interface{}, err error) {
	return ev.literalStringer.String(l.Text)
}

func (ev *Evaluator) evalIdentExpression(i ast.Ident) (o interface{}, err error) {
	name := i.Name

	var ok bool
	if o, ok = ev.scope.Value(name); !ok {
		err = newEvalErrorf(i.StartLine, i.StartCol, "identifier not found in scope: %s", name)
	}
	return
}

func (ev *Evaluator) evalIfExpression(i ast.IfExpression) (o interface{}, err error) {
	for _, c := range i.Conditionals {
		cond := true

		if c.Condition != nil {
			var v interface{}
			if v, err = ev.eval(c.Condition); err != nil {
				return
			}

			if cond, err = toBool(v); err != nil {
				err = newEvalErrorf(c.Condition.Line(), c.Condition.Col(), "condition expression type in if expression is not bool: %s", v)
				return
			}
		}

		if cond {
			var os []interface{}
			if os, err = ev.evalBlockCaptureAll(c.Block); err != nil {
				break
			}

			o = toSingleOrSliceObject(os)

			break
		}
	}

	return
}

func (ev *Evaluator) evalForExpression(f ast.ForExpression) (o interface{}, err error) {
	name := f.Ident.Name
	if ev.scope.HasValue(name) {
		err = newEvalErrorf(f.Ident.StartLine, f.Ident.StartCol, "identifier in for statement already in use: %s", name)
		return
	}

	var statusName *string
	if f.StatusIdent != nil {
		statusName = &f.StatusIdent.Name
	}

	if statusName != nil && ev.scope.HasValue(*statusName) {
		err = newEvalErrorf(f.Ident.StartLine, f.Ident.StartCol, "status identifier in for statement already in use: %s", *statusName)
		return
	}

	var r interface{}
	if r, err = ev.eval(f.RangeExpr); err != nil {
		return
	}

	var rg ranger.Ranger
	var ok bool
	if rg, ok = r.(ranger.Ranger); !ok {
		err = newEvalErrorf(f.RangeExpr.Line(), f.RangeExpr.Col(), "range expression in for statement did not produce a ranger.Ranger: %T", r)
		return
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

		var loopOs []interface{}
		if loopOs, err = ev.evalBlockCaptureAll(f.Block); err != nil {
			break
		}

		os = append(os, loopOs...)

		if ev.breakRequested {
			ev.breakRequested = false
			break
		}

		ev.continueRequested = false
	}

	if err != nil {
		return
	}

	o = toSingleOrSliceObject(os)

	return
}

func (ev *Evaluator) evalCallExpression(c ast.CallExpression) (o interface{}, err error) {
	var f interface{}
	if f, err = ev.eval(c.Callee); err != nil {
		return
	}

	fValue := reflect.ValueOf(f)
	if fValue.Kind() != reflect.Func {
		err = newEvalErrorf(c.Callee.Line(), c.Callee.Col(), "callee expression in call expression is not a function: %T", f)
		return
	}

	fValueType := fValue.Type()
	numExpectedParams := fValueType.NumIn()

	if len(c.Params) > numExpectedParams {
		err = newEvalErrorf(c.StartLine, c.StartCol, "too many arguments for function call")
		return
	}

	params := []reflect.Value{}

	for i, e := range c.Params {
		var po interface{}
		if po, err = ev.eval(e); err != nil {
			break
		}

		pType := fValueType.In(i)
		if po != nil {
			pValue := reflect.ValueOf(po)
			if !pValue.Type().ConvertibleTo(pType) {
				err = newEvalErrorf(e.Line(), e.Col(), "cannot convert argument of type %T to required type %s", po, pType)
				break
			}

			pValue = pValue.Convert(pType)
			params = append(params, pValue)
		} else {
			nilValue := reflect.New(pType).Elem()
			params = append(params, nilValue)
		}
	}

	if err != nil {
		return
	}

	for i := len(c.Params); i < numExpectedParams; i++ {
		pType := fValueType.In(i)

		var v interface{}
		for _, ra := range ev.argumentResolvers {
			if v, err = ra.Resolve(pType); v != nil || err != nil {
				break
			}
		}

		if err != nil {
			break
		}

		if v != nil {
			vValue := reflect.ValueOf(v)
			pValue := vValue.Convert(pType)
			params = append(params, pValue)
		} else {
			nilValue := reflect.New(pType).Elem()
			params = append(params, nilValue)
		}
	}

	if err != nil {
		return
	}

	if len(params) != numExpectedParams {
		err = newEvalErrorf(c.StartLine, c.StartCol, "not enough arguments for function call")
		return
	}

	rs := fValue.Call(params)

	if len(rs) > 0 {
		lastR := rs[len(rs)-1].Interface()
		if lastR != nil {
			if lastErr, ok := lastR.(error); ok {
				err = lastErr
				return
			}
		}

		o = rs[0].Interface()
	}

	return
}

func (ev *Evaluator) evalCaptureExpression(c ast.CaptureExpression) (o interface{}, err error) {
	var os []interface{}
	if os, err = ev.evalBlockCaptureAll(c.Block); err == nil {
		o = toSingleOrSliceObject(os)
	}
	return
}

func (ev *Evaluator) evalHashExpression(h ast.HashExpression) (o interface{}, err error) {
	values := map[string]interface{}{}

	for key, expr := range h.Values {
		if _, ok := values[key]; ok {
			err = newEvalErrorf(h.StartLine, h.StartCol, "duplicate key in hash expression: %s", key)
			break
		}

		var v interface{}
		if v, err = ev.eval(expr); err != nil {
			break
		}

		values[key] = v
	}

	if err != nil {
		return
	}

	o = values

	return
}

func defaultLiteral(s string) (interface{}, error) {
	return s, nil
}

func defaultResolve(t reflect.Type) (interface{}, error) {
	return nil, nil
}

func toSingleOrSliceObject(objs []interface{}) (s interface{}) {
	if len(objs) > 1 {
		s = objs
	} else if len(objs) == 1 {
		s = objs[0]
	}
	return
}
