package main

import (
	"encoding/csv"
	"fmt"
)

type replaceItem struct {
	find    []byte
	replace []byte
}

type replaceIndex []replaceItem

func (x *replaceIndex) init(r *csv.Reader) {

	var i int
	for i = 0; ; i++ {
		inputRow, e := r.Read()
		if e != nil {
			// i = i - 1
			break
		}
		// TODO Not sure how to handle EOF / non-EOF errors
		//		if err != nil {
		//			return i, err
		//		}

		fmt.Printf("Index item: %v, %v", []byte(inputRow[0]), []byte(inputRow[1]))
		*x = append(*x, replaceItem{[]byte(inputRow[0]), []byte(inputRow[1])})
	}
}
