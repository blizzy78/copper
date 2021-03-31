package evaluator

import (
	"reflect"

	"github.com/blizzy78/copper/ast"
)

func (ev *Evaluator) evalFieldExpression(f ast.FieldExpression) (interface{}, error) {
	index, err := ev.eval(f.Index)
	if err != nil {
		return nil, err
	}

	name, err := toString(index)
	if err != nil {
		return nil, newEvalErrorf(f.Index.Line(), f.Index.Col(), "type of index expression in field expression is not string: %T", index)
	}

	callee, err := ev.eval(f.Callee)
	if err != nil {
		return nil, err
	}

	calleeValue := reflect.ValueOf(callee)
	if callee == nil || (calleeValue.Kind() == reflect.Ptr && calleeValue.IsNil()) {
		return nil, newEvalErrorf(f.StartLine, f.StartCol, "cannot get field or function '%s' from nil object", name)
	}

	switch calleeValue.Kind() {
	case reflect.Map:
		hash, err := toMap(callee)
		if err != nil {
			return nil, err
		}

		return evalFieldExpressionHash(hash, name, f.StartLine, f.StartCol)

	default:
		return evalFieldExpressionNative(callee, name, f.StartLine, f.StartCol)
	}
}

func evalFieldExpressionNative(i interface{}, name string, line int, col int) (interface{}, error) {
	iValue := reflect.ValueOf(i)
	switch iValue.Kind() {
	case reflect.Ptr:
		return evalFieldExpressionNativePtr(i, iValue, name, line, col)
	default:
		return evalFieldExpressionNativeDirect(i, iValue, name, line, col)
	}
}

func evalFieldExpressionNativeDirect(s interface{}, sValue reflect.Value, name string, line int, col int) (interface{}, error) {
	o := tryEvalFieldExpressionNativeDirectField(sValue, name)
	if o == nil {
		o = tryEvalFieldExpressionNativeDirectFunc(sValue, name)
	}
	if o == nil {
		return nil, newEvalErrorf(line, col, "field or function not found in object of type %T: %s", s, name)
	}
	return o, nil
}

func evalFieldExpressionNativePtr(s interface{}, sValue reflect.Value, name string, line int, col int) (interface{}, error) {
	o := tryEvalFieldExpressionNativePtrField(sValue, name)
	if o == nil {
		o = tryEvalFieldExpressionNativePtrFunc(sValue, name)
	}
	if o == nil {
		return nil, newEvalErrorf(line, col, "field or function not found in object of type %T: %s", s, name)
	}
	return o, nil
}

func tryEvalFieldExpressionNativeDirectField(sValue reflect.Value, name string) interface{} {
	if sValue.Kind() != reflect.Struct {
		return nil
	}
	if _, ok := sValue.Type().FieldByName(name); !ok {
		return nil
	}
	return sValue.FieldByName(name).Interface()
}

func tryEvalFieldExpressionNativePtrField(sValue reflect.Value, name string) interface{} {
	sValue = sValue.Elem()
	if sValue.Kind() != reflect.Struct {
		return nil
	}
	if _, ok := sValue.Type().FieldByName(name); !ok {
		return nil
	}
	return sValue.FieldByName(name).Interface()
}

func tryEvalFieldExpressionNativeDirectFunc(sValue reflect.Value, name string) interface{} {
	if _, ok := sValue.Type().MethodByName(name); !ok {
		return nil
	}
	return sValue.MethodByName(name).Interface()
}

func tryEvalFieldExpressionNativePtrFunc(sValue reflect.Value, name string) interface{} {
	if _, ok := sValue.Type().MethodByName(name); !ok {
		return nil
	}
	return sValue.MethodByName(name).Interface()
}

func evalFieldExpressionHash(hash map[string]interface{}, name string, line int, col int) (interface{}, error) {
	o, ok := hash[name]
	if !ok {
		return nil, newEvalErrorf(line, col, "key not found in map: %s", name)
	}
	return o, nil
}
