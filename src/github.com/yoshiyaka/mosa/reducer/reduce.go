package reducer

import (
	"fmt"

	"github.com/yoshiyaka/mosa/common"
	. "github.com/yoshiyaka/mosa/manifest2"
)

type ErrorType int

const (
	ErrorTypeUnresolvableVariable = iota
	ErrorTypeCyclicVariable
	ErrorTypeMultipleDefinition
)

type Error struct {
	Type ErrorType
	File string
	Line int
}

func (e *Error) Error() string {
	msg := ""
	switch e.Type {
	case ErrorTypeCyclicVariable:
		msg = "Cyclic dependency for variable"
	case ErrorTypeMultipleDefinition:
		msg = "Multiple definition for variable"
	case ErrorTypeUnresolvableVariable:
		msg = "Reference to non-defined variable"
	default:
		msg = "Unknown"
	}

	return fmt.Sprintf("Error at %s:%d: %s", e.File, e.Line, msg)
}

// Resolves the specified manifest and converts into a number of steps the need
// to be executed in order to reach it.
func Reduce(ast *File) []*common.Step {
	return nil
}

// Resolves all variables in the class and converts them to values. For
// instance, consider the following manifest:
//
//  class C {
//  	$foo = 'bar'
// 		$baz = $foo
//
//		package { $baz: }
//	}
//
// After this function is run, the class would be returned as:
//
//  class C {
//  	$foo = 'bar'
// 		$baz = 'bar'
//
//		package { 'bar': }
//	}
func resolveVariables(c *Class) (Class, error) {
	retClass := *c

	varsByName := map[Variable]*Def{}
	for _, def := range c.Defs {
		if _, exists := varsByName[def.Name]; exists {
			return retClass, &Error{
				Line: 0,
				Type: ErrorTypeMultipleDefinition,
			}
		}

		varsByName[def.Name] = &def
	}

	newDefs := make([]Def, len(c.Defs))
	for i, def := range c.Defs {
		switch def.Val.(type) {
		case Variable:
			varName := def.Val.(Variable)
			val, err := resolveVariable(
				varName, varsByName, map[Variable]bool{},
			)
			if err != nil {
				return retClass, err
			}
			newDefs[i] = Def{
				Name: Variable(varName),
				Val:  val,
			}
		default:
			newDefs[i] = def
		}
	}

	retClass.Defs = newDefs

	return retClass, nil
}

func resolveVariable(name Variable, varsByName map[Variable]*Def, seenNames map[Variable]bool) (Value, error) {
	foundVar, found := varsByName[name]
	if !found {
		return nil, &Error{
			Line: 0,
			Type: ErrorTypeUnresolvableVariable,
		}
	}

	if _, seen := seenNames[name]; seen {
		return nil, &Error{
			Line: 0,
			Type: ErrorTypeCyclicVariable,
		}
	}

	seenNames[name] = true

	if variable, isVar := foundVar.Val.(Variable); isVar {
		return resolveVariable(variable, varsByName, seenNames)
	} else {
		// This is an actual value
		return foundVar.Val, nil
	}
}
