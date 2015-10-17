package reducer

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/yoshiyaka/mosa/manifest"
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
		expectedFile, err := manifest.Lex(
			"expected.ms", strings.NewReader(test.expectedManifest),
		)
		if err != nil {
			t.Log(test.inputManifest)
			t.Fatal(err)
		}

		realFile, realErr := manifest.Lex(
			"real.ms", strings.NewReader(test.inputManifest),
		)
		if realErr != nil {
			t.Fatal(realErr)
		}

		resolver := newClassResolver(&realFile.Classes[0])
		if reducedClass, err := resolver.Resolve(); err != nil {
			t.Log(test.inputManifest)
			t.Fatal(err)
		} else {
			c := expectedFile.Classes[0]
			c.Filename = "real.ms"
			if !reflect.DeepEqual(c, reducedClass) {
				t.Logf("%#v", c)
				t.Logf("%#v", reducedClass)
				t.Fatal(
					"Got bad manifest, expected", c.String(),
					"got", reducedClass.String(),
				)
			}

			if len(resolver.varDefsByName) != 0 && false {
				t.Fatal(
					"Not all variables were resolved in", test.inputManifest,
					resolver.varDefsByName,
				)
			}
		}
	}
}

var resolveFileTest = []struct {
	inputManifest,
	expectedManifest string
}{
	{
		`
		node 'x' {}
		class A{}`,
		``,
	},

	{
		`
		node 'x' {
			class { 'A': }
		}
		
		class A {
			$foo = 'A'
			$bar = $foo
			file { $bar: }
		}`,
		`file { 'A': }`,
	},

	{
		`
		node 'x' {
			class { 'A': }
			class { 'B': }
		}
		
		class A {
			$foo = 'A'
			$bar = $foo
			file { $bar: }
		}
		class B {
			$foo = 'B'
			$bar = $foo
			file { $bar: }
		}`,
		`
		file { 'A': }
		file { 'B': }
		`,
	},

	{
		`
		node 'localhost' {
			class { 'Webserver':
				docroot => '/home/www',
			}
		}
		
		class Webserver(
			$docroot = '/var/www',
			$workers = 8,
		){
			$server = 'nginx'
			
			package { $server: ensure => installed }
			
			file { '/etc/nginx/conf.d/workers.conf':
				ensure => 'present',
				content => $workers,
				depends => package[$server],
			}
			
			file { $docroot: ensure => 'directory', }
			
			service { $server:
				ensure => 'running',
				depends => [
					file['/etc/nginx/conf.d/workers.conf'],
					package[$server],
				],
			}
		}`,
		`
		package { 'nginx': ensure => installed }

		file { '/etc/nginx/conf.d/workers.conf':
			ensure => 'present',
			content => 8,
			depends => package['nginx'],
		}
		
		file { '/home/www': ensure => 'directory', }
		
		service { 'nginx':
			ensure => 'running',
			depends => [
				file['/etc/nginx/conf.d/workers.conf'],
				package['nginx'],
			],
		}`,
	},

	{
		`
		// Defining the same package multiple times is okay, as long as only one
		// of the declarations is realized.
		node 'n' {
			class { 'A': }
		}
		class A {
			package { 'foo': from => 'A', }
		}
		class B {
			package { 'foo': from => 'B', }
		}
		`,

		`package { 'foo': from => 'A', }`,
	},

	{
		`
		// Nested cyclic realization
		node 'n' {
			class { 'A': 
				subclass => 'B',
				b_var => 'foo',
			}
		}
		class A($subclass, $b_var) {
			decl { 'a_decl': }
			class { $subclass:
				var => $b_var,
			}
		}
		class B($var) {
			decl { 'b_decl':
				var => $var,
			}
		}
		`,

		`
		decl { 'a_decl': }
		decl { 'b_decl':
			var => 'foo',
		}
		`,
	},
}

func TestResolveFile(t *testing.T) {
	for _, test := range resolveFileTest {
		expectedWrapper := fmt.Sprintf(`
			node 'x' {
				class { '__X': }
			}
			class __X {
				%s
			}
			`, test.expectedManifest,
		)
		expectedFile, err := manifest.Lex(
			"expected.ms", strings.NewReader(expectedWrapper),
		)
		if err != nil {
			t.Log(expectedWrapper)
			t.Fatal(err)
		}

		realFile, realErr := manifest.Lex(
			"real.ms", strings.NewReader(test.inputManifest),
		)
		if realErr != nil {
			t.Log(test.inputManifest)
			t.Fatal(realErr)
		}

		if reducedDecls, err := Reduce(realFile); err != nil {
			t.Log(test.inputManifest)
			t.Fatal(err)
		} else if decls := expectedFile.Classes[0].Declarations; !reflect.DeepEqual(decls, reducedDecls) {
			t.Logf("%#v", decls)
			t.Logf("%#v", reducedDecls)

			declsClass := &manifest.Class{Declarations: decls}
			reducedDeclsClass := &manifest.Class{Declarations: reducedDecls}

			t.Fatalf(
				"Got bad manifest, expected\n>>%s<< but got\n>>%s<<",
				declsClass.String(), reducedDeclsClass.String(),
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
		&Err{Line: 3, Type: ErrorTypeCyclicVariable},
	},

	{
		"Nested cyclic variables $foo -> $bar -> $baz -> $foo",
		`class C {
			$foo = $bar
			$bar = $baz
			$baz = $foo
		}`,
		&Err{Line: 3, Type: ErrorTypeCyclicVariable},
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
		ast, err := manifest.Lex(
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

var badDefsTest = []struct {
	manifest    string
	expectedErr string
}{
	{
		`
		// Multiple classes with the same name
		class A {}
		class A {}
		`,
		`Can't redfined class 'A' at real.ms:4 which is already defined at real.ms:3`,
	},

	{
		`
		// Multiple realization of the same class
		node 'x' {
			class { 'A': }
			class { 'A': }
		}
		class A {}
		`,
		`Can't realize class 'A' multiple times at real.ms:5, first defined at real.ms:4`,
	},

	{
		`
		// Multiple realization of the same declaration
		node 'n' {
			class { 'A': }
			class { 'B': }
		}
		class A {
			package { 'foo': from => 'A', }
		}
		class B {
			package { 'foo': from => 'B', }
		}
		`,

		`A nice error`,
	},

	{
		`
		// Cyclic realization
		node 'n' {
			class { 'A': }
		}
		class A {
			class { 'A': }
		}
		`,
		`An error`,
	},

	{
		`
		// Nested cyclic realization
		node 'n' {
			class { 'A': }
		}
		class A {
			class { 'B': }
		}
		class B {
			class { 'A': }
		}
		`,
		`An error`,
	},

	{
		`
		// Nested cyclic realization with variables
		node 'n' {
			class { 'A':
				subclass => 'B',
			}
		}
		class A($subclass,) {
			class { $subclass: }
		}
		class B {
			class { 'A': }
		}
		`,
		`An error`,
	},

	{
		`
		// Realizing a declaration with a non-string (number) name
		node 'n' {
			class { 'A': }
		}
		class A {
			$number = 5
			decl { $number: }
		}
		`,
		`An error`,
	},

	{
		`
		// Realizing a declaration with a non-string (array) name
		node 'n' {
			class { 'A': }
		}
		class A {
			$array = []
			decl { $array: }
		}
		`,
		`An error`,
	},
}

func TestBadDefs(t *testing.T) {
	for _, test := range badDefsTest {
		realFile, realErr := manifest.Lex(
			"real.ms", strings.NewReader(test.manifest),
		)
		if realErr != nil {
			t.Log(test.manifest)
			t.Fatal(realErr)
		}

		if _, err := Reduce(realFile); err == nil {
			t.Log(realFile)
			t.Error("Got no error for bad file")
		} else if err.Error() != test.expectedErr {
			t.Log(test.manifest)
			t.Fatal("Got bad error:", err)
		}
	}
}
