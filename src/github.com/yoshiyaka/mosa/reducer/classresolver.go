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

	var err error
	retClass.Block, err = cr.resolveBlock(&c.Block)
	if err != nil {
		return retClass, err
	}

	return retClass, nil
}

func (cr *classResolver) resolveBlock(block *Block) (Block, error) {
	retBlock := *block

	for _, def := range block.VariableDefs {
		if _, exists := cr.ls.varDefsByName[def.VariableName.Str]; exists {
			return retBlock, &Err{
				Line:       def.LineNum,
				Type:       ErrorTypeMultipleDefinition,
				SymbolName: string(def.VariableName.Str),
			}
		}

		cr.ls.varDefsByName[def.VariableName.Str] = def
	}

	// Resolve top-level variables defined
	newDefs := make([]VariableDef, len(block.VariableDefs))
	for i, def := range block.VariableDefs {
		var err error
		def.Val, err = cr.ls.resolveValue(def.Val, def.LineNum)
		if err != nil {
			return retBlock, err
		}
		newDefs[i] = def
	}
	retBlock.VariableDefs = newDefs

	retBlock.Ifs = make([]If, len(block.Ifs))
	for i, _if := range block.Ifs {
		var err error
		retBlock.Ifs[i], err = cr.resolveIf(&_if)
		if err != nil {
			return retBlock, err
		}
	}

	retBlock.Declarations = make([]Declaration, len(block.Declarations))
	for i, decl := range block.Declarations {
		var err error
		retBlock.Declarations[i], err = cr.resolveDeclaration(&decl)
		if err != nil {
			return retBlock, err
		}
	}

	return retBlock, nil
}

func (cr *classResolver) resolveIf(_if *If) (If, error) {
	retIf := *_if

	var boolean bool
	if boolVal, err := cr.ls.resolveValue(_if.Expression, _if.LineNum); err != nil {
		return retIf, err
	} else if realBool, ok := boolVal.(Bool); !ok {
		return retIf, fmt.Errorf(
			"Expressions in if-statements must be boolean at %s:%d",
			cr.original.Filename, _if.LineNum,
		)
	} else {
		boolean = bool(realBool)
	}

	if boolean {
		var err error
		retIf.Block, err = cr.resolveBlock(&_if.Block)
		if err != nil {
			return retIf, err
		}
	} else if _if.Else != nil {
		if block, err := cr.resolveBlock(_if.Else); err != nil {
			return retIf, err
		} else {
			retIf.Else = &block
		}
	}

	return retIf, nil
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
	if resolvedName, err := cr.ls.resolveValue(ret.Scalar, decl.LineNum); err != nil {
		return ret, err
	} else if n, ok := resolvedName.(QuotedString); !ok {
		return ret, fmt.Errorf(
			"Can't realize declaration of type %s with non-string name at %s:%d",
			ret.Type, cr.original.Filename, ret.LineNum,
		)
	} else {
		name = string(n)
		ret.Scalar = n
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
