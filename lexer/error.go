package lexer

import (
	"errors"
	"fmt"
)

type parseError struct {
	err  error
	msg  string
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

func newParseErrorf(line int, col int, format string, args ...interface{}) *parseError {
	return &parseError{
		msg:  fmt.Sprintf(format, args...),
		line: line,
		col:  col,
	}
}

func IsParseError(e error) bool {
	var pe *parseError
	return errors.As(e, &pe)
}

func (e parseError) Error() string {
	if e.msg != "" {
		return fmt.Sprintf("parse error at line %d, column %d: %s", e.line, e.col, e.msg)
	}
	return fmt.Sprintf("parse error at line %d, column %d: %v", e.line, e.col, e.err)
}
