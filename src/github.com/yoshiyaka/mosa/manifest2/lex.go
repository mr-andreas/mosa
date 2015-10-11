package manifest2

// #cgo LDFLAGS: -lfl
// typedef struct {
//   int code;
//   const char *error;
//   int line;
// } t_error;
// extern t_error doparse(char *);
// extern int line_num;
//
// #include "types.h"
import "C"
import (
	"fmt"
	"io"
	"io/ioutil"
)

var (
	ht       handleTable
	lastFile *File
)

type stringable interface {
	String() string
}

type File struct {
	Classes []Class
	Defines []Define
	Nodes   []Node
}

func (f *File) String() string {
	s := ""
	for _, class := range f.Classes {
		s = class.String() + "\n"
	}

	return s
}

type DefineType int

const (
	DefineTypeSingle DefineType = iota
	DefineTypeMultiple
)

type Define struct {
	LineNum int
	Name    string
	Type    DefineType
}

type Node struct {
	LineNum      int
	Name         string
	ArgDefs      []ArgDef
	VariableDefs []VariableDef
	Declarations []Declaration
}

type Class struct {
	LineNum      int
	Name         string
	ArgDefs      []ArgDef
	VariableDefs []VariableDef
	Declarations []Declaration
}

func (n *Node) String() string {
	defs := ""
	decls := ""
	for _, def := range n.VariableDefs {
		defs += fmt.Sprintf("\t%s\n", def.String())
	}

	for _, decl := range n.Declarations {
		decls += fmt.Sprintf("\t%s\n", decl.String())
	}

	return fmt.Sprintf("node '%s' {\n%s\n%s\n}\n", n.Name, defs, decls)
}

func (c *Class) String() string {
	defs := ""
	decls := ""
	for _, def := range c.VariableDefs {
		defs += fmt.Sprintf("\t\t%s\n", def.String())
	}

	for _, decl := range c.Declarations {
		decls += fmt.Sprintf("\t\t%s\n", decl.String())
	}

	return fmt.Sprintf("\tclass %s {\n%s\n%s\n\t}\n", c.Name, defs, decls)
}

type ArgDef struct {
	LineNum      int
	VariableName VariableName
	Val          Value
}

func (d *ArgDef) String() string {
	return fmt.Sprintf("%s = %s", d.VariableName, valToStr(d.Val))
}

type VariableDef struct {
	LineNum      int
	VariableName VariableName
	Val          Value
}

func (d *VariableDef) String() string {
	return fmt.Sprintf("%s = %s", d.VariableName, valToStr(d.Val))
}

func valToStr(i interface{}) string {
	switch i.(type) {
	case int, int64:
		return fmt.Sprintf("%d", i)
	case stringable:
		return i.(stringable).String()
	default:
		return i.(string)
	}
}

type QuotedString string

func (qs QuotedString) String() string { return fmt.Sprintf("'%s'", string(qs)) }

type VariableName struct {
	LineNum int
	Str     string
}

func (vn VariableName) String() string { return vn.Str }

// A used type, for instance package { 'nginx': ensure => 'latest' }
type Declaration struct {
	LineNum int

	// The type of declaration, 'package' in the example above
	Type string

	// The name of the declaration, 'nginx' in the example above
	Scalar Value

	// All properties for the declaration, ensure => 'latest' in the example
	// above.
	Props []Prop
}

func (d *Declaration) String() string {
	props := ""
	for _, prop := range d.Props {
		props += fmt.Sprintf("\t\t\t%s\n", prop.String())
	}

	return fmt.Sprintf("\t\t%s { %s:\n%s\n\t\t}\n", d.Type, d.Scalar, props)
}

// A property in declaration, for instance ensure => 'latest'
type Prop struct {
	LineNum int
	Name    string
	Value   Value
}

func (p *Prop) String() string {
	return fmt.Sprintf("%s => %s", p.Name, p.Value)
}

// A value, for instance 1, 'foo', $bar or [ 1, 'five', ]
type Value interface{}

// An array of strings, number or references, for instance
// [ 1, 'foo', package[$webserver], ]
type Array []interface{}

func (a Array) String() string {
	str := "["
	for _, val := range a {
		str += fmt.Sprintf(" %s,", valToStr(val))
	}
	str += " ]"

	return str
}

// A reference, for instance package['nginx'] or package[$webserver]
type Reference struct {
	LineNum int
	Type    string
	Scalar  Value
}

//export NilArray
func NilArray(typ C.ASTTYPE) goHandle {
	switch typ {
	case C.ASTTYPE_DEFS:
		return ht.Add([]interface{}{})
	case C.ASTTYPE_CLASSES:
		return ht.Add([]Class{})
	case C.ASTTYPE_PROPLIST:
		return ht.Add([]Prop{})
	case C.ASTTYPE_ARRAY:
		return ht.Add(Array{})
	case C.ASTTYPE_ARRAY_INTERFACE:
		return ht.Add([]interface{}{})
	case C.ASTTYPE_ARGDEFS:
		return ht.Add([]ArgDef{})
	}

	fmt.Printf("%#v\n", typ)
	panic("Bad type")
}

//export AppendArray
func AppendArray(arrayHandle, newValue goHandle) goHandle {
	array := ht.Get(arrayHandle)
	switch array.(type) {
	case []VariableDef:
		return ht.Add(append(array.([]VariableDef), ht.Get(newValue).(VariableDef)))
	case []Class:
		return ht.Add(append(array.([]Class), ht.Get(newValue).(Class)))
	case []Prop:
		return ht.Add(append(array.([]Prop), ht.Get(newValue).(Prop)))
	case []interface{}:
		return ht.Add(append(array.([]interface{}), ht.Get(newValue)))
	case Array:
		return ht.Add(append(array.(Array), ht.Get(newValue)))
	case []ArgDef:
		return ht.Add(append(array.([]ArgDef), ht.Get(newValue).(ArgDef)))
	}

	fmt.Printf("%#v\n", array)
	panic("Bad type")
}

//export NewFile
func NewFile(classesAndDefines goHandle) goHandle {
	lastFile = &File{}

	for _, classOrDefine := range ht.Get(classesAndDefines).([]interface{}) {
		switch classOrDefine.(type) {
		case Class:
			lastFile.Classes = append(lastFile.Classes, classOrDefine.(Class))
		case Define:
			lastFile.Defines = append(lastFile.Defines, classOrDefine.(Define))
		case Node:
			lastFile.Nodes = append(lastFile.Nodes, classOrDefine.(Node))
		default:
			panic("Found top-level object which is not class or define")
		}
	}

	return ht.Add(lastFile)
}

//export NewClass
func NewClass(lineNum C.int, identifier *C.char, argDefsH, defsAndDeclsH goHandle) goHandle {
	argDefs := ht.Get(argDefsH).([]ArgDef)
	defsAndDecls := ht.Get(defsAndDeclsH).([]interface{})

	defs := []VariableDef{}
	decls := []Declaration{}

	for _, val := range defsAndDecls {
		switch val.(type) {
		case VariableDef:
			defs = append(defs, val.(VariableDef))
		case Declaration:
			decls = append(decls, val.(Declaration))
		default:
			panic("Value is neither def nor decl")
		}
	}

	return ht.Add(Class{
		LineNum:      int(lineNum),
		Name:         C.GoString(identifier),
		ArgDefs:      argDefs,
		VariableDefs: defs,
		Declarations: decls,
	})
}

//export SawNode
func SawNode(lineNum C.int, name *C.char, defsAndDeclsHandle goHandle) goHandle {
	defsAndDecls := ht.Get(defsAndDeclsHandle).([]interface{})

	defs := []VariableDef{}
	decls := []Declaration{}

	for _, val := range defsAndDecls {
		switch val.(type) {
		case VariableDef:
			defs = append(defs, val.(VariableDef))
		case Declaration:
			decls = append(decls, val.(Declaration))
		default:
			panic("Value is neither def nor decl")
		}
	}

	return ht.Add(Node{
		LineNum:      int(lineNum),
		Name:         C.GoString(name),
		VariableDefs: defs,
		Declarations: decls,
	})
}

//export SawVariableDef
func SawVariableDef(lineNum C.int, varName *C.char, val goHandle) goHandle {
	return ht.Add(VariableDef{
		int(lineNum),
		VariableName{int(lineNum), C.GoString(varName)},
		ht.Get(val),
	})
}

//export SawQuotedString
func SawQuotedString(lineNum C.int, val *C.char) goHandle {
	return ht.Add(QuotedString(C.GoString(val)))
}

//export SawInt
func SawInt(lineNum C.int, val int) goHandle {
	return ht.Add(val)
}

//export SawVariableName
func SawVariableName(lineNum C.int, name *C.char) goHandle {
	return ht.Add(VariableName{int(lineNum), C.GoString(name)})
}

//export SawDeclaration
func SawDeclaration(lineNum C.int, typ *C.char, scalar, proplist goHandle) goHandle {
	return ht.Add(Declaration{
		LineNum: int(lineNum),
		Type:    C.GoString(typ),
		Scalar:  ht.Get(scalar).(Value),
		Props:   ht.Get(proplist).([]Prop),
	})
}

//export SawProp
func SawProp(lineNum C.int, propName *C.char, value goHandle) goHandle {
	return ht.Add(Prop{
		LineNum: int(lineNum),
		Name:    C.GoString(propName),
		Value:   ht.Get(value),
	})
}

//export SawReference
func SawReference(lineNum C.int, typ *C.char, scalar goHandle) goHandle {
	return ht.Add(Reference{
		LineNum: int(lineNum),
		Type:    C.GoString(typ),
		Scalar:  ht.Get(scalar),
	})
}

//export SawDefine
func SawDefine(lineNum C.int, modifier, name *C.char) goHandle {
	var dt DefineType
	switch C.GoString(modifier) {
	case "single":
		dt = DefineTypeSingle
	case "multiple":
		dt = DefineTypeMultiple
	default:
		return -1
	}

	return ht.Add(Define{
		LineNum: int(lineNum),
		Name:    C.GoString(name),
		Type:    dt,
	})
}

//export SawArgDef
func SawArgDef(lineNum C.int, varName *C.char, val goHandle) goHandle {
	v := Value(nil)
	if val != 0 {
		v = ht.Get(val).(Value)
	}

	return ht.Add(ArgDef{
		LineNum:      int(lineNum),
		VariableName: VariableName{int(lineNum), C.GoString(varName)},
		Val:          v,
	})
}

func Lex(filename string, r io.Reader) (*File, error) {
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	ret := C.doparse(C.CString(string(buf)))
	if ret.code != 0 {
		return nil, fmt.Errorf(
			"%s:%d: %s", filename, C.line_num, C.GoString(ret.error),
		)
	} else {
		return lastFile, nil
	}
}
