package resolver

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/yoshiyaka/mosa/ast"
	"github.com/yoshiyaka/mosa/parser"
)

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
		// Variable definition of same name in header and body
		node 'x' {
			class { 'A': }
		}
		class A($foo = "bar",) {
			$foo = "baz"
			file { $foo: }
		}
		define single file($name,) {}
		`,
		`file { 'baz': }`,
	},

	{

		`
		// Multiple definitions of the same name
		node 'x' {
			class { 'A': }
		}
		class A {
			$foo = 1
			$foo = 'bar'
			file { $foo: }
		}
		define single file($name,) {}
		`,
		`file { 'bar': }`,
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
			$bar = 'barVal'
			$fooArray = [ $bar, ]
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

	{
		`
		// Expressions
		node 'x' { class { 'A': } }
		class A {
			$str = 'a' + 'bar'
			exec { $str: }
		}
		`,
		`
		exec { 'abar': }
		`,
	},

	{
		`
		// Expressions
		node 'x' { class { 'A': } }
		class A {
			$number = 2
			mytype { "foo" + 'bar':
				workers => 5 + 6 * $number,
				array => [
					$number, 'string', 2+3,
				],
			}
		}
		define single mytype($name, $workers, $array,) {}
		`,
		`
		mytype { 'foobar': 
			workers => 17,
			array => [ 2, 'string', 5, ],
		}
		`,
	},

	{
		`
		// If-statements
		node 'x' { class { 'A': } }
		class A {
			if true {
				exec { 'foo': }
			}
		}
		`,

		`
		exec { 'foo': }
		`,
	},

	{
		`
		// If-statements
		node 'x' { class { 'A': } }
		class A {
			if false {
				exec { 'foo': }
			} else {
				exec { 'bar': }
			}
		}
		`,

		`
		exec { 'bar': }
		`,
	},

	{
		`
		// If-statements
		node 'x' { class { 'A': } }
		class A {
			$myval = "foo"
			
			if $myval == 'foo' {
				$bar = 'fromif'
			} else {
				$bar = 'fromelse'
			}
			
			exec { $bar: }
		}
		`,

		`
		exec { 'fromif': }
		`,
	},

	{
		`
		// If-statements
		node 'x' { class { 'A': } }
		class A {
			if true {
				$bar = 'baz'
			}
			$myval = "foo$bar"
			
			exec { $myval: }
		}
		`,

		`
		exec { 'foobaz': }
		`,
	},

	{
		`
		// If-statements
		node 'x' { class { 'A': } }
		class A {
			$myval = "foo"
			
			if $myval != 'foo' {
				$bar = 'fromif'
			} else {
				if $myval == "baz" {
					$bar = 'fromelseif'
				} else {
					$bar = 'fromelseelse'
				}
			}
			
			exec { $bar: }
		}
		`,

		`
		exec { 'fromelseelse': }
		`,
	},

	{
		`
		// If-statements in defines
		node 'x' { class { 'A': } }
		class A {
			mytype { 'foo': }
		}
		define single mytype($name,) {
			if $name == "foo" {
				exec { "name is foo": }
			}
		}
		`,

		`
		exec { 'name is foo': }
		mytype { 'foo': }
		`,
	},

	{
		`
		// Exec with unless
		node 'x' {
			exec { 'kde':
				unless => 'gnome',
			}
			exec { 'bash':
				unless => "zsh",
			}
		}
		`,

		`
		exec { 'kde':
			unless => 'gnome',
		}
		exec { 'bash':
			unless => 'zsh',
		}
		`,
	},

	{
		`define single user($name, $password,) {
			exec { "useradd -p '$password' $name":
				unless => "
				cat /etc/passwd | grep 
				-q 
				'^$name:'",
			}
		}`,
		``,
	},

	{
		`
		// Realize declaration with array
		node 'n' {
			exec { [ "bar", "baz", ]:
				stdin => "foo",
			}
		}`,
		`
		exec { 'bar': stdin => 'foo', }
		exec { 'baz': stdin => 'foo', }
		`,
	},

	{
		`
		// Realize declaration with array in define
		node 'n' {
			t { "t": }
		}
		
		define single t($name,) {
			exec { [ "bar", "baz", ]:
				stdin => "foo",
			}
		}`,
		`
		exec { 'bar': stdin => 'foo', }
		exec { 'baz': stdin => 'foo', }
		t { 't': }
		`,
	},

	{
		`
		// Realize define with array
		node 'n' {
			t { [ "bar", "baz", ]:
				stdin => "foo",
			}
		}
		
		define single t($name, $stdin,) {}
		
		`,
		`
		t { 'bar': stdin => 'foo', }
		t { 'baz': stdin => 'foo', }
		`,
	},

	{
		`
		// Realizing a declaration with an empty array
		node 'n' {
			class { 'A': }
		}
		class A {
			$array = []
			decl { $array: }
		}
		`,
		``,
	},

	{
		`node 'n' {
			t { 'bar': foo => 'bar', }
		}
		define single t($name,$foo,) {
			if $name == $foo && $foo != 'cat' {
				exec { $name: }	
			}
		}`,
		`
		exec { 'bar': }
		t { 'bar': foo => 'bar', }
		`,
	},

	{
		`node 'n' {
			$dep = 'a'
			
			exec { 'a': }
			exec { 'b':
				depends => exec["$dep"],
			}
		}
		`,
		`
		exec { 'a': }
		exec { 'b': require => exec['a'], }
		`,
	},
}

func TestResolveFile(t *testing.T) {
	for _, test := range resolveFileTest {
		t.Run("", func(t *testing.T) {
			expectedWrapper := fmt.Sprintf(`
			node 'x' {
				class { '__X': }
			}
			class __X {
				%s
			}
			`, test.expectedManifest,
			)
			expectedAST := ast.NewAST()
			err := parser.Parse(
				expectedAST, "expected.ms", strings.NewReader(expectedWrapper),
			)
			if err != nil {
				t.Log(expectedWrapper)
				t.Fatal(err)
			}

			realAST := ast.NewAST()
			realErr := parser.Parse(
				realAST, "real.ms", strings.NewReader(test.inputManifest),
			)
			if realErr != nil {
				t.Log(test.inputManifest)
				t.Fatal(realErr)
			}

			if resolvedDecls, err := Resolve(realAST); err != nil {
				t.Log(test.inputManifest)
				t.Error(err)
			} else {
				decls := expectedAST.Classes[0].Block.Statements
				r := make([]ast.Statement, len(resolvedDecls))
				for i, decl := range resolvedDecls {
					_copy := decl
					r[i] = &_copy
				}

				if !ast.StatementsEquals(decls, r) {
					declsClass := &ast.Class{Block: ast.Block{Statements: decls}}
					resolvedDeclsClass := &ast.Class{Block: ast.Block{
						Statements: r,
					}}
					t.Error(resolvedDecls)
					//					t.Log(expectedWrapper)

					t.Fatalf(
						"Got bad manifest, expected\n>>%s<< but got\n>>%s<<",
						declsClass.String(), resolvedDeclsClass.String(),
					)
				}
			}
		})
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
		&Err{Line: 2, Type: ErrorTypeUnresolvableVariable},
	},

	{
		"Cyclic array",
		`class C {
			$a = [ $a, ]
		}`,
		&Err{Line: 2, Type: ErrorTypeUnresolvableVariable},
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
		"Cyclic variables",
		`class C {
			$foo = $foo
		}`,
		&Err{
			Line: 2, Type: ErrorTypeUnresolvableVariable, SymbolName: "$foo",
		},
	},

	{
		"Multiple definitions of the same name in header",
		`class C($foo = 4, $foo = 5,) {}`,
		&Err{Line: 1, Type: ErrorTypeMultipleDefinition},
	},
}

func TestResolveBadVariable(t *testing.T) {
	for _, test := range badVariableTest {
		t.Run(test.comment, func(t *testing.T) {
			ast := ast.NewAST()
			err := parser.Parse(
				ast, "err.ms", strings.NewReader(test.inputManifest),
			)
			if err != nil {
				t.Log(test.inputManifest)
				t.Fatal(err)
			}

			gs := newGlobalState()
			resolver := newClassResolver(
				gs, &ast.Classes[0], nil, "err.ms", ast.Classes[0].LineNum,
			)
			resolveErr := resolver.resolve()
			if resolveErr == nil {
				t.Log(test.inputManifest)
				t.Log(resolver.ls)
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
						t.Logf("Expected %#v", cyclicE)
						t.Logf("Got      %#v", re)
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
		})
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
		// Realizing class with an undefined parameter
		node 'n' {
			class { 'A':
				undefined => 5,
			}
		}
		class A {}
		`,
		`Unsupported argument 'undefined' sent to type at real.ms:5`,
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
		`Reference keys must be strings (got ast.Array) at real.ms:9`,
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

	{
		`
		// Non-bool expression in if
		node 'n' {
			if "five" {}
		}
		`,
		`Expressions in if-statements must be boolean at real.ms:4`,
	},

	{
		`
		// Non-string argument to exec's unless
		node 'n' {
			exec { "foo":
				unless => 5,
			}
		}
		`,
		`Value for parameter 'unless' must be of type string at real.ms:5`,
	},
}

func TestBadDefs(t *testing.T) {
	for _, test := range badDefsTest {
		realAST := ast.NewAST()
		realErr := parser.Parse(
			realAST, "real.ms", strings.NewReader(test.manifest),
		)
		if realErr != nil {
			t.Log(test.manifest)
			t.Fatal(realErr)
		}

		if _, err := Resolve(realAST); err == nil {
			t.Log(test.manifest)
			t.Error("Got no error for bad file")
		} else if err.Error() != test.expectedErr {
			t.Log(test.manifest)
			t.Error("Got bad error:", err)
		}
	}
}

var badExpressionManifests = []struct {
	expression    string
	expectedError string
}{
	{
		`$foo = 5 / 'notanumber'`,
		``,
	},
	{
		`$foo = true > false`,
		`Comparing bools`,
	},
	{
		`$foo = 5 > true`,
		`Comparing bools`,
	},
	{
		`$foo = 5 > (4 > 3)`,
		`Comparing bools`,
	},
	{
		`$foo = 5 / 'foo'`,
		`Math on strings`,
	},
	{
		`$foo = "bar" / 'foo'`,
		`Math on strings`,
	},
	{
		`$foo = 5 * 'foo'`,
		`Math on strings`,
	},
	{
		`$foo = "foo" * 'foo'`,
		`Math on strings`,
	},
	{
		`$foo = 5 + []`,
		`Math on arrays`,
	},
	{
		`$foo = [] + 'foo'`,
		`Math on arrays`,
	},
}

func TestBadExpressionManifests(t *testing.T) {
	for _, test := range badExpressionManifests {
		realManifest := fmt.Sprintf("class c { %s }", test.expression)

		ast := ast.NewAST()
		if err := parser.Parse(ast, "test.ms", strings.NewReader(realManifest)); err != nil {
			t.Log(realManifest)
			t.Error(err)
			continue
		}

		if _, err := Resolve(ast); err == nil || err.Error() != test.expectedError {
			t.Log(test.expression)
			t.Error("Got bad error:", err)
		}
	}
}
