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
	"github.com/gobuffalo/nulls"
)

// Safe converts v to a string and returns it as a safe string.
func Safe(v interface{}) template.SafeString {
	return template.SafeString(toString(v))
}

// HTML converts v to a string, escapes any special characters for HTML-safe output, and returns
// it as a safe string.
func HTML(v interface{}) template.SafeString {
	return template.SafeString(html.EscapeString(toString(v)))
}

// Len returns the length of v. If v is a string, slice, or array, it returns v's length.
// Len panics if v is neither of those types, or if v is nil.
func Len(v interface{}) int {
	if v == nil {
		panic(errors.New("cannot get length of nil"))
	}

	value := reflect.ValueOf(v)
	switch value.Kind() {
	case reflect.Array:
		fallthrough
	case reflect.String:
		fallthrough
	case reflect.Slice:
		return value.Len()
	default:
		panic(fmt.Errorf("cannot get length of unsupported type: %T", v))
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

func toString(v interface{}) (s string) {
	if str, ok := v.(fmt.Stringer); ok {
		return str.String()
	}

	var ok bool

	switch value := v.(type) {
	case nil:
		ok = true
	case string:
		s = value
		ok = true
	case bool:
		if value {
			s = "true"
		} else {
			s = "false"
		}
		ok = true
	case int:
		s = strconv.FormatInt(int64(value), 10)
		ok = true
	case int8:
		s = strconv.FormatInt(int64(value), 10)
		ok = true
	case int16:
		s = strconv.FormatInt(int64(value), 10)
		ok = true
	case int32:
		s = strconv.FormatInt(int64(value), 10)
		ok = true
	case int64:
		s = strconv.FormatInt(value, 10)
		ok = true
	case uint:
		s = strconv.FormatUint(uint64(value), 10)
		ok = true
	case uint8:
		s = strconv.FormatUint(uint64(value), 10)
		ok = true
	case uint16:
		s = strconv.FormatUint(uint64(value), 10)
		ok = true
	case uint32:
		s = strconv.FormatUint(uint64(value), 10)
		ok = true
	case uint64:
		s = strconv.FormatUint(value, 10)
		ok = true
	case []interface{}:
		buf := strings.Builder{}
		for _, el := range value {
			es := toString(el)
			buf.WriteString(es)
		}
		s = buf.String()
		ok = true
	case []string:
		buf := strings.Builder{}
		for _, es := range value {
			buf.WriteString(es)
		}
		s = buf.String()
		ok = true
	case nulls.String:
		if value.Valid {
			s = value.String
		}
		ok = true
	case nulls.Int64:
		if value.Valid {
			s = strconv.FormatInt(value.Int64, 10)
		}
		ok = true
	}

	if !ok {
		s = fmt.Sprintf("[?TYPE? %T]", v)
	}

	return
}
