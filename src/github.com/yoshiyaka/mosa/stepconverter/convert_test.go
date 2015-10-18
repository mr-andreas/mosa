package stepconverter

import (
	"reflect"
	"strings"
	"testing"

	. "github.com/yoshiyaka/mosa/common"
	"github.com/yoshiyaka/mosa/manifest"
	"github.com/yoshiyaka/mosa/reducer"
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
		}`,
		[]Step{},
	},

	{
		`
		node 'x' { class { 'A': } }
		class A {
			package { 'foo': }
		}`,
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
		}`,
		[]Step{
			Step{
				Type: "package",
				Item: "foo",
				Args: map[string]interface{}{
					"ensure": manifest.QuotedString("latest"),
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
		}`,
		[]Step{
			Step{
				Type: "package",
				Item: "foo",
				Args: map[string]interface{}{
					"ensure": manifest.QuotedString("latest"),
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
		}`,
		[]Step{
			Step{
				Type: "package",
				Item: "foo",
				Args: map[string]interface{}{
					"ensure": manifest.QuotedString("latest"),
				},
				Depends: map[string][]string{
					"file": []string{"undefined", "anotherfile"},
				},
			},
			Step{
				Type: "file",
				Item: "anotherfile",
				Args: map[string]interface{}{
					"ensure":  manifest.QuotedString("present"),
					"content": manifest.QuotedString("string content"),
				},
				Depends: map[string][]string{"file": []string{"undefined"}},
			},
		},
	},
}

func TestConvert(t *testing.T) {
	for _, test := range convertTests {
		ast, astErr := manifest.Lex("test.ms", strings.NewReader(test.manifest))
		if astErr != nil {
			t.Fatal(astErr)
		}

		reduced, reducedErr := reducer.Reduce(ast)
		if reducedErr != nil {
			t.Fatal(reducedErr)
		}

		if steps, err := Convert(reduced); err != nil {
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
		`,
		`Good error here`,
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
		`,
		`Good error here`,
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
		`,
		`Good error here`,
	},
}

func TestConvertInvalidManifests(t *testing.T) {
	for _, test := range invalidManifests {
		ast, astErr := manifest.Lex("test.ms", strings.NewReader(test.manifest))
		if astErr != nil {
			t.Log(test.manifest)
			t.Fatal(astErr)
		}

		reduced, reducedErr := reducer.Reduce(ast)
		if reducedErr != nil {
			t.Log(test.manifest)
			t.Fatal(reducedErr)
		}

		if _, err := Convert(reduced); err == nil {
			t.Log(test.manifest)
			t.Error("Got no error")
		} else if err.Error() != test.err {
			t.Log(test.manifest)
			t.Error("Got bad error:", err.Error())
		}
	}
}
