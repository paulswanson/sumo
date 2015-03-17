package main

import (
	"index/suffixarray"
	"runtime"
	"sort"
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
func producer(input <-chan workTask, index *replaceIndex) []delta {

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
func makeDeltas(t workTask, index *replaceIndex, id int) []delta {

	s := make([]delta, 0)
	lineIndex := suffixarray.New(t.lineValue)

	for i := 0; i < index.len(); i++ {
		results := lineIndex.Lookup(index.readItem(i).find, -1)
		for _, p := range results {
			s = append(s, delta{line: t.lineNumber, pos: p, index: i})
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
