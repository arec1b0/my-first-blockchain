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

func main() {
	genesisBlock := Block{0, time.Now().String(), "Genesis", "", ""}
	genesisBlock.Hash = calculateHash(genesisBlock)

	blockchain := []Block{genesisBlock}
	blockchain = append(blockchain, generateBlock(blockchain[len(blockchain)-1], "Second Block"))

	for _, block := range blockchain {
		fmt.Printf("%+v\n", block)
	}
}
