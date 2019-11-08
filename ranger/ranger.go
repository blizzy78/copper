package ranger

import (
	"errors"
	"fmt"
	"reflect"
)

// Ranger iterates over a set of values and returns the current value for each iteration.
type Ranger interface {
	// Next advances to the next value in the set. It returns whether it was successful in advancing.
	Next() (ok bool)

	// Value returns the current value in the set. Value panics if Next has not been called or if
	// there are no more values.
	Value() interface{}
}

type intRanger struct {
	minInclusive int
	maxExclusive int
	current      int
}

type sliceRanger struct {
	s     []interface{}
	index int
}

// New returns a Ranger that iterates over a slice or an array. New panics if v is nil, or if it is not a slice or an array.
func New(v interface{}) Ranger {
	if s, err := toSlice(v); err != nil {
		panic(err)
	} else {
		return &sliceRanger{
			s:     s,
			index: -1,
		}
	}
}

// NewInt returns a Ranger that iterates over a range of integer values. NewInt panics if maxExclusive is not
// greater than minInclusive.
func NewInt(minInclusive int, maxExclusive int) Ranger {
	if maxExclusive <= minInclusive {
		panic(errors.New("maxExclusive must be greater than minInclusive"))
	}

	return &intRanger{
		minInclusive: minInclusive,
		maxExclusive: maxExclusive,
		current:      minInclusive - 1,
	}
}

// NewFromTo returns a Ranger that iterates over a range of integer values. NewFromTo panics if maxInclusive is
// less than minInclusive.
func NewFromTo(minInclusive int, maxInclusive int) Ranger {
	return NewInt(minInclusive, maxInclusive+1)
}

func (i *intRanger) Next() (ok bool) {
	c := i.current + 1
	if ok = c < i.maxExclusive; ok {
		i.current = c
	}
	return
}

func (i *intRanger) Value() (v interface{}) {
	return i.current
}

func (s *sliceRanger) Next() (ok bool) {
	i := s.index + 1
	if ok = i < len(s.s); ok {
		s.index = i
	}
	return
}

func (s *sliceRanger) Value() (v interface{}) {
	return s.s[s.index]
}

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
		for i := 0; i < l; i++ {
			el := value.Index(i).Interface()
			s = append(s, el)
		}

	default:
		err = fmt.Errorf("cannot convert unsupported type to slice: %T", v)
	}

	return
}
