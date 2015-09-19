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
					Name: "Test",
					Defs: []Def{},
					Declarations: []Declaration{
						{
							Type: "package",
							Name: "pkg3",
							Props: []Prop{
								{
									Name:  "depends",
									Value: Array{},
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
					Name: "Test",
					Defs: []Def{},
					Declarations: []Declaration{
						{
							Type: "package",
							Name: "pkg3",
							Props: []Prop{
								{
									Name: "depends",
									Value: Array{
										Reference{
											Type:   "package",
											Scalar: "pkg1",
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
					Name: "Test",
					Defs: []Def{},
					Declarations: []Declaration{
						{
							Type: "package",
							Name: "pkg3",
							Props: []Prop{
								{
									Name: "depends",
									Value: Array{
										Reference{
											Type:   "package",
											Scalar: "pkg1",
										},
										Reference{
											Type:   "package",
											Scalar: "pkg2",
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
					Name: "Arrays",
					Defs: []Def{
						{
							Name: "$a1",
							Val:  Array{},
						},
						{
							Name: "$a2",
							Val:  Array{"foo"},
						},
						{
							Name: "$a3",
							Val:  Array{"foo", "bar"},
						},
						{
							Name: "$a4",
							Val:  Array{Variable("$a1")},
						},
						{
							Name: "$a5",
							Val:  Array{Variable("$a1"), Variable("$a2")},
						},
						{
							Name: "$a6",
							Val:  Array{Variable("$a1"), "foo"},
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
					Name: "Test",
					Defs: []Def{
						{
							Name: "$webserver",
							Val:  "nginx",
						},
					},
					Declarations: []Declaration{
						{
							Type: "package",
							Name: "pkg3",
							Props: []Prop{
								{
									Name: "depends",
									Value: Array{
										Reference{
											Type:   "package",
											Scalar: "$webserver",
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

	//	{
	//		`class Test {
	//			package { 'pkg1':
	//		    	depends => [
	//		        - deb[pkg2]
	//		        - file[file1]

	//		deb[pkg2]:
	//		    depends:
	//		        - file[file1]
	//		        - file[file2]

	//		file[file1]:
	//		    depends:
	//		        - shell[cmd1]

	//		file[file2]:
	//		    depends:
	//		        - shell[cmd2]

	//		shell[cmd1]:
	//		    depends:
	//		        - shell[cmd2]

	//		shell[cmd2]:
	//		`

	//		&File{
	//			Classes: []Class{
	//				{
	//					Name: "Test",
	//					Defs: []Def{
	//						{
	//							Name: "$foo",
	//							Val:  "bar",
	//						},

	//						{
	//							Name: "$baz",
	//							Val:  "$foo",
	//						},
	//					},
	//					Declarations: []Declaration{},
	//				},

	//				{
	//					Name: "Class2",
	//					Defs: []Def{
	//						{
	//							Name: "$good",
	//							Val:  "text",
	//						},
	//					},
	//					Declarations: []Declaration{},
	//				},
	//			},
	//		},
	//	},
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
