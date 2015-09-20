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

type File struct {
	Classes []Class
}

func (f *File) String() string {
	s := ""
	for _, class := range f.Classes {
		s = class.String() + "\n"
	}

	return s
}

type Class struct {
	Name         string
	Defs         []Def
	Declarations []Declaration
}

func (c *Class) String() string {
	defs := ""
	decls := ""
	for _, def := range c.Defs {
		defs += fmt.Sprintf("\t\t%s\n", def.String())
	}

	for _, decl := range c.Declarations {
		decls += fmt.Sprintf("\t\t%s\n", decl.String())
	}

	return fmt.Sprintf("\tclass %s {\n%s\n%s\n\t}\n", c.Name, defs, decls)
}

type Def struct {
	Name Variable
	Val  Value
}

func (d *Def) String() string {
	return fmt.Sprintf("%s = %s", d.Name, d.Val)
}

type QuotedString string
type Variable string

// A used type, for instance package { 'nginx': ensure => 'latest' }
type Declaration struct {
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
	Name  string
	Value Value
}

func (p *Prop) String() string {
	return fmt.Sprintf("%s => %s", p.Name, p.Value)
}

// A value, for instance 1, 'foo', $bar or [ 1, 'five', ]
type Value interface{}

// An array of strings, number or references, for instance
// [ 1, 'foo', package[$webserver], ]
type Array []interface{}

// A reference, for instance package['nginx'] or package[$webserver]
type Reference struct {
	Type   string
	Scalar Value
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
	}

	fmt.Printf("%#v\n", typ)
	panic("Bad type")
}

//export AppendArray
func AppendArray(arrayHandle, newValue goHandle) goHandle {
	array := ht.Get(arrayHandle)
	switch array.(type) {
	case []Def:
		return ht.Add(append(array.([]Def), ht.Get(newValue).(Def)))
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

//export NewFile
func NewFile(classes goHandle) goHandle {
	lastFile = &File{
		Classes: ht.Get(classes).([]Class),
	}
	return ht.Add(lastFile)
}

//export NewClass
func NewClass(identifier *C.char, defsAndDeclsHandle goHandle) goHandle {
	defsAndDecls := ht.Get(defsAndDeclsHandle).([]interface{})

	defs := []Def{}
	decls := []Declaration{}

	for _, val := range defsAndDecls {
		switch val.(type) {
		case Def:
			defs = append(defs, val.(Def))
		case Declaration:
			decls = append(decls, val.(Declaration))
		default:
			panic("Value is neither def nor decl")
		}
	}

	return ht.Add(Class{
		Name:         C.GoString(identifier),
		Defs:         defs,
		Declarations: decls,
	})
}

//export SawDef
func SawDef(varName *C.char, val goHandle) goHandle {
	return ht.Add(Def{
		Variable(C.GoString(varName)),
		ht.Get(val),
	})
}

//export SawQuotedString
func SawQuotedString(val *C.char) goHandle {
	return ht.Add(QuotedString(C.GoString(val)))
}

//export SawInt
func SawInt(val int) goHandle {
	return ht.Add(val)
}

//export SawVariable
func SawVariable(name *C.char) goHandle {
	return ht.Add(Variable(C.GoString(name)))
}

//export SawDeclaration
func SawDeclaration(typ *C.char, scalar, proplist goHandle) goHandle {
	return ht.Add(Declaration{
		Type:   C.GoString(typ),
		Scalar: ht.Get(scalar).(Value),
		Props:  ht.Get(proplist).([]Prop),
	})
}

//export SawProp
func SawProp(propName *C.char, value goHandle) goHandle {
	return ht.Add(Prop{
		Name:  C.GoString(propName),
		Value: ht.Get(value),
	})
}

//export SawReference
func SawReference(typ *C.char, scalar goHandle) goHandle {
	return ht.Add(Reference{
		Type:   C.GoString(typ),
		Scalar: ht.Get(scalar),
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
