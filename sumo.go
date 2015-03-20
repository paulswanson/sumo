package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
)

func main() {

	flag.Parse()
	if flag.NArg() < 3 {
		fmt.Println("Please specifiy: index input output")
		return
	}

	// TODO A no write, stdout only option is necessary for sanity

	// Read and build the master replace index
	indexFile, err := os.Open(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
		return
	}
	defer indexFile.Close()

	fmt.Printf("Building index ...\n")
	masterIndex, err := newReplaceIndex(indexFile)
	if err != nil {
		switch err {
		case io.EOF:
			fmt.Printf("The index file is empty, can't continue.\n")
			return
		default:
			log.Fatal(err)
			return
		}
	}

	// Read in the input file
	inputFile, err := os.Open(flag.Arg(1))
	if err != nil {
		log.Fatal(err)
		return
	}
	defer inputFile.Close()

	fmt.Printf("Reading input file into memory ...\n")
	inputData, err := ioutil.ReadAll(inputFile)
	if err != nil {
		log.Fatal(err)
		return
	}

	fmt.Printf("Processing input data ...\n")
	var off, end int
	inputChan := make(chan line)

	// Send line at a time as work task to generate deltas
	go func() {
		for {
			i := bytes.IndexByte(inputData[off:], '\n')
			end = off + i + 1
			if i < 0 {
				err = io.EOF
				break
			}
			inputChan <- line{off, inputData[off:end]}
			off = end
		}
		close(inputChan)
	}()

	// Collect the resultant deltas
	deltas := producer(inputChan, &masterIndex)

	if len(deltas) == 0 {
		fmt.Printf("No matches found\n")
		return
	}

	// open output file
	outFile, err := os.Create(flag.Arg(2))
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := outFile.Close(); err != nil {
			panic(err)
		}
	}()

	fmt.Printf("Writing to output file ...\n")
	// make a write buffer
	w := bufio.NewWriter(outFile)

	var o int

	// for each delta process the corresponding line in inputData
	for _, d := range deltas {

		// TODO This approach could lead to large slices in memory

		// read up to offset as slice, and write
		_, err = w.Write(inputData[o:d.off])
		if err != nil {
			log.Fatal(err)
			return
		}

		// Write replacement value
		_, err = w.Write(masterIndex.readItem(d.index).replace)
		if err != nil {
			log.Fatal(err)
			return
		}

		o = d.off + len(masterIndex.readItem(d.index).find)
		w.Flush()

	}

	// write out remainder of slice
	_, err = w.Write(inputData[o:])
	if err != nil {
		log.Fatal(err)
		return
	}
	w.Flush()
	fmt.Printf("Done.\n")
}
