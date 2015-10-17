package reducer

import (
	"fmt"

	. "github.com/yoshiyaka/mosa/manifest"
)

// Resolves a whole manifest
type resolver struct {
	ast *File

	classesByName map[string]*Class

	// All realized declarations, mapped by type and name
	realizedDeclarations map[string]map[string]*Declaration

	// All realized classes, mapped by name
	realizedClasses map[string]*Class
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

	r.realizedClasses = map[string]*Class{}
	r.realizedDeclarations = map[string]map[string]*Declaration{}
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

	return nil, nil
}

func (r *resolver) populateClassesByName() error {
	r.classesByName = map[string]*Class{}

	for i, class := range r.ast.Classes {
		if existingClass, exists := r.classesByName[class.Name]; exists {
			return fmt.Errorf(
				"Can't redfined class '%s' at %s:%d which is already defined at %s:%d",
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

	return nil
}
