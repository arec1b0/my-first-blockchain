// Fixed version of main.go and main_test.go merged correctly
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
	Index                 int
	Timestamp             int64
	Data                  []byte
	PrevHash              []byte
	Hash                  []byte
	Nonce                 int
	explicitlyInitialized bool
}

func serializeBlock(block *Block) []byte {
	var buf bytes.Buffer

	buf.WriteByte(0x01)
	if block.explicitlyInitialized {
		buf.WriteByte(0xFF)
	} else {
		buf.WriteByte(0x00)
	}

	binary.Write(&buf, binary.LittleEndian, int64(block.Index))
	binary.Write(&buf, binary.LittleEndian, int64(block.Timestamp))
	binary.Write(&buf, binary.LittleEndian, int64(block.Nonce))

	binary.Write(&buf, binary.LittleEndian, int32(len(block.Data)))
	buf.Write(block.Data)

	binary.Write(&buf, binary.LittleEndian, int32(len(block.PrevHash)))
	buf.Write(block.PrevHash)

	return buf.Bytes()
}

func calculateHash(block *Block) []byte {
	bytes := serializeBlock(block)
	hash := sha256.Sum256(bytes)
	return hash[:]
}

func proofOfWork(block *Block, difficulty int) ([]byte, int) {
	prefix := strings.Repeat("0", difficulty)
	nonce := 0
	var hash []byte
	for {
		block.Nonce = nonce
		hash = calculateHash(block)
		if strings.HasPrefix(fmt.Sprintf("%x", hash), prefix) {
			break
		}
		nonce++
	}
	return hash, nonce
}

func generateBlock(prevBlock *Block, data string, difficulty int) *Block {
	newBlock := &Block{
		Index:     prevBlock.Index + 1,
		Timestamp: time.Now().Unix(),
		Data:      []byte(data),
		PrevHash:  prevBlock.Hash,
	}
	hash, nonce := proofOfWork(newBlock, difficulty)
	newBlock.Hash = hash
	newBlock.Nonce = nonce
	return newBlock
}

func isChainValidCached(chain []*Block) bool {
	hashCache := make(map[int][]byte)
	for i := 1; i < len(chain); i++ {
		prevBlock := chain[i-1]
		currBlock := chain[i]

		prevHash, ok := hashCache[prevBlock.Index]
		if !ok {
			prevHash = calculateHash(prevBlock)
			hashCache[prevBlock.Index] = prevHash
		}

		currHash, ok := hashCache[currBlock.Index]
		if !ok {
			currHash = calculateHash(currBlock)
			hashCache[currBlock.Index] = currHash
		}

		if !bytes.Equal(currBlock.PrevHash, prevHash) {
			return false
		}
		if !bytes.Equal(currBlock.Hash, currHash) {
			return false
		}
	}
	return true
}

func main() {
	genesisBlock := &Block{
		Index:     0,
		Timestamp: time.Now().Unix(),
		Data:      []byte("Genesis"),
		PrevHash:  []byte{},
	}
	genesisBlock.Hash = calculateHash(genesisBlock)

	blockchain := []*Block{genesisBlock}
	difficulty := 4

	blockchain = append(blockchain, generateBlock(blockchain[len(blockchain)-1], "Second Block", difficulty))
	blockchain = append(blockchain, generateBlock(blockchain[len(blockchain)-1], "Third Block", difficulty))

	fmt.Println("Blockchain:")
	for _, block := range blockchain {
		fmt.Printf("Index: %d, Data: %s, Hash: %s\n", block.Index, string(block.Data), fmt.Sprintf("%x", block.Hash)[:10]+"...")
	}

	fmt.Printf("\nIs blockchain valid? %t\n", isChainValidCached(blockchain))
}
