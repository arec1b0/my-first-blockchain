package main

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"
)

// cloneBlock returns a deep copy of the given block for test mutation
func cloneBlock(b Block) Block {
	return Block{
		Index:     b.Index,
		Timestamp: b.Timestamp,
		Data:      append([]byte{}, b.Data...),
		PrevHash:  append([]byte{}, b.PrevHash...),
		Nonce:     b.Nonce,
	}
}

// TestCalculateHash_AdversarialCollisions checks if different block contents produce unique hashes
func TestCalculateHash_AdversarialCollisions(t *testing.T) {
	base := Block{
		Index:     5,
		Timestamp: 1111222233,
		Data:      []byte("foo|bar||baz"),
		PrevHash:  []byte("feedcafe"),
		Nonce:     1337,
	}
	calculateHash(&base)

	cases := []struct {
		name string
		a, b Block
	}{
		{
			"Delimiter Injection: Data contains PrevHash as prefix",
			base,
			func() Block {
				blk := cloneBlock(base)
				blk.Data = append(base.PrevHash, base.Data...)
				return blk
			}(),
		},
		{
			"Delimiter Injection: Data and PrevHash with null bytes",
			base,
			func() Block {
				blk := cloneBlock(base)
				blk.Data = []byte("foo\x00bar")
				blk.PrevHash = []byte("baz\x00qux")
				return blk
			}(),
		},
		{
			"Length Prefix Edge: Data and PrevHash same bytes, different length prefix",
			func() Block {
				blk := cloneBlock(base)
				blk.Data = []byte("AA")
				blk.PrevHash = []byte("A")
				return blk
			}(),
			func() Block {
				blk := cloneBlock(base)
				blk.Data = []byte("A")
				blk.PrevHash = []byte("AA")
				return blk
			}(),
		},
		{
			"Leading Zeros in Data: Data with and without leading zeros",
			func() Block {
				blk := cloneBlock(base)
				blk.Data = []byte("\x00\x00foobar")
				return blk
			}(),
			func() Block {
				blk := cloneBlock(base)
				blk.Data = []byte("foobar")
				return blk
			}(),
		},
		{
			"Unicode vs. ASCII: visually similar but different bytes",
			func() Block {
				blk := cloneBlock(base)
				blk.Data = []byte("é") // U+0065 U+0301
				return blk
			}(),
			func() Block {
				blk := cloneBlock(base)
				blk.Data = []byte("\u00e9") // U+00E9
				return blk
			}(),
		},
		{
			"Different Nonce Values",
			func() Block {
				return Block{Index: 0, Timestamp: 0, Data: []byte{}, PrevHash: []byte{}, Nonce: 0}
			}(),
			func() Block {
				return Block{Index: 0, Timestamp: 0, Data: []byte{}, PrevHash: []byte{}, Nonce: 1}
			}(),
		},
		{
			"Identical After Stripping Non-printables",
			func() Block {
				blk := cloneBlock(base)
				blk.Data = []byte("foo\nbar")
				return blk
			}(),
			func() Block {
				blk := cloneBlock(base)
				blk.Data = []byte("foo\rbar")
				return blk
			}(),
		},
		{
			"Data and PrevHash swapped, same combined bytes",
			func() Block {
				blk := cloneBlock(base)
				blk.Data, blk.PrevHash = base.PrevHash, base.Data
				return blk
			}(),
			base,
		},
		{
			"Different Index but other fields match",
			func() Block {
				blk := cloneBlock(base)
				blk.Index++
				return blk
			}(),
			base,
		},
	}

	for _, tc := range cases {
		hashA := calculateHash(&tc.a)
		hashB := calculateHash(&tc.b)
		if bytes.Equal(hashA, hashB) {
			var buf bytes.Buffer
			buf.WriteString("Hash collision detected for case '" + tc.name + "':\n")
			buf.WriteString(fmt.Sprintf("Block A: %+v\n", tc.a))
			buf.WriteString(fmt.Sprintf("Block B: %+v\n", tc.b))
			buf.WriteString("Hash: " + hex.EncodeToString(hashA) + "\n")
			buf.WriteString("BlockA bytes: " + hex.EncodeToString(tc.a.Data) + "\n")
			buf.WriteString("BlockB bytes: " + hex.EncodeToString(tc.b.Data) + "\n")
			t.Error(buf.String())
		}
	}
}

// makeBlockchain creates a sample blockchain of the given size with a specified difficulty
func makeBlockchain(size int, difficulty int) []*Block {
	genesis := &Block{
		Index:     0,
		Timestamp: 0,
		Data:      []byte("Genesis"),
		PrevHash:  []byte{},
	}
	// Genesis block hash is calculated without PoW in this model
	genesis.Hash = calculateHash(genesis)

	chain := []*Block{genesis}
	ctx := context.Background()

	for i := 1; i < size; i++ {
		block, err := generateBlock(ctx, chain[i-1], fmt.Sprintf("Block %d", i), difficulty)
		if err != nil {
			// Tests should fail hard if block generation fails
			panic(fmt.Sprintf("test blockchain generation failed: %v", err))
		}
		chain = append(chain, block)
	}
	return chain
}

// BenchmarkChainValidation measures performance of isChainValidCached
func BenchmarkChainValidation(b *testing.B) {
	sizes := []int{100, 1000, 5000, 10000}
	const difficulty = 1 // Use a low, constant difficulty for benchmarks

	for _, size := range sizes {
		chain := makeBlockchain(size, difficulty)
		b.Run(fmt.Sprintf("N=%d", size), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				ok := isChainValidCached(chain, difficulty)
				if !ok {
					b.Fatal("Chain is invalid, check your logic!")
				}
			}
		})
	}
}

// TestWriteChainJSON checks that a chain can be written to a JSON file.
func TestWriteChainJSON(t *testing.T) {
	chain := makeBlockchain(3, 1) // Use low difficulty for test speed
	tmp, err := os.CreateTemp("", "chain*.json")
	if err != nil {
		t.Fatal(err)
	}
	path := tmp.Name()
	tmp.Close()
	defer os.Remove(path)

	if err := writeChainJSON(chain, path); err != nil {
		t.Fatalf("writeChainJSON failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	var decoded []*Block
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}
	if len(decoded) != len(chain) {
		t.Fatalf("expected %d blocks, got %d", len(chain), len(decoded))
	}
}

// TestValidateChain_InvalidPoW verifies that a chain with a block not meeting PoW is invalid.
func TestValidateChain_InvalidPoW(t *testing.T) {
	const difficulty = 4
	chain := makeBlockchain(3, difficulty)

	// Tamper with the last block by replacing its hash with one that does not meet the PoW difficulty.
	// This simulates a fraudulent block.
	invalidBlock := chain[2]
	invalidBlock.Nonce = 0 // Reset nonce to find a hash without PoW
	invalidBlock.Hash = calculateHash(invalidBlock)

	// Ensure our test setup is correct: the new hash should NOT meet the difficulty.
	if validateDifficulty(invalidBlock.Hash, difficulty) {
		t.Fatalf("Test setup failed: manually calculated hash unexpectedly met PoW difficulty. Increase difficulty.")
	}
	chain[2] = invalidBlock

	// The cached (sequential) validator should fail
	if isChainValidCached(chain, difficulty) {
		t.Error("Expected chain to be invalid due to faulty PoW, but isChainValidCached returned true")
	}

	// The concurrent validator should also fail
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if isChainValidConcurrent(ctx, chain, difficulty) {
		t.Error("Expected chain to be invalid due to faulty PoW, but isChainValidConcurrent returned true")
	}
}