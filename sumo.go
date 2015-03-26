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

	write := func(w *bufio.Writer, b []byte) {
		_, err := w.Write(b)
		if err != nil {
			log.Fatal(err)
			return
		}
	}

	fmt.Printf("Writing to output file ...\n")
	w := bufio.NewWriter(outFile)

	var o int

	// Write out to file, making replacements as we go
	for _, d := range deltas {

		// Read up to the next found word
		write(w, inputData[o:d.off])

		// Write replacement value
		write(w, masterIndex.readItem(d.index).replace)

		// Jump forward to end of replaced word
		o = d.off + len(masterIndex.readItem(d.index).find)
		w.Flush()

	}

	// Write the remaining input data
	write(w, inputData[o:])
	w.Flush()
	fmt.Printf("Done.\n")
}
