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

			exec { $baz: }
		}`,

		`class C {
  			$foo = 'bar'
			$baz = 'bar'

			exec { 'bar': }
		}`,
	},

	{
		`class C {
  			$foo = 'foostr'
			$bar = 'barstr'
 			$baz = "$foo x $bar"

			exec { $baz: }
		}`,

		`class C {
  			$foo = 'foostr'
			$bar = 'barstr'
			$baz = 'foostr x barstr'

			exec { 'foostr x barstr': }
		}`,
	},

	{
		`class C {
  			$foo = 'foostr'
			$bar = "$foo barstr"
 			$baz = "$foo x $bar"

			exec { $baz: }
		}`,

		`class C {
  			$foo = 'foostr'
			$bar = 'foostr barstr'
			$baz = 'foostr x foostr barstr'

			exec { 'foostr x foostr barstr': }
		}`,
	},

	{
		`
		class C {
  			$foo = 'bar'

			package { 'baz': fooVal => $foo, }
		}
		define single package($name, $fooVal,){}
		`,

		`class C {
  			$foo = 'bar'

			package { 'baz': fooVal => 'bar', }
		}`,
	},

	{
		`class C {
  			$foo = 'bar'
 			$baz = $foo

			package { $baz: fooVal => $baz, }
		}
		define single package($name, $fooVal,){}
		`,

		`class C {
  			$foo = 'bar'
			$baz = 'bar'

			package { 'bar': fooVal => 'bar', }
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

	{
		`class C {
			$foo = 'foo'
			bar { 'baz':
				val => ref[$foo],
			}
		}
		define single bar($name,$val,){}
		`,
		`class C {
			$foo = 'foo'
			bar { 'baz':
				val => ref['foo'],
			}
		}`,
	},

	{
		`class C {
			$foo = 'foo'
			$ref = [ ref[$foo], ]
		}`,
		`class C {
			$foo = 'foo'
			$ref = [ ref['foo'], ]
		}`,
	},

	{
		`class C {
			$foo = 'foo'
			$ref = [ [ ref[$foo], ], ]
		}`,
		`class C {
			$foo = 'foo'
			$ref = [ [ ref['foo'], ], ]
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

		gs := newGlobalState()
		gs.populateClassesByName(realFile.Classes)
		gs.populateDefinesByName(realFile.Defines)
		resolver := newClassResolver(
			gs, &realFile.Classes[0], nil, "real.ms", realFile.Classes[0].LineNum,
		)
		if reducedClass, err := resolver.resolve(); err != nil {
			t.Log(test.inputManifest)
			t.Fatal(err)
		} else {
			c := expectedFile.Classes[0]
			c.Filename = "real.ms"
			if !c.Equals(&reducedClass) {
				t.Logf("%#v", c)
				t.Logf("%#v", reducedClass)
				t.Fatal(
					"Got bad manifest, expected", c.String(),
					"got", reducedClass.String(),
				)
			}

			if len(resolver.ls.varDefsByName) != 0 && false {
				t.Fatal(
					"Not all variables were resolved in", test.inputManifest,
					resolver.ls.varDefsByName,
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
		}
		
		define single file($name,) {}
		`,
		`file { 'A': }`,
	},

	{
		`
		node 'x' {
			class { 'A': }
		}

		class A {
			$foo = 'fooVal'
			file { 'filename':
				value => $foo,
			}
		}

		define single file($name, $value,) {}
		`,
		`file { 'filename':
			value => 'fooVal',
		}`,
	},

	{
		`
		node 'x' {
			class { 'A': }
		}

		class A {
			$fooArray = [ $bar, ]
			$bar = 'barVal'
			file { 'filename':
				value => $fooArray,
			}
		}

		define single file($name, $value,) {}
		`,
		`file { 'filename':
			value => [ 'barVal', ],
		}`,
	},

	{
		`
		node 'x' {
			class { 'A': }
		}

		class A {
			$fileVar = 'f1'
			file { $fileVar: }
			file { 'f2':
				depends => file[$fileVar],
			}
		}

		define single file($name,) {}
		`,
		`
		file { 'f1': }
		file { 'f2': depends => file['f1'], }
		`,
	},

	{
		`
		node 'x' {
			class { 'A': }
		}

		class A {
			$fileVar = 'f1'
			file { $fileVar: }
			file { 'f2':
				depends => [ file[$fileVar], ],
			}
		}

		define single file($name, ) {}
		`,
		`
		file { 'f1': }
		file { 'f2': depends => [ file['f1'], ], }
		`,
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
		}

		define single file($name, ) {}
		`,
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

			package { $server: ensure => 'installed', }

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
		}

		define single file($name, $ensure, $content = '',) {}
		define single package($name, $ensure,) {}
		define single service($name, $ensure,) {}
		`,
		`
		package { 'nginx': ensure => 'installed', }

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

		define single package($name, $from,) {}
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
		class A($subclass, $b_var,) {
			decl { 'a_decl': }
			class { $subclass:
				var => $b_var,
			}
		}
		class B($var,) {
			decl { 'b_decl':
				var => $var,
			}
		}

		define single decl($name, $var = '',) {}
		`,

		`
		decl { 'a_decl': }
		decl { 'b_decl':
			var => 'foo',
		}
		`,
	},

	{
		`
		// Realize empty define
		node 'x' { class { 'A': } }
		class A {
			mytype { 'foo': }
		}
		define single mytype($name,){}
		`,
		`
		mytype { 'foo': }
		`,
	},

	{
		`
		// Realize simple define
		node 'x' { class { 'A': } }
		class A {
			mytype { 'foo': }
		}
		define single mytype($name,){
			exec { 'echo foo': }
		}
		`,
		`
		exec { 'echo foo': }
		mytype { 'foo': }
		`,
	},

	{
		`
		// Realize simple define with interpolated string
		node 'x' { class { 'A': } }
		class A {
			mytype { "foostr": }
		}
		define single mytype($name,){
			exec { "echo $name": }
		}
		`,
		`
		exec { 'echo foostr': }
		mytype { 'foostr': }
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
		} else if decls := expectedFile.Classes[0].Declarations; !manifest.DeclarationsEquals(decls, reducedDecls) {
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
		"Cyclic interpolated string",
		`class C {
			$foo = "$foo"
		}`,
		&Err{Line: 2, Type: ErrorTypeCyclicVariable},
	},

	{
		"Cyclic array",
		`class C {
			$a = [ $a, ]
		}`,
		&Err{Line: 2, Type: ErrorTypeCyclicVariable},
	},

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
		"Nested cyclic variables in string $foo -> $bar -> $baz -> $foo",
		`class C {
			$foo = $bar
			$bar = "$baz"
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

	{
		"Multiple definitions of the same name in header",
		`class C($foo = 4, $foo = 5,) {}`,
		&Err{Line: 1, Type: ErrorTypeMultipleDefinition},
	},
	{
		"Variable definition of same name in header and body",
		`class C($foo = 5,) {
			$foo = 4
		}`,
		&Err{Line: 2, Type: ErrorTypeMultipleDefinition},
	},
}

func TestResolveBadVariable(t *testing.T) {
	for _, test := range badVariableTest {
		ast, err := manifest.Lex(
			"err.ms", strings.NewReader(test.inputManifest),
		)
		if err != nil {
			t.Log(test.inputManifest)
			t.Fatal(err)
		}

		gs := newGlobalState()
		resolver := newClassResolver(
			gs, &ast.Classes[0], nil, "err.ms", ast.Classes[0].LineNum,
		)
		resolved, resolveErr := resolver.resolve()
		if resolveErr == nil {
			t.Log(test.inputManifest)
			t.Log(&resolved)
			t.Fatal("Got no error for", test.comment)
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
		`Can't redefine class 'A' at real.ms:4 which is already defined at real.ms:3`,
	},

	{
		`
		// Reference to undefined class in node
		node 'x' {
			class { 'Undefined': }
		}
		`,
		`Reference to undefined class 'Undefined' at real.ms:4`,
	},

	{
		`
		// Reference to undefined class in class
		node 'x' {
			class { 'A': }
		}
		class A {
			class { 'Undefined': }
		}
		`,
		`Reference to undefined class 'Undefined' at real.ms:7`,
	},

	{
		`
		// Reference to undefined class in class by var
		node 'x' {
			class { 'A': }
		}
		class A {
			$var = 'VarValue'
			class { $var: }
		}
		`,
		`Reference to undefined class 'VarValue' at real.ms:8`,
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
		`class['A'] realized twice at real.ms:5. Previously realized at real.ms:4`,
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
		
		define single package($name, $from,){}
		`,

		`package['foo'] realized twice at real.ms:11. Previously realized at real.ms:8`,
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
		`class['A'] realized twice at real.ms:7. Previously realized at real.ms:4`,
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
		`class['A'] realized twice at real.ms:10. Previously realized at real.ms:4`,
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
		`class['A'] realized twice at real.ms:12. Previously realized at real.ms:4`,
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
		`Can't realize declaration of type decl with non-string name at real.ms:8`,
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
		`Can't realize declaration of type decl with non-string name at real.ms:8`,
	},

	{
		`
		// Realizing class with an undefined parameter
		node 'n' {
			class { 'A':
				undefined => 5,
			}
		}
		class A {}
		`,
		`Unsupported argument 'undefined' sent to class at real.ms:5`,
	},

	{
		`
		// Realizing class without supplying a required parameter
		node 'n' {
			class { 'A': }
		}
		class A($required,) {}
		`,
		`Required argument 'required' not supplied at real.ms:4`,
	},

	{
		`
		// A reference with an array value
		node 'n' {
			class { 'A': }
		}
		class A {
			$array = []
			file { 'x':
				ref => file[$array],
			}
		}
		`,
		`Reference keys must be strings at real.ms:9 - the value of $array is not.`,
	},

	{
		`
		// Reference to undefined type
		node 'n' {
			myType { 'A': }
		}
		`,
		`Reference to undefined type 'myType' at real.ms:4`,
	},

	{
		`
		// Nested multiple definition of the same type
		node 'n' {
			testtype { 'bar': }
			exec { 'bar': }
		}
		define single testtype($name,) {
			exec { $name: }
		}
		`,
		`exec['bar'] realized twice at real.ms:5. Previously realized at real.ms:8`,
	},

	{
		`
		// Single define without name parameter
		define single testtype($names,) {}
		`,
		`Missing required argument $name when defining type 'testtype' at real.ms:3`,
	},

	{
		`
		// Multiple define without names parameter
		define multiple testtype($name,) {}
		`,
		`Missing required argument $names when defining type 'testtype' at real.ms:3`,
	},

	{
		`
		// Define same type multiple times
		define single x($name,){}
		define single x($name,){}
		`,
		`Can't redefine type 'x' at real.ms:4 which is already defined at real.ms:3`,
	},

	{
		`
		// Supply name in props
		define single x($name,){}
		class A {
			x { 'y':
				name => 'y',
			}
		}
		node 'x' { class { 'A': } }
		`,
		`'name' may not be passed as an argument in real.ms:6`,
	},

	{
		`
		// Supply names in props
		define multiple x($names,){}
		class A {
			x { 'y':
				names => 'y',
			}
		}
		node 'x' { class { 'A': } }
		`,
		`'names' may not be passed as an argument in real.ms:6`,
	},

	{
		`
		// Cyclic defines
		define single foo($name,) {
			bar { $name: }
		}
		define single bar($name,) {
			foo { $name: }
		}
		class A {
			foo { 'baz': }
		}
		node 'x' { class { 'A': } }
		`,
		`foo['baz'] realized twice at real.ms:7. Previously realized at real.ms:10`,
	},

	{
		`
		// Class inside define
		node 'n' {
			class { 'B': }
		}
		class A {}
		class B {
			x { 'test': }
		}
		define single x($name,) {
			class { 'A': }
		}
		`,
		`Can't realize classes inside of a define at real.ms:11`,
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
			t.Log(test.manifest)
			t.Error("Got no error for bad file")
		} else if err.Error() != test.expectedErr {
			t.Log(test.manifest)
			t.Error("Got bad error:", err)
		}
	}
}
