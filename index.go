package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"sync"
)

type replaceItem struct {
	find    []byte
	replace []byte
}

// type replaceIndex []replaceItem

type replaceIndex struct {
	mu    sync.RWMutex
	index []replaceItem
}

// TODO This func also needs to return an error - file read or csv error
func newReplaceIndex(r *csv.Reader) (ri replaceIndex, err error) {
	ri.index = make([]replaceItem, 0)
	for i := 0; ; i++ {
		inputRow, err := r.Read()
		if err != nil {
			// Return io.EOF only if there are no records
			if err == io.EOF && i > 0 {
				return ri, nil
			} else {
				return ri, err
			}
		}

		fmt.Printf("Index item: %v, %v\n", []byte(inputRow[0]), []byte(inputRow[1]))
		ri.index = append(ri.index, replaceItem{find: []byte(inputRow[0]), replace: []byte(inputRow[1])})
	}
}

func (r *replaceIndex) readItem(i int) replaceItem {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.index[i]
}

func (r *replaceIndex) writeItem(i int, ri replaceItem) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.index[i] = ri
}

func (r *replaceIndex) len() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.index)
}
