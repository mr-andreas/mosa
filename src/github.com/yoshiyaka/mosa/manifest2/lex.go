package manifest2

// #cgo LDFLAGS: -lfl
// typedef struct {
//   int code;
//   const char *error;
//   int line;
// } t_error;
// extern t_error doparse(char *);
// extern int line_num;
import "C"
import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
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

//export NilDefs
func NilDefs() goHandle {
	return ht.Add(&Defs{})
}

//export SawDef
func SawDef(name, val *C.char) goHandle {
	fmt.Println("GOLANG SawDef", C.GoString(name), C.GoString(val))

	return ht.Add(&Def{
		C.GoString(name),
		C.GoString(val),
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
		js, _ := json.MarshalIndent(lastFile, "", "  ")
		fmt.Println(string(js))
		return lastFile, nil
	}
}
