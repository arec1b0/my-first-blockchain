package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

// Block represents a single record in the blockchain.
// Fields are kept in raw byte form to avoid encoding pitfalls.
type Block struct {
	Index     int    `json:"index"`
	Timestamp int64  `json:"timestamp"`
	Data      []byte `json:"data"`
	PrevHash  []byte `json:"prev_hash"`
	Hash      []byte `json:"hash"`
	Nonce     int    `json:"nonce"`

	explicitlyInitialized bool `json:"-"`
}

// serializeBlock converts a block into a deterministic byte slice.
// The format is intentionally simple to avoid ambiguities when hashing.
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

// calculateHash returns a SHA-256 hash of the serialized block.
func calculateHash(block *Block) []byte {
	bytes := serializeBlock(block)
	hash := sha256.Sum256(bytes)
	return hash[:]
}

// proofOfWork finds a valid hash that satisfies the difficulty constraint.
// It returns the discovered hash and the nonce used to generate it.
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

// generateBlock creates a new block referencing the previous one
// and performs proof-of-work to finalize its hash.
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

// isChainValidCached validates a chain by caching intermediate hashes
// to avoid redundant hash computations.
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

// writeChainJSON saves the blockchain to a JSON file.
// The file will be overwritten if it already exists.
func writeChainJSON(chain []*Block, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(chain)
}

// newGenesisBlock returns the first block of the chain.
func newGenesisBlock() *Block {
	b := &Block{
		Index:     0,
		Timestamp: time.Now().Unix(),
		Data:      []byte("Genesis"),
		PrevHash:  []byte{},
	}
	b.Hash = calculateHash(b)
	return b
}

// main demonstrates block creation and chain validation.
func main() {
	blocks := flag.Int("blocks", 2, "number of additional blocks to generate")
	difficulty := flag.Int("difficulty", 4, "proof-of-work difficulty")
	output := flag.String("output", "", "optional path to write blockchain as JSON")
	flag.Parse()

	blockchain := []*Block{newGenesisBlock()}

	for i := 1; i <= *blocks; i++ {
		block := generateBlock(blockchain[len(blockchain)-1], fmt.Sprintf("Block %d", i), *difficulty)
		blockchain = append(blockchain, block)
	}

	fmt.Println("Blockchain:")
	for _, block := range blockchain {
		fmt.Printf("Index: %d, Data: %s, Hash: %s\n", block.Index, string(block.Data), fmt.Sprintf("%x", block.Hash)[:10]+"...")
	}

	fmt.Printf("\nIs blockchain valid? %t\n", isChainValidCached(blockchain))

	if *output != "" {
		if err := writeChainJSON(blockchain, *output); err != nil {
			fmt.Printf("Error writing JSON: %v\n", err)
		} else {
			fmt.Printf("Blockchain written to %s\n", *output)
		}
	}
}
