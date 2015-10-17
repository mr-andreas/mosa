package reducer

import (
	"fmt"
	"strings"

	. "github.com/yoshiyaka/mosa/manifest"
)

type ErrorType int

const (
	ErrorTypeUnresolvableVariable ErrorType = iota
	ErrorTypeCyclicVariable
	ErrorTypeMultipleDefinition
)

type Err struct {
	Type       ErrorType
	File       string
	Line       int
	SymbolName string
}

func (e *Err) Error() string {
	msg := ""
	switch e.Type {
	case ErrorTypeCyclicVariable:
		msg = "Cyclic dependency for variable " + e.SymbolName
	case ErrorTypeMultipleDefinition:
		msg = "Multiple definition for variable " + e.SymbolName
	case ErrorTypeUnresolvableVariable:
		msg = "Reference to non-defined variable " + e.SymbolName
	default:
		msg = "Unknown"
	}

	return fmt.Sprintf("Error at %s:%d: %s", e.File, e.Line, msg)
}

type CyclicError struct {
	Err
	Cycle []string
}

func (ce *CyclicError) Error() string {
	msg := ce.Err.Error()
	msg += fmt.Sprintf(" (%s)", strings.Join(ce.Cycle, " -> "))
	return msg
}

// Resolves all variables and references in the specified manifest.
func Reduce(ast *File) (*File, error) {
	retFile := *ast
	retFile.Classes = make([]Class, len(ast.Classes))

	for i, class := range ast.Classes {
		var err error
		resolver := newClassResolver(&class)
		retFile.Classes[i], err = resolver.Resolve()
		if err != nil {
			return nil, err
		}
	}

	return &retFile, nil
}
