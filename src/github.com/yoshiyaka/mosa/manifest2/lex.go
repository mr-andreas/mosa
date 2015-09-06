package manifest2

// #cgo LDFLAGS: -lfl
// extern int yylex();
// extern int doparse();
import "C"
import (
	"encoding/json"
	"fmt"
)

var (
	ht       handleTable
	lastFile *File
)

type File struct {
	Classes *Classes
}

type Classes struct {
	Classes []*Class
}

type Class struct {
	Name string
	Defs *Defs
}

type Defs struct {
	Defs []*Def
}

type Def struct {
	Name,
	Val string
}

//export GoFunc
func GoFunc() {
	fmt.Println("A GO FUNC")
}

//export NewFile
func NewFile(classes goHandle) goHandle {
	lastFile = &File{
		Classes: ht.Get(classes).(*Classes),
	}
	return ht.Add(lastFile)
}

//export AddClasses
func AddClasses(classesH, classH goHandle) {
	defs := ht.Get(classesH).(*Classes)
	defs.Classes = append(defs.Classes, ht.Get(classH).(*Class))
}

//export NewClasses
func NewClasses(class goHandle) goHandle {
	return ht.Add(&Classes{
		Classes: []*Class{ht.Get(class).(*Class)},
	})
}

//export NewClass
func NewClass(identifier *C.char, defs goHandle) goHandle {
	return ht.Add(&Class{
		Name: C.GoString(identifier),
		Defs: ht.Get(defs).(*Defs),
	})
}

//export AddDefs
func AddDefs(defsH, defH goHandle) {
	defs := ht.Get(defsH).(*Defs)
	defs.Defs = append(defs.Defs, ht.Get(defH).(*Def))
}

//export NewDefs
func NewDefs(def goHandle) goHandle {
	return ht.Add(&Defs{
		Defs: []*Def{ht.Get(def).(*Def)},
	})
}

//export SawDef
func SawDef(name, val *C.char) goHandle {
	fmt.Println("GOLANG SawDef", C.GoString(name), C.GoString(val))

	return ht.Add(&Def{
		C.GoString(name),
		C.GoString(val),
	})
}

func Lex() {
	ret := C.doparse()

	fmt.Println("yylex returned", ret)

	js, _ := json.MarshalIndent(lastFile, "", "  ")
	fmt.Println(string(js))

	//	panic(ret)
}
