package evaluator

import (
	"errors"
	"fmt"
	"reflect"
)

// toInt64 converts v to an int64. v may be any of int, int8, int16, int32, int64, or a type derived from those.
func toInt64(v interface{}) (int64, error) {
	if v == nil {
		return 0, errors.New("cannot convert nil to int64")
	}

	value := reflect.ValueOf(v)
	switch value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return value.Int(), nil
	default:
		return 0, fmt.Errorf("cannot convert unsupported type to int64: %T", v)
	}
}

// toString converts v to a string. v may be a string or a type derived from it.
func toString(v interface{}) (string, error) {
	if v == nil {
		return "", errors.New("cannot convert nil to string")
	}

	value := reflect.ValueOf(v)
	if value.Kind() != reflect.String {
		return "", fmt.Errorf("cannot convert unsupported type to string: %T", v)
	}

	return value.String(), nil
}

// toSlice converts the slice or array v to a slice. v may be a slice or array or a type derived from those.
func toSlice(v interface{}) ([]interface{}, error) {
	if v == nil {
		return nil, errors.New("cannot convert nil to slice")
	}

	value := reflect.ValueOf(v)

	switch value.Kind() {
	case reflect.Array, reflect.Slice:
		l := value.Len()
		s := make([]interface{}, l)
		for i := 0; i < l; i++ {
			s[i] = value.Index(i).Interface()
		}
		return s, nil

	default:
		return nil, fmt.Errorf("cannot convert unsupported type to slice: %T", v)
	}
}

// toMap converts the map v to a map indexed by string keys. Keys of v are passed through toString.
func toMap(v interface{}) (map[string]interface{}, error) {
	if v == nil {
		return nil, errors.New("cannot convert nil to map")
	}

	value := reflect.ValueOf(v)
	if value.Kind() != reflect.Map {
		return nil, fmt.Errorf("cannot convert unsupported type to map: %T", v)

	}

	m := map[string]interface{}{}
	i := value.MapRange()
	for i.Next() {
		kv := i.Key().Interface()

		k, err := toString(kv)
		if err != nil {
			return nil, fmt.Errorf("cannot convert map key of unsupported type to string: %T", kv)
		}

		m[k] = i.Value().Interface()
	}

	return m, nil
}

// toBool converts v to a bool. v may be a bool or a type derived from it.
func toBool(v interface{}) (bool, error) {
	if v == nil {
		return false, errors.New("cannot convert nil to bool")
	}

	value := reflect.ValueOf(v)
	if value.Kind() != reflect.Bool {
		return false, fmt.Errorf("cannot convert unsupported type to bool: %T", v)
	}

	return value.Bool(), nil
}
