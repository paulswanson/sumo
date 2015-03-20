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
func producer(input <-chan line, index *replaceIndex) []delta {

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

	fmt.Printf("Collecting deltas... \n")
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
func makeDeltas(t line, index *replaceIndex, id int) []delta {

	s := make([]delta, 0)

	lineIndex := suffixarray.New(t.value)

	for i := 0; i < index.len(); i++ {
		results := lineIndex.Lookup(index.readItem(i).find, -1)
		if len(results) > 0 {
			for _, p := range results {
				// It's not a match if it's a partial word
				x := p + len(index.readItem(i).find) - 1
				if x < len(t.value) {
					if syntax.IsWordChar(rune(t.value[x])) && !syntax.IsWordChar(rune(t.value[x+1])) {
						d := delta{off: t.off + p, index: i}
						s = append(s, d)
					}
				} else {
					d := delta{off: t.off + p, index: i}
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
