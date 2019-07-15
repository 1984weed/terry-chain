package main

import (
	"fmt"
	"time"
)

type Block struct {
	Index     int
	Previous  string
	Timestamp int64
	data      interface{}
	hash      string
}

func generateNextBlock(blockData string) {
	previousBlock := getltestBlock()
	nextIndex := previousBlock.Index + 1
	newxtTimestamp = time.Now()
}

func main() {
	fmt.Print("======")
}
