package resolver

import (
	"fmt"

	. "github.com/yoshiyaka/mosa/ast"
)

type blockResolver struct {
	block *Block

	ls *localState
	gs *globalState

	// If false, declarations like
	//  class { "myclass": }
	// are not allowed. This is set to true if the block is inside of a define.
	allowClassRealizations bool
}

func newBlockResolver(b *Block, ls *localState, gs *globalState, allowClassRealizations bool) *blockResolver {
	return &blockResolver{
		block: b,
		ls:    ls,
		gs:    gs,
		allowClassRealizations: allowClassRealizations,
	}
}

func (br *blockResolver) resolve() error {
	for _, stmt := range br.block.Statements {
		switch v := stmt.(type) {
		case *VariableDef:
			if val, err := br.ls.resolveValue(v.Val, v.LineNum); err != nil {
				return err
			} else {
				br.ls.resolvedVars[v.VariableName.Str] = val
			}
			//			newStmts[i] = val

		case *If:
			if err := br.resolveIf(v); err != nil {
				return err
			}

		case *Declaration:
			if _, err := br.resolveDeclaration(v); err != nil {
				return err
			}
		}
	}

	return nil
}

func (cr *blockResolver) resolveDeclaration(decl *Declaration) ([]Declaration, error) {
	var ret []Declaration

	var resolvedNames Value
	if v, err := cr.ls.resolveValue(decl.Scalar, decl.LineNum); err != nil {
		return ret, err
	} else {
		resolvedNames = v
	}

	var props []Prop
	if p, err := cr.ls.resolveProps(decl.Props); err != nil {
		return ret, err
	} else {
		props = p
	}

	var names []QuotedString
	if n, ok := resolvedNames.(QuotedString); ok {
		names = []QuotedString{n}
	} else if namesArray, ok := resolvedNames.(Array); ok {
		names = make([]QuotedString, len(namesArray))
		for i, name := range namesArray {
			if n, ok := name.(QuotedString); !ok {
				return ret, fmt.Errorf(
					"Can't realize declaration of type %s with non-string name at %s:%d",
					decl.Type, cr.block.Filename, decl.LineNum,
				)
			} else {
				names[i] = n
			}
		}
	} else {
		return ret, fmt.Errorf(
			"Can't realize declaration of type %s with non-string name at %s:%d",
			decl.Type, cr.block.Filename, decl.LineNum,
		)
	}

	ret = make([]Declaration, 0, len(names))
	for _, name := range names {
		if previous := cr.gs.lockRealization(decl, string(name), cr.block.Filename, decl.LineNum); previous != nil {
			return ret, fmt.Errorf(
				"%s[%s] realized twice at %s:%d. Previously realized at %s:%d",
				decl.Type, name, cr.block.Filename, decl.LineNum,
				previous.file, previous.line,
			)
		}

		declCopy := *decl
		declCopy.Props = props
		declCopy.Scalar = name

		if declCopy.Type == "class" {
			if err := cr.realizeClass(string(name), &declCopy); err != nil {
				return ret, err
			}
		} else {
			if err := cr.realizeDeclaration(string(name), &declCopy); err != nil {
				return ret, err
			}
		}
		ret = append(ret, declCopy)
	}

	return ret, nil
}

func (cr *blockResolver) realizeDeclaration(name string, decl *Declaration) error {
	def, defOk := cr.gs.definesByName[decl.Type]
	if !defOk {
		return fmt.Errorf(
			"Reference to undefined type '%s' at %s:%d",
			decl.Type, cr.block.Filename, decl.LineNum,
		)
	}

	dr := newDeclarationResolver(
		def, decl.Scalar, decl.Props, cr.gs, cr.block.Filename,
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
		file: cr.block.Filename,
		line: decl.LineNum,
	}
	cr.gs.realizedDeclarationsInOrder = append(
		cr.gs.realizedDeclarationsInOrder, *decl,
	)

	return nil
}

func (br *blockResolver) realizeClass(name string, decl *Declaration) error {
	if !br.allowClassRealizations {
		return fmt.Errorf(
			"Can't realize classes inside of a define at %s:%d",
			br.block.Filename, decl.LineNum,
		)
	}

	if class, ok := br.gs.classesByName[name]; !ok {
		return fmt.Errorf(
			"Reference to undefined class '%s' at %s:%d",
			string(name), br.block.Filename, decl.LineNum,
		)
	} else if oldDef, defined := br.gs.realizedClasses[name]; defined {
		return fmt.Errorf(
			"Class %s realized twice at %s:%d. Previously realized at %s:%d",
			string(name), br.block.Filename, decl.LineNum,
			oldDef.file, oldDef.line,
		)
	} else {
		br.gs.realizedClasses[string(name)] = realizedClass{
			c:    class,
			file: br.block.Filename,
			line: decl.LineNum,
		}
		nestedResolver := newClassResolver(
			br.gs, class, decl.Props, br.block.Filename, decl.LineNum,
		)
		return nestedResolver.resolve()
	}
}

func (br *blockResolver) resolveIf(_if *If) error {
	var boolean bool
	if boolVal, err := br.ls.resolveValue(_if.Expression, _if.LineNum); err != nil {
		return err
	} else if realBool, ok := boolVal.(Bool); !ok {
		return fmt.Errorf(
			"Expressions in if-statements must be boolean at %s:%d",
			br.block.Filename, _if.LineNum,
		)
	} else {
		boolean = bool(realBool)
	}

	if boolean {
		br := newBlockResolver(
			&_if.Block, br.ls, br.gs, br.allowClassRealizations,
		)

		if err := br.resolve(); err != nil {
			return err
		}
	} else if _if.Else != nil {
		br := newBlockResolver(
			_if.Else, br.ls, br.gs, br.allowClassRealizations,
		)

		if err := br.resolve(); err != nil {
			return err
		}
	}

	return nil
}
