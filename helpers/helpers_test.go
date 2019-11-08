package helpers

import (
	"testing"

	"github.com/matryer/is"

	"github.com/blizzy78/copper/scope"
)

func TestSafe(t *testing.T) {
	is := is.New(t)

	tests := []struct {
		input    interface{}
		expected string
	}{
		{"foo", "foo"},
		{123, "123"},
		{true, "true"},
		{[]string{"a", "b", "c"}, "abc"},
		{[]interface{}{"a", "<b>", int(1), int8(2), int16(3), int32(4), int64(5), uint(6), uint8(7), uint16(8), uint32(9), uint64(10), nil, true, false}, "a<b>12345678910truefalse"},
	}

	for _, test := range tests {
		actual := Safe(test.input)
		is.Equal(string(actual), test.expected)
	}
}

func TestHTML(t *testing.T) {
	is := is.New(t)

	tests := []struct {
		input    interface{}
		expected string
	}{
		{"foo", "foo"},
		{"<foo>", "&lt;foo&gt;"},
		{123, "123"},
		{true, "true"},
		{[]string{"a", "<b>", "c"}, "a&lt;b&gt;c"},
		{[]interface{}{"a", "<b>", int(1), int8(2), int16(3), int32(4), int64(5), uint(6), uint8(7), uint16(8), uint32(9), uint64(10), nil, true, false}, "a&lt;b&gt;12345678910truefalse"},
	}

	for _, test := range tests {
		actual := HTML(test.input)
		is.Equal(string(actual), test.expected)
	}
}

func TestLen(t *testing.T) {
	is := is.New(t)

	tests := []struct {
		input    interface{}
		expected int
	}{
		{[5]int{}, 5},
		{[]int{1, 2, 3}, 3},
		{[0]int{}, 0},
		{[]int{}, 0},
		{"foo", 3},
		{"", 0},
	}

	for _, test := range tests {
		actual := Len(test.input)
		is.Equal(actual, test.expected)
	}
}

func TestHas(t *testing.T) {
	is := is.New(t)

	s := scope.Scope{}
	s.Set("foo", true)

	is.True(Has("foo", &s))
	is.True(!Has("bar", &s))
}
