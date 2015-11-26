package stepconverter

import (
	"reflect"
	"strings"
	"testing"

	. "github.com/yoshiyaka/mosa/ast"
	. "github.com/yoshiyaka/mosa/common"
	"github.com/yoshiyaka/mosa/parser"
	"github.com/yoshiyaka/mosa/resolver"
)

var convertTests = []struct {
	manifest      string
	expectedSteps []Step
}{
	{
		``,
		[]Step{},
	},

	{
		`class A {}`,
		[]Step{},
	},

	{
		`class A {
			package { 'foo': }
		}
		define single package($name,) {}`,
		[]Step{},
	},

	{
		`
		node 'x' { class { 'A': } }
		class A {
			package { 'foo': }
		}
		define single package($name,) {}`,
		[]Step{
			Step{
				Type:    "package",
				Item:    "foo",
				Args:    map[string]interface{}{},
				Depends: nil,
			},
		},
	},

	{
		`
		node 'x' { class { 'A': } }
		class A {
			package { 'foo':
				ensure => 'latest',
			}
		}
		define single package($name, $ensure,) {}`,
		[]Step{
			Step{
				Type: "package",
				Item: "foo",
				Args: map[string]interface{}{
					"ensure": QuotedString("latest"),
				},
				Depends: nil,
			},
		},
	},

	{
		`
		node 'x' { class { 'A': } }
		class A {
			package { 'foo':
				ensure => 'latest',
				
				// It's ok referencing stuff not defined at this time, we're
				// just converting to steps, not checking dependencies yet.
				depends => file['undefined'], 
			}
		}
		define single package($name, $ensure,) {}
		`,
		[]Step{
			Step{
				Type: "package",
				Item: "foo",
				Args: map[string]interface{}{
					"ensure": QuotedString("latest"),
				},
				Depends: map[string][]string{"file": []string{"undefined"}},
			},
		},
	},

	{
		`
		node 'x' { class { 'A': } }
		class A {
			$content = 'string content'
			
			package { 'foo':
				ensure => 'latest',
				
				depends => [
					file['undefined'], 
					file['anotherfile'],
				],
			}
			
			file { 'anotherfile':
				ensure => 'present',
				content => $content,
				depends => file['undefined'],
			}
		}
		define single file($name, $ensure, $content,) {}
		define single package($name, $ensure,) {}
		`,
		[]Step{
			Step{
				Type: "package",
				Item: "foo",
				Args: map[string]interface{}{
					"ensure": QuotedString("latest"),
				},
				Depends: map[string][]string{
					"file": []string{"undefined", "anotherfile"},
				},
			},
			Step{
				Type: "file",
				Item: "anotherfile",
				Args: map[string]interface{}{
					"ensure":  QuotedString("present"),
					"content": QuotedString("string content"),
				},
				Depends: map[string][]string{"file": []string{"undefined"}},
			},
		},
	},
}

func TestConvert(t *testing.T) {
	for _, test := range convertTests {
		ast := NewAST()
		astErr := parser.Parse(ast, "test.ms", strings.NewReader(test.manifest))
		if astErr != nil {
			t.Fatal(astErr)
		}

		resolved, resolvedErr := resolver.Resolve(ast)
		if resolvedErr != nil {
			t.Log(test.manifest)
			t.Fatal(resolvedErr)
		}

		if steps, err := Convert(resolved); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(steps, test.expectedSteps) {
			t.Logf("%#v", test.expectedSteps)
			t.Logf("%#v", steps)
			t.Error("For", test.manifest, "got bad steps:", steps)
		}
	}
}

var invalidManifests = []struct {
	manifest string
	err      string
}{
	{
		`
		node 'x' {
			class { 'A': }
		}
		class A {
			file { 'foo':
				depends => 'bar',
			}
		}
		define single file($name,) {}
		`,
		`depends must be a reference or an array of references at test.ms:7`,
	},
	{
		`
		node 'x' {
			class { 'A': }
		}
		class A {
			file { 'foo':
				depends => [
					file['bar'],
					'not_a_reference',
				],
			}
		}
		define single file($name,) {}
		`,
		`depends must be a reference or an array of references at test.ms:7`,
	},

	{
		`
		node 'x' {
			class { 'A': }
		}
		class A {
			file { 'foo':
				depends => [
					file['bar'],
					[ file['baz'], ], // Too nested reference
				],
			}
		}
		define single file($name,) {}
		`,
		`depends must be a reference or an array of references at test.ms:7`,
	},
}

func TestConvertInvalidManifests(t *testing.T) {
	for _, test := range invalidManifests {
		ast := NewAST()
		astErr := parser.Parse(ast, "test.ms", strings.NewReader(test.manifest))
		if astErr != nil {
			t.Log(test.manifest)
			t.Fatal(astErr)
		}

		resolved, resolvedErr := resolver.Resolve(ast)
		if resolvedErr != nil {
			t.Log(test.manifest)
			t.Fatal(resolvedErr)
		}

		if _, err := Convert(resolved); err == nil {
			t.Log(test.manifest)
			t.Error("Got no error")
		} else if err.Error() != test.err {
			t.Log(test.manifest)
			t.Error("Got bad error:", err.Error())
		}
	}
}
