package manifest2

import (
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
			Classes: &Classes{
				Classes: []*Class{
					{
						Name: "Test",
						Defs: &Defs{},
					},
				},
			},
		},
	},
	{
		`
		class Test {
			foo = bar,
			baz = yup
		}
		
		class Class2 {
			good = text
		}
		`,

		&File{
			Classes: &Classes{
				Classes: []*Class{
					{
						Name: "Test",
						Defs: &Defs{
							Defs: []*Def{
								{
									Name: "foo",
									Val:  "bar",
								},

								{
									Name: "baz",
									Val:  "yup",
								},
							},
						},
					},

					{
						Name: "Class2",
						Defs: &Defs{
							Defs: []*Def{
								{
									Name: "good",
									Val:  "text",
								},
							},
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
			t.Fatal(err)
		} else {
			if !reflect.DeepEqual(file, test.ast) {
				t.Error(test.manifest)
			}
		}
	}
}
