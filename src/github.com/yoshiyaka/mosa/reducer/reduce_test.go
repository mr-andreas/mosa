package reducer

import (
	"reflect"
	"strings"
	"testing"

	"github.com/yoshiyaka/mosa/manifest2"
)

var resolveVariablesTest = []struct {
	inputManifest,
	expectedManifest string
}{
	{
		`class C {}`,
		`class C {}`,
	},

	{
		`class C {
			$foo = 'x'
			$bar = $foo
		}`,
		`class C {
			$foo = 'x'
			$bar = 'x'
		}`,
	},

	{
		`class C {
			$foo = 'x'
			$bar = '$foo'
		}`,
		`class C {
			$foo = 'x'
			$bar = '$foo'
		}`,
	},

	{
		`class C {
  			$foo = 'bar'
 			$baz = $foo

			package { $baz: }
		}`,

		`class C {
  			$foo = 'bar'
			$baz = 'bar'

			package { 'bar': }
		}`,
	},
}

func TestResolveVariable(t *testing.T) {
	for _, test := range resolveVariablesTest {
		expectedFile, err := manifest2.Lex(
			"expected.ms", strings.NewReader(test.expectedManifest),
		)
		if err != nil {
			t.Fatal(err)
		}

		realFile, realErr := manifest2.Lex(
			"real.ms", strings.NewReader(test.inputManifest),
		)
		if realErr != nil {
			t.Fatal(realErr)
		}

		if !reflect.DeepEqual(expectedFile, realFile) {
			t.Error(realFile.String())
		}
	}
}

var badVariableTest = []struct {
	comment       string
	inputManifest string
	expectedError Error
}{
	{
		"Non-existing variable",
		`class C { $foo = $bar }`,
		Error{Line: 1, Type: ErrorTypeUnresolvableVariable},
	},
}

func TestResolveBadVariable(t *testing.T) {

}
