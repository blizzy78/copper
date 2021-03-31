package helpers

import (
	"errors"
	"fmt"
	"html"
	"reflect"
	"strconv"
	"strings"

	"github.com/blizzy78/copper/scope"
	"github.com/blizzy78/copper/template"
)

var errUnsupportedTypeOrNil = errors.New("unsupported type or nil")

// Safe converts v to a string and returns it as a safe string.
func Safe(v interface{}) template.SafeString {
	return template.SafeString(toString(v))
}

// HTML converts v to a string, escapes any special characters for HTML-safe output, and returns
// it as a safe string.
func HTML(v interface{}) template.SafeString {
	return template.SafeString(html.EscapeString(toString(v)))
}

// Len returns the length of v. If v is a string, slice, or array, it returns len(v).
// Len panics if v is neither of those types, or if v is nil.
func Len(v interface{}) int {
	if v == nil {
		panic(errUnsupportedTypeOrNil)
	}

	value := reflect.ValueOf(v)
	switch value.Kind() {
	case reflect.String, reflect.Slice, reflect.Array:
		return value.Len()
	default:
		panic(errUnsupportedTypeOrNil)
	}
}

// Has returns whether the scope s stores a value identified by name.
// The s argument is usually filled using an evaluator.ResolveArgumentFunc.
func Has(name string, s *scope.Scope) bool {
	return s.HasValue(name)
}

// HasPrefix is equivalent to calling strings.HasPrefix(s, w).
func HasPrefix(s string, w string) bool {
	return strings.HasPrefix(s, w)
}

// HasSuffix is equivalent to calling strings.HasSuffix(s, w).
func HasSuffix(s string, w string) bool {
	return strings.HasSuffix(s, w)
}

func toString(v interface{}) string { //nolint:gocyclo
	if s, ok := v.(fmt.Stringer); ok {
		return s.String()
	}

	switch value := v.(type) {
	case nil:
		return ""
	case string:
		return value
	case bool:
		if !value {
			return "false"
		}
		return "true"
	case int:
		return strconv.FormatInt(int64(value), 10)
	case int8:
		return strconv.FormatInt(int64(value), 10)
	case int16:
		return strconv.FormatInt(int64(value), 10)
	case int32:
		return strconv.FormatInt(int64(value), 10)
	case int64:
		return strconv.FormatInt(value, 10)
	case uint:
		return strconv.FormatUint(uint64(value), 10)
	case uint8:
		return strconv.FormatUint(uint64(value), 10)
	case uint16:
		return strconv.FormatUint(uint64(value), 10)
	case uint32:
		return strconv.FormatUint(uint64(value), 10)
	case uint64:
		return strconv.FormatUint(value, 10)
	case []interface{}:
		buf := strings.Builder{}
		for _, el := range value {
			_, _ = buf.WriteString(toString(el))
		}
		return buf.String()
	case []string:
		buf := strings.Builder{}
		for _, es := range value {
			buf.WriteString(es)
		}
		return buf.String()
	default:
		return fmt.Sprintf("[?TYPE? %T]", v)
	}
}
