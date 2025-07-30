package main

import (
	"bytes"
	"context"
	"fmt"
	"testing"
)

// BenchmarkStressGenerateBlockDifficulty4 measures PoW generation with higher difficulty.
func BenchmarkStressGenerateBlockDifficulty4(b *testing.B) {
	prev := &Block{Hash: []byte("prev")}
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		_, _ = generateBlock(ctx, prev, fmt.Sprintf("data-%d", i), 4)
	}
}

// BenchmarkStressGenerateBlockLargeData benchmarks creating blocks with large payloads.
func BenchmarkStressGenerateBlockLargeData(b *testing.B) {
	prev := &Block{Hash: []byte("prev")}
	data := bytes.Repeat([]byte("a"), 1<<20) // 1 MB
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		_, _ = generateBlock(ctx, prev, string(data), 1)
	}
}

// BenchmarkStressCalculateHashLargeData hashes a block with a large data field repeatedly.
func BenchmarkStressCalculateHashLargeData(b *testing.B) {
	blk := &Block{Data: bytes.Repeat([]byte("a"), 512*1024)} // 512 KB
	for i := 0; i < b.N; i++ {
		calculateHash(blk)
	}
}

// BenchmarkStressCalculateHashStreaming tests the streaming hash function
func BenchmarkStressCalculateHashStreaming(b *testing.B) {
	blk := &Block{Data: bytes.Repeat([]byte("a"), 512*1024)} // 512 KB
	for i := 0; i < b.N; i++ {
		calculateHashStreaming(blk)
	}
}

// BenchmarkStressSerializeLargeBlock serializes a large block repeatedly.
func BenchmarkStressSerializeLargeBlock(b *testing.B) {
	blk := &Block{Data: bytes.Repeat([]byte("a"), 1<<20)} // 1 MB
	for i := 0; i < b.N; i++ {
		serializeBlock(blk)
	}
}

// BenchmarkStressValidateLargeChain validates a large blockchain for each iteration.
func BenchmarkStressValidateLargeChain(b *testing.B) {
	chain := makeBlockchain(20000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if !isChainValidCached(chain) {
			b.Fatal("invalid chain")
		}
	}
}

// BenchmarkStressValidateLargeChainConcurrent tests concurrent validation
func BenchmarkStressValidateLargeChainConcurrent(b *testing.B) {
	chain := makeBlockchain(20000)
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if !isChainValidConcurrent(ctx, chain) {
			b.Fatal("invalid chain")
		}
	}
}

// BenchmarkStressGenerateBlockDifficulty2 measures PoW generation with lower difficulty.
func BenchmarkStressGenerateBlockDifficulty2(b *testing.B) {
	prev := &Block{Hash: []byte("prev")}
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		_, _ = generateBlock(ctx, prev, fmt.Sprintf("data-%d", i), 2)
	}
}

// BenchmarkStressValidateSmallChain tests validation on smaller chains
func BenchmarkStressValidateSmallChain(b *testing.B) {
	chain := makeBlockchain(100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if !isChainValidCached(chain) {
			b.Fatal("invalid chain")
		}
	}
}
