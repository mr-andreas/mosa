package reducer

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

// Reduces the whole manifest to a number of concrete declarations. All
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
func Reduce(ast *AST) ([]Declaration, error) {
	r := newResolver(ast)
	return r.resolve()
}
