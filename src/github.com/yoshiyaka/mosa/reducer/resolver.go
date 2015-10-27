package reducer

import . "github.com/yoshiyaka/mosa/manifest"

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
	if err := r.gs.populateDefinesByName(r.ast.Defines); err != nil {
		return nil, err
	}

	for _, node := range r.ast.Nodes {
		if err := r.resolveNode(&node); err != nil {
			return nil, err
		}
	}

	return r.gs.realizedDeclarationsInOrder, nil
}

func (r *resolver) resolveNode(node *Node) error {
	castedClass := Class(*node)
	return r.realizeClassesRecursive(&castedClass, nil, "", 0)
}

func (r *resolver) realizeClassesRecursive(c *Class, args []Prop, file string, line int) error {
	classResolver := newClassResolver(r.gs, c, args, file, line)
	if _, err := classResolver.resolve(); err != nil {
		return err
	}

	return nil
}
