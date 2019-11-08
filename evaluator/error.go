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

func (e *evalError) Error() string {
	return fmt.Sprintf("evaluation error at line %d, column %d: %v", e.line, e.col, e.err)
}

func IsEvaluationError(e error) bool {
	var ee *evalError
	return errors.As(e, &ee)
}

func ErrorLocation(e error) (line int, col int, ok bool) {
	var ee *evalError
	if errors.As(e, &ee) {
		line = ee.line
		col = ee.col
		ok = true
	}
	return
}
