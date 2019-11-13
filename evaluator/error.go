package evaluator

import (
	"errors"
	"fmt"
)

type evalError struct {
	err  error
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

func newEvalErrorf(line int, col int, s string, args ...interface{}) *evalError {
	var e error
	if len(args) > 0 {
		e = fmt.Errorf(s, args...)
	} else {
		e = errors.New(s)
	}
	return newEvalError(e, line, col)
}

// IsEvaluationError returns whether e is an evaluation error that occurred in the evaluator.
func IsEvaluationError(e error) bool {
	var ee *evalError
	return errors.As(e, &ee)
}

// ErrorLocation returns the location in the template where the evaluation error e occurred.
// ok will be true if e actually was an evaluation error.
func ErrorLocation(e error) (line int, col int, ok bool) {
	var ee *evalError
	if errors.As(e, &ee) {
		line = ee.line
		col = ee.col
		ok = true
	}
	return
}

func (e *evalError) Error() string {
	return fmt.Sprintf("evaluation error at line %d, column %d: %v", e.line, e.col, e.err)
}
