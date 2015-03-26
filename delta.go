package main

import (
	"fmt"
	"index/suffixarray"
	"regexp/syntax"
	"runtime"
	"sort"
	"sync"
)

type line struct {
	off   int    // Offset of line
	value []byte // One line from source slice
}

type delta struct {
	off   int // Offset of existing value
	index int // Index of replacement value
}

// TODO Are there any errors that need handling?
func getDeltas(input <-chan line, index *replaceIndex) []delta {

	var wg sync.WaitGroup
	runtime.GOMAXPROCS(runtime.NumCPU()) // Maximum CPU utilisation please!
	maxGophers := runtime.GOMAXPROCS(0)
	wg.Add(maxGophers)

	deltaChan := make(chan []delta)
	done := make(chan bool)
	defer close(done)

	fmt.Printf("Launching %v gophers ...\n", maxGophers)

	for i := 0; i < maxGophers; i++ {
		go func() {
			for t := range input {
				select {
				case deltaChan <- makeDeltas(t, index, i):
				case <-done:
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

	fmt.Printf("Generating deltas... \n")
	deltas := make([]delta, 0)
	for d := range deltaChan {
		if len(d) > 0 {
			deltas = append(deltas, d...)
		}
	}

	fmt.Printf("Got %v deltas.\n", len(deltas))

	sort.Sort(ByLine(deltas))
	return deltas
}

// Make deltas
func makeDeltas(l line, index *replaceIndex, id int) []delta {

	s := make([]delta, 0)

	lineIndex := suffixarray.New(l.value)

	// Look for each replacement word in the line
	for i := 0; i < index.len(); i++ {

		// Find this word in the line
		results := lineIndex.Lookup(index.readItem(i).find, -1)
		if len(results) > 0 {
			for _, p := range results {

				// Is it a full word match or not?
				if (p < len(l.value) && !syntax.IsWordChar(rune(l.value[p+len(index.readItem(i).find)])+1)) || (p == len(l.value)) {
					d := delta{off: l.off + p, index: i}
					s = append(s, d)
				}
			}
		}
	}

	// TODO Optional memory optimiser. Use less system RAM at the expense of CPU.
	// debug.FreeOSMemory()

	return s
}

type ByLine []delta

func (d ByLine) Len() int {
	return len(d)
}

func (d ByLine) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}

func (d ByLine) Less(i, j int) bool {
	return d[i].off < d[j].off
}
