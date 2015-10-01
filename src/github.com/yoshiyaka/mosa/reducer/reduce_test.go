package reducer

import (
	"reflect"
	"strings"
	"testing"

	"github.com/yoshiyaka/mosa/manifest2"
)

var resolveClassTest = []struct {
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

			package { 'baz': name => $foo, }
		}`,

		`class C {
  			$foo = 'bar'

			package { 'baz': name => 'bar', }
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

			package { 'bar': name => 'bar', }
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
			$baz = [ 'baz', $bar, ]
		}`,
		`class C {
			$foo = 'foo'
			$bar = [ 'foo', 1, 'z', ]
			$baz = [ 'baz', [ 'foo', 1, 'z', ], ]
		}`,
	},
}

func TestResolveClass(t *testing.T) {
	for _, test := range resolveClassTest {
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

		resolver := newClassResolver(&realFile.Classes[0])
		if reducedClass, err := resolver.Resolve(); err != nil {
			t.Log(test.inputManifest)
			t.Fatal(err)
		} else if c := expectedFile.Classes[0]; !reflect.DeepEqual(c, reducedClass) {
			//			t.Logf("%#v", expectedFile)
			//			t.Logf("%#v", reducedFile)
			t.Fatal(
				"Got bad manifest, expected", c.String(),
				"got", reducedClass.String(),
			)
		}
	}
}

var resolveFileTest = []struct {
	inputManifest,
	expectedManifest string
}{

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

func TestResolveFile(t *testing.T) {
	for _, test := range resolveFileTest {
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

		if reducedFile, err := Reduce(realFile); err != nil {
			t.Log(test.inputManifest)
			t.Fatal(err)
		} else if !reflect.DeepEqual(expectedFile, &reducedFile) {
			//			t.Logf("%#v", expectedFile)
			//			t.Logf("%#v", &reducedFile)
			t.Error(
				"Got bad manifest, expected", expectedFile.String(),
				"got", reducedFile.String(),
			)
		}
	}
}

var badVariableTest = []struct {
	comment       string
	inputManifest string
	expectedError error
}{
	{
		"Non-existing variable",
		`class C { $foo = $bar }`,
		&Err{Line: 1, Type: ErrorTypeUnresolvableVariable},
	},

	{
		"Non-existing variable",
		`class C {
			file { $undefined: }
		}`,
		&Err{Line: 2, Type: ErrorTypeUnresolvableVariable},
	},

	{
		"Non-existing variable",
		`class C {
			file { '/etc/issue': content => $text, }
		}`,
		&Err{Line: 2, Type: ErrorTypeUnresolvableVariable},
	},

	{
		"Non-existing nested variable",
		`class C {
			$foo = $bar
			$bar = $baz
		}`,
		&Err{Line: 3, Type: ErrorTypeUnresolvableVariable},
	},

	{
		"Cyclic variables",
		`class C {
			$foo = $foo
		}`,
		&CyclicError{
			Err:   Err{Line: 2, Type: ErrorTypeCyclicVariable},
			Cycle: []string{"$foo", "$foo"},
		},
	},

	{
		"Cyclic variables",
		`class C {
			$foo = $bar
			$bar = $foo
		}`,
		&Err{Line: 2, Type: ErrorTypeCyclicVariable},
	},

	{
		"Nested cyclic variables $foo -> $bar -> $baz -> $foo",
		`class C {
			$foo = $bar
			$bar = $baz
			$baz = $foo
		}`,
		&Err{Line: 2, Type: ErrorTypeCyclicVariable},
	},

	{
		"Nested cyclic variables with arrays",
		`class C {
			$foo = $bar
			$bar = [ 1, 'foo', $foo, ]
		}`,
		&Err{Line: 2, Type: ErrorTypeCyclicVariable},
	},

	{
		"Multiple definitions of the same name",
		`class C {
			$foo = 1
			$foo = 1
		}`,
		&Err{Line: 3, Type: ErrorTypeMultipleDefinition},
	},

	{
		"Multiple definitions of the same name",
		`class C {
			$foo = 1
			$foo = 'bar'
		}`,
		&Err{Line: 3, Type: ErrorTypeMultipleDefinition},
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

		resolver := newClassResolver(&ast.Classes[0])
		_, resolveErr := resolver.Resolve()
		if resolveErr == nil {
			t.Log(test.inputManifest)
			t.Error("Got no error for", test.comment)
		} else {
			var e, expE *Err
			if ce, ok := resolveErr.(*CyclicError); ok {
				e = &ce.Err
			} else {
				e = resolveErr.(*Err)
			}

			if ce, ok := test.expectedError.(*CyclicError); ok {
				expE = &ce.Err
			} else {
				expE = test.expectedError.(*Err)
			}

			if cyclicE, ok := test.expectedError.(*CyclicError); ok {
				if re, cyclic := resolveErr.(*CyclicError); !cyclic {
					t.Log(test.inputManifest)
					t.Errorf(
						"%s: Got non-cyclic error: %s", test.comment, resolveErr,
					)
				} else if !reflect.DeepEqual(cyclicE.Cycle, re.Cycle) {
					t.Log(test.inputManifest)
					t.Errorf("%s: Got bad cycle error: %s", test.comment, e)
				}
			}

			if e.Line != expE.Line || e.Type != expE.Type {
				t.Log(test.inputManifest)
				t.Errorf(
					"%s: Got bad error: %s. Expected %s", test.comment, e, expE,
				)
			}
		}
	}
}
