package manifest

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

var lexTests = []struct {
	manifest string
	ast      *AST
}{
	{
		`class Test {}`,

		&AST{
			Classes: []Class{
				{
					LineNum:      1,
					Name:         "Test",
					ArgDefs:      []VariableDef{},
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

		&AST{
			Classes: []Class{
				{
					LineNum:      2,
					Name:         "Test",
					ArgDefs:      []VariableDef{},
					VariableDefs: []VariableDef{},
					Declarations: []Declaration{},
				},
				{
					LineNum:      4,
					Name:         "Bar",
					ArgDefs:      []VariableDef{},
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

		&AST{
			Classes: []Class{
				{
					LineNum: 2,
					Name:    "Test",
					ArgDefs: []VariableDef{},
					VariableDefs: []VariableDef{
						{
							LineNum:      3,
							VariableName: VariableName{3, "$prop"},
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

		&AST{
			Classes: []Class{
				{
					LineNum: 2,
					Name:    "Test",
					ArgDefs: []VariableDef{},
					VariableDefs: []VariableDef{
						{
							LineNum:      3,
							VariableName: VariableName{3, "$prop"},
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
		// Comment
		class Test { // Another comment
			$prop = 'x' // Comment here
			// And another comment
		}
		`,

		&AST{
			Classes: []Class{
				{
					LineNum: 3,
					Name:    "Test",
					ArgDefs: []VariableDef{},
					VariableDefs: []VariableDef{
						{
							LineNum:      4,
							VariableName: VariableName{4, "$prop"},
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
			package { 'pkg-name':
			}
		}
		`,

		&AST{
			Classes: []Class{
				{
					LineNum:      2,
					Name:         "Test",
					ArgDefs:      []VariableDef{},
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

		&AST{
			Classes: []Class{
				{
					LineNum:      2,
					Name:         "Test",
					ArgDefs:      []VariableDef{},
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

		&AST{
			Classes: []Class{
				{
					LineNum:      1,
					Name:         "Test",
					ArgDefs:      []VariableDef{},
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

		&AST{
			Classes: []Class{
				{
					LineNum: 2,
					Name:    "Test",
					ArgDefs: []VariableDef{},
					VariableDefs: []VariableDef{
						{
							LineNum:      3,
							VariableName: VariableName{3, "$foo"},
							Val:          "bar",
						},

						{
							LineNum:      4,
							VariableName: VariableName{4, "$baz"},
							Val:          VariableName{4, "$foo"},
						},
					},
					Declarations: []Declaration{},
				},

				{
					LineNum: 7,
					Name:    "Class2",
					ArgDefs: []VariableDef{},
					VariableDefs: []VariableDef{
						{
							LineNum:      8,
							VariableName: VariableName{8, "$good"},
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

		&AST{
			Classes: []Class{
				{
					LineNum:      2,
					Name:         "Test",
					ArgDefs:      []VariableDef{},
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

		&AST{
			Classes: []Class{
				{
					LineNum:      2,
					Name:         "Test",
					ArgDefs:      []VariableDef{},
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

		&AST{
			Classes: []Class{
				{
					LineNum:      2,
					Name:         "Test",
					ArgDefs:      []VariableDef{},
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

		&AST{
			Classes: []Class{
				{
					LineNum: 2,
					Name:    "Arrays",
					ArgDefs: []VariableDef{},
					VariableDefs: []VariableDef{
						{
							LineNum:      3,
							VariableName: VariableName{3, "$a1"},
							Val:          Array{},
						},
						{
							LineNum:      4,
							VariableName: VariableName{4, "$a2"},
							Val:          Array{"foo"},
						},
						{
							LineNum:      5,
							VariableName: VariableName{5, "$a3"},
							Val:          Array{"foo", "bar"},
						},
						{
							LineNum:      6,
							VariableName: VariableName{6, "$a4"},
							Val:          Array{VariableName{6, "$a1"}},
						},
						{
							LineNum:      7,
							VariableName: VariableName{7, "$a5"},
							Val: Array{
								VariableName{7, "$a1"}, VariableName{7, "$a2"},
							},
						},
						{
							LineNum:      8,
							VariableName: VariableName{8, "$a6"},
							Val:          Array{VariableName{8, "$a1"}, "foo"},
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

		&AST{
			Classes: []Class{
				{
					LineNum: 2,
					Name:    "Test",
					ArgDefs: []VariableDef{},
					VariableDefs: []VariableDef{
						{
							LineNum:      3,
							VariableName: VariableName{3, "$webserver"},
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
											Scalar:  VariableName{5, "$webserver"},
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

		&AST{
			Classes: []Class{
				{
					LineNum: 2,
					Name:    "Test",
					ArgDefs: []VariableDef{},
					VariableDefs: []VariableDef{
						{
							LineNum:      3,
							VariableName: VariableName{3, "$webserver"},
							Val:          "nginx",
						},
					},
					Declarations: []Declaration{
						{
							LineNum: 5,
							Type:    "package",
							Scalar:  VariableName{5, "$webserver"},
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
										Scalar:  VariableName{10, "$webserver"},
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

		&AST{
			Classes: []Class{
				{
					LineNum:      2,
					Name:         "Deps",
					ArgDefs:      []VariableDef{},
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
		`define multiple package($names,) {
			$foo = 'x'
		}`,
		&AST{
			Defines: []Define{
				{
					Name:    "package",
					LineNum: 1,
					ArgDefs: []VariableDef{
						{
							LineNum:      1,
							Val:          nil,
							VariableName: VariableName{1, "$names"},
						},
					},
					Type: DefineTypeMultiple,
					VariableDefs: []VariableDef{
						{
							LineNum:      2,
							VariableName: VariableName{2, "$foo"},
							Val:          "x",
						},
					},
					Declarations: []Declaration{},
				},
			},
		},
	},

	{
		`define multiple package($names,) {}`,
		&AST{
			Defines: []Define{
				{
					Name:    "package",
					LineNum: 1,
					ArgDefs: []VariableDef{
						{
							LineNum:      1,
							Val:          nil,
							VariableName: VariableName{1, "$names"},
						},
					},
					Type:         DefineTypeMultiple,
					VariableDefs: []VariableDef{},
					Declarations: []Declaration{},
				},
			},
		},
	},

	{
		`node 'localhost' {
			$foo = 'x'
		}`,
		&AST{
			Nodes: []Node{
				{
					Name:    "localhost",
					LineNum: 1,
					VariableDefs: []VariableDef{
						{
							LineNum:      2,
							VariableName: VariableName{2, "$foo"},
							Val:          "x",
						},
					},
					Declarations: []Declaration{},
				},
			},
		},
	},

	{
		`node 'localhost' {}`,
		&AST{
			Nodes: []Node{
				{
					Name:         "localhost",
					LineNum:      1,
					VariableDefs: []VariableDef{},
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
		&AST{
			Nodes: []Node{
				{
					Name:    "localhost",
					LineNum: 1,
					VariableDefs: []VariableDef{
						{
							LineNum:      2,
							VariableName: VariableName{2, "$foo"},
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

		&AST{
			Classes: []Class{
				{
					LineNum:      1,
					Name:         "Test",
					ArgDefs:      []VariableDef{},
					VariableDefs: []VariableDef{},
					Declarations: []Declaration{},
				},
			},
		},
	},

	{
		`class Test($foo,) {}`,

		&AST{
			Classes: []Class{
				{
					LineNum: 1,
					Name:    "Test",
					ArgDefs: []VariableDef{
						{
							LineNum:      1,
							Val:          nil,
							VariableName: VariableName{1, "$foo"},
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

		&AST{
			Classes: []Class{
				{
					LineNum: 1,
					Name:    "Test",
					ArgDefs: []VariableDef{
						{
							LineNum:      2,
							VariableName: VariableName{2, "$foo"},
							Val:          nil,
						},
						{
							LineNum:      3,
							VariableName: VariableName{3, "$bar"},
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

		&AST{
			Classes: []Class{
				{
					LineNum: 1,
					Name:    "Test",
					ArgDefs: []VariableDef{
						{
							LineNum:      1,
							VariableName: VariableName{1, "$foo"},
							Val:          5,
						},
						{
							LineNum:      1,
							VariableName: VariableName{1, "$bar"},
							Val:          "x",
						},
						{
							LineNum:      1,
							VariableName: VariableName{1, "$baz"},
							Val:          Array{1, 2},
						},
						{
							LineNum:      1,
							VariableName: VariableName{1, "$a"},
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
		`
		// InterpolatedString
		class Test {
			$foo = "string"
		}`,

		&AST{
			Classes: []Class{
				{
					LineNum: 3,
					Name:    "Test",
					ArgDefs: []VariableDef{},
					VariableDefs: []VariableDef{
						{
							LineNum:      4,
							VariableName: VariableName{4, "$foo"},
							Val: InterpolatedString{
								LineNum:  4,
								Segments: []interface{}{"string"},
							},
						},
					},
					Declarations: []Declaration{},
				},
			},
		},
	},

	{
		`class Test($foo = "/home/$bar",) {}`,

		&AST{
			Classes: []Class{
				{
					LineNum: 1,
					Name:    "Test",
					ArgDefs: []VariableDef{
						{
							LineNum:      1,
							VariableName: VariableName{1, "$foo"},
							Val: InterpolatedString{
								LineNum: 1,
								Segments: []interface{}{
									"/home/",
									VariableName{
										LineNum: 1,
										Str:     "$bar",
									},
								},
							},
						},
					},
					VariableDefs: []VariableDef{},
					Declarations: []Declaration{},
				},
			},
		},
	},

	{
		`class Test {
			$a = "${b}"
		}`,

		&AST{
			Classes: []Class{
				{
					LineNum: 1,
					Name:    "Test",
					ArgDefs: []VariableDef{},
					VariableDefs: []VariableDef{
						{
							LineNum:      2,
							VariableName: VariableName{2, "$a"},
							Val: InterpolatedString{
								LineNum: 2,
								Segments: []interface{}{
									VariableName{LineNum: 2, Str: "$b"},
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
		// Different interpolated strings
		class T {
			$a = ""
			$b = "$foo$bar"
			$c = "$multi
			$line"
			$d = "${foo}bar"
			$e = "bar${foo}"
			$f = "bar{baz}"
			$g = "bar{ba$z}"
			$h = "bar{${foo}}"
			$i = "bar${{foo}}"
			$j = "bar{{$foo}}"
		}`,

		&AST{
			Classes: []Class{
				{
					LineNum: 3,
					Name:    "T",
					ArgDefs: []VariableDef{},
					VariableDefs: []VariableDef{
						{
							LineNum:      4,
							VariableName: VariableName{4, "$a"},
							Val: InterpolatedString{
								LineNum:  4,
								Segments: nil,
							},
						},
						{
							LineNum:      5,
							VariableName: VariableName{5, "$b"},
							Val: InterpolatedString{
								LineNum: 5,
								Segments: []interface{}{
									VariableName{LineNum: 5, Str: "$foo"},
									VariableName{LineNum: 5, Str: "$bar"},
								},
							},
						},
						{
							LineNum:      6,
							VariableName: VariableName{6, "$c"},
							Val: InterpolatedString{
								LineNum: 6,
								Segments: []interface{}{
									VariableName{LineNum: 6, Str: "$multi"},
									string("\n\t\t\t"),
									VariableName{LineNum: 7, Str: "$line"},
								},
							},
						},
						{
							LineNum:      8,
							VariableName: VariableName{8, "$d"},
							Val: InterpolatedString{
								LineNum: 8,
								Segments: []interface{}{
									VariableName{LineNum: 8, Str: "$foo"},
									"bar",
								},
							},
						},
						{
							LineNum:      9,
							VariableName: VariableName{9, "$e"},
							Val: InterpolatedString{
								LineNum: 9,
								Segments: []interface{}{
									"bar",
									VariableName{LineNum: 9, Str: "$foo"},
								},
							},
						},
						{
							LineNum:      10,
							VariableName: VariableName{10, "$f"},
							Val: InterpolatedString{
								LineNum:  10,
								Segments: []interface{}{"bar{baz}"},
							},
						},
						{
							LineNum:      11,
							VariableName: VariableName{11, "$g"},
							Val: InterpolatedString{
								LineNum: 11,
								Segments: []interface{}{
									"bar{ba",
									VariableName{LineNum: 11, Str: "$z"},
									"}",
								},
							},
						},
						{
							LineNum:      12,
							VariableName: VariableName{12, "$h"},
							Val: InterpolatedString{
								LineNum: 12,
								Segments: []interface{}{
									"bar{",
									VariableName{LineNum: 12, Str: "$foo"},
									"}",
								},
							},
						},
						{
							LineNum:      13,
							VariableName: VariableName{13, "$i"},
							Val: InterpolatedString{
								LineNum:  13,
								Segments: []interface{}{"bar", "$", "{{foo}}"},
							},
						},
						{
							LineNum:      14,
							VariableName: VariableName{14, "$j"},
							Val: InterpolatedString{
								LineNum: 14,
								Segments: []interface{}{
									"bar{{",
									VariableName{LineNum: 14, Str: "$foo"},
									"}}",
								},
							},
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
		for i, _ := range test.ast.Classes {
			test.ast.Classes[i].Filename = "test.manifest"
		}
		for i, _ := range test.ast.Defines {
			test.ast.Defines[i].Filename = "test.manifest"
		}
		for i, _ := range test.ast.Nodes {
			test.ast.Nodes[i].Filename = "test.manifest"
		}

		ast := NewAST()
		if err := Lex(ast, "test.manifest", strings.NewReader(test.manifest)); err != nil {
			t.Log(test.manifest)
			t.Error(err)
		} else {
			if !equalsAsJson(ast, test.ast) {
				t.Logf("%#v", test.ast)
				t.Logf("%#v", ast)
				js2, _ := json.MarshalIndent(test.ast, "", "  ")
				t.Log(string(js2))
				js, _ := json.MarshalIndent(ast, "", "  ")
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
		ast := NewAST()
		if err := Lex(ast, "test.manifest", strings.NewReader(test.manifest)); err == nil {
			t.Error("Bad manifest didn't fail:", test.manifest)
		}
	}
}

// Makes sure that a second call to yyparse() does not return the error of a
// previous run.
func TestParseGoodAfterBad(t *testing.T) {
	goods := []string{
		"",
		"define single a(){}",
		"class x{}",
		"define multiple a(){}",
		"node 'x'{}",
	}

	for _, good := range goods {
		// Parse a bad grammar file
		ast := NewAST()
		if err := Lex(ast, "bad.ms", strings.NewReader("node{}")); err == nil {
			t.Fatal("Bad grammar parsed")
		}

		// Now parse valid grammar and make sure we don't get an error
		ast2 := NewAST()
		if err := Lex(ast2, "bad.ms", strings.NewReader(good)); err != nil {
			t.Log(good)
			t.Error("Got error when parsing good grammar:", err)
		}
	}
}

func TestParseMultipleFiles(t *testing.T) {
	testMs := `
		node 'n' { 
			class { 'A': }
		}`
	test2Ms := `
		class A { 
			exec { 'ls': }
		}`

	ast := NewAST()
	expectedAst := &AST{
		Nodes: []Node{
			{
				Filename:     "test.ms",
				LineNum:      2,
				Name:         "n",
				VariableDefs: []VariableDef{},
				Declarations: []Declaration{
					{
						LineNum: 3,
						Type:    "class",
						Scalar:  QuotedString("A"),
						Props:   []Prop{},
					},
				},
			},
		},
		Classes: []Class{
			{
				Filename:     "test2.ms",
				LineNum:      2,
				Name:         "A",
				ArgDefs:      []VariableDef{},
				VariableDefs: []VariableDef{},
				Declarations: []Declaration{
					{
						LineNum: 3,
						Type:    "exec",
						Scalar:  QuotedString("ls"),
						Props:   []Prop{},
					},
				},
			},
		},
	}

	if err := Lex(ast, "test.ms", strings.NewReader(testMs)); err != nil {
		t.Log(testMs)
		t.Fatal(err)
	}
	if err := Lex(ast, "test2.ms", strings.NewReader(test2Ms)); err != nil {
		t.Log(test2Ms)
		t.Fatal(err)
	}

	if !equalsAsJson(ast, expectedAst) {
		t.Logf("%#v", expectedAst)
		t.Logf("%#v", ast)
		js2, _ := json.MarshalIndent(expectedAst, "", "  ")
		t.Log(string(js2))
		js, _ := json.MarshalIndent(ast, "", "  ")
		t.Log(string(js))
		t.Fail()
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
