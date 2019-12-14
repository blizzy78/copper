package evaluator

import (
	"github.com/blizzy78/copper/ast"
	"github.com/blizzy78/copper/scope"
)

func (ev *Evaluator) evalProgram(p ast.Program) (o interface{}, err error) {
	o, err = ev.evalStatements(p.Statements)
	return
}

func (ev *Evaluator) evalBlock(b ast.Block) (o interface{}, err error) {
	var os []interface{}
	if os, err = ev.evalBlockCaptureAll(b); err != nil {
		return
	}

	if len(os) > 0 {
		o = os[len(os)-1]
	}

	return
}

func (ev *Evaluator) evalBlockCaptureAll(b ast.Block) (os []interface{}, err error) {
	defer func(oldScope *scope.Scope) {
		ev.scope = oldScope
	}(ev.scope)

	ev.scope = &scope.Scope{
		Parent: ev.scope,
	}

	os, err = ev.evalStatementsCaptureAll(b.Statements)

	return
}

func (ev *Evaluator) evalStatements(st []ast.Statement) (o interface{}, err error) {
	var os []interface{}
	if os, err = ev.evalStatementsCaptureAll(st); err != nil {
		return
	}

	if len(os) > 0 {
		o = os[len(os)-1]
	}

	return
}

func (ev *Evaluator) evalStatementsCaptureAll(st []ast.Statement) (os []interface{}, err error) {
	os = make([]interface{}, len(st))

	for i, st := range st {
		var o interface{}
		if o, err = ev.evalStatement(st); err != nil {
			break
		}

		if ev.breakRequested {
			if ev.loopLevel <= 0 {
				err = newEvalErrorf(st.Line(), st.Col(), "break outside of loop")
			}

			break
		}

		if ev.continueRequested {
			if ev.loopLevel <= 0 {
				err = newEvalErrorf(st.Line(), st.Col(), "continue outside of loop")
			}

			break
		}

		os[i] = o
	}

	return
}

func (ev *Evaluator) evalStatement(st ast.Statement) (o interface{}, err error) {
	switch stmt := st.(type) {
	case *ast.ExpressionStatement:
		o, err = ev.evalExpressionStatement(*stmt)
	case *ast.LetStatement:
		err = ev.evalLetStatement(*stmt)
	case *ast.BreakStatement:
		ev.evalBreakStatement()
	case *ast.ContinueStatement:
		ev.evalContinueStatement()
	default:
		panic(newEvalErrorf(st.Line(), st.Col(), "unknown statement type: %T", st))
	}
	return
}

func (ev *Evaluator) evalExpressionStatement(e ast.ExpressionStatement) (o interface{}, err error) {
	o, err = ev.eval(e.Expression)
	return
}

func (ev *Evaluator) evalLetStatement(l ast.LetStatement) (err error) {
	var o interface{}
	if o, err = ev.eval(l.Expression); err == nil {
		name := l.Ident.Name
		ev.scope.Set(name, o)
	}
	return
}

func (ev *Evaluator) evalBreakStatement() {
	ev.breakRequested = true
}

func (ev *Evaluator) evalContinueStatement() {
	ev.continueRequested = true
}
