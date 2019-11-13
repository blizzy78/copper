package parser

import (
	"errors"
	"fmt"
)

type parseError struct {
	err  error
	line int
	col  int
}

func newParseError(e error, line int, col int) *parseError {
	return &parseError{
		err:  e,
		line: line,
		col:  col,
	}
}

func newParseErrorf(line int, col int, s string, args ...interface{}) *parseError {
	var e error
	if len(args) > 0 {
		e = fmt.Errorf(s, args...)
	} else {
		e = errors.New(s)
	}
	return newParseError(e, line, col)
}

// IsParseError returns whether e is a parse error that occurred in the parser.
func IsParseError(e error) bool {
	var pe *parseError
	return errors.As(e, &pe)
}

func (e *parseError) Error() string {
	return fmt.Sprintf("parse error at line %d, column %d: %v", e.line, e.col, e.err)
}
