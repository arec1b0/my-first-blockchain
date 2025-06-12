package main

import (
	"crypto/sha256"
	"fmt"
	"time"
)

type Block struct {
	Index     int
	Timestamp string
	Data      string
	PrevHash  string
	Hash      string
}

func calculateHash(block Block) string {
	record := fmt.Sprintf("%d%s%s%s", block.Index, block.Timestamp, block.Data, block.PrevHash)
	return fmt.Sprintf("%x", sha256.Sum256([]byte(record)))
}

func generateBlock(prevBlock Block, data string) Block {
	newBlock := Block{
		Index:     prevBlock.Index + 1,
		Timestamp: time.Now().String(),
		Data:      data,
		PrevHash:  prevBlock.Hash,
	}
	newBlock.Hash = calculateHash(newBlock)
	return newBlock
}

func IsBlockValid(newBlock, prevBlock Block) bool {
	if newBlock.Index != prevBlock.Index+1 {
		return false
	}
	if newBlock.PrevHash != prevBlock.Hash {
		return false
	}
	if calculateHash(newBlock) != newBlock.Hash {
		return false
	}
	return true // Fixed: should return true if all validations pass
}

func IsChainValid(chain []Block) bool {
	for i := 1; i < len(chain); i++ {
		if !IsBlockValid(chain[i], chain[i-1]) { // Fixed: negate the condition
			return false
		}
	}
	return true // Fixed: return true if all blocks are valid
}

func main() {
	// Create genesis block
	genesisBlock := Block{0, time.Now().String(), "Genesis", "", ""}
	genesisBlock.Hash = calculateHash(genesisBlock)

	// Initialize blockchain
	blockchain := []Block{genesisBlock}

	// Add more blocks
	blockchain = append(blockchain, generateBlock(blockchain[len(blockchain)-1], "Second Block"))
	blockchain = append(blockchain, generateBlock(blockchain[len(blockchain)-1], "Third Block"))

	// Display blockchain
	fmt.Println("Blockchain:")
	for _, block := range blockchain {
		fmt.Printf("Index: %d, Data: %s, Hash: %s\n", block.Index, block.Data, block.Hash[:10]+"...")
	}

	// Validate blockchain
	fmt.Printf("\nIs blockchain valid? %t\n", IsChainValid(blockchain))
}
