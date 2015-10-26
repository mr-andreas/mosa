package reducer

import . "github.com/yoshiyaka/mosa/manifest"

// Resolves variable references in a class. The object holds the internal state
// of all variables used during the resolving, and should only be used once for
// each class.
type classResolver struct {
	// The class we're resolving
	original *Class

	// Args to resolve the class with
	args []Prop

	ls *localState

	realizedInFile string
	realizedAtLine int
}

func newClassResolver(class *Class, withArgs []Prop, realizedIn string, at int) *classResolver {
	return &classResolver{
		original:       class,
		args:           withArgs,
		realizedInFile: realizedIn,
		realizedAtLine: at,
		ls:             newLocalState(class.Filename, realizedIn, at),
	}
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
func (cr *classResolver) resolve() (Class, error) {
	c := cr.original
	retClass := *cr.original

	// Start by loading all top-level variables defined
	if err := cr.ls.setVarsFromArgs(cr.args, cr.original.ArgDefs); err != nil {
		return retClass, err
	}
	for _, def := range c.VariableDefs {
		if _, exists := cr.ls.varDefsByName[def.VariableName.Str]; exists {
			return retClass, &Err{
				Line:       def.LineNum,
				Type:       ErrorTypeMultipleDefinition,
				SymbolName: string(def.VariableName.Str),
			}
		}

		cr.ls.varDefsByName[def.VariableName.Str] = def
	}

	// Resolve top-level variables defined
	newDefs := make([]VariableDef, len(c.VariableDefs))
	for i, def := range c.VariableDefs {
		var err error
		def.Val, err = cr.ls.resolveValue(def.Val, def.LineNum)
		if err != nil {
			return retClass, err
		}
		newDefs[i] = def
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
//
//  package { $webserver:
//  	ensure => present,
//  	workers => $workers,
//  }
//
// Would be resolved into
//
//  package { 'nginx':
//  	ensure => present,
//  	workers => 5,
//  }
//
// when $webserver = 'nginx' and $workers = 5 are defined in the class.
func (cr *classResolver) resolveDeclaration(decl *Declaration) (Declaration, error) {
	ret := *decl

	if variable, ok := decl.Scalar.(VariableName); ok {
		// The current value points to a variable, for instance foo => $bar.
		// Resolve it.
		if v, err := cr.ls.resolveVariable(variable, decl.LineNum); err != nil {
			return ret, err
		} else {
			ret.Scalar = v
		}
	}

	var err error
	ret.Props, err = cr.ls.resolveProps(decl.Props)
	if err != nil {
		return ret, err
	}

	return ret, nil
}
