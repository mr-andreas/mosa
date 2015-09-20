package reducer

import (
	"fmt"

	"github.com/yoshiyaka/mosa/common"
	"github.com/yoshiyaka/mosa/manifest2"
)

type ErrorType int

const (
	ErrorTypeUnresolvableVariable = iota
	ErrorTypeCyclicVariable
)

type Error struct {
	Type    ErrorType
	File    string
	Line    int
	Message string
}

func (e *Error) Error() string {
	return fmt.Sprintf("Error at %s:%d: %s", e.File, e.Line, e.Message)
}

// Resolves the specified manifest and converts into a number of steps the need
// to be executed in order to reach it.
func Reduce(ast *manifest2.File) []*common.Step {
	return nil
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
func resolveVariables(c *manifest2.Class) (manifest2.Class, error) {

}
