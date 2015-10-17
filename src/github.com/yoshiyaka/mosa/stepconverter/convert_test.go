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
