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
	}

	is.True(!r.Next()) // no more values
}

func TestNewFromTo(t *testing.T) {
	is := is.New(t)

	r := NewFromTo(1, 5)

	for i := 1; i <= 5; i++ {
		is.True(r.Next()) // have value
		is.Equal(r.Value().(int), i)
	}

	is.True(!r.Next()) // no more values
}

func TestNew(t *testing.T) {
	is := is.New(t)

	r := New([]int{1, 2, 3, 4, 5})

	for i := 1; i <= 5; i++ {
		is.True(r.Next()) // have value
		is.Equal(r.Value().(int), i)
	}

	is.True(!r.Next()) // no more values
}
