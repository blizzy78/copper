package lexer

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

func IsParseError(e error) bool {
	var pe *parseError
	return errors.As(e, &pe)
}

func (e *parseError) Error() string {
	return fmt.Sprintf("parse error at line %d, column %d: %v", e.line, e.col, e.err)
}
