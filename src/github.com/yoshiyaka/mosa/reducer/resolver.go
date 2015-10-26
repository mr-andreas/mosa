package reducer

import (
	"fmt"

	. "github.com/yoshiyaka/mosa/manifest"
)

// Resolves a whole manifest
type resolver struct {
	ast *File

	gs *globalState
}

func newResolver(ast *File) *resolver {
	return &resolver{
		ast: ast,
		gs:  newGlobalState(),
	}
}

func (r *resolver) resolve() ([]Declaration, error) {
	if err := r.gs.populateClassesByName(r.ast.Classes); err != nil {
		return nil, err
	}

	for _, node := range r.ast.Nodes {
		if err := r.resolveNode(&node); err != nil {
			return nil, err
		}
	}

	//	retFile := *ast
	//	retFile.Classes = make([]Class, len(ast.Classes))

	//	for i, class := range ast.Classes {
	//		var err error
	//		resolver := newClassResolver(&class)
	//		retFile.Classes[i], err = resolver.Resolve()
	//		if err != nil {
	//			return nil, err
	//		}
	//	}

	//	return &retFile, nil

	return r.gs.realizedDeclarationsInOrder, nil
}

func (r *resolver) resolveNode(node *Node) error {
	castedClass := Class(*node)
	return r.realizeClassesRecursive(&castedClass, nil, "", 0)
}

func (r *resolver) realizeClassesRecursive(c *Class, args []Prop, file string, line int) error {
	classResolver := newClassResolver(c, args, file, line)
	if newClass, err := classResolver.resolve(); err != nil {
		return err
	} else {
		for i, decl := range newClass.Declarations {
			if name, ok := decl.Scalar.(QuotedString); !ok {
				return fmt.Errorf(
					"Can't realize declaration of type %s with non-string name at %s:%d",
					decl.Type, c.Filename, decl.LineNum,
				)
			} else {
				if r.gs.realizedDeclarations[decl.Type] == nil {
					r.gs.realizedDeclarations[decl.Type] = map[string]realizedDeclaration{}
				}

				if decl.Type == "class" {
					if nestedClass, ok := r.gs.classesByName[string(name)]; !ok {
						return fmt.Errorf(
							"Reference to undefined class '%s' at %s:%d",
							string(name), c.Filename, decl.LineNum,
						)
					} else if oldDef, defined := r.gs.realizedClasses[string(name)]; defined {
						return fmt.Errorf(
							"Class %s realized twice at %s:%d. Previously realized at %s:%d",
							string(name), c.Filename, decl.LineNum,
							oldDef.file, oldDef.line,
						)
					} else {
						r.gs.realizedClasses[string(name)] = realizedClass{
							c:    nestedClass,
							file: c.Filename,
							line: decl.LineNum,
						}
						if err := r.realizeClassesRecursive(nestedClass, decl.Props, c.Filename, decl.LineNum); err != nil {
							return err
						}
					}
				} else {
					if oldDef, ok := r.gs.realizedDeclarations[decl.Type][string(name)]; ok {
						return fmt.Errorf(
							"Declaration %s[%s] realized twice at %s:%d. Previously realized at %s:%d",
							decl.Type, string(name), c.Filename, decl.LineNum,
							oldDef.file, oldDef.line,
						)
					} else {
						r.gs.realizedDeclarations[decl.Type][string(name)] = realizedDeclaration{
							d:    &newClass.Declarations[i],
							file: c.Filename,
							line: decl.LineNum,
						}
						r.gs.realizedDeclarationsInOrder = append(
							r.gs.realizedDeclarationsInOrder, newClass.Declarations[i],
						)
					}
				}
			}
		}
	}

	return nil
}
