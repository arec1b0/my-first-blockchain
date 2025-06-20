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
	// explicitlyInitialized helps distinguish between zero-value and explicitly initialized blocks
	explicitlyInitialized bool
}

// serializeBlock converts a Block to a deterministic byte array,
// with special handling for zero values vs. explicitly set empty fields
func serializeBlock(block *Block) []byte {
	var buf bytes.Buffer

	// Add a version byte for future compatibility
	buf.WriteByte(0x01)

	// Write the explicitlyInitialized flag first
	if block.explicitlyInitialized {
		buf.WriteByte(0xFF)
	} else {
		buf.WriteByte(0x00)
	}

	// Write field values with type information
	binary.Write(&buf, binary.LittleEndian, int64(block.Index))
	binary.Write(&buf, binary.LittleEndian, int64(block.Timestamp))
	binary.Write(&buf, binary.LittleEndian, int64(block.Nonce))

	// For strings, write length prefix and data
	binary.Write(&buf, binary.LittleEndian, int32(len(block.Data)))
	buf.WriteString(block.Data)

	binary.Write(&buf, binary.LittleEndian, int32(len(block.PrevHash)))
	buf.WriteString(block.PrevHash)

	return buf.Bytes()
}

func calculateHash(block *Block) string {
	// Get deterministic serialization
	bytes := serializeBlock(block)

	// Hash the serialized data
	hash := sha256.Sum256(bytes)
	return fmt.Sprintf("%x", hash)
}

// Генерация нового блока с Proof-of-Work
func generateBlock(prevBlock *Block, data string, difficulty int) *Block {
	newBlock := &Block{
		Index:     prevBlock.Index + 1,
		Timestamp: time.Now().Unix(),
		Data:      data,
		PrevHash:  prevBlock.Hash,
	}

	hash, nonce := proofOfWork(newBlock, difficulty)
	newBlock.Hash = hash
	newBlock.Nonce = nonce
	return newBlock
}

// Валидация одного блока
func isBlockValid(newBlock, prevBlock *Block) bool {
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

// Валидация всей цепочки
func isChainValid(chain *[]Block) bool {
	for i := 1; i < len(*chain); i++ {
		if !isBlockValid(&(*chain)[i], &(*chain)[i-1]) {
			return false
		}
	}
	return true
}

// Proof-of-Work
func proofOfWork(block *Block, difficulty int) (string, int) {
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
	// Создание genesis-блока
	genesisBlock := &Block{
		Index:     0,
		Timestamp: time.Now().Unix(),
		Data:      "Genesis",
		PrevHash:  "",
	}
	genesisBlock.Hash = calculateHash(genesisBlock)

	// Инициализация цепочки
	blockchain := []Block{*genesisBlock}
	difficulty := 4

	// Добавление блоков
	blockchain = append(blockchain, *generateBlock(&blockchain[len(blockchain)-1], "Second Block", difficulty))
	blockchain = append(blockchain, *generateBlock(&blockchain[len(blockchain)-1], "Third Block", difficulty))

	// Вывод блоков
	fmt.Println("Blockchain:")
	for _, block := range blockchain {
		fmt.Printf("Index: %d, Data: %s, Hash: %s\n", block.Index, block.Data, block.Hash[:10]+"...")
	}

	// Валидация цепочки
	fmt.Printf("\nIs blockchain valid? %t\n", isChainValid(&blockchain))
}
