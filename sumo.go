package main

import (
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

	inputData, err := ioutil.ReadAll(inputFile)
	if err != nil {
		log.Fatal(err)
		return
	}

	// TODO Using bytes.Buffer could cause huge memory issues on big files
	// inputData should always be referenced via a pointer
	buf := bytes.NewBuffer(inputData)
	inputChan := make(chan workTask)
	var lineCount int

	// Feed the producer line at a time
	go func() {
		for {
			// TODO Probably best just to index the delims in inputData
			b, err := buf.ReadBytes('\n')
			if err != nil {
				break
			}
			inputChan <- workTask{b, lineCount}
			lineCount++
		}
		close(inputChan)
	}()

	// Collect the resultant deltas
	deltas := producer(inputChan, &masterIndex)

	// for each delta process the corresponding line in inputData

	for i, d := range deltas {
		fmt.Printf("Delta %v: %v\n", i, d)
	}
	//	err := writeNewFile(inputData, deltas, flag.Arg(1))
	//	if err != nil {
	//		log.Fatal(err)
	//		return
	//	}

	//	out := csv.NewWriter(os.Stdout)
	//	out.Write(headerRow)
	//	out.Flush()
}
