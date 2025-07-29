package block_test

import (
	"bytes"
	"fmt"
	"testing"

	block "my-first-blockchain/block"
	bc "my-first-blockchain/blockchain"
)

// BenchmarkStressGenerateBlockDifficulty4 measures PoW generation with higher difficulty.
func BenchmarkStressGenerateBlockDifficulty4(b *testing.B) {
	prev := &block.Block{Hash: []byte("prev")}
	for i := 0; i < b.N; i++ {
		_ = bc.GenerateBlock(prev, fmt.Sprintf("data-%d", i), 4)
	}
}

// BenchmarkStressGenerateBlockLargeData benchmarks creating blocks with large payloads.
func BenchmarkStressGenerateBlockLargeData(b *testing.B) {
	prev := &block.Block{Hash: []byte("prev")}
	data := bytes.Repeat([]byte("a"), 1<<20) // 1 MB
	for i := 0; i < b.N; i++ {
		_ = bc.GenerateBlock(prev, string(data), 1)
	}
}

// BenchmarkStressCalculateHashLargeData hashes a block with a large data field repeatedly.
func BenchmarkStressCalculateHashLargeData(b *testing.B) {
	blk := &block.Block{Data: bytes.Repeat([]byte("a"), 512*1024)} // 512 KB
	for i := 0; i < b.N; i++ {
		block.CalculateHash(blk)
	}
}

// BenchmarkStressSerializeLargeBlock serializes a large block repeatedly.
func BenchmarkStressSerializeLargeBlock(b *testing.B) {
	blk := &block.Block{Data: bytes.Repeat([]byte("a"), 1<<20)} // 1 MB
	for i := 0; i < b.N; i++ {
		block.Serialize(blk)
	}
}

// BenchmarkStressValidateLargeChain validates a large blockchain for each iteration.
func BenchmarkStressValidateLargeChain(b *testing.B) {
	chain := makeBlockchain(20000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if !bc.IsChainValidCached(chain) {
			b.Fatal("invalid chain")
		}
	}
}
