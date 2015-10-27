package reducer

import (
	"fmt"

	. "github.com/yoshiyaka/mosa/manifest"
)

// Realizes a given define
type declarationResolver struct {
	// The define we're resolving
	define *Define

	// The name of the realization. In package { 'bash': }, name would be 'bash'
	name Value

	// Args to resolve the define with
	args []Prop

	ls *localState
	gs *globalState

	realizedInFile string
	realizedAtLine int
}

func newDeclarationResolver(d *Define, name Value, withArgs []Prop, gs *globalState, realizedIn string, at int) *declarationResolver {
	return &declarationResolver{
		define:         d,
		name:           name,
		args:           withArgs,
		realizedInFile: realizedIn,
		realizedAtLine: at,
		ls:             newLocalState(d.Filename, realizedIn, at),
	}
}

func (cr *declarationResolver) resolve() (Define, error) {
	def := cr.define

	retClass := *cr.define

	nameKey := "name"
	if def.Type == DefineTypeMultiple {
		nameKey = "names"
	}
	for _, arg := range cr.args {
		if arg.Name == nameKey {
			return retClass, fmt.Errorf(
				"'%s' may not be passed as an argument in %s:%d",
				nameKey, cr.realizedInFile, arg.LineNum,
			)
		}
	}

	cr.args = append(cr.args, Prop{
		LineNum: 0,
		Name:    nameKey,
		Value:   cr.name,
	})

	// Start by loading all top-level variables defined
	if err := cr.ls.setVarsFromArgs(cr.args, def.ArgDefs); err != nil {
		return retClass, err
	}

	for _, def := range def.VariableDefs {
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
	newDefs := make([]VariableDef, len(def.VariableDefs))
	for i, def := range def.VariableDefs {
		var err error
		def.Val, err = cr.ls.resolveValue(def.Val, def.LineNum)
		if err != nil {
			return retClass, err
		}
		newDefs[i] = def
	}

	retClass.VariableDefs = newDefs

	retClass.Declarations = make([]Declaration, len(def.Declarations))
	for i, decl := range def.Declarations {
		if decl.Type == "class" {
			return retClass, fmt.Errorf(
				"Can't realize classes inside of a define at %s:%d",
				cr.define.Filename, decl.LineNum,
			)
		}

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
func (cr *declarationResolver) resolveDeclaration(decl *Declaration) (Declaration, error) {
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
