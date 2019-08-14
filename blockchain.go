package main

import (
	"crypto/sha256"
	"fmt"
	"time"
)

type Block struct {
	Index        int         `json:"index,omitempty"`
	PreviousHash string      `json:"previousHash,omitempty"`
	Timestamp    int64       `json:"timestamp,omitempty"`
	Data         interface{} `json:"data,omitempty"`
	Hash         string      `json:"hash,omitempty"`
}

func GenerageBlock(index int, previousHash string, timestamp int64, data interface{}, hash string) *Block {
	return &Block{
		Index:        index,
		Hash:         hash,
		PreviousHash: previousHash,
		Timestamp:    timestamp,
		Data:         data,
	}
}

func generateNextBlock(blockData string) *Block {
	previousBlock := getLatestBlock()
	nextIndex := previousBlock.Index + 1
	nextTimestamp := time.Now()
	nextHash := calculateHash(nextIndex, previousBlock.Hash, nextTimestamp.Unix(), blockData)

	return &Block{
		Index:        nextIndex,
		Hash:         nextHash,
		PreviousHash: previousBlock.Hash,
		Timestamp:    nextTimestamp.Unix(),
		Data:         blockData,
	}
}

func calculateHash(index int, prevHash string, nextTimestamp int64, blockData string) string {
	h := sha256.New()

	s := fmt.Sprintf("%s%s%d%s", index, prevHash, nextTimestamp, blockData)
	h.Write([]byte(s))

	// This gets the finalized hash result as a byte
	// slice. The argument to `Sum` can be used to append
	// to an existing byte slice: it usually isn't needed.
	bs := h.Sum(nil)

	return string(bs)
}

var genesisBlock = GenerageBlock(0, "816534932c2b7154836da6afc367695e6337db8a921823784c14378abed4f7d7", 1465154705, nil, "my genesis block!!")
var blockChain = []Block{*genesisBlock}

func isValidNewBlock(newBlock Block, previousBlock Block) bool {
	if previousBlock.Index+1 != newBlock.Index {
		return false
	} else if previousBlock.Hash != newBlock.PreviousHash {
		return false
	} else if calculateHashForBlock(newBlock) != newBlock.Hash {
		return false
	}
	return true
}

func replaceChain(newBlocks []Block) {
	if isValidChain(newBlocks) && len(newBlocks) > len(GetBlockChain()) {
		blockChain = newBlocks
		// broadcastLatest()
	} else {
		fmt.Println("Received blockchain invalid")
	}
}

func isValidChain(blockchainToValidate []Block) bool {
	isValidGenesis := func(block Block) bool {
		return block == *genesisBlock
	}

	if !isValidGenesis(blockchainToValidate[0]) {
		return false
	}

	for i := 1; i < len(blockchainToValidate); i++ {
		if !isValidNewBlock(blockchainToValidate[i], blockchainToValidate[i-1]) {
			return false
		}
	}
	return true
}

func GetBlockChain() []Block {
	return blockChain
}

func getLatestBlock() Block {
	return blockChain[len(blockChain)-1]
}

func calculateHashForBlock(block Block) string {
	return calculateHash(block.Index, block.PreviousHash, block.Timestamp, fmt.Sprintf("%v", block.Data))
}

func addBlockToChain(newBlock Block) bool {
	if isValidNewBlock(newBlock, getLatestBlock()) {
		blockChain = append(blockChain, newBlock)
		return true
	}
	return false
}
