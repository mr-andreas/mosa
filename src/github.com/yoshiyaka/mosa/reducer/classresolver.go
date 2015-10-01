package reducer

import . "github.com/yoshiyaka/mosa/manifest2"

// Resolves variable references in a class. The object holds the internal state
// of all variables used during the resolving, and should only be used once for
// each class.
type classResolver struct {
	// The class we're resolving
	original *Class

	// Contains a map of all top level variable definitions seen in the class.
	varDefsByName map[VariableName]VariableDef
}

func newClassResolver(class *Class) *classResolver {
	return &classResolver{
		original: class,
	}
}

func (cr *classResolver) resolveProps(props []Prop) ([]Prop, error) {
	ret := make([]Prop, len(props))

	for i, prop := range props {
		if varName, pointsToVar := prop.Value.(VariableName); pointsToVar {
			var err error
			prop.Value, err = cr.resolveVariable(varName, prop.LineNum)
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

func (cr *classResolver) resolveVariable(v VariableName, lineNum int) (Value, error) {
	return cr.resolveVariableRecursive(
		v, lineNum, nil, map[VariableName]bool{},
	)
}

func (cr *classResolver) resolveArray(a Array, lineNum int) (Array, error) {
	return cr.resolveArrayRecursive(
		a, lineNum, nil, map[VariableName]bool{},
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
// seenNames is keeps track of all variables already seen during the current
// recursion. Used to detect cyclic dependencies.
func (cr *classResolver) resolveVariableRecursive(lookingFor VariableName, lineNum int, chain []*VariableDef, seenNames map[VariableName]bool) (Value, error) {
	foundVar, found := cr.varDefsByName[lookingFor]
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
			return cr.resolveArrayRecursive(
				array, lineNum, chain, seenNamesCopy,
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

	return cr.resolveVariableRecursive(
		foundVar.Val.(VariableName), foundVar.LineNum, append(chain, &foundVar),
		seenNames,
	)
}

func (cr *classResolver) resolveArrayRecursive(a Array, lineNum int, chain []*VariableDef, seenNames map[VariableName]bool) (Array, error) {
	newArray := make(Array, len(a))

	for i, val := range a {
		if varName, isVar := val.(VariableName); isVar {
			// This array entry is a variable name, resolve it.
			var err error
			seenNamesCopy := map[VariableName]bool{}
			for key, val := range seenNames {
				seenNamesCopy[key] = val
			}

			newArray[i], err = cr.resolveVariableRecursive(
				varName, lineNum, chain, seenNamesCopy,
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
func (cr *classResolver) Resolve() (Class, error) {
	c := cr.original
	retClass := *cr.original

	cr.varDefsByName = map[VariableName]VariableDef{}
	for _, def := range c.VariableDefs {
		if _, exists := cr.varDefsByName[def.VariableName]; exists {
			return retClass, &Err{
				Line:       def.LineNum,
				Type:       ErrorTypeMultipleDefinition,
				SymbolName: string(def.VariableName),
			}
		}

		cr.varDefsByName[def.VariableName] = def
	}

	newDefs := make([]VariableDef, len(c.VariableDefs))
	for i, def := range c.VariableDefs {
		switch def.Val.(type) {
		case VariableName:
			val, err := cr.resolveVariableRecursive(
				def.Val.(VariableName), def.LineNum, []*VariableDef{&def},
				map[VariableName]bool{def.VariableName: true},
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
			val, err := cr.resolveArray(
				def.Val.(Array), def.LineNum,
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
		retClass.Declarations[i], err = cr.resolveDeclaration(&decl)
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
func (cr *classResolver) resolveDeclaration(decl *Declaration) (Declaration, error) {
	ret := *decl

	if variable, ok := decl.Scalar.(VariableName); ok {
		// The current value points to a variable, for instance foo => $bar.
		// Resolve it.
		if v, err := cr.resolveVariable(variable, decl.LineNum); err != nil {
			return ret, err
		} else {
			ret.Scalar = v
		}
	}

	var err error
	ret.Props, err = cr.resolveProps(decl.Props)
	if err != nil {
		return ret, err
	}

	return ret, nil
}
