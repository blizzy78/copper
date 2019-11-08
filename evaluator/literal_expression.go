package evaluator

import "github.com/blizzy78/copper/ast"

func evalNilLiteral() interface{} {
	return nil
}

func evalIntLiteral(i ast.IntLiteral) interface{} {
	return i.Value
}

func evalBoolLiteral(b ast.BoolLiteral) interface{} {
	return b.Value
}

func evalStringLiteral(s ast.StringLiteral) interface{} {
	return s.Value
}
