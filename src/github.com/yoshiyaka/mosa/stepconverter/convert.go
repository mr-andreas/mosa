package stepconverter

import (
	"github.com/yoshiyaka/mosa/common"
	. "github.com/yoshiyaka/mosa/manifest"
)

// Converts the specified manifest into a number of concrete steps that needs to
// be execute in order to fullfill it. The manifest should already have all
// references resolved by the reducer.
func Convert(ast *File) ([]common.Step, error) {
	return nil, nil
}
