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
					LineNum:      2,
					Name:         "Test",
					Defs:         []Def{},
					Declarations: []Declaration{},
				},
				{
					LineNum:      4,
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
					LineNum: 2,
					Name:    "Test",
					Defs: []Def{
						{
							LineNum: 3,
							Name:    Variable("$prop"),
							Val:     Value("x"),
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
					Defs: []Def{
						{
							LineNum: 3,
							Name:    Variable("$prop"),
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
					LineNum: 2,
					Name:    "Test",
					Defs:    []Def{},
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
					LineNum: 2,
					Name:    "Test",
					Defs:    []Def{},
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
					Defs: []Def{
						{
							LineNum: 3,
							Name:    "$foo",
							Val:     "bar",
						},

						{
							LineNum: 4,
							Name:    "$baz",
							Val:     "$foo",
						},
					},
					Declarations: []Declaration{},
				},

				{
					LineNum: 7,
					Name:    "Class2",
					Defs: []Def{
						{
							LineNum: 8,
							Name:    "$good",
							Val:     "text",
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
					LineNum: 2,
					Name:    "Test",
					Defs:    []Def{},
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
					LineNum: 2,
					Name:    "Test",
					Defs:    []Def{},
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
					LineNum: 2,
					Name:    "Test",
					Defs:    []Def{},
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
					Defs: []Def{
						{
							LineNum: 3,
							Name:    "$a1",
							Val:     Array{},
						},
						{
							LineNum: 4,
							Name:    "$a2",
							Val:     Array{"foo"},
						},
						{
							LineNum: 5,
							Name:    "$a3",
							Val:     Array{"foo", "bar"},
						},
						{
							LineNum: 6,
							Name:    "$a4",
							Val:     Array{Variable("$a1")},
						},
						{
							LineNum: 7,
							Name:    "$a5",
							Val:     Array{Variable("$a1"), Variable("$a2")},
						},
						{
							LineNum: 8,
							Name:    "$a6",
							Val:     Array{Variable("$a1"), "foo"},
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
					Defs: []Def{
						{
							LineNum: 3,
							Name:    "$webserver",
							Val:     "nginx",
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
					Defs: []Def{
						{
							LineNum: 3,
							Name:    Variable("$webserver"),
							Val:     "nginx",
						},
					},
					Declarations: []Declaration{
						{
							LineNum: 5,
							Type:    "package",
							Scalar:  Variable("$webserver"),
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
										Scalar:  Variable("$webserver"),
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
					LineNum: 2,
					Name:    "Deps",
					Defs:    []Def{},
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
