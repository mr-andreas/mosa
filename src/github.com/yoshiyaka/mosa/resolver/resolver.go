package resolver

import (
	"fmt"
	"strings"

	. "github.com/yoshiyaka/mosa/ast"
)

type ErrorType int

const (
	ErrorTypeUnresolvableVariable ErrorType = iota
	ErrorTypeCyclicVariable
	ErrorTypeMultipleDefinition
)

type Err struct {
	Type       ErrorType
	File       string
	Line       int
	SymbolName string
}

func (e *Err) Error() string {
	msg := ""
	switch e.Type {
	case ErrorTypeCyclicVariable:
		msg = "Cyclic dependency for variable " + e.SymbolName
	case ErrorTypeMultipleDefinition:
		msg = "Multiple definition for variable " + e.SymbolName
	case ErrorTypeUnresolvableVariable:
		msg = "Reference to non-defined variable " + e.SymbolName
	default:
		msg = "Unknown"
	}

	return fmt.Sprintf("Error at %s:%d: %s", e.File, e.Line, msg)
}

type CyclicError struct {
	Err
	Cycle []string
}

func (ce *CyclicError) Error() string {
	msg := ce.Err.Error()
	msg += fmt.Sprintf(" (%s)", strings.Join(ce.Cycle, " -> "))
	return msg
}

// Resolves the whole manifest to a number of concrete declarations. All
// parameters in the returned decalartions will be concrete values. For
// instance, if called with the following manifest:
//
//  node 'localhost' {
//  	class { 'Webserver':
//  		docroot => '/home/www',
//  	}
//  }
//
//  class Webserver(
//  	$docroot = '/var/www',
//  	$workers = 8,
//  ){
//  	$server = 'nginx'
//
//  	package { $server: ensure => installed }
//
//  	file { '/etc/nginx/conf.d/workers.conf':
//  		ensure => 'present',
//  		content => "workers = $workers",
//  		depends => package[$server],
//  	}
//
//  	file { $docroot: ensure => 'directory', }
//
//  	service { $server:
//  		ensure => 'running',
//  		depends => [
//  			file['/etc/nginx/conf.d/workers.conf'],
//  			package[$server],
//  		],
//  	}
//  }
// The following declarations would be returned. Note that all variables are
// gone:
//  package { 'nginx': ensure => installed }
//
//  file { '/etc/nginx/conf.d/workers.conf':
//  	ensure => 'present',
//  	content => "workers = 8",
//  	depends => package['nginx'],
//  }
//
//  file { '/home/www': ensure => 'directory', }
//
//  service { 'nginx':
//  	ensure => 'running',
//  	depends => [
//  		file['/etc/nginx/conf.d/workers.conf'],
//  		package['nginx'],
//  	],
//  }
//
func Resolve(ast *AST) ([]Declaration, error) {
	r := newResolver(ast)
	return r.resolve()
}

// Resolves a whole manifest
type resolver struct {
	ast *AST

	gs *globalState
}

func newResolver(ast *AST) *resolver {
	return &resolver{
		ast: ast,
		gs:  newGlobalState(),
	}
}

func (r *resolver) resolve() ([]Declaration, error) {
	if err := r.gs.populateClassesByName(r.ast.Classes); err != nil {
		return nil, err
	}
	if err := r.gs.populateDefinesByName(r.ast.Defines); err != nil {
		return nil, err
	}

	for _, node := range r.ast.Nodes {
		if err := r.resolveNode(&node); err != nil {
			return nil, err
		}
	}

	if err := checkDeclarationsValidity(r.gs.realizedDeclarationsInOrder); err != nil {
		return nil, err
	}

	return r.gs.realizedDeclarationsInOrder, nil
}

func (r *resolver) resolveNode(node *Node) error {
	castedClass := Class(*node)
	return r.realizeClassesRecursive(&castedClass, nil, "", 0)
}

func (r *resolver) realizeClassesRecursive(c *Class, args []Prop, file string, line int) error {
	classResolver := newClassResolver(r.gs, c, args, file, line)
	if _, err := classResolver.resolve(); err != nil {
		return err
	}

	return nil
}

// Late checks for manifest validity, such as enforcing that all
// 'unless'-arguments to the built in exec define are of type string.
func checkDeclarationsValidity(d []Declaration) error {
	for _, decl := range d {
		if decl.Type != "exec" {
			continue
		}

		for _, prop := range decl.Props {
			if prop.Name != "unless" {
				continue
			}

			if _, isString := prop.Value.(QuotedString); !isString {
				return fmt.Errorf(
					"Value for parameter 'unless' must be of type string at %s:%d",
					decl.Filename, prop.LineNum,
				)
			}
		}
	}

	return nil
}
