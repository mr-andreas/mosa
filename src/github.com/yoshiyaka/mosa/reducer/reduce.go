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

	varsByName := map[VariableName]*VariableDef{}
	for _, def := range c.VariableDefs {
		if _, exists := varsByName[def.VariableName]; exists {
			return retClass, &Err{
				Line:       def.LineNum,
				Type:       ErrorTypeMultipleDefinition,
				SymbolName: string(def.VariableName),
			}
		}

		varsByName[def.VariableName] = &def
	}

	newDefs := make([]VariableDef, len(c.VariableDefs))
	for i, def := range c.VariableDefs {
		switch def.Val.(type) {
		case VariableName:
			varName := def.Val.(VariableName)
			val, err := resolveVariable(
				&def, []*VariableDef{&def}, varsByName, map[VariableName]bool{},
			)
			if err != nil {
				return retClass, err
			}
			newDefs[i] = VariableDef{
				VariableName: VariableName(varName),
				Val:          val,
			}
		default:
			newDefs[i] = def
		}
	}

	retClass.VariableDefs = newDefs

	//	newDecls := make([]Declaration, len(c.Declarations))
	//	for i, decl := range c.Declarations {

	//	}

	return retClass, nil
}

// Resolves all variables used in a declaration. For instance
//  package { $webserver:
//  	ensure => present,
//  	workers => $workers,
//  }
// Would be resolved into
//  package { 'nginx':
//  	ensure => present,
//  	workers => 5,
//  }
// when $webserver = 'nginx' and $workers = 5 are defined in the class.
//
// varsByName should contain a map of all top level variable definitions seen in
// the class.
func resolveDeclaration(decl *Declaration, varsByName map[VariableName]*VariableDef) (Declaration, error) {
	ret := *decl

	//	if variable, ok := decl.Scalar.(Variable); ok {
	//		// The current value points to a variable, for instance foo => $bar.
	//		// Resolve it.
	//		if v, err := resolveVariable(variable, nil, varsByName, nil); err != nil {
	//			return ret, err
	//		} else {
	//			ret.Scalar = v
	//		}
	//	}

	return ret, nil
}

// Recursively resolves a variable's actual value.
//
// chain will keep the chain used to define the variable, for instance if
// a manifest looks like
//  $foo = $bar
//  $bar = 3
// chain will contain [ $foo, $bar ]. This is used when printing errors about
// cyclic dependencies.
//
// varsByName should contain a map of all top level variable definitions seen in
// the class.
//
// seenNames is keeps track of all variables already seen during the current
// recursion. Used to detect cyclic dependencies.
func resolveVariable(varDef *VariableDef, chain []*VariableDef, varsByName map[VariableName]*VariableDef, seenNames map[VariableName]bool) (Value, error) {
	seenNames[varDef.VariableName] = true

	foundVar, found := varsByName[varDef.Val.(VariableName)]
	if !found {
		return nil, &Err{
			Line:       varDef.LineNum,
			Type:       ErrorTypeUnresolvableVariable,
			SymbolName: string(varDef.VariableName),
		}
	}

	if _, seen := seenNames[varDef.Val.(VariableName)]; seen {
		cycle := make([]string, len(chain)+1)
		for i, def := range chain {
			cycle[i] = string(def.VariableName)
		}
		cycle[len(cycle)-1] = string(varDef.VariableName)

		return nil, &CyclicError{
			Err: Err{
				Line:       chain[0].LineNum,
				Type:       ErrorTypeCyclicVariable,
				SymbolName: string(chain[0].VariableName),
			},
			Cycle: cycle,
		}
	}

	if _, isVar := foundVar.Val.(VariableName); isVar {
		return resolveVariable(foundVar, append(chain, varDef), varsByName, seenNames)
	} else {
		// This is an actual value
		return foundVar.Val, nil
	}
}
