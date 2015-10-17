package reducer

import (
	"fmt"

	. "github.com/yoshiyaka/mosa/manifest"
)

type realizedDeclaration struct {
	d    *Declaration
	file string
	line int
}

type realizedClass struct {
	c    *Class
	file string
	line int
}

// Resolves a whole manifest
type resolver struct {
	ast *File

	classesByName map[string]*Class

	// All realized declarations, mapped by type and name
	realizedDeclarations map[string]map[string]realizedDeclaration

	// Allows us to fetch all realized declarations in the order they were
	// defined. Not strictly necessary since the language is declarative, but it
	// makes unit testing a whole lot easier.
	realizedDecalarationsInOrder []Declaration

	// All realized classes, mapped by name
	realizedClasses map[string]realizedClass
}

func newResolver(ast *File) *resolver {
	return &resolver{
		ast: ast,
	}
}

func (r *resolver) resolve() ([]Declaration, error) {
	if err := r.populateClassesByName(); err != nil {
		return nil, err
	}

	r.realizedClasses = map[string]realizedClass{}
	r.realizedDeclarations = map[string]map[string]realizedDeclaration{}
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

	return r.realizedDecalarationsInOrder, nil
}

func (r *resolver) populateClassesByName() error {
	r.classesByName = map[string]*Class{}

	for i, class := range r.ast.Classes {
		if existingClass, exists := r.classesByName[class.Name]; exists {
			return fmt.Errorf(
				"Can't redefine class '%s' at %s:%d which is already defined at %s:%d",
				class.Name,
				class.Filename, class.LineNum,
				existingClass.Filename, existingClass.LineNum,
			)
		} else {
			r.classesByName[class.Name] = &r.ast.Classes[i]
		}
	}

	return nil
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
				if r.realizedDeclarations[decl.Type] == nil {
					r.realizedDeclarations[decl.Type] = map[string]realizedDeclaration{}
				}

				if decl.Type == "class" {
					if nestedClass, ok := r.classesByName[string(name)]; !ok {
						return fmt.Errorf(
							"Reference to undefined class '%s' at %s:%d",
							string(name), c.Filename, decl.LineNum,
						)
					} else if oldDef, defined := r.realizedClasses[string(name)]; defined {
						return fmt.Errorf(
							"Class %s realized twice at %s:%d. Previously realized at %s:%d",
							string(name), c.Filename, decl.LineNum,
							oldDef.file, oldDef.line,
						)
					} else {
						r.realizedClasses[string(name)] = realizedClass{
							c:    nestedClass,
							file: c.Filename,
							line: decl.LineNum,
						}
						if err := r.realizeClassesRecursive(nestedClass, decl.Props, c.Filename, decl.LineNum); err != nil {
							return err
						}
					}
				} else {
					r.realizedDeclarations[decl.Type][string(name)] = realizedDeclaration{
						d:    &newClass.Declarations[i],
						file: c.Filename,
						line: decl.LineNum,
					}
					r.realizedDecalarationsInOrder = append(
						r.realizedDecalarationsInOrder, newClass.Declarations[i],
					)
				}
			}
		}
	}

	return nil
}
