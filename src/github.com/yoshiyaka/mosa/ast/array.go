package ast

import "fmt"

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
