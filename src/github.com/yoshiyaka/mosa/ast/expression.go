package ast

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
