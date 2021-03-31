package scope

import (
	"testing"

	"github.com/matryer/is"
)

func TestScope(t *testing.T) {
	is := is.New(t)

	s := Scope{}
	s.Set("x", 5)

	testIntValue(&s, "x", 5, is)
}

func TestScope_Parent(t *testing.T) {
	is := is.New(t)

	a := Scope{}
	a.Set("x", 3)

	b := Scope{
		Parent: &a,
	}

	c := Scope{
		Parent: &b,
	}
	c.Set("y", 42)

	testIntValue(&a, "x", 3, is) // self
	testIntValue(&b, "x", 3, is) // transitive through empty scope
	testIntValue(&c, "x", 3, is) // double transitive through non-empty scope

	c.Set("x", 33)

	testIntValue(&a, "x", 33, is) // self
	testIntValue(&b, "x", 33, is) // transitive through empty scope
	testIntValue(&c, "x", 33, is) // double transitive through non-empty scope
}

func TestScope_Lock(t *testing.T) {
	is := is.New(t)

	s := Scope{}
	s.Set("x", 5)

	s.Lock()

	s.Set("x", 42)

	testIntValue(&s, "x", 5, is) // no change
}

func TestScope_ClearSelf(t *testing.T) {
	is := is.New(t)

	s := Scope{}
	s.Set("x", 5)

	s.ClearSelf()

	testNoValue(&s, "x", is) // removed
}

func testIntValue(s *Scope, name string, v int, is *is.I) { //nolint:unparam
	is.True(s.HasValue(name))
	actual, ok := s.Value(name)
	is.Equal(actual.(int), v)
	is.True(ok)
}

func testNoValue(s *Scope, name string, is *is.I) {
	is.True(!s.HasValue(name))
	_, ok := s.Value(name)
	is.True(!ok)
}
