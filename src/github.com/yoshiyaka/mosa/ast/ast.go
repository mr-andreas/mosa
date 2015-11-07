package ast

import (
	"fmt"
	"reflect"
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

type VariableName struct {
	LineNum int
	Str     string
}

func (vn VariableName) String() string { return vn.Str }

// A used type, for instance package { 'nginx': ensure => 'latest' }
type Declaration struct {
	Filename string
	LineNum  int

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

func valToStr(i interface{}) string {
	switch i.(type) {
	case int, int64:
		return fmt.Sprintf("%d", i)
	case stringable:
		return i.(stringable).String()
	case Expression:
		return i.(Expression).String()
	default:
		return i.(string)
	}
}
