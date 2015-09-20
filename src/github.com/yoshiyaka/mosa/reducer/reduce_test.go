package reducer

import "testing"

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
