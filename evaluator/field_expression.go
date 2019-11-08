package evaluator

import (
	"reflect"

	"github.com/blizzy78/copper/ast"
)

func (ev *Evaluator) evalFieldExpression(f ast.FieldExpression) (o interface{}, err error) {
	var index interface{}
	if index, err = ev.eval(f.Index); err != nil {
		return
	}

	var name string
	if name, err = toString(index); err != nil {
		err = newEvalErrorf(f.Index.Line(), f.Index.Col(), "type of index expression in field expression is not string: %T", index)
		return
	}

	var callee interface{}
	if callee, err = ev.eval(f.Callee); err != nil {
		return
	}

	calleeValue := reflect.ValueOf(callee)
	if callee == nil || (calleeValue.Kind() == reflect.Ptr && calleeValue.IsNil()) {
		err = newEvalErrorf(f.StartLine, f.StartCol, "cannot get field or function '%s' from nil object", name)
		return
	}

	switch calleeValue.Kind() {
	case reflect.Map:
		var hash map[string]interface{}
		if hash, err = toMap(callee); err != nil {
			return
		}

		o, err = evalFieldExpressionHash(hash, name, f.StartLine, f.StartCol)

	default:
		o, err = evalFieldExpressionNative(callee, name, f.StartLine, f.StartCol)
	}

	return
}

func evalFieldExpressionNative(i interface{}, name string, line int, col int) (o interface{}, err error) {
	iValue := reflect.ValueOf(i)

	switch iValue.Kind() {
	case reflect.Ptr:
		o, err = evalFieldExpressionNativePtr(i, iValue, name, line, col)
	default:
		o, err = evalFieldExpressionNativeDirect(i, iValue, name, line, col)
	}

	return
}

func evalFieldExpressionNativeDirect(s interface{}, sValue reflect.Value, name string, line int, col int) (o interface{}, err error) {
	if o = tryEvalFieldExpressionNativeDirectField(s, sValue, name); o == nil {
		o = tryEvalFieldExpressionNativeDirectFunc(s, sValue, name)
	}

	if o == nil {
		err = newEvalErrorf(line, col, "field or function not found in object of type %T: %s", s, name)
		return
	}

	return
}

func evalFieldExpressionNativePtr(s interface{}, sValue reflect.Value, name string, line int, col int) (o interface{}, err error) {
	if o = tryEvalFieldExpressionNativePtrField(s, sValue, name); o == nil {
		o = tryEvalFieldExpressionNativePtrFunc(s, sValue, name)
	}

	if o == nil {
		err = newEvalErrorf(line, col, "field or function not found in object of type %T: %s", s, name)
		return
	}

	return
}

func tryEvalFieldExpressionNativeDirectField(s interface{}, sValue reflect.Value, name string) interface{} {
	if sValue.Kind() == reflect.Struct {
		if _, ok := sValue.Type().FieldByName(name); ok {
			return sValue.FieldByName(name).Interface()
		}
	}
	return nil
}

func tryEvalFieldExpressionNativePtrField(s interface{}, sValue reflect.Value, name string) interface{} {
	sValue = sValue.Elem()
	if sValue.Kind() == reflect.Struct {
		if _, ok := sValue.Type().FieldByName(name); ok {
			return sValue.FieldByName(name).Interface()
		}
	}
	return nil
}

func tryEvalFieldExpressionNativeDirectFunc(s interface{}, sValue reflect.Value, name string) interface{} {
	if _, ok := sValue.Type().MethodByName(name); ok {
		return sValue.MethodByName(name).Interface()
	}
	return nil
}

func tryEvalFieldExpressionNativePtrFunc(s interface{}, sValue reflect.Value, name string) interface{} {
	if _, ok := sValue.Type().MethodByName(name); ok {
		return sValue.MethodByName(name).Interface()
	}
	return nil
}

func evalFieldExpressionHash(hash map[string]interface{}, name string, line int, col int) (o interface{}, err error) {
	var ok bool
	if o, ok = hash[name]; !ok {
		err = newEvalErrorf(line, col, "key not found in map: %s", name)
	}
	return
}
