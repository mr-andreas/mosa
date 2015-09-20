package reducer

import (
	"fmt"
	"strings"

	"github.com/yoshiyaka/mosa/common"
	. "github.com/yoshiyaka/mosa/manifest2"
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
			return retClass, &Err{
				Line:       def.LineNum,
				Type:       ErrorTypeMultipleDefinition,
				SymbolName: string(def.Name),
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
				&def, []*Def{&def}, varsByName, map[Variable]bool{},
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

func resolveVariable(varDef *Def, chain []*Def, varsByName map[Variable]*Def, seenNames map[Variable]bool) (Value, error) {
	seenNames[varDef.Name] = true

	foundVar, found := varsByName[varDef.Val.(Variable)]
	if !found {
		return nil, &Err{
			Line:       varDef.LineNum,
			Type:       ErrorTypeUnresolvableVariable,
			SymbolName: string(varDef.Name),
		}
	}

	if _, seen := seenNames[varDef.Val.(Variable)]; seen {
		cycle := make([]string, len(chain)+1)
		for i, def := range chain {
			cycle[i] = string(def.Name)
		}
		cycle[len(cycle)-1] = string(varDef.Name)

		return nil, &CyclicError{
			Err: Err{
				Line:       chain[0].LineNum,
				Type:       ErrorTypeCyclicVariable,
				SymbolName: string(chain[0].Name),
			},
			Cycle: cycle,
		}
	}

	if _, isVar := foundVar.Val.(Variable); isVar {
		return resolveVariable(foundVar, append(chain, varDef), varsByName, seenNames)
	} else {
		// This is an actual value
		return foundVar.Val, nil
	}
}
