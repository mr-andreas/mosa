package reducer

import (
	"fmt"

	. "github.com/yoshiyaka/mosa/manifest"
)

var (
	defineExec = Define{
		Filename: "<builtin>",
		LineNum:  0,
		Name:     "exec",
		ArgDefs: []VariableDef{
			VariableDef{
				LineNum: 0,
				VariableName: VariableName{
					LineNum: 0,
					Str:     "$name",
				},
			},
		},
		Type: DefineTypeSingle,
	}
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

// Holds the global state for the complete manifest. This includes stuff such
// as all classes and types available, and the result of all classes and defines
// realized.
type globalState struct {
	classesByName map[string]*Class
	definesByName map[string]*Define

	// All realized declarations, mapped by type and name
	realizedDeclarations map[string]map[string]realizedDeclaration

	// Allows us to fetch all realized declarations in the order they were
	// defined. Not strictly necessary since the language is declarative, but it
	// makes unit testing a whole lot easier.
	realizedDeclarationsInOrder []Declaration

	// All realized classes, mapped by name
	realizedClasses map[string]realizedClass
}

func newGlobalState() *globalState {
	return &globalState{
		realizedDeclarations: map[string]map[string]realizedDeclaration{},
		realizedClasses:      map[string]realizedClass{},
	}
}

func (r *globalState) populateClassesByName(classes []Class) error {
	r.classesByName = map[string]*Class{}

	for i, class := range classes {
		if existingClass, exists := r.classesByName[class.Name]; exists {
			return fmt.Errorf(
				"Can't redefine class '%s' at %s:%d which is already defined at %s:%d",
				class.Name,
				class.Filename, class.LineNum,
				existingClass.Filename, existingClass.LineNum,
			)
		} else {
			r.classesByName[class.Name] = &classes[i]
		}
	}

	return nil
}

func (r *globalState) populateDefinesByName(defines []Define) error {
	r.definesByName = map[string]*Define{}

	// The built in exec type is always available
	r.definesByName["exec"] = &defineExec

	for i, def := range defines {
		if existingDef, exists := r.definesByName[def.Name]; exists {
			return fmt.Errorf(
				"Can't redefine type '%s' at %s:%d which is already defined at %s:%d",
				def.Name, def.Filename, def.LineNum,
				existingDef.Filename, existingDef.LineNum,
			)
		} else {
			nameKey := "$name"
			if def.Type == DefineTypeMultiple {
				nameKey = "$names"
			}

			foundNameKey := false
			for _, arg := range def.ArgDefs {
				if arg.VariableName.Str == nameKey {
					foundNameKey = true
					break
				}
			}
			if !foundNameKey {
				return fmt.Errorf(
					"Missing required argument %s when defining type '%s' at %s:%d",
					nameKey, def.Name, def.Filename, def.LineNum,
				)
			}

			r.definesByName[def.Name] = &defines[i]
		}
	}

	return nil
}
