package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sync"
	"time"
)

// Buffer pool for reusing byte buffers to reduce allocations
var bufferPool = sync.Pool{
	New: func() interface{} {
		return &bytes.Buffer{}
	},
}

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
// Optimized version with buffer pooling to reduce allocations.
func serializeBlock(block *Block) []byte {
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)

	// Pre-allocate buffer capacity to avoid reallocations
	estimatedSize := 1 + 1 + 8 + 8 + 8 + 4 + len(block.Data) + 4 + len(block.PrevHash)
	if buf.Cap() < estimatedSize {
		buf.Grow(estimatedSize - buf.Len())
	}

	buf.WriteByte(0x01)
	if block.explicitlyInitialized {
		buf.WriteByte(0xFF)
	} else {
		buf.WriteByte(0x00)
	}

	binary.Write(buf, binary.LittleEndian, int64(block.Index))
	binary.Write(buf, binary.LittleEndian, int64(block.Timestamp))
	binary.Write(buf, binary.LittleEndian, int64(block.Nonce))

	binary.Write(buf, binary.LittleEndian, int32(len(block.Data)))
	buf.Write(block.Data)

	binary.Write(buf, binary.LittleEndian, int32(len(block.PrevHash)))
	buf.Write(block.PrevHash)

	// Return a copy since we're reusing the buffer
	result := make([]byte, buf.Len())
	copy(result, buf.Bytes())
	return result
}

// calculateHashStreaming computes hash for large blocks using streaming
// to avoid keeping entire serialized block in memory
func calculateHashStreaming(block *Block) []byte {
	hasher := sha256.New()
	
	// Write header data directly to hasher
	hasher.Write([]byte{0x01})
	if block.explicitlyInitialized {
		hasher.Write([]byte{0xFF})
	} else {
		hasher.Write([]byte{0x00})
	}
	
	// Write fixed-size fields
	var tmpBuf [8]byte
	binary.LittleEndian.PutUint64(tmpBuf[:], uint64(block.Index))
	hasher.Write(tmpBuf[:])
	binary.LittleEndian.PutUint64(tmpBuf[:], uint64(block.Timestamp))
	hasher.Write(tmpBuf[:])
	binary.LittleEndian.PutUint64(tmpBuf[:], uint64(block.Nonce))
	hasher.Write(tmpBuf[:])
	
	// Write data length and data
	var lenBuf [4]byte
	binary.LittleEndian.PutUint32(lenBuf[:], uint32(len(block.Data)))
	hasher.Write(lenBuf[:])
	hasher.Write(block.Data)
	
	// Write prev hash length and hash
	binary.LittleEndian.PutUint32(lenBuf[:], uint32(len(block.PrevHash)))
	hasher.Write(lenBuf[:])
	hasher.Write(block.PrevHash)
	
	return hasher.Sum(nil)
}

// calculateHash returns a SHA-256 hash of the serialized block.
// Uses streaming for large blocks to reduce memory usage.
func calculateHash(block *Block) []byte {
	// Use streaming hash for large blocks to reduce memory pressure
	if len(block.Data) > 64*1024 { // 64KB threshold
		return calculateHashStreaming(block)
	}
	
	bytes := serializeBlock(block)
	hash := sha256.Sum256(bytes)
	return hash[:]
}

// proofOfWork finds a valid hash that satisfies the difficulty constraint.
// It returns the discovered hash and the nonce used to generate it.
// Optimized version with reduced string operations and early termination.
func proofOfWork(block *Block, difficulty int) ([]byte, int) {
	// Pre-calculate target for comparison instead of string operations
	target := make([]byte, (difficulty+1)/2)
	if difficulty%2 == 1 {
		target[difficulty/2] = 0xF0
	}
	
	nonce := 0
	var hash []byte
	
	for {
		block.Nonce = nonce
		hash = calculateHash(block)
		
		// Compare bytes directly instead of hex string conversion
		valid := true
		for i := 0; i < difficulty/2; i++ {
			if hash[i] != 0 {
				valid = false
				break
			}
		}
		if valid && difficulty%2 == 1 && hash[difficulty/2] >= 0x10 {
			valid = false
		}
		
		if valid {
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
// Optimized version with better memory management and early exits.
func isChainValidCached(chain []*Block) bool {
	if len(chain) == 0 {
		return true
	}
	
	// Pre-allocate cache with expected capacity
	hashCache := make(map[int][]byte, len(chain))
	
	for i := 1; i < len(chain); i++ {
		prevBlock := chain[i-1]
		currBlock := chain[i]

		// Get or compute previous block hash
		prevHash, ok := hashCache[prevBlock.Index]
		if !ok {
			prevHash = calculateHash(prevBlock)
			hashCache[prevBlock.Index] = prevHash
		}

		// Early exit if previous hash doesn't match
		if !bytes.Equal(currBlock.PrevHash, prevHash) {
			return false
		}

		// Get or compute current block hash
		currHash, ok := hashCache[currBlock.Index]
		if !ok {
			currHash = calculateHash(currBlock)
			hashCache[currBlock.Index] = currHash
		}

		// Early exit if current hash doesn't match
		if !bytes.Equal(currBlock.Hash, currHash) {
			return false
		}
	}
	return true
}

// isChainValidConcurrent validates a chain using concurrent processing
// for better performance on large chains
func isChainValidConcurrent(chain []*Block) bool {
	if len(chain) == 0 {
		return true
	}
	
	// Use concurrent validation for large chains
	if len(chain) < 1000 {
		return isChainValidCached(chain)
	}
	
	// Channel for validation results
	type validationResult struct {
		index int
		valid bool
		err   error
	}
	
	const maxWorkers = 4
	workers := maxWorkers
	if len(chain) < workers {
		workers = len(chain)
	}
	
	results := make(chan validationResult, len(chain)-1)
	hashCache := make(map[int][]byte, len(chain))
	
	// Pre-compute first block hash
	hashCache[0] = calculateHash(chain[0])
	
	// Worker function
	validateBlock := func(i int) {
		prevBlock := chain[i-1]
		currBlock := chain[i]
		
		prevHash := calculateHash(prevBlock)
		currHash := calculateHash(currBlock)
		
		valid := bytes.Equal(currBlock.PrevHash, prevHash) && 
				bytes.Equal(currBlock.Hash, currHash)
		
		results <- validationResult{index: i, valid: valid}
	}
	
	// Launch workers
	for i := 1; i < len(chain); i++ {
		go validateBlock(i)
	}
	
	// Collect results
	for i := 1; i < len(chain); i++ {
		result := <-results
		if !result.valid {
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
	concurrent := flag.Bool("concurrent", false, "use concurrent validation for large chains")
	flag.Parse()

	blockchain := []*Block{newGenesisBlock()}

	fmt.Printf("Generating %d blocks with difficulty %d...\n", *blocks, *difficulty)
	start := time.Now()
	
	for i := 1; i <= *blocks; i++ {
		block := generateBlock(blockchain[len(blockchain)-1], fmt.Sprintf("Block %d", i), *difficulty)
		blockchain = append(blockchain, block)
		if i%100 == 0 || i == *blocks {
			fmt.Printf("Generated %d/%d blocks\n", i, *blocks)
		}
	}
	
	generationTime := time.Since(start)
	fmt.Printf("Generation completed in %v\n", generationTime)

	fmt.Println("\nBlockchain:")
	for _, block := range blockchain {
		fmt.Printf("Index: %d, Data: %s, Hash: %s\n", block.Index, string(block.Data), fmt.Sprintf("%x", block.Hash)[:10]+"...")
	}

	// Validate using appropriate method
	fmt.Print("\nValidating blockchain...")
	validationStart := time.Now()
	var isValid bool
	
	if *concurrent && len(blockchain) >= 1000 {
		isValid = isChainValidConcurrent(blockchain)
		fmt.Printf(" (using concurrent validation)")
	} else {
		isValid = isChainValidCached(blockchain)
		fmt.Printf(" (using cached validation)")
	}
	
	validationTime := time.Since(validationStart)
	fmt.Printf("\nIs blockchain valid? %t (validated in %v)\n", isValid, validationTime)

	if *output != "" {
		if err := writeChainJSON(blockchain, *output); err != nil {
			fmt.Printf("Error writing JSON: %v\n", err)
		} else {
			fmt.Printf("Blockchain written to %s\n", *output)
		}
	}
}
