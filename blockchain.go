package main

import (
	"crypto/sha256"
	"fmt"
	"math"
	"strings"
	"time"
)

const (
	// in seconds
	blockGenerationInterval = 10
	// in blocks
	difficultyAdjustmentInterval = 10
)

func getDifficulty(aBlockChain []Block) int {
	latestBlock := aBlockChain[len(aBlockChain)-1]
	if latestBlock.Index%difficultyAdjustmentInterval == 0 && latestBlock.Index != 0 {
		return getAdjustedDifficulty(latestBlock, aBlockChain)
	}

	return latestBlock.Difficulty
}

func getAdjustedDifficulty(latestBlock Block, aBlockchain []Block) int {
	prevAdjustmentBlock := aBlockchain[len(blockChain)-difficultyAdjustmentInterval]
	timeExpected := int64(blockGenerationInterval * difficultyAdjustmentInterval)
	timeTaken := latestBlock.Timestamp - prevAdjustmentBlock.Timestamp

	if timeTaken < timeExpected/2 {
		return prevAdjustmentBlock.Difficulty + 1
	}
	if timeTaken > timeExpected*2 {
		return prevAdjustmentBlock.Difficulty - 1
	}
	return prevAdjustmentBlock.Difficulty
}

func getCurrentTimestamp() int64 {
	now := time.Now()
	return int64(math.Round(float64(now.Unix() / 1000)))
}

// Block is a one of chain
type Block struct {
	Index        int    `json:"index,omitempty"`
	PreviousHash string `json:"previousHash,omitempty"`
	Timestamp    int64  `json:"timestamp,omitempty"`
	Data         string `json:"data,omitempty"`
	Hash         string `json:"hash,omitempty"`
	Difficulty   int    `json:"difficulty"`
	Nonce        int    `json:"nonce"`
}

// GenerageBlock generates a Block by information
func GenerageBlock(index int, previousHash string, timestamp int64, data string, hash string, difficulty int, nonce int) *Block {
	return &Block{
		Index:        index,
		Hash:         hash,
		PreviousHash: previousHash,
		Timestamp:    timestamp,
		Data:         data,
		Difficulty:   difficulty,
		Nonce:        nonce,
	}
}

// GenerateNextBlock generates a next Block
func GenerateNextBlock(blockData string) *Block {
	previousBlock := GetLatestBlock()
	difficulty := getDifficulty(GetBlockChain())

	nextIndex := previousBlock.Index + 1
	nextTimestamp := getCurrentTimestamp()

	newBlock := findBlock(nextIndex, previousBlock.Hash, nextTimestamp, blockData, difficulty)

	addBlockToChain(*newBlock)

	// Broadcast
	return newBlock
}

func findBlock(index int, previousHash string, timestamp int64, data string, difficulty int) *Block {
	nonce := 0
	for true {
		hash := calculateHash(index, previousHash, timestamp, data, difficulty, nonce)
		if hashMatchesDifficulty(hash, difficulty) {
			return GenerageBlock(index, previousHash, timestamp, data, hash, difficulty, nonce)
		}
		nonce++
	}
	return nil
}

func getAccumulatedDifficulty(aBlockchain []Block) float64 {
	result := float64(0)
	for _, block := range aBlockchain {
		result += math.Pow(2, float64(block.Difficulty))
	}

	return result
}

func isValidTimestamp(newBlock Block, previousBlock Block) bool {
	return (previousBlock.Timestamp-60 < newBlock.Timestamp) && newBlock.Timestamp-60 < getCurrentTimestamp()
}

func calculateHash(index int, prevHash string, nextTimestamp int64, blockData string, difficulty, nonce int) string {
	h := sha256.New()

	s := fmt.Sprintf("%d%s%d%s%d%d", index, prevHash, nextTimestamp, blockData, difficulty, nonce)
	h.Write([]byte(s))

	bs := h.Sum(nil)

	return string(bs)
}

func hashMatchesDifficulty(hash string, difficulty int) bool {
	hashInBinary := hexToBinary(hash)
	requiredPrefix := ""
	for i := 0; i < difficulty; i++ {
		requiredPrefix += "0"
	}

	return strings.HasPrefix(hashInBinary, requiredPrefix) // true
}

var genesisBlock = GenerageBlock(0, "816534932c2b7154836da6afc367695e6337db8a921823784c14378abed4f7d7", 1465154705, "", "my genesis block!!", 0, 0)
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

// ReplaceChain handle whether to relace to new chain or ignore new chain
func ReplaceChain(newBlocks []Block) {
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

// GetBlockChain gets blockchain
func GetBlockChain() []Block {
	return blockChain
}

// GetLatestBlock gets the latest block in chain
func GetLatestBlock() Block {
	return blockChain[len(blockChain)-1]
}

func calculateHashForBlock(block Block) string {
	return calculateHash(block.Index, block.PreviousHash, block.Timestamp, fmt.Sprintf("%v", block.Data), block.Difficulty, block.Nonce)
}

func addBlockToChain(newBlock Block) bool {
	if isValidNewBlock(newBlock, GetLatestBlock()) {
		blockChain = append(blockChain, newBlock)
		return true
	}
	return false
}

// // the unspent txOut of genesis block is set to unspentTxOuts on startup
var unspentTxOuts []UnspentTxOut = ProcessTransactions(blockchain[0].data, []UnspentTxOut{}, 0)

// const getBlockchain = (): Block[] => blockchain;

func getUnspentTxOuts() []UnspentTxOut {
	b := append(unspentTxOuts[:0:0], unspentTxOuts...)
	return b
}

// const getUnspentTxOuts = (): UnspentTxOut[] => _.cloneDeep(unspentTxOuts);

func sendTransaction(address string, amount int) *Transaction {
	wallet, err := GetPrivateFromWallet()

	if err != nil {
		return nil
	}

	tx, err := createTransaction(address, amount, wallet, getUnspentTxOuts(), getTransactionPool())

	if err != nil {
		return nil
	}

	addToTransactionPool(tx, getUnspentTxOuts())

	return tx
}

// const sendTransaction = (address: string, amount: number): Transaction => {
//     const tx: Transaction = createTransaction(address, amount, getPrivateFromWallet(), getUnspentTxOuts(), getTransactionPool());
//     addToTransactionPool(tx, getUnspentTxOuts());
//     broadCastTransactionPool();
//     return tx;
// };
