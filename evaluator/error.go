package evaluator

import (
	"errors"
	"fmt"
)

type evalError struct {
	err  error
	msg  string
	line int
	col  int
}

func newEvalError(e error, line int, col int) *evalError {
	return &evalError{
		err:  e,
		line: line,
		col:  col,
	}
}

func newEvalErrorf(line int, col int, format string, args ...interface{}) *evalError {
	return &evalError{
		msg:  fmt.Sprintf(format, args...),
		line: line,
		col:  col,
	}
}

// IsEvaluationError returns whether e is an evaluation error that occurred in the evaluator.
func IsEvaluationError(e error) bool {
	var ee *evalError
	return errors.As(e, &ee)
}

// ErrorLocation returns the location in the template where the evaluation error e occurred.
// ok will be true if e actually was an evaluation error.
func ErrorLocation(e error) (int, int, bool) {
	var ee *evalError
	if !errors.As(e, &ee) {
		return 0, 0, false
	}
	return ee.line, ee.col, true
}

func (e evalError) Error() string {
	if e.msg != "" {
		return fmt.Sprintf("evaluation error at line %d, column %d: %s", e.line, e.col, e.msg)
	}
	return fmt.Sprintf("evaluation error at line %d, column %d: %v", e.line, e.col, e.err)
}
