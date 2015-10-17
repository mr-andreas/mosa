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
				Type: "package",
				Item: "foo",
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
				Type:    "package",
				Item:    "foo",
				Args:    map[string]interface{}{"ensure": "latest"},
				Depends: map[string][]string{"file": []string{"undefined"}},
			},
		},
	},

	{
		`
		node 'x' { class { 'A': } }
		class A {
			$content = 'string conent'
			
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
				Args: map[string]interface{}{"ensure": "latest"},
				Depends: map[string][]string{
					"file": []string{"undefined", "anotherfile"},
				},
			},
			Step{
				Type: "file",
				Item: "anotherfile",
				Args: map[string]interface{}{
					"ensure":  "present",
					"content": "string content",
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
			t.Error("For", test.manifest, "got bad steps:", steps)
		}
	}
}
