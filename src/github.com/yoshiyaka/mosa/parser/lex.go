package parser

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

	. "github.com/yoshiyaka/mosa/ast"
)

var (
	ht         handleTable
	currentAST *AST
)

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
		Filename: curFilename,
		LineNum:  int(lineNum),
		Type:     C.GoString(typ),
		Scalar:   ht.Get(scalar).(Value),
		Props:    ht.Get(proplist).([]Prop),
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
func Parse(ast *AST, filename string, r io.Reader) error {
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
