package ranger

import (
	"errors"
	"fmt"
	"reflect"
)

// Ranger iterates over a set of values and returns the current value for each iteration.
type Ranger interface {
	// Next advances the ranger to the next value in the set. It returns whether it was successful in advancing.
	Next() bool

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
// If v is a hash, the ranger will produce HashEntry elements.
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
// greater than or equal to minInclusive.
func NewInt(minInclusive int, maxExclusive int) Ranger {
	if maxExclusive < minInclusive {
		panic(errors.New("maxExclusive must be greater than or equal to minInclusive"))
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

// Next implements Ranger.
func (i *intRanger) Next() bool {
	c := i.current + 1
	if c >= i.maxExclusive {
		return false
	}
	i.current = c
	return true
}

// Value implements Ranger.
func (i *intRanger) Value() interface{} {
	return i.current
}

// Status implements Ranger.
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

// Next implements Ranger.
func (s *sliceRanger) Next() bool {
	i := s.index + 1
	if i >= len(s.s) {
		return false
	}
	s.index = i
	return true
}

// Value implements Ranger.
func (s *sliceRanger) Value() interface{} {
	return s.s[s.index]
}

// Status implements Ranger.
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

// Next implements Ranger.
func (h *hashRanger) Next() bool {
	i := h.index + 1
	if i >= len(h.keys) {
		return false
	}
	h.index = i
	return true
}

// Value implements Ranger.
func (h *hashRanger) Value() interface{} {
	k := h.keys[h.index]
	return HashEntry{
		Key:   k,
		Value: h.h[k],
	}
}

// Status implements Ranger.
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
