package manifest2

// #cgo LDFLAGS: -lfl
// extern int yylex();
// extern int doparse();
import "C"
import "fmt"

func Lex() {
	ret := C.doparse()

	fmt.Println("yylex returned", ret)
	//	panic(ret)
}
