package ast

import "fmt"

// Operation. Supported values are: + - * / < > != ==
type ExpOp string

// A binary expression tree, for instance $foo + 5 or 1 == 2.
type Expression struct {
	LineNum int

	Operation ExpOp

	// Left and right hand of the operator. May be either Expression, or a value
	// (such as VariableName or QuotedString).
	Left  Value
	Right Value
}

func (e Expression) String() string {
	left := valToStr(e.Left)
	if _, ok := e.Left.(Expression); ok {
		left = "(" + left + ")"
	}

	right := valToStr(e.Right)
	if _, ok := e.Right.(Expression); ok {
		right = "(" + right + ")"
	}

	return fmt.Sprintf("%s %s %s", left, e.Operation, right)
}

func ExpressionEquals(e1, e2 *Expression) bool {
	return e1.Operation == e2.Operation &&
		ValueEquals(e1.Left, e2.Left) &&
		ValueEquals(e1.Right, e2.Right)
}
