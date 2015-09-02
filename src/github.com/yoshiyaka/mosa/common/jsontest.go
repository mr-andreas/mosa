package common

import (
	"encoding/json"
	"reflect"
)

// Returns whether the JSON encoded strings j1 and j2 are equal JSON objects.
// For instance JSONEquals(`{"foo":"bar","baz":x"}`, `{"baz":x","foo":"bar"}`)
// will return true.
// If j1 or j2 is not valid JSON, this function will panic.
func JSONEquals(j1, j2 string) bool {
	var j1obj, j2obj interface{}

	if err := json.Unmarshal([]byte(j1), &j1obj); err != nil {
		panic(err)
	}
	if err := json.Unmarshal([]byte(j2), &j2obj); err != nil {
		panic(err)
	}

	return reflect.DeepEqual(j1obj, j2obj)
}

// Returns wether the JSON representations of i1 and i2 equals.
func EqualsAsJSON(i1, i2 interface{}) bool {
	j1, j1err := json.Marshal(i1)
	j2, j2err := json.Marshal(i2)

	if j1err != nil || j2err != nil {
		return false
	}

	return JSONEquals(string(j1), string(j2))
}
