package manifest2

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

var lexTests = []struct {
	manifest string
	ast      *File
}{
	{
		`
		class Test {
		}
		`,

		&File{
			Classes: []Class{
				{
					Name: "Test",
					Defs: []Def{},
				},
			},
		},
	},

	{
		`
		class Test {
		}
		class Bar {}
		`,

		&File{
			Classes: []Class{
				{
					Name: "Test",
					Defs: []Def{},
				},
				{
					Name: "Bar",
					Defs: []Def{},
				},
			},
		},
	},

	{
		`
		class Test {
			$prop = 'x'
		}
		`,

		&File{
			Classes: []Class{
				{
					Name: "Test",
					Defs: []Def{
						{
							Name: "$prop",
							Val:  "'x'",
						},
					},
				},
			},
		},
	},

	{
		`
		class Test {
			$foo = 'bar',
			$baz = $foo
		}
		
		class Class2 {
			$good = 'text'
		}
		`,

		&File{
			Classes: []Class{
				{
					Name: "Test",
					Defs: []Def{
						{
							Name: "$foo",
							Val:  "'bar'",
						},

						{
							Name: "$baz",
							Val:  "$foo",
						},
					},
				},

				{
					Name: "Class2",
					Defs: []Def{
						{
							Name: "$good",
							Val:  "'text'",
						},
					},
				},
			},
		},
	},
}

func TestLex(t *testing.T) {
	for _, test := range lexTests {
		if file, err := Lex("test.manifest", strings.NewReader(test.manifest)); err != nil {
			t.Log(test.manifest)
			t.Fatal(err)
		} else {
			if !reflect.DeepEqual(file, test.ast) {
				js, _ := json.MarshalIndent(lastFile, "", "  ")
				t.Log(string(js))
				t.Error(test.manifest)
			}
		}
	}
}
