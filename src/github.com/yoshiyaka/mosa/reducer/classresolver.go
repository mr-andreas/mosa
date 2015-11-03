package reducer

import (
	"fmt"

	. "github.com/yoshiyaka/mosa/ast"
)

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

	if v, err := cr.ls.resolveValue(decl.Scalar, decl.LineNum); err != nil {
		return ret, err
	} else {
		ret.Scalar = v
	}

	var err error
	ret.Props, err = cr.ls.resolveProps(decl.Props)
	if err != nil {
		return ret, err
	}

	var name string
	if n, ok := ret.Scalar.(QuotedString); !ok {
		return ret, fmt.Errorf(
			"Can't realize declaration of type %s with non-string name at %s:%d",
			ret.Type, cr.original.Filename, ret.LineNum,
		)
	} else {
		name = string(n)
	}

	if previous := cr.gs.lockRealization(decl, name, cr.original.Filename, decl.LineNum); previous != nil {
		return ret, fmt.Errorf(
			"%s['%s'] realized twice at %s:%d. Previously realized at %s:%d",
			decl.Type, name, cr.original.Filename, decl.LineNum,
			previous.file, previous.line,
		)
	}

	if ret.Type == "class" {
		if err := cr.realizeClass(name, &ret); err != nil {
			return ret, err
		}
	} else {
		if err := cr.realizeDeclaration(name, &ret); err != nil {
			return ret, err
		}
	}

	return ret, nil
}

func (cr *classResolver) realizeClass(name string, decl *Declaration) error {
	if class, ok := cr.gs.classesByName[name]; !ok {
		return fmt.Errorf(
			"Reference to undefined class '%s' at %s:%d",
			string(name), cr.original.Filename, decl.LineNum,
		)
	} else if oldDef, defined := cr.gs.realizedClasses[name]; defined {
		return fmt.Errorf(
			"Class %s realized twice at %s:%d. Previously realized at %s:%d",
			string(name), cr.original.Filename, decl.LineNum,
			oldDef.file, oldDef.line,
		)
	} else {
		cr.gs.realizedClasses[string(name)] = realizedClass{
			c:    class,
			file: cr.original.Filename,
			line: decl.LineNum,
		}
		nestedResolver := newClassResolver(
			cr.gs, class, decl.Props, cr.original.Filename, decl.LineNum,
		)
		_, err := nestedResolver.resolve()
		return err
	}
}

func (cr *classResolver) realizeDeclaration(name string, decl *Declaration) error {
	def, defOk := cr.gs.definesByName[decl.Type]
	if !defOk {
		return fmt.Errorf(
			"Reference to undefined type '%s' at %s:%d",
			decl.Type, cr.original.Filename, decl.LineNum,
		)
	}

	dr := newDeclarationResolver(
		def, decl.Scalar, decl.Props, cr.gs, cr.original.Filename,
		decl.LineNum,
	)
	if _, err := dr.resolve(); err != nil {
		return err
	}

	if cr.gs.realizedDeclarations[decl.Type] == nil {
		cr.gs.realizedDeclarations[decl.Type] = map[string]realizedDeclaration{}
	}

	cr.gs.realizedDeclarations[decl.Type][name] = realizedDeclaration{
		d:    decl,
		file: cr.original.Filename,
		line: decl.LineNum,
	}
	cr.gs.realizedDeclarationsInOrder = append(
		cr.gs.realizedDeclarationsInOrder, *decl,
	)

	return nil
}
