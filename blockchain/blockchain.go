package blockchain

import (
	"bytes"
	"encoding/json"
	"os"
	"time"

	"my-first-blockchain/block"
	"my-first-blockchain/pow"
)

// GenerateBlock creates a new block referencing the previous one
// and performs proof-of-work to finalize its hash.
func GenerateBlock(prevBlock *block.Block, data string, difficulty int) *block.Block {
	newBlock := &block.Block{
		Index:     prevBlock.Index + 1,
		Timestamp: time.Now().Unix(),
		Data:      []byte(data),
		PrevHash:  prevBlock.Hash,
	}
	hash, nonce := pow.ProofOfWork(newBlock, difficulty)
	newBlock.Hash = hash
	newBlock.Nonce = nonce
	return newBlock
}

// IsChainValidCached validates a chain by caching intermediate hashes
// to avoid redundant hash computations.
func IsChainValidCached(chain []*block.Block) bool {
	hashCache := make(map[int][]byte)
	for i := 1; i < len(chain); i++ {
		prevBlock := chain[i-1]
		currBlock := chain[i]

		prevHash, ok := hashCache[prevBlock.Index]
		if !ok {
			prevHash = block.CalculateHash(prevBlock)
			hashCache[prevBlock.Index] = prevHash
		}

		currHash, ok := hashCache[currBlock.Index]
		if !ok {
			currHash = block.CalculateHash(currBlock)
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

// WriteChainJSON saves the blockchain to a JSON file.
// The file will be overwritten if it already exists.
func WriteChainJSON(chain []*block.Block, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(chain)
}

// NewGenesisBlock returns the first block of the chain.
func NewGenesisBlock() *block.Block {
	b := &block.Block{
		Index:     0,
		Timestamp: time.Now().Unix(),
		Data:      []byte("Genesis"),
		PrevHash:  []byte{},
	}
	b.Hash = block.CalculateHash(b)
	return b
}
