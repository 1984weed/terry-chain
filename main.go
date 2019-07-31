package main

import (
	"fmt"
)

type Block struct {
	Index     int
	Previous  string
	Timestamp int64
	Data      string
	Hash      string
}

func NewBlock(index int, hash, previousHash string, timestamp int64, data string) *Block {
	return &Block{
		Index:     index,
		Previous:  previousHash,
		Timestamp: timestamp,
		Data:      data,
		Hash:      hash,
	}
}

func CalculateHash (index: number, previousHash: string, timestamp: number, data: string): string {
	return CryptoJS.SHA256(index + previousHash + timestamp + data).toString();
}


// func generateNextBlock(blockData string) {
// 	previousBlock := getltestBlock()
// 	nextIndex := previousBlock.Index + 1
// 	newxtTimestamp = time.Now()
// }

func main() {
	fmt.Print("======")
}
