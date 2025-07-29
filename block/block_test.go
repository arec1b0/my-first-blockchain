package block_test

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	block "my-first-blockchain/block"
	bc "my-first-blockchain/blockchain"
)

// cloneBlock returns a deep copy of the given block for test mutation
func cloneBlock(b block.Block) block.Block {
	return block.Block{
		Index:     b.Index,
		Timestamp: b.Timestamp,
		Data:      append([]byte{}, b.Data...),
		PrevHash:  append([]byte{}, b.PrevHash...),
		Nonce:     b.Nonce,
	}
}

// TestCalculateHash_AdversarialCollisions checks if different block contents produce unique hashes
func TestCalculateHash_AdversarialCollisions(t *testing.T) {
	base := block.Block{
		Index:     5,
		Timestamp: 1111222233,
		Data:      []byte("foo|bar||baz"),
		PrevHash:  []byte("feedcafe"),
		Nonce:     1337,
	}
	block.CalculateHash(&base)

	cases := []struct {
		name string
		a, b block.Block
	}{
		{
			"Delimiter Injection: Data contains PrevHash as prefix",
			base,
			func() block.Block {
				blk := cloneBlock(base)
				blk.Data = append(base.PrevHash, base.Data...)
				return blk
			}(),
		},
		{
			"Delimiter Injection: Data and PrevHash with null bytes",
			base,
			func() block.Block {
				blk := cloneBlock(base)
				blk.Data = []byte("foo\x00bar")
				blk.PrevHash = []byte("baz\x00qux")
				return blk
			}(),
		},
		{
			"Length Prefix Edge: Data and PrevHash same bytes, different length prefix",
			func() block.Block {
				blk := cloneBlock(base)
				blk.Data = []byte("AA")
				blk.PrevHash = []byte("A")
				return blk
			}(),
			func() block.Block {
				blk := cloneBlock(base)
				blk.Data = []byte("A")
				blk.PrevHash = []byte("AA")
				return blk
			}(),
		},
		{
			"Leading Zeros in Data: Data with and without leading zeros",
			func() block.Block {
				blk := cloneBlock(base)
				blk.Data = []byte("\x00\x00foobar")
				return blk
			}(),
			func() block.Block {
				blk := cloneBlock(base)
				blk.Data = []byte("foobar")
				return blk
			}(),
		},
		{
			"Unicode vs. ASCII: visually similar but different bytes",
			func() block.Block {
				blk := cloneBlock(base)
				blk.Data = []byte("é") // U+0065 U+0301
				return blk
			}(),
			func() block.Block {
				blk := cloneBlock(base)
				blk.Data = []byte("\u00e9") // U+00E9
				return blk
			}(),
		},
		{
			"All Fields Zero vs. Explicit Empty",
			func() block.Block {
				return block.Block{}
			}(),
			func() block.Block {
				blk := block.NewExplicitBlock()
				return blk
			}(),
		},
		{
			"Identical After Stripping Non-printables",
			func() block.Block {
				blk := cloneBlock(base)
				blk.Data = []byte("foo\nbar")
				return blk
			}(),
			func() block.Block {
				blk := cloneBlock(base)
				blk.Data = []byte("foo\rbar")
				return blk
			}(),
		},
		{
			"Data and PrevHash swapped, same combined bytes",
			func() block.Block {
				blk := cloneBlock(base)
				blk.Data, blk.PrevHash = base.PrevHash, base.Data
				return blk
			}(),
			base,
		},
		{
			"Different Index but other fields match",
			func() block.Block {
				blk := cloneBlock(base)
				blk.Index++
				return blk
			}(),
			base,
		},
	}

	for _, tc := range cases {
		hashA := block.CalculateHash(&tc.a)
		hashB := block.CalculateHash(&tc.b)
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

// makeBlockchain creates a sample blockchain of the given size
func makeBlockchain(size int) []*block.Block {
	genesis := &block.Block{
		Index:     0,
		Timestamp: 0,
		Data:      []byte("Genesis"),
		PrevHash:  []byte{},
	}
	genesis.Hash = block.CalculateHash(genesis)

	chain := []*block.Block{genesis}
	for i := 1; i < size; i++ {
		blk := &block.Block{
			Index:     i,
			Timestamp: int64(i),
			Data:      []byte(fmt.Sprintf("Block %d", i)),
			PrevHash:  chain[i-1].Hash,
		}
		blk.Hash = block.CalculateHash(blk)
		chain = append(chain, blk)
	}
	return chain
}

// BenchmarkChainValidation measures performance of bc.IsChainValidCached
func BenchmarkChainValidation(b *testing.B) {
	sizes := []int{100, 1000, 5000, 10000}

	for _, size := range sizes {
		chain := makeBlockchain(size)
		b.Run(fmt.Sprintf("N=%d", size), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				ok := bc.IsChainValidCached(chain)
				if !ok {
					b.Fatal("Chain is invalid, check your logic!")
				}
			}
		})
	}
}

// TestWriteChainJSON checks that a chain can be written to a JSON file.
func TestWriteChainJSON(t *testing.T) {
	chain := makeBlockchain(3)
	tmp, err := os.CreateTemp("", "chain*.json")
	if err != nil {
		t.Fatal(err)
	}
	path := tmp.Name()
	tmp.Close()
	defer os.Remove(path)

	if err := bc.WriteChainJSON(chain, path); err != nil {
		t.Fatalf("bc.WriteChainJSON failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	var decoded []*block.Block
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}
	if len(decoded) != len(chain) {
		t.Fatalf("expected %d blocks, got %d", len(chain), len(decoded))
	}
}
