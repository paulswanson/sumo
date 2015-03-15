package main

import (
	"encoding/csv"
	"fmt"
	"sync"
)

type replaceItem struct {
	mu      sync.Mutex
	find    []byte
	replace []byte
}

type replaceIndex []replaceItem

// type replaceIndex struct {
//	mu    sync.Mutex
//	index []replaceItem
//}

// TODO This func also needs to return an error - file read or csv error 
func newReplaceIndex(r *csv.Reader) replaceIndex {
	x := make(replaceIndex, 0)
	for i := 0; ; i++ {
		inputRow, err := r.Read()
		if inputRow == nil && err != nil {
			// TODO return nil, err
			break
		}

		fmt.Printf("Index item: %v, %v\n", []byte(inputRow[0]), []byte(inputRow[1]))
		x = append(x, replaceItem{find: []byte(inputRow[0]), replace: []byte(inputRow[1])})
	}
	return x
}

func (r replaceIndex) readItem(i int) replaceItem {
	r[i].mu.Lock()
	defer r[i].mu.Unlock()
	return r[i]
}

//func (r *replaceIndex) writeItem(i int, r replaceItem) {
//	r.mu.Lock()
//	defer r.mu.Unlock()
//	r.index[i] = r
//}
