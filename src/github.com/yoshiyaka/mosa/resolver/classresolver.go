package resolver

import . "github.com/yoshiyaka/mosa/ast"

// Resolves variable references in a class. The object holds the internal state
// of all variables used during the resolving, and should only be used once for
// each class.
type classResolver struct {
	// The class we're resolving
	original *Class

	// Args to resolve the class with
	args []Prop

	ls *localState
	gs *globalState

	realizedInFile string
	realizedAtLine int
}

func newClassResolver(gs *globalState, class *Class, withArgs []Prop, realizedIn string, at int) *classResolver {
	return &classResolver{
		original:       class,
		args:           withArgs,
		realizedInFile: realizedIn,
		realizedAtLine: at,
		ls:             newLocalState(class.Filename, realizedIn, at),
		gs:             gs,
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
func (cr *classResolver) resolve() error {
	c := cr.original

	// Start by loading all top-level variables defined
	if err := cr.ls.setVarsFromArgs(cr.args, cr.original.ArgDefs); err != nil {
		return err
	}

	br := newBlockResolver(&c.Block, cr.ls, cr.gs, true)

	var err error
	if err = br.resolve(); err != nil {
		return err
	}

	return nil
}
