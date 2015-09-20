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
			$bar = $foo
			$foo = 'x'
		}`,
		`class C {
			$bar = 'x'
			$foo = 'x'
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

	{
		`class C {
  			$foo = 'bar'
 			$baz = $foo

			package { $baz: name => $baz, }
		}`,

		`class C {
  			$foo = 'bar'
			$baz = 'bar'

			package { 'bar': name => $bar, }
		}`,
	},

	{
		`class C {
			$foo = 'x'
			$bar = [ $foo, ]
		}`,
		`class C {
			$foo = 'x'
			$bar = [ 'x', ]
		}`,
	},

	{
		`class C {
			$foo = 'foo'
			$bar = [ $foo, 1, 'z', ]
			$baz = [ 'baz', $foo, ]
		}`,
		`class C {
			$foo = 'foo'
			$bar = [ 'foo', 1, 'z', ]
			$baz = [ 'baz', [ 'foo', 1, 'z', ], ]
		}`,
	},

	{
		`class A {
			$foo = 'A'
			$bar = $foo
		}
		class B {
			$foo = 'B'
			$bar = $foo		
		}`,
		`class A {
			$foo = 'A'
			$bar = 'A'
		}
		class B {
			$foo = 'B'
			$bar = 'B'
		}`,
	},
}

func TestResolveVariable(t *testing.T) {
	for _, test := range resolveVariablesTest {
		expectedFile, err := manifest2.Lex(
			"expected.ms", strings.NewReader(test.expectedManifest),
		)
		if err != nil {
			t.Log(test.inputManifest)
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

	{
		"Non-existing variable",
		`class C {
			file { $undefined: }
		}`,
		Error{Line: 2, Type: ErrorTypeUnresolvableVariable},
	},

	{
		"Non-existing variable",
		`class C {
			file { '/etc/issue': content => $text, }
		}`,
		Error{Line: 2, Type: ErrorTypeUnresolvableVariable},
	},

	{
		"Non-existing nested variable",
		`class C {
			$foo = $bar
			$bar = $baz
		}`,
		Error{Line: 3, Type: ErrorTypeUnresolvableVariable},
	},

	{
		"Cyclic variables",
		`class C {
			$foo = $$foo
		}`,
		Error{Line: 2, Type: ErrorTypeCyclicVariable},
	},

	{
		"Cyclic variables",
		`class C {
			$foo = $bar
			$bar = $foo
		}`,
		Error{Line: 2, Type: ErrorTypeCyclicVariable},
	},

	{
		"Nested cyclic variables",
		`class C {
			$foo = $bar
			$bar = $baz
			$baz = $foo
		}`,
		Error{Line: 2, Type: ErrorTypeCyclicVariable},
	},

	{
		"Nested cyclic variables with arrays",
		`class C {
			$foo = $bar
			$bar = [ 1, 'foo', $foo, ]
		}`,
		Error{Line: 2, Type: ErrorTypeCyclicVariable},
	},

	{
		"Multiple definitions of the same name",
		`class C {
			$foo = 1
			$foo = 1
		}`,
		Error{Line: 2, Type: ErrorTypeMultipleDefinition},
	},

	{
		"Multiple definitions of the same name",
		`class C {
			$foo = 1
			$foo = 'bar'
		}`,
		Error{Line: 2, Type: ErrorTypeMultipleDefinition},
	},
}

func TestResolveBadVariable(t *testing.T) {
	for _, test := range badVariableTest {
		ast, err := manifest2.Lex(
			"err.ms", strings.NewReader(test.inputManifest),
		)
		if err != nil {
			t.Fatal(err)
		}

		_, resolveErr := resolveVariables(&ast.Classes[0])
		if resolveErr == nil {
			t.Error("Got no error for", test.comment)
		} else if e, ok := resolveErr.(*Error); !ok {
			t.Errorf("%s: Error was not *Error: %s", test.comment, resolveErr)
		} else {
			if e.Line != test.expectedError.Line || e.Type != test.expectedError.Type {
				t.Errorf("%s: Got bad error: %s", test.comment, e)
			}
		}
	}
}
