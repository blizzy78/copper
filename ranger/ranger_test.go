package ranger

import (
	"testing"

	"github.com/matryer/is"
)

func TestNewInt(t *testing.T) {
	is := is.New(t)

	r := NewInt(1, 6)

	for i := 1; i < 6; i++ {
		is.True(r.Next()) // have value
		is.Equal(r.Value().(int), i)

		s := r.Status()
		is.Equal(s.Index, i-1)
		is.Equal(s.First, i == 1)
		is.Equal(s.Last, i == 5)
		is.Equal(s.Even, s.Index%2 == 0)
		is.Equal(s.Odd, !s.Even)
		is.Equal(s.HasMore, i < 5)
	}

	is.True(!r.Next()) // no more values
}

func TestNewInt_Empty(t *testing.T) {
	is := is.New(t)

	r := NewInt(1, 1)

	is.True(!r.Next()) // no more values
}

func TestNewFromTo(t *testing.T) {
	is := is.New(t)

	r := NewFromTo(1, 5)

	for i := 1; i <= 5; i++ {
		is.True(r.Next()) // have value
		is.Equal(r.Value().(int), i)

		s := r.Status()
		is.Equal(s.Index, i-1)
		is.Equal(s.First, i == 1)
		is.Equal(s.Last, i == 5)
		is.Equal(s.Even, s.Index%2 == 0)
		is.Equal(s.Odd, !s.Even)
		is.Equal(s.HasMore, i < 5)
	}

	is.True(!r.Next()) // no more values
}

func TestNew_Slice(t *testing.T) {
	is := is.New(t)

	r := New([]int{1, 2, 3, 4, 5})

	for i := 1; i <= 5; i++ {
		is.True(r.Next()) // have value
		is.Equal(r.Value().(int), i)

		s := r.Status()
		is.Equal(s.Index, i-1)
		is.Equal(s.First, i == 1)
		is.Equal(s.Last, i == 5)
		is.Equal(s.Even, s.Index%2 == 0)
		is.Equal(s.Odd, !s.Even)
		is.Equal(s.HasMore, i < 5)
	}

	is.True(!r.Next()) // no more values
}

func TestNew_SliceEmpty(t *testing.T) {
	is := is.New(t)

	r := New([]int{})

	is.True(!r.Next()) // no more values
}

func TestNew_Hash(t *testing.T) {
	is := is.New(t)

	r := New(map[string]interface{}{
		"a": 1,
		"b": 2,
		"c": 3,
	})

	for i := 1; i <= 3; i++ {
		is.True(r.Next())

		e := r.Value().(HashEntry)
		is.True(e.Key == "a" || e.Key == "b" || e.Key == "c")
		v := e.Value.(int)
		is.True(v == 1 || v == 2 || v == 3)

		s := r.Status()
		is.Equal(s.Index, i-1)
		is.Equal(s.First, i == 1)
		is.Equal(s.Last, i == 3)
		is.Equal(s.Even, s.Index%2 == 0)
		is.Equal(s.Odd, !s.Even)
		is.Equal(s.HasMore, i < 3)
	}

	is.True(!r.Next()) // no more values
}

func TestNew_HashEmpty(t *testing.T) {
	is := is.New(t)

	r := New(map[string]interface{}{})

	is.True(!r.Next()) // no more values
}
