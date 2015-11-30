package resolver

import (
	"fmt"

	. "github.com/yoshiyaka/mosa/ast"
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
		gs:             gs,
	}
}

func (cr *declarationResolver) resolve() (Define, error) {
	retClass := *cr.define

	nameKey := "name"
	if cr.define.Type == DefineTypeMultiple {
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
	if err := cr.ls.setVarsFromArgs(cr.args, cr.define.ArgDefs); err != nil {
		return retClass, err
	}

	br := newBlockResolver(&cr.define.Block, cr.ls, cr.gs, false)
	var err error
	retClass.Block, err = br.resolve()
	if err != nil {
		return retClass, err
	}

	return retClass, nil
}
