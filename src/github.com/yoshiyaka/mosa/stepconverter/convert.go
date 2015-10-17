package stepconverter

import (
	"fmt"

	"github.com/yoshiyaka/mosa/common"
	. "github.com/yoshiyaka/mosa/manifest"
)

// Converts the specified manifest into a number of concrete steps that needs to
// be execute in order to fullfill it. The manifest should already have all
// references resolved by the reducer.
func Convert(declarations []Declaration) ([]common.Step, error) {
	steps := make([]common.Step, len(declarations))
	for i, decl := range declarations {
		args := map[string]interface{}{}
		var depends map[string][]string

		for _, prop := range decl.Props {
			if prop.Name == "depends" {
				var err error
				depends, err = propAsReferenceList(&prop)
				if err != nil {
					return nil, err
				}
			} else {
				args[prop.Name] = prop.Value
			}
		}

		steps[i] = common.Step{
			Type:    decl.Type,
			Item:    string(decl.Scalar.(QuotedString)),
			Depends: depends,
			Args:    args,
		}
	}

	return steps, nil
}

func propAsReferenceList(depends *Prop) (map[string][]string, error) {
	switch depends.Value.(type) {
	case Reference:
		r := depends.Value.(Reference)
		return map[string][]string{
			r.Type: []string{string(r.Scalar.(QuotedString))},
		}, nil
	case Array:
		ret := map[string][]string{}
		for _, val := range depends.Value.(Array) {
			if ref, ok := val.(Reference); !ok {
				return nil, fmt.Errorf(
					"depends must be a reference or an array of references at %s:%d",
					"foooo", depends.LineNum,
				)
			} else {
				if ret[ref.Type] == nil {
					ret[ref.Type] = []string{
						string(ref.Scalar.(QuotedString)),
					}
				} else {
					ret[ref.Type] = append(
						ret[ref.Type], string(ref.Scalar.(QuotedString)),
					)
				}
			}
		}
		return ret, nil
	default:
		return nil, fmt.Errorf(
			"depends must be a reference or an array of references at %s:%d",
			"foooo", depends.LineNum,
		)
	}
}
