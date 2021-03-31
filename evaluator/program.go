package evaluator

import (
	"github.com/blizzy78/copper/ast"
	"github.com/blizzy78/copper/scope"
)

func (ev *Evaluator) evalProgram(p ast.Program) (interface{}, error) {
	return ev.evalStatements(p.Statements)
}

func (ev *Evaluator) evalBlock(b ast.Block) (interface{}, error) {
	os, err := ev.evalBlockCaptureAll(b)
	if err != nil {
		return nil, err
	}

	if len(os) == 0 {
		return nil, nil
	}

	return os[len(os)-1], nil
}

func (ev *Evaluator) evalBlockCaptureAll(b ast.Block) ([]interface{}, error) {
	defer func(oldScope *scope.Scope) {
		ev.scope = oldScope
	}(ev.scope)

	ev.scope = &scope.Scope{
		Parent: ev.scope,
	}

	return ev.evalStatementsCaptureAll(b.Statements)
}

func (ev *Evaluator) evalStatements(st []ast.Statement) (interface{}, error) {
	os, err := ev.evalStatementsCaptureAll(st)
	if err != nil {
		return nil, err
	}

	if len(os) == 0 {
		return nil, nil
	}

	return os[len(os)-1], nil
}

func (ev *Evaluator) evalStatementsCaptureAll(st []ast.Statement) ([]interface{}, error) {
	os := make([]interface{}, len(st))

	for i, st := range st {
		o, err := ev.evalStatement(st)
		if err != nil {
			return nil, err
		}

		if ev.breakRequested {
			if ev.loopLevel <= 0 {
				return nil, newEvalErrorf(st.Line(), st.Col(), "break outside of loop")
			}
			break
		}

		if ev.continueRequested {
			if ev.loopLevel <= 0 {
				return nil, newEvalErrorf(st.Line(), st.Col(), "continue outside of loop")
			}
			break
		}

		os[i] = o
	}

	return os, nil
}

func (ev *Evaluator) evalStatement(st ast.Statement) (interface{}, error) {
	switch stmt := st.(type) {
	case *ast.ExpressionStatement:
		return ev.evalExpressionStatement(*stmt)
	case *ast.LetStatement:
		return nil, ev.evalLetStatement(*stmt)
	case *ast.BreakStatement:
		ev.evalBreakStatement()
		return nil, nil
	case *ast.ContinueStatement:
		ev.evalContinueStatement()
		return nil, nil
	default:
		panic(newEvalErrorf(st.Line(), st.Col(), "unknown statement type: %T", st))
	}
}

func (ev *Evaluator) evalExpressionStatement(e ast.ExpressionStatement) (interface{}, error) {
	return ev.eval(e.Expression)
}

func (ev *Evaluator) evalLetStatement(l ast.LetStatement) error {
	o, err := ev.eval(l.Expression)
	if err != nil {
		return err
	}
	name := l.Ident.Name
	ev.scope.Set(name, o)
	return nil
}

func (ev *Evaluator) evalBreakStatement() {
	ev.breakRequested = true
}

func (ev *Evaluator) evalContinueStatement() {
	ev.continueRequested = true
}
