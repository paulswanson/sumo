package main

import (
	"index/suffixarray"
	"runtime"
	"sort"
	"sync"
)

type line struct {
	off   int    // Sequence number of line
	value []byte // One line from source slice
}

type delta struct {
	line  int // Line offset
	off   int // Offset of existing value
	index int // Index of replacement value
}

// TODO Are there any errors that need handling?
func producer(input <-chan line, index *replaceIndex) []delta {

	var wg sync.WaitGroup
	runtime.GOMAXPROCS(runtime.NumCPU()) // Maximum CPU utilisation please!
	maxGophers := runtime.GOMAXPROCS(0) * 2
	wg.Add(maxGophers)

	deltaChan := make(chan []delta)
	done := make(chan bool)
	defer close(done)

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

	deltas := make([]delta, 0)
	for d := range deltaChan {
		deltas = append(deltas, d...)
	}

	sort.Sort(ByLine(deltas))
	return deltas
}

// Make deltas
func makeDeltas(t line, index *replaceIndex, id int) []delta {

	s := make([]delta, 0)
	lineIndex := suffixarray.New(t.value)

	for i := 0; i < index.len(); i++ {
		results := lineIndex.Lookup(index.readItem(i).find, -1)
		for _, p := range results {
			s = append(s, delta{line: t.off, off: p, index: i})
		}
	}
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
	return d[i].line < d[j].line
}
