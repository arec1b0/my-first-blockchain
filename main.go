package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"strings"
	"time"
)

type Block struct {
	Index     int
	Timestamp int64
	Data      string
	PrevHash  string
	Hash      string
	Nonce     int
}

func calculateHash(block Block) string {
	var buf bytes.Buffer

	binary.Write(&buf, binary.LittleEndian, int64(block.Index))
	binary.Write(&buf, binary.LittleEndian, int64(block.Timestamp))
	binary.Write(&buf, binary.LittleEndian, int64(block.Nonce))
	binary.Write(&buf, binary.LittleEndian, int32(len(block.Data)))
	buf.WriteString(block.Data)
	binary.Write(&buf, binary.LittleEndian, int32(len(block.PrevHash)))
	buf.WriteString(block.PrevHash)

	hash := sha256.Sum256(buf.Bytes())
	return fmt.Sprintf("%x", hash)
}

func generateBlock(prevBlock Block, data string, difficulty int) Block {
	newBlock := Block{
		Index:     prevBlock.Index + 1,
		Timestamp: time.Now().Unix(),
		Data:      data,
		PrevHash:  prevBlock.Hash,
	}
	hash, nonce := ProofOfWork(newBlock, difficulty)
	newBlock.Hash = hash
	newBlock.Nonce = nonce
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
	return true 
}

func IsChainValid(chain []Block) bool {
	for i := 1; i < len(chain); i++ {
		if !IsBlockValid(chain[i], chain[i-1]) { 
			return false
		}
	}
	return true 
}

func ProofOfWork(block Block, difficulty int) (string, int) {
	prefix := strings.Repeat("0", difficulty)
	nonce := 0
	var hash string

	for {
		block.Nonce = nonce
		hash = calculateHash(block)
		if strings.HasPrefix(hash, prefix) {
			break
		}
		nonce++
	}
	return hash, nonce
}

func main() {
	// Create genesis block
	genesisBlock := Block{
		Index:     0,
		Timestamp: time.Now().Unix(),
		Data:      "Genesis",
		Nonce:     0,
	}
	genesisBlock.Hash = calculateHash(genesisBlock)

	// Initialize blockchain
	blockchain := []Block{genesisBlock}

	// Difficulty level for PoW
	difficulty := 4

	// Add more blocks
	blockchain = append(blockchain, generateBlock(blockchain[len(blockchain)-1], "Second Block", difficulty))
	blockchain = append(blockchain, generateBlock(blockchain[len(blockchain)-1], "Third Block", difficulty))

	// Display blockchain
	fmt.Println("Blockchain:")
	for _, block := range blockchain {
		fmt.Printf("Index: %d, Data: %s, Hash: %s\n", block.Index, block.Data, block.Hash[:10]+"...")
	}

	// Validate blockchain
	fmt.Printf("\nIs blockchain valid? %t\n", IsChainValid(blockchain))
}
