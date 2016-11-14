package reducer

import (
	"fmt"
	"strings"
	"testing"

	. "github.com/yoshiyaka/mosa/ast"
	"github.com/yoshiyaka/mosa/parser"
	"github.com/yoshiyaka/mosa/resolver"
)

var reducerTests = []struct {
	inputManifest    string
	expectedManifest string
}{
	{
		``,
		``,
	},

	{
		`node 'n' {
			exec { 'foo': }
		}
		`,
		`exec { 'foo': }`,
	},

	{
		`node 'n' {
			exec { 'foo':
				unless => "/bin/false",
			}
		}
		`,
		`exec { 'foo':
			unless => '/bin/false',
		}`,
	},

	{
		`node 'n' {
			exec { 'foo':
				unless => "/bin/true",
			}
		}
		`,
		``,
	},

	{
		`node 'n' {
			t { 'foo': }
		}
		define single t($name,) {
			exec { "t-foo": }
		}
		`,
		`exec { 't-foo': }`,
	},

	{
		`node 'n' {
			a { 'foo': }
		}
		define single a($name,) {
			b { "a-$name": }
		}
		define single b($name,) {
			exec { "b-$name": }
		}
		`,
		`exec { 'b-a-foo': }`,
	},
}

func TestReducer(t *testing.T) {
	for _, test := range reducerTests {
		inputDecls, inputErr := parseDecls(test.inputManifest)
		if inputErr != nil {
			t.Log(test.inputManifest)
			t.Error(inputErr)
			continue
		}

		realExpectedManifest := fmt.Sprintf(
			`node 'n' { %s }`, test.expectedManifest,
		)
		expectedDecls, expectedErr := parseDecls(realExpectedManifest)
		if expectedErr != nil {
			t.Log(test.expectedManifest)
			t.Error(expectedErr)
			continue
		}

		reduced, reducedErr := Reduce(inputDecls)
		if reducedErr != nil {
			t.Log(test.inputManifest)
			t.Error(reducedErr)
			continue
		}

		if !DeclarationsEquals(reduced, expectedDecls) {
			reducedStmts := make([]Statement, len(reduced))
			for i, _ := range reduced {
				reducedStmts[i] = &reduced[i]
			}

			b := &Block{Statements: reducedStmts}
			t.Errorf(
				"Bad manifest encountered. Wanted \n%s but got \n%s",
				test.expectedManifest, b.String(),
			)
		}
	}
}

func parseDecls(manifest string) ([]Declaration, error) {
	var ast AST
	if err := parser.Parse(&ast, "t.ms", strings.NewReader(manifest)); err != nil {
		return nil, err
	}

	if decls, err := resolver.Resolve(&ast); err != nil {
		return nil, err
	} else {
		return decls, nil
	}
}
