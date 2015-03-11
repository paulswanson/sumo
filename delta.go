package main

import (
	"fmt"
	"index/suffixarray"
	"runtime"
	"sync"
)

type workTask struct {
	lineValue  []byte // One line from source slice
	lineNumber int    // Sequence number of line
}

type delta struct {
	line  int // Line number of source []byte
	pos   int // Position of current value
	index int // Index value of replacement
}

func producer(input <-chan workTask, index replaceIndex) []delta {
	// TODO this func really needs to handle & return errors
	done := make(chan bool)
	defer close(done)

	deltaChan := make(chan []delta)

	var wg sync.WaitGroup
	runtime.GOMAXPROCS(runtime.NumCPU()) // Maximum CPU utilisation please!
	numWorkers := runtime.GOMAXPROCS(0) * 2
	wg.Add(numWorkers)

	for i := 0; i < numWorkers; i++ {
		go func() {
			// worker(done, input, deltaChan, index, i)
			for t := range input {
				fmt.Printf("worker(%v): got %v\n", i, t)
				select {
				case deltaChan <- makeDeltas(t, index, i):
				case <-done: // TODO Is this even required / useful
					return
				}
			}
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(deltaChan)
	}()

	deltas := make([]delta, 0)
	for d := range deltaChan {
		deltas = append(deltas, d...)
	}

	return deltas
}

//func worker(done <-chan bool, input <-chan workTask, output chan<- []delta, index replaceIndex, id int) {
//	for t := range input {
//		fmt.Printf("worker(%v): got %v\n", id, t)
//		select {
//		case output <- makeDeltas(t, index, id):
//		case <-done:
//			return
//		}
//	}
//}

// Make deltas
func makeDeltas(t workTask, index replaceIndex, id int) []delta {

	s := make([]delta, 0)

	fmt.Printf("%v: Building suffixarray on: %v\n", id, string(t.lineValue))
	lineIndex := suffixarray.New(t.lineValue)

	for i := 0; i < len(index); i++ {

		fmt.Printf("makeDeltas(%v) find: %v \n", id, string(index[i].find))
		results := lineIndex.Lookup(index[i].find, -1)
		fmt.Printf("makeDeltas(%v) found: %v\n", id, results)

		for p := range results {
			s = append(s, delta{t.lineNumber, p, i})
			fmt.Printf("makeDeltas(%v) adding: %v\n", id, p)
		}
	}
	return s
}
