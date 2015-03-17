package main

import (
	"bytes"
	"encoding/csv"
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

	// fmt.Printf("Master index: Building\n")

	masterIndex, err := newReplaceIndex(csv.NewReader(indexFile))
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

	// fmt.Printf("Master index: Done\n")

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

	buf := bytes.NewBuffer(inputData)
	inputChan := make(chan workTask)
	var lineCount int

	// Feed the producer line at a time
	go func() {
		for {
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
	fmt.Printf("Final deltas: %v\n", deltas)

	buf = bytes.NewBuffer(inputData)

	// Output the file, applying the deltas
	for i := 0; i < len(inputData); i++ {
		b, err := buf.ReadBytes('\n')
		if err != nil {
			break
		}
		// TODO Apply delta to []byte and output
		fmt.Sprintf("%v\n", b)
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
