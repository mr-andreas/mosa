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
					Name:         "Test",
					Defs:         []Def{},
					Declarations: []Declaration{},
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
					Name:         "Test",
					Defs:         []Def{},
					Declarations: []Declaration{},
				},
				{
					Name:         "Bar",
					Defs:         []Def{},
					Declarations: []Declaration{},
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
							Name: Variable("$prop"),
							Val:  Value("x"),
						},
					},
					Declarations: []Declaration{},
				},
			},
		},
	},

	{
		`
		class Test {
			package { 'pkg-name':
			}
		}
		`,

		&File{
			Classes: []Class{
				{
					Name: "Test",
					Defs: []Def{},
					Declarations: []Declaration{
						{
							Type:  "package",
							Name:  "pkg-name",
							Props: []Prop{},
						},
					},
				},
			},
		},
	},

	{
		`
		class Test {
			package { 'pkg':
				foo => 'bar',
			}
		}
		`,

		&File{
			Classes: []Class{
				{
					Name: "Test",
					Defs: []Def{},
					Declarations: []Declaration{
						{
							Type: "package",
							Name: "pkg",
							Props: []Prop{
								{
									Name:  "foo",
									Value: "bar",
								},
							},
						},
					},
				},
			},
		},
	},

	{
		`
		class Test {
			$foo = 'bar'
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
							Val:  "bar",
						},

						{
							Name: "$baz",
							Val:  "$foo",
						},
					},
					Declarations: []Declaration{},
				},

				{
					Name: "Class2",
					Defs: []Def{
						{
							Name: "$good",
							Val:  "text",
						},
					},
					Declarations: []Declaration{},
				},
			},
		},
	},
}

func TestLex(t *testing.T) {
	for _, test := range lexTests {
		if file, err := Lex("test.manifest", strings.NewReader(test.manifest)); err != nil {
			t.Log(test.manifest)
			t.Error(err)
		} else {
			if !equalsAsJson(file, test.ast) {
				t.Logf("%#v", test.ast)
				t.Logf("%#v", lastFile)
				js2, _ := json.MarshalIndent(test.ast, "", "  ")
				t.Log(string(js2))
				js, _ := json.MarshalIndent(lastFile, "", "  ")
				t.Log(string(js))
				t.Fatal(test.manifest)
			}
		}
	}
}

func equalsAsJson(i1, i2 interface{}) bool {
	j1, err1 := json.Marshal(i1)
	j2, err2 := json.Marshal(i2)

	if err1 != nil || err2 != nil {
		return false
	}

	var unserialized1, unserialized2 interface{}
	json.Unmarshal(j1, &unserialized1)
	json.Unmarshal(j2, &unserialized2)

	return reflect.DeepEqual(unserialized1, unserialized2)
}
