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

func TestNew(t *testing.T) {
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
