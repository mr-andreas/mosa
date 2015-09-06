package manifest2

import "C"

type goHandle int

type handleTable struct {
	table []interface{}
}

func (ht *handleTable) Add(i interface{}) goHandle {
	ht.table = append(ht.table, i)
	return goHandle(len(ht.table) - 1)
}

func (ht *handleTable) Get(i goHandle) interface{} {
	return ht.table[i]
}
