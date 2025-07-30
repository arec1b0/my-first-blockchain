package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
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
}

// ValidationResult represents the result of block validation
type ValidationResult struct {
	Index int
	Valid bool
	Error error
}

// HashCache provides thread-safe hash caching
type HashCache struct {
	mu    sync.RWMutex
	cache map[int][]byte
}

// NewHashCache creates a new thread-safe hash cache
func NewHashCache(capacity int) *HashCache {
	return &HashCache{
		cache: make(map[int][]byte, capacity),
	}
}

// Get retrieves a hash from cache
func (hc *HashCache) Get(index int) ([]byte, bool) {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	hash, exists := hc.cache[index]
	if exists {
		// Return a copy to prevent modification
		result := make([]byte, len(hash))
		copy(result, hash)
		return result, true
	}
	return nil, false
}

// Set stores a hash in cache
func (hc *HashCache) Set(index int, hash []byte) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	// Store a copy to prevent external modification
	hashCopy := make([]byte, len(hash))
	copy(hashCopy, hash)
	hc.cache[index] = hashCopy
}

// serializeBlockHeader serializes the block header without data for efficiency
func serializeBlockHeader(block *Block, buf *bytes.Buffer) {
	buf.WriteByte(0x01) // Version marker
	buf.WriteByte(0x00) // Reserved for future use
	
	binary.Write(buf, binary.LittleEndian, int64(block.Index))
	binary.Write(buf, binary.LittleEndian, int64(block.Timestamp))
	binary.Write(buf, binary.LittleEndian, int64(block.Nonce))
}

// serializeBlock converts a block into a deterministic byte slice.
// The format is intentionally simple to avoid ambiguities when hashing.
// Optimized version with buffer pooling to reduce allocations.
func serializeBlock(block *Block) []byte {
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)

	// Pre-allocate buffer capacity to avoid reallocations
	estimatedSize := 32 + len(block.Data) + len(block.PrevHash) // Conservative estimate
	if buf.Cap() < estimatedSize {
		buf.Grow(estimatedSize - buf.Len())
	}

	serializeBlockHeader(block, buf)

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
	hasher.Write([]byte{0x01, 0x00}) // Version and reserved byte
	
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

// validateDifficulty checks if a hash meets the difficulty requirement
func validateDifficulty(hash []byte, difficulty int) bool {
	// Check whole bytes first (more efficient)
	wholeBytes := difficulty / 2
	for i := 0; i < wholeBytes; i++ {
		if hash[i] != 0 {
			return false
		}
	}
	
	// Check remaining nibble if odd difficulty
	if difficulty%2 == 1 && wholeBytes < len(hash) {
		return hash[wholeBytes] < 0x10
	}
	
	return true
}

// proofOfWork finds a valid hash that satisfies the difficulty constraint.
// It returns the discovered hash and the nonce used to generate it.
// Supports cancellation via context.
func proofOfWork(ctx context.Context, block *Block, difficulty int) ([]byte, int, error) {
	if difficulty < 0 || difficulty > 64 {
		return nil, 0, errors.New("invalid difficulty level")
	}
	
	nonce := 0
	var hash []byte
	
	// Check for cancellation every 1000 iterations to avoid overhead
	const checkInterval = 1000
	
	for {
		// Check for cancellation periodically
		if nonce%checkInterval == 0 {
			select {
			case <-ctx.Done():
				return nil, 0, ctx.Err()
			default:
			}
		}
		
		block.Nonce = nonce
		hash = calculateHash(block)
		
		if validateDifficulty(hash, difficulty) {
			return hash, nonce, nil
		}
		nonce++
	}
}

// generateBlock creates a new block referencing the previous one
// and performs proof-of-work to finalize its hash.
func generateBlock(ctx context.Context, prevBlock *Block, data string, difficulty int) (*Block, error) {
	newBlock := &Block{
		Index:     prevBlock.Index + 1,
		Timestamp: time.Now().Unix(),
		Data:      []byte(data),
		PrevHash:  prevBlock.Hash,
	}
	
	hash, nonce, err := proofOfWork(ctx, newBlock, difficulty)
	if err != nil {
		return nil, fmt.Errorf("proof of work failed: %w", err)
	}
	
	newBlock.Hash = hash
	newBlock.Nonce = nonce
	return newBlock, nil
}

// validateBlockPair validates a single block against its predecessor
func validateBlockPair(prevBlock, currBlock *Block, hashCache *HashCache) error {
	// Get or compute previous block hash
	prevHash, ok := hashCache.Get(prevBlock.Index)
	if !ok {
		prevHash = calculateHash(prevBlock)
		hashCache.Set(prevBlock.Index, prevHash)
	}

	// Check previous hash link
	if !bytes.Equal(currBlock.PrevHash, prevHash) {
		return fmt.Errorf("block %d: invalid previous hash", currBlock.Index)
	}

	// Get or compute current block hash
	currHash, ok := hashCache.Get(currBlock.Index)
	if !ok {
		currHash = calculateHash(currBlock)
		hashCache.Set(currBlock.Index, currHash)
	}

	// Check current hash
	if !bytes.Equal(currBlock.Hash, currHash) {
		return fmt.Errorf("block %d: invalid hash", currBlock.Index)
	}

	return nil
}

// isChainValidCached validates a chain by caching intermediate hashes
// to avoid redundant hash computations.
// Optimized version with better memory management and early exits.
func isChainValidCached(chain []*Block) bool {
	if len(chain) == 0 {
		return true
	}
	
	hashCache := NewHashCache(len(chain))
	
	for i := 1; i < len(chain); i++ {
		if err := validateBlockPair(chain[i-1], chain[i], hashCache); err != nil {
			return false
		}
	}
	return true
}

// validateChainConcurrent validates blocks concurrently with proper error handling
func validateChainConcurrent(ctx context.Context, chain []*Block, maxWorkers int) error {
	if len(chain) == 0 {
		return nil
	}
	
	if len(chain) < maxWorkers {
		maxWorkers = len(chain) - 1
	}
	
	if maxWorkers <= 0 {
		return nil
	}
	
	// Channel for validation jobs and results
	jobs := make(chan int, len(chain)-1)
	results := make(chan ValidationResult, len(chain)-1)
	
	hashCache := NewHashCache(len(chain))
	
	// Worker function
	worker := func() {
		for i := range jobs {
			select {
			case <-ctx.Done():
				results <- ValidationResult{Index: i, Valid: false, Error: ctx.Err()}
				return
			default:
			}
			
			err := validateBlockPair(chain[i-1], chain[i], hashCache)
			results <- ValidationResult{
				Index: i,
				Valid: err == nil,
				Error: err,
			}
		}
	}
	
	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			worker()
		}()
	}
	
	// Send jobs
	go func() {
		defer close(jobs)
		for i := 1; i < len(chain); i++ {
			select {
			case jobs <- i:
			case <-ctx.Done():
				return
			}
		}
	}()
	
	// Wait for workers to finish
	go func() {
		wg.Wait()
		close(results)
	}()
	
	// Collect results
	for i := 1; i < len(chain); i++ {
		select {
		case result := <-results:
			if !result.Valid {
				return result.Error
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	
	return nil
}

// isChainValidConcurrent validates a chain using concurrent processing
// for better performance on large chains
func isChainValidConcurrent(ctx context.Context, chain []*Block) bool {
	// Use concurrent validation for large chains
	if len(chain) < 1000 {
		return isChainValidCached(chain)
	}
	
	const maxWorkers = 4
	err := validateChainConcurrent(ctx, chain, maxWorkers)
	return err == nil
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
	timeout := flag.Duration("timeout", 30*time.Minute, "timeout for long-running operations")
	flag.Parse()

	// Validate input parameters
	if *blocks < 0 {
		fmt.Printf("Error: blocks must be non-negative\n")
		os.Exit(1)
	}
	if *difficulty < 0 || *difficulty > 32 {
		fmt.Printf("Error: difficulty must be between 0 and 32\n")
		os.Exit(1)
	}

	blockchain := []*Block{newGenesisBlock()}

	fmt.Printf("Generating %d blocks with difficulty %d (timeout: %v)...\n", *blocks, *difficulty, *timeout)
	start := time.Now()
	
	// Create context with timeout for cancellation
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()
	
	for i := 1; i <= *blocks; i++ {
		block, err := generateBlock(ctx, blockchain[len(blockchain)-1], fmt.Sprintf("Block %d", i), *difficulty)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				fmt.Printf("Timeout exceeded while generating block %d\n", i)
			} else {
				fmt.Printf("Error generating block %d: %v\n", i, err)
			}
			os.Exit(1)
		}
		blockchain = append(blockchain, block)
		if i%100 == 0 || i == *blocks {
			fmt.Printf("Generated %d/%d blocks\n", i, *blocks)
		}
	}
	
	generationTime := time.Since(start)
	fmt.Printf("Generation completed in %v (avg: %v per block)\n", 
		generationTime, generationTime/time.Duration(*blocks))

	fmt.Println("\nBlockchain:")
	displayLimit := 10
	if len(blockchain) > displayLimit {
		fmt.Printf("Showing first %d and last %d blocks:\n", displayLimit/2, displayLimit/2)
		for _, block := range blockchain[:displayLimit/2] {
			fmt.Printf("Index: %d, Data: %s, Hash: %s\n", 
				block.Index, string(block.Data), fmt.Sprintf("%x", block.Hash)[:10]+"...")
		}
		fmt.Printf("... (%d blocks omitted) ...\n", len(blockchain)-displayLimit)
		for _, block := range blockchain[len(blockchain)-displayLimit/2:] {
			fmt.Printf("Index: %d, Data: %s, Hash: %s\n", 
				block.Index, string(block.Data), fmt.Sprintf("%x", block.Hash)[:10]+"...")
		}
	} else {
		for _, block := range blockchain {
			fmt.Printf("Index: %d, Data: %s, Hash: %s\n", 
				block.Index, string(block.Data), fmt.Sprintf("%x", block.Hash)[:10]+"...")
		}
	}

	// Validate using appropriate method
	fmt.Print("\nValidating blockchain...")
	validationStart := time.Now()
	var isValid bool
	
	// Create new context for validation (separate from generation timeout)
	validationCtx, validationCancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer validationCancel()
	
	if *concurrent && len(blockchain) >= 1000 {
		isValid = isChainValidConcurrent(validationCtx, blockchain)
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
			os.Exit(1)
		} else {
			fmt.Printf("Blockchain written to %s\n", *output)
		}
	}

	// Performance summary
	fmt.Printf("\nPerformance Summary:\n")
	fmt.Printf("- Total blocks: %d\n", len(blockchain))
	fmt.Printf("- Average generation time: %v/block\n", generationTime/time.Duration(*blocks))
	fmt.Printf("- Validation time: %v\n", validationTime)
	fmt.Printf("- Total runtime: %v\n", time.Since(start))
}
