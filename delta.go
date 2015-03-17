package main

import (
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

// TODO Are there any errors that need handling?
func producer(input <-chan workTask, index replaceIndex) []delta {

	var wg sync.WaitGroup
	runtime.GOMAXPROCS(runtime.NumCPU()) // Maximum CPU utilisation please!
	maxGophers := runtime.GOMAXPROCS(0) * 2
	wg.Add(maxGophers)

	deltaChan := make(chan []delta)
	done := make(chan bool)
	defer close(done)

	for i := 0; i < maxGophers; i++ {
		go func() {
			// worker(done, input, deltaChan, index, i)
			for t := range input {
				// fmt.Printf("worker(%v): got %v\n", i, t)
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

// Make deltas
func makeDeltas(t workTask, index replaceIndex, id int) []delta {

	s := make([]delta, 0)

	// fmt.Printf("%v: Building suffixarray on: %v\n", id, string(t.lineValue))
	lineIndex := suffixarray.New(t.lineValue)

	for i := 0; i < index.len(); i++ {

		// fmt.Printf("makeDeltas(%v) find: %v \n", id, string(index.readItem(i).find))
		results := lineIndex.Lookup(index.readItem(i).find, -1)
		// fmt.Printf("makeDeltas(%v) found: %v\n", id, results)

		for _, p := range results {
			s = append(s, delta{line: t.lineNumber, pos: p, index: i})
		}
	}
	return s
}
