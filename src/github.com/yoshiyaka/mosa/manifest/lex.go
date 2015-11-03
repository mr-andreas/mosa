package manifest

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
	"reflect"
)

var (
	ht         handleTable
	currentAST *AST
)

type stringable interface {
	String() string
}

type AST struct {
	Classes []Class
	Defines []Define
	Nodes   []Node
}

func NewAST() *AST {
	return &AST{}
}

func (f *AST) String() string {
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
	Filename     string
	LineNum      int
	Name         string
	ArgDefs      []VariableDef
	VariableDefs []VariableDef
	Declarations []Declaration
	Type         DefineType
}

type Node Class

type Class struct {
	Filename     string
	LineNum      int
	Name         string
	ArgDefs      []VariableDef
	VariableDefs []VariableDef
	Declarations []Declaration
}

// Returns whether the classes are equal. Line numbers and filenames are not
// taken into consideration.
func (c *Class) Equals(c2 *Class) bool {
	return c.Name == c2.Name &&
		VariableDefsEquals(c.ArgDefs, c2.ArgDefs) &&
		VariableDefsEquals(c.VariableDefs, c2.VariableDefs) &&
		DeclarationsEquals(c.Declarations, c2.Declarations)
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

type VariableDef struct {
	LineNum      int
	VariableName VariableName
	Val          Value
}

func (v *VariableDef) Equals(v2 *VariableDef) bool {
	return ValueEquals(v.Val, v2.Val)
}

// Returns whether the variable def lists are equals. Order is important.
func VariableDefsEquals(v1, v2 []VariableDef) bool {
	if len(v1) != len(v2) {
		return false
	}

	for i, arg := range v1 {
		if !arg.Equals(&v2[i]) {
			return false
		}
	}

	return true
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

// A double-quoted interpolated string which may contain variables. For instance
// "php5-$module" or "/home/$user".
//
// It consists of a number of segments which are parsed directly in bison, where
// each segment is either a raw string, or a variable name. For instance, the
// string "/home/$user/.config-{$app}" will be interpreted as
// [ "/home/", $user, "/.config-", $app ].
type InterpolatedString struct {
	LineNum int

	// Each segment will be either a string, or a VariableName.
	Segments []interface{}
}

func (is *InterpolatedString) String() string {
	str := `"`
	for _, seg := range is.Segments {
		switch seg.(type) {
		case string:
			str += seg.(string)
		case VariableName:
			str += seg.(VariableName).Str
		default:
			panic("Bad segment type")
		}
	}
	str += `"`

	return str
}

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

func (d *Declaration) Equals(d2 *Declaration) bool {
	return d.Type == d2.Type &&
		ValueEquals(d.Scalar, d2.Scalar) &&
		PropsEquals(d.Props, d2.Props)
}

// Returns whether the declaration lists are equal. Order is important.
func DeclarationsEquals(d1, d2 []Declaration) bool {
	if len(d1) != len(d2) {
		return false
	}

	for i, decl := range d1 {
		if !decl.Equals(&d2[i]) {
			return false
		}
	}

	return true
}

func (d *Declaration) String() string {
	props := ""
	for _, prop := range d.Props {
		props += fmt.Sprintf("\n\t\t\t%s,", prop.String())
	}
	if len(d.Props) > 0 {
		props += "\n\t\t"
	}

	return fmt.Sprintf("%s { %s: %s}\n", d.Type, d.Scalar, props)
}

// A property in declaration, for instance ensure => 'latest'
type Prop struct {
	LineNum int
	Name    string
	Value   Value
}

func (p *Prop) Equals(p2 *Prop) bool {
	return ValueEquals(p.Value, p2.Value)
}

// Returns whether the props lists are equal. Order is important.
func PropsEquals(p1, p2 []Prop) bool {
	if len(p1) != len(p2) {
		return false
	}

	for i, prop := range p1 {
		if !prop.Equals(&p2[i]) {
			return false
		}
	}

	return true
}

func (p *Prop) String() string {
	if intVal, ok := p.Value.(int); ok {
		return fmt.Sprintf("%s => %d", p.Name, intVal)
	} else {
		return fmt.Sprintf("%s => %s", p.Name, p.Value)
	}
}

// A value, for instance 1, 'foo', $bar or [ 1, 'five', ]
type Value interface{}

func ValueEquals(v1, v2 Value) bool {
	switch v1.(type) {
	case Reference:
		if ref2, ok := v2.(Reference); ok {
			v1ref := v1.(Reference)
			return ref2.Equals(&v1ref)
		} else {
			return false
		}
	case Array:
		if a2, ok := v2.(Array); ok {
			return ArrayEquals(v1.(Array), a2)
		} else {
			return false
		}
	default:
		return reflect.DeepEqual(v1, v2)
	}
}

// An array of strings, number or references, for instance
// [ 1, 'foo', package[$webserver], ]
type Array []interface{}

func ArrayEquals(a1, a2 Array) bool {
	if len(a1) != len(a2) {
		return false
	}

	for i, _ := range a1 {
		if !ValueEquals(a1[i], a2[i]) {
			return false
		}
	}

	return true
}

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

func (r Reference) String() string {
	return fmt.Sprintf("%s[%s]", r.Type, r.Scalar)
}

func (r *Reference) Equals(r2 *Reference) bool {
	return ValueEquals(r.Scalar, r2.Scalar)
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
		return ht.Add([]VariableDef{})
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
	}

	fmt.Printf("%#v\n", array)
	panic("Bad type")
}

//export SawBody
func SawBody(classesAndDefines goHandle) {
	for _, classOrDefine := range ht.Get(classesAndDefines).([]interface{}) {
		switch classOrDefine.(type) {
		case Class:
			currentAST.Classes = append(currentAST.Classes, classOrDefine.(Class))
		case Define:
			currentAST.Defines = append(currentAST.Defines, classOrDefine.(Define))
		case Node:
			currentAST.Nodes = append(currentAST.Nodes, classOrDefine.(Node))
		default:
			panic("Found top-level object which is not class or define")
		}
	}
}

//export NewClass
func NewClass(lineNum C.int, identifier *C.char, argDefsH, defsAndDeclsH goHandle) goHandle {
	argDefs := ht.Get(argDefsH).([]VariableDef)
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
		Filename:     curFilename,
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
		Filename:     curFilename,
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

//export EmptyInterpolatedString
func EmptyInterpolatedString(lineNum C.int) goHandle {
	return ht.Add(InterpolatedString{
		LineNum: int(lineNum),
	})
}

//export AppendInterpolatedString
func AppendInterpolatedString(ipStrH goHandle, val goHandle) goHandle {
	ipStr := ht.Get(ipStrH).(InterpolatedString)
	ipStr.Segments = append(ipStr.Segments, ht.Get(val))
	return ht.Add(ipStr)
}

//export SawString
func SawString(val *C.char) goHandle {
	return ht.Add(C.GoString(val))
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
func SawDefine(lineNum C.int, modifier, name *C.char, argDefsH, defsAndDeclsH goHandle) goHandle {
	var dt DefineType
	switch C.GoString(modifier) {
	case "single":
		dt = DefineTypeSingle
	case "multiple":
		dt = DefineTypeMultiple
	default:
		return -1
	}

	argDefs := ht.Get(argDefsH).([]VariableDef)
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

	return ht.Add(Define{
		Filename:     curFilename,
		LineNum:      int(lineNum),
		Name:         C.GoString(name),
		ArgDefs:      argDefs,
		VariableDefs: defs,
		Declarations: decls,
		Type:         dt,
	})
}

//export SawArgDef
func SawArgDef(lineNum C.int, varName *C.char, val goHandle) goHandle {
	v := Value(nil)
	if val != 0 {
		v = ht.Get(val).(Value)
	}

	return ht.Add(VariableDef{
		LineNum:      int(lineNum),
		VariableName: VariableName{int(lineNum), C.GoString(varName)},
		Val:          v,
	})
}

var curFilename string

// Please note that as of current, Lex() is /NOT/ reentrant.
// This function will parse r and store the output into ast.
func Lex(ast *AST, filename string, r io.Reader) error {
	curFilename = filename
	currentAST = ast

	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	ret := C.doparse(C.CString(string(buf)))
	if ret.code != 0 {
		return fmt.Errorf(
			"%s:%d: %s", filename, C.line_num, C.GoString(ret.error),
		)
	} else {
		return nil
	}
}
