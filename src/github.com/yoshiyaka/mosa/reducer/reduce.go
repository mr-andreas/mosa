package reducer

import (
	"fmt"
	"strings"

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

// Resolves all variables and references in the specified manifest.
func Reduce(ast *File) (File, error) {
	retFile := *ast
	retFile.Classes = make([]Class, len(ast.Classes))

	for i, class := range ast.Classes {
		var err error
		retFile.Classes[i], err = resolveVariables(&class)
		if err != nil {
			return retFile, err
		}
	}

	return retFile, nil
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

	varsByName := map[VariableName]VariableDef{}
	for _, def := range c.VariableDefs {
		if _, exists := varsByName[def.VariableName]; exists {
			return retClass, &Err{
				Line:       def.LineNum,
				Type:       ErrorTypeMultipleDefinition,
				SymbolName: string(def.VariableName),
			}
		}

		varsByName[def.VariableName] = def
	}

	newDefs := make([]VariableDef, len(c.VariableDefs))
	for i, def := range c.VariableDefs {
		switch def.Val.(type) {
		case VariableName:
			val, err := resolveVariableRecursive(
				def.Val.(VariableName), def.LineNum, []*VariableDef{&def},
				varsByName, map[VariableName]bool{def.VariableName: true},
			)
			if err != nil {
				return retClass, err
			}
			newDefs[i] = VariableDef{
				LineNum:      def.LineNum,
				VariableName: def.VariableName,
				Val:          val,
			}
		case Array:
			val, err := resolveArray(
				def.Val.(Array), def.LineNum, varsByName,
			)
			if err != nil {
				return retClass, err
			}
			newDefs[i] = VariableDef{
				LineNum:      def.LineNum,
				VariableName: def.VariableName,
				Val:          val,
			}
		default:
			newDefs[i] = def
		}
	}

	retClass.VariableDefs = newDefs

	retClass.Declarations = make([]Declaration, len(c.Declarations))
	for i, decl := range c.Declarations {
		var err error
		retClass.Declarations[i], err = resolveDeclaration(&decl, varsByName)
		if err != nil {
			return retClass, err
		}
	}

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
func resolveDeclaration(decl *Declaration, varsByName map[VariableName]VariableDef) (Declaration, error) {
	ret := *decl

	if variable, ok := decl.Scalar.(VariableName); ok {
		// The current value points to a variable, for instance foo => $bar.
		// Resolve it.
		if v, err := resolveVariable(variable, decl.LineNum, varsByName); err != nil {
			return ret, err
		} else {
			ret.Scalar = v
		}
	}

	var err error
	ret.Props, err = resolveProps(decl.Props, varsByName)
	if err != nil {
		return ret, err
	}

	return ret, nil
}

func resolveProps(props []Prop, varsByName map[VariableName]VariableDef) ([]Prop, error) {
	ret := make([]Prop, len(props))

	for i, prop := range props {
		if varName, pointsToVar := prop.Value.(VariableName); pointsToVar {
			var err error
			prop.Value, err = resolveVariable(varName, prop.LineNum, varsByName)
			if err != nil {
				return nil, err
			}
			ret[i] = prop
		} else {
			ret[i] = prop
		}
	}

	return ret, nil
}

func resolveVariable(v VariableName, lineNum int, varsByName map[VariableName]VariableDef) (Value, error) {
	return resolveVariableRecursive(
		v, lineNum, nil, varsByName, map[VariableName]bool{},
	)
}

func resolveArray(a Array, lineNum int, varsByName map[VariableName]VariableDef) (Array, error) {
	return resolveArrayRecursive(
		a, lineNum, nil, varsByName, map[VariableName]bool{},
	)
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
func resolveVariableRecursive(lookingFor VariableName, lineNum int, chain []*VariableDef, varsByName map[VariableName]VariableDef, seenNames map[VariableName]bool) (Value, error) {
	foundVar, found := varsByName[lookingFor]
	if !found {
		return nil, &Err{
			Line:       lineNum,
			Type:       ErrorTypeUnresolvableVariable,
			SymbolName: string(lookingFor),
		}
	}

	if _, isVar := foundVar.Val.(VariableName); !isVar {
		// This is an actual value
		if array, isArray := foundVar.Val.(Array); isArray {
			// The value pointed to is an array. Resolve all values in the array
			// aswell.
			seenNamesCopy := map[VariableName]bool{}
			for key, val := range seenNames {
				seenNamesCopy[key] = val
			}
			return resolveArrayRecursive(
				array, lineNum, chain, varsByName, seenNamesCopy,
			)
		} else {
			return foundVar.Val, nil
		}
	}

	if _, seen := seenNames[lookingFor]; seen {
		cycle := make([]string, len(chain)+1)
		for i, def := range chain {
			cycle[i] = string(def.VariableName)
		}
		cycle[len(cycle)-1] = string(lookingFor)

		return nil, &CyclicError{
			Err: Err{
				Line:       chain[0].LineNum,
				Type:       ErrorTypeCyclicVariable,
				SymbolName: string(chain[0].VariableName),
			},
			Cycle: cycle,
		}
	}

	seenNames[lookingFor] = true

	return resolveVariableRecursive(
		foundVar.Val.(VariableName), foundVar.LineNum, append(chain, &foundVar),
		varsByName, seenNames,
	)
}

func resolveArrayRecursive(a Array, lineNum int, chain []*VariableDef, varsByName map[VariableName]VariableDef, seenNames map[VariableName]bool) (Array, error) {
	newArray := make(Array, len(a))

	for i, val := range a {
		if varName, isVar := val.(VariableName); isVar {
			// This array entry is a variable name, resolve it.
			var err error
			seenNamesCopy := map[VariableName]bool{}
			for key, val := range seenNames {
				seenNamesCopy[key] = val
			}

			newArray[i], err = resolveVariableRecursive(
				varName, lineNum, chain, varsByName, seenNamesCopy,
			)
			if err != nil {
				return nil, err
			}
		} else {
			newArray[i] = val
		}
	}

	return newArray, nil
}
