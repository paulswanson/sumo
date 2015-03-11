package main

import (
	"bytes"
	"encoding/csv"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

func main() {

	flag.Parse()
	if flag.NArg() < 3 {
		fmt.Println("Error, please specifiy: index input output")
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

	var masterIndex replaceIndex = make(replaceIndex, 0)
	c := csv.NewReader(indexFile)

	fmt.Printf("Master index: Building\n")
	// TODO Need to reinstate error checking properly
	masterIndex.init(c)
	//	indexSize, err := masterIndex.init(c)
	//	if err != nil {
	//		log.Fatal(err)
	//		return
	//	}
	fmt.Printf("Master index: Done\n")

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
	go func() {
		var lineCount int
		for {
			b, err := buf.ReadBytes('\n')
			if err != nil { // TODO Investigate final return / EOF scenario
				break
			}
			inputChan <- workTask{b, lineCount}
			lineCount++
		}
		close(inputChan)
	}()

	var deltas []delta
	deltas = producer(inputChan, masterIndex)

	fmt.Printf("Final deltas: %v\n", deltas)
	//	err := writeNewFile(inputData, deltas, flag.Arg(1))
	//	if err != nil {
	//		log.Fatal(err)
	//		return
	//	}

	//	out := csv.NewWriter(os.Stdout)
	//	out.Write(headerRow)
	//	out.Flush()
}
