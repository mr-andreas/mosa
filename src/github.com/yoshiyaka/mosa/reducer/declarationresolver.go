package reducer

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

	var err error
	retClass.Block, err = cr.resolveBlock(&cr.define.Block)
	if err != nil {
		return retClass, err
	}

	return retClass, nil
}

func (cr *declarationResolver) resolveBlock(block *Block) (Block, error) {
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
	for _, decl := range block.Declarations {
		if decl.Type == "class" {
			return retBlock, fmt.Errorf(
				"Can't realize classes inside of a define at %s:%d",
				cr.define.Filename, decl.LineNum,
			)
		}

		if _, err := cr.resolveDeclaration(&decl); err != nil {
			return retBlock, err
		}
	}

	return retBlock, nil
}

func (cr *declarationResolver) resolveIf(_if *If) (If, error) {
	retIf := *_if

	var boolean bool
	if boolVal, err := cr.ls.resolveValue(_if.Expression, _if.LineNum); err != nil {
		return retIf, err
	} else if realBool, ok := boolVal.(Bool); !ok {
		return retIf, fmt.Errorf(
			"Expressions in if-statements must be boolean at %s:%d",
			cr.define.Filename, _if.LineNum,
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

func (cr *declarationResolver) resolveDeclaration(decl *Declaration) (Declaration, error) {
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
			ret.Type, cr.define.Filename, ret.LineNum,
		)
	} else {
		name = string(n)
		ret.Scalar = n
	}

	if previous := cr.gs.lockRealization(decl, name, cr.define.Filename, decl.LineNum); previous != nil {
		return ret, fmt.Errorf(
			"%s['%s'] realized twice at %s:%d. Previously realized at %s:%d",
			decl.Type, name, cr.define.Filename, decl.LineNum,
			previous.file, previous.line,
		)
	}

	if err := cr.realizeDeclaration(name, &ret); err != nil {
		return ret, err
	}

	return ret, nil
}

func (cr *declarationResolver) realizeDeclaration(name string, decl *Declaration) error {
	def, defOk := cr.gs.definesByName[decl.Type]
	if !defOk {
		return fmt.Errorf(
			"Reference to undefined type '%s' at %s:%d",
			decl.Type, cr.define.Filename, decl.LineNum,
		)
	}

	dr := newDeclarationResolver(
		def, decl.Scalar, decl.Props, cr.gs, cr.define.Filename,
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
		file: cr.define.Filename,
		line: decl.LineNum,
	}
	cr.gs.realizedDeclarationsInOrder = append(
		cr.gs.realizedDeclarationsInOrder, *decl,
	)

	return nil
}
