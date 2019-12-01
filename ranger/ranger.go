package ranger

import (
	"errors"
	"fmt"
	"reflect"
)

// Ranger iterates over a set of values and returns the current value for each iteration.
type Ranger interface {
	// Next advances the ranger to the next value in the set. It returns whether it was successful in advancing.
	Next() (ok bool)

	// Value returns the current value in the set. Value panics if Next has not been called or if
	// there are no more values.
	Value() interface{}

	// Status returns the current iteration status. Status panics if Next has not been called or if
	// there are no more values.
	Status() Status
}

type Status struct {
	Index   int
	First   bool
	Last    bool
	Even    bool
	Odd     bool
	HasMore bool
}

type HashEntry struct {
	Key   string
	Value interface{}
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

type hashRanger struct {
	h     map[string]interface{}
	keys  []string
	index int
}

// New returns a ranger that iterates over a slice, an array, or a hash. New panics if v is nil, or if it is of another type.
func New(v interface{}) Ranger {
	if h, ok := v.(map[string]interface{}); ok {
		return &hashRanger{
			h:     h,
			keys:  keys(h),
			index: -1,
		}
	}

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

func (i *intRanger) Status() Status {
	index := i.current - i.minInclusive
	lastIndex := i.maxExclusive - i.minInclusive - 1
	even := index%2 == 0
	return Status{
		Index:   index,
		First:   index == 0,
		Last:    index == lastIndex,
		Even:    even,
		Odd:     !even,
		HasMore: index < lastIndex,
	}
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

func (s *sliceRanger) Status() Status {
	even := s.index%2 == 0
	lastIndex := len(s.s) - 1
	return Status{
		Index:   s.index,
		First:   s.index == 0,
		Last:    s.index == lastIndex,
		Even:    even,
		Odd:     !even,
		HasMore: s.index < lastIndex,
	}
}

func (h *hashRanger) Next() (ok bool) {
	i := h.index + 1
	if ok = i < len(h.keys); ok {
		h.index = i
	}
	return
}

func (h *hashRanger) Value() (v interface{}) {
	k := h.keys[h.index]
	return HashEntry{
		Key:   k,
		Value: h.h[k],
	}
}

func (h *hashRanger) Status() Status {
	lastIndex := len(h.keys) - 1
	even := h.index%2 == 0
	return Status{
		Index:   h.index,
		First:   h.index == 0,
		Last:    h.index == lastIndex,
		Even:    even,
		Odd:     !even,
		HasMore: h.index < lastIndex,
	}
}

func keys(h map[string]interface{}) []string {
	keys := make([]string, len(h))
	index := 0
	for k := range h {
		keys[index] = k
		index++
	}
	return keys
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
