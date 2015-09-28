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
		`class Test {}`,

		&File{
			Classes: []Class{
				{
					LineNum:      1,
					Name:         "Test",
					ArgDefs:      []ArgDef{},
					VariableDefs: []VariableDef{},
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
					LineNum:      2,
					Name:         "Test",
					ArgDefs:      []ArgDef{},
					VariableDefs: []VariableDef{},
					Declarations: []Declaration{},
				},
				{
					LineNum:      4,
					Name:         "Bar",
					ArgDefs:      []ArgDef{},
					VariableDefs: []VariableDef{},
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
					LineNum: 2,
					Name:    "Test",
					ArgDefs: []ArgDef{},
					VariableDefs: []VariableDef{
						{
							LineNum:      3,
							VariableName: VariableName("$prop"),
							Val:          Value("x"),
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
			$prop = [ 'x', 1, [ 'y', ], ]
		}
		`,

		&File{
			Classes: []Class{
				{
					LineNum: 2,
					Name:    "Test",
					ArgDefs: []ArgDef{},
					VariableDefs: []VariableDef{
						{
							LineNum:      3,
							VariableName: VariableName("$prop"),
							Val: Array{
								Value("x"),
								Value(1),
								Array{
									Value("y"),
								},
							},
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
					LineNum:      2,
					Name:         "Test",
					ArgDefs:      []ArgDef{},
					VariableDefs: []VariableDef{},
					Declarations: []Declaration{
						{
							LineNum: 3,
							Type:    "package",
							Scalar:  "pkg-name",
							Props:   []Prop{},
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
					LineNum:      2,
					Name:         "Test",
					ArgDefs:      []ArgDef{},
					VariableDefs: []VariableDef{},
					Declarations: []Declaration{
						{
							LineNum: 3,
							Type:    "package",
							Scalar:  "pkg",
							Props: []Prop{
								{
									LineNum: 4,
									Name:    "foo",
									Value:   "bar",
								},
							},
						},
					},
				},
			},
		},
	},

	{
		`class Test {
			package { 'pkg':
				class => 'foo',
				define => 'bar',
				node => 'baz',
			}
		}`,

		&File{
			Classes: []Class{
				{
					LineNum:      1,
					Name:         "Test",
					ArgDefs:      []ArgDef{},
					VariableDefs: []VariableDef{},
					Declarations: []Declaration{
						{
							LineNum: 2,
							Type:    "package",
							Scalar:  "pkg",
							Props: []Prop{
								{
									LineNum: 3,
									Name:    "class",
									Value:   "foo",
								},
								{
									LineNum: 4,
									Name:    "define",
									Value:   "bar",
								},
								{
									LineNum: 5,
									Name:    "node",
									Value:   "baz",
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
					LineNum: 2,
					Name:    "Test",
					ArgDefs: []ArgDef{},
					VariableDefs: []VariableDef{
						{
							LineNum:      3,
							VariableName: "$foo",
							Val:          "bar",
						},

						{
							LineNum:      4,
							VariableName: "$baz",
							Val:          "$foo",
						},
					},
					Declarations: []Declaration{},
				},

				{
					LineNum: 7,
					Name:    "Class2",
					ArgDefs: []ArgDef{},
					VariableDefs: []VariableDef{
						{
							LineNum:      8,
							VariableName: "$good",
							Val:          "text",
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
			package { 'pkg3':
		    	depends => [],
			}
		}
		`,

		&File{
			Classes: []Class{
				{
					LineNum:      2,
					Name:         "Test",
					ArgDefs:      []ArgDef{},
					VariableDefs: []VariableDef{},
					Declarations: []Declaration{
						{
							LineNum: 3,
							Type:    "package",
							Scalar:  "pkg3",
							Props: []Prop{
								{
									LineNum: 4,
									Name:    "depends",
									Value:   Array{},
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
			package { 'pkg3':
		    	depends => [ package['pkg1'], ],
			}
		}
		`,

		&File{
			Classes: []Class{
				{
					LineNum:      2,
					Name:         "Test",
					ArgDefs:      []ArgDef{},
					VariableDefs: []VariableDef{},
					Declarations: []Declaration{
						{
							LineNum: 3,
							Type:    "package",
							Scalar:  "pkg3",
							Props: []Prop{
								{
									LineNum: 4,
									Name:    "depends",
									Value: Array{
										Reference{
											LineNum: 4,
											Type:    "package",
											Scalar:  "pkg1",
										},
									},
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
			package { 'pkg3':
		    	depends => [
					package['pkg1'], 
					package['pkg2'],
				],
			}
		}
		`,

		&File{
			Classes: []Class{
				{
					LineNum:      2,
					Name:         "Test",
					ArgDefs:      []ArgDef{},
					VariableDefs: []VariableDef{},
					Declarations: []Declaration{
						{
							LineNum: 3,
							Type:    "package",
							Scalar:  "pkg3",
							Props: []Prop{
								{
									LineNum: 4,
									Name:    "depends",
									Value: Array{
										Reference{
											LineNum: 5,
											Type:    "package",
											Scalar:  "pkg1",
										},
										Reference{
											LineNum: 6,
											Type:    "package",
											Scalar:  "pkg2",
										},
									},
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
		class Arrays {
			$a1 = []
			$a2 = [ 'foo', ]
			$a3 = [ 'foo', 'bar', ]
			$a4 = [ $a1, ]
			$a5 = [ $a1, $a2, ]
			$a6 = [ $a1, 'foo', ]			
		}
		`,

		&File{
			Classes: []Class{
				{
					LineNum: 2,
					Name:    "Arrays",
					ArgDefs: []ArgDef{},
					VariableDefs: []VariableDef{
						{
							LineNum:      3,
							VariableName: "$a1",
							Val:          Array{},
						},
						{
							LineNum:      4,
							VariableName: "$a2",
							Val:          Array{"foo"},
						},
						{
							LineNum:      5,
							VariableName: "$a3",
							Val:          Array{"foo", "bar"},
						},
						{
							LineNum:      6,
							VariableName: "$a4",
							Val:          Array{VariableName("$a1")},
						},
						{
							LineNum:      7,
							VariableName: "$a5",
							Val:          Array{VariableName("$a1"), VariableName("$a2")},
						},
						{
							LineNum:      8,
							VariableName: "$a6",
							Val:          Array{VariableName("$a1"), "foo"},
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
			$webserver = 'nginx'
			package { 'pkg3':
		    	depends => [ package[$webserver], ],
			}
		}
		`,

		&File{
			Classes: []Class{
				{
					LineNum: 2,
					Name:    "Test",
					ArgDefs: []ArgDef{},
					VariableDefs: []VariableDef{
						{
							LineNum:      3,
							VariableName: "$webserver",
							Val:          "nginx",
						},
					},
					Declarations: []Declaration{
						{
							LineNum: 4,
							Type:    "package",
							Scalar:  "pkg3",
							Props: []Prop{
								{
									Name:    "depends",
									LineNum: 5,
									Value: Array{
										Reference{
											LineNum: 5,
											Type:    "package",
											Scalar:  "$webserver",
										},
									},
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
			$webserver = 'nginx'
			
			package { $webserver:
		    	ensure => 'latest',
			}
			
			package { 'website':
				depends => package[$webserver],
			}
		}
		`,

		&File{
			Classes: []Class{
				{
					LineNum: 2,
					Name:    "Test",
					ArgDefs: []ArgDef{},
					VariableDefs: []VariableDef{
						{
							LineNum:      3,
							VariableName: VariableName("$webserver"),
							Val:          "nginx",
						},
					},
					Declarations: []Declaration{
						{
							LineNum: 5,
							Type:    "package",
							Scalar:  VariableName("$webserver"),
							Props: []Prop{
								{
									LineNum: 6,
									Name:    "ensure",
									Value:   "latest",
								},
							},
						},
						{
							LineNum: 9,
							Type:    "package",
							Scalar:  "website",
							Props: []Prop{
								{
									LineNum: 10,
									Name:    "depends",
									Value: Reference{
										LineNum: 10,
										Type:    "package",
										Scalar:  VariableName("$webserver"),
									},
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
		class Deps {
			package { 'pkg1':
		    	depends => [ deb['pkg2'], file['file1'], ],
			}

			package { 'pkg2':
			    depends => [ file['file1'], file['file2'], ],
			}

			file { 'file1':
			    depends => shell['cmd1'],
			}

			file { 'file2':
			    depends => shell['cmd2'],
			}

			shell { 'cmd1':
			    depends => shell['cmd2'],
			}

			shell { 'cmd2': }
		}
		`,

		&File{
			Classes: []Class{
				{
					LineNum:      2,
					Name:         "Deps",
					ArgDefs:      []ArgDef{},
					VariableDefs: []VariableDef{},
					Declarations: []Declaration{
						{
							LineNum: 3,
							Type:    "package",
							Scalar:  "pkg1",
							Props: []Prop{
								{
									LineNum: 4,
									Name:    "depends",
									Value: Array{
										Reference{4, "deb", "pkg2"},
										Reference{4, "file", "file1"},
									},
								},
							},
						},
						{
							LineNum: 7,
							Type:    "package",
							Scalar:  "pkg2",
							Props: []Prop{
								{
									LineNum: 8,
									Name:    "depends",
									Value: Array{
										Reference{8, "file", "file1"},
										Reference{8, "file", "file2"},
									},
								},
							},
						},
						{
							LineNum: 11,
							Type:    "file",
							Scalar:  "file1",
							Props: []Prop{
								{
									LineNum: 12,
									Name:    "depends",
									Value:   Reference{12, "shell", "cmd1"},
								},
							},
						},
						{
							LineNum: 15,
							Type:    "file",
							Scalar:  "file2",
							Props: []Prop{
								{
									LineNum: 16,
									Name:    "depends",
									Value:   Reference{16, "shell", "cmd2"},
								},
							},
						},
						{
							LineNum: 19,
							Type:    "shell",
							Scalar:  "cmd1",
							Props: []Prop{
								{
									LineNum: 20,
									Name:    "depends",
									Value:   Reference{20, "shell", "cmd2"},
								},
							},
						},
						{
							LineNum: 23,
							Type:    "shell",
							Scalar:  "cmd2",
							Props:   []Prop{},
						},
					},
				},
			},
		},
	},

	{
		`define multiple package {
			$foo = 'x'
		}`,
		&File{
			Defines: []Define{
				{
					Name:    "package",
					LineNum: 1,
					Type:    DefineTypeMultiple,
				},
			},
		},
	},

	{
		`node 'localhost' {
			$foo = 'x'
		}`,
		&File{
			Nodes: []Node{
				{
					Name:    "localhost",
					LineNum: 1,
					VariableDefs: []VariableDef{
						{
							LineNum:      2,
							VariableName: VariableName("$foo"),
							Val:          "x",
						},
					},
					Declarations: []Declaration{},
				},
			},
		},
	},

	{
		`node 'localhost' {
			$foo = 'x'
			
			decl { 'x': foo => 5, }
		}`,
		&File{
			Nodes: []Node{
				{
					Name:    "localhost",
					LineNum: 1,
					VariableDefs: []VariableDef{
						{
							LineNum:      2,
							VariableName: VariableName("$foo"),
							Val:          "x",
						},
					},
					Declarations: []Declaration{
						{
							LineNum: 4,
							Type:    "decl",
							Scalar:  "x",
							Props: []Prop{
								{
									LineNum: 4,
									Name:    "foo",
									Value:   5,
								},
							},
						},
					},
				},
			},
		},
	},

	{
		`class Test() {}`,

		&File{
			Classes: []Class{
				{
					LineNum:      1,
					Name:         "Test",
					ArgDefs:      []ArgDef{},
					VariableDefs: []VariableDef{},
					Declarations: []Declaration{},
				},
			},
		},
	},

	{
		`class Test($foo,) {}`,

		&File{
			Classes: []Class{
				{
					LineNum: 1,
					Name:    "Test",
					ArgDefs: []ArgDef{
						{
							LineNum:      1,
							VariableName: "$foo",
							Val:          nil,
						},
					},
					VariableDefs: []VariableDef{},
					Declarations: []Declaration{},
				},
			},
		},
	},

	{
		`class Test(
		$foo, 
		$bar,
		) {}`,

		&File{
			Classes: []Class{
				{
					LineNum: 1,
					Name:    "Test",
					ArgDefs: []ArgDef{
						{
							LineNum:      2,
							VariableName: "$foo",
							Val:          nil,
						},
						{
							LineNum:      3,
							VariableName: "$bar",
							Val:          nil,
						},
					},
					VariableDefs: []VariableDef{},
					Declarations: []Declaration{},
				},
			},
		},
	},

	{
		`class Test($foo = 5, $bar = 'x', $baz = [ 1, 2, ], $a,) {}`,

		&File{
			Classes: []Class{
				{
					LineNum: 1,
					Name:    "Test",
					ArgDefs: []ArgDef{
						{
							LineNum:      1,
							VariableName: "$foo",
							Val:          5,
						},
						{
							LineNum:      1,
							VariableName: "$bar",
							Val:          "x",
						},
						{
							LineNum:      1,
							VariableName: "$baz",
							Val:          Array{1, 2},
						},
						{
							LineNum:      1,
							VariableName: "$a",
							Val:          nil,
						},
					},
					VariableDefs: []VariableDef{},
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

var badLexTests = []struct {
	manifest string
}{
	{`class`},
	{`class foo`},
	{`class foo {`},
	{`class foo }`},
	{`class bar {}}`},
	{`foo`},
	{`define package {}`},
	{`define foobar package {}`},
	{`define foobar package { $x = 5 }`},
	{`define single multiple package {}`},
	{`define single multiple package { $x = 5 }`},
	{`define multiple package {}`},
	{`define multiple package($nonamevar) {}`},
	{`node {}`},
	{`node badname {}`},
}

func TestBadLex(t *testing.T) {
	for _, test := range badLexTests {
		if _, err := Lex("test.manifest", strings.NewReader(test.manifest)); err == nil {
			t.Error("Bad manifest didn't fail:", test.manifest)
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
