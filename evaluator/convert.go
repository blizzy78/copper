package evaluator

import (
	"errors"
	"fmt"
	"reflect"
)

// toInt64 converts v to an int64. v may be any of int, int8, int16, int32, int64, or a type derived from those.
func toInt64(v interface{}) (i int64, err error) {
	if v == nil {
		err = errors.New("cannot convert nil to int64")
		return
	}

	value := reflect.ValueOf(v)

	switch value.Kind() {
	case reflect.Int:
		fallthrough
	case reflect.Int8:
		fallthrough
	case reflect.Int16:
		fallthrough
	case reflect.Int32:
		fallthrough
	case reflect.Int64:
		i = value.Int()

	default:
		err = fmt.Errorf("cannot convert unsupported type to int64: %T", v)
	}

	return
}

// toString converts v to a string. v may be a string or a type derived from it.
func toString(v interface{}) (s string, err error) {
	if v == nil {
		err = errors.New("cannot convert nil to string")
		return
	}

	value := reflect.ValueOf(v)

	switch value.Kind() {
	case reflect.String:
		s = value.String()

	default:
		err = fmt.Errorf("cannot convert unsupported type to string: %T", v)
	}

	return
}

// toSlice converts the slice or array v to a slice. v may be a slice or array or a type derived from those.
func toSlice(v interface{}) (s []interface{}, err error) {
	if v == nil {
		err = errors.New("cannot convert nil to slice")
		return
	}

	value := reflect.ValueOf(v)

	switch value.Kind() {
	case reflect.Array:
		fallthrough
	case reflect.Slice:
		l := value.Len()
		s = make([]interface{}, l)
		for i := 0; i < l; i++ {
			s[i] = value.Index(i).Interface()
		}

	default:
		err = fmt.Errorf("cannot convert unsupported type to slice: %T", v)
	}

	return
}

// toMap converts the map v to a map indexed by string keys. Keys of v are passed through toString.
func toMap(v interface{}) (m map[string]interface{}, err error) {
	if v == nil {
		err = errors.New("cannot convert nil to map")
		return
	}

	value := reflect.ValueOf(v)

	switch value.Kind() {
	case reflect.Map:
		m = map[string]interface{}{}

		i := value.MapRange()
		for i.Next() {
			kv := i.Key().Interface()

			var k string

			if k, err = toString(kv); err != nil {
				err = fmt.Errorf("cannot convert map key of unsupported type to string: %T", kv)
				break
			}

			m[k] = i.Value().Interface()
		}

	default:
		err = fmt.Errorf("cannot convert unsupported type to map: %T", v)
	}

	return
}

// toBool converts v to a bool. v may be a bool or a type derived from it.
func toBool(v interface{}) (b bool, err error) {
	if v == nil {
		err = errors.New("cannot convert nil to bool")
		return
	}

	value := reflect.ValueOf(v)

	switch value.Kind() {
	case reflect.Bool:
		b = value.Bool()

	default:
		err = fmt.Errorf("cannot convert unsupported type to bool: %T", v)
	}

	return
}
