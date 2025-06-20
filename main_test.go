package main

import (
	"bytes"
	"encoding/hex"
	"testing"
)

// Helper to clone a block (for mutation)
func cloneBlock(b Block) Block {
	return Block{
		Index:     b.Index,
		Timestamp: b.Timestamp,
		Data:      b.Data,
		PrevHash:  b.PrevHash,
		Nonce:     b.Nonce,
	}
}

func TestCalculateHash_AdversarialCollisions(t *testing.T) {
	// Baseline block for all tests
	base := Block{
		Index:     5,
		Timestamp: 1111222233,
		Data:      "foo|bar||baz",
		PrevHash:  "feedcafe",
		Nonce:     1337,
	}
	calculateHash(&base)

	cases := []struct {
		name string
		a, b Block
	}{
		// Same field values, different field order not possible in strongly typed Go,
		// but try same field content with field delimiters embedded
		{
			"Delimiter Injection: Data contains PrevHash as prefix",
			base,
			func() Block {
				blk := cloneBlock(base)
				blk.Data = base.PrevHash + base.Data
				return blk
			}(),
		},
		{
			"Delimiter Injection: Data and PrevHash with null bytes",
			base,
			func() Block {
				blk := cloneBlock(base)
				blk.Data = "foo\x00bar"
				blk.PrevHash = "baz\x00qux"
				return blk
			}(),
		},
		{
			"Length Prefix Edge: Data and PrevHash same bytes, different length prefix",
			func() Block {
				blk := cloneBlock(base)
				blk.Data = "AA"
				blk.PrevHash = "A"
				return blk
			}(),
			func() Block {
				blk := cloneBlock(base)
				blk.Data = "A"
				blk.PrevHash = "AA"
				return blk
			}(),
		},
		{
			"Leading Zeros in Data: Data with and without leading zeros",
			func() Block {
				blk := cloneBlock(base)
				blk.Data = "\x00\x00foobar"
				return blk
			}(),
			func() Block {
				blk := cloneBlock(base)
				blk.Data = "foobar"
				return blk
			}(),
		},
		{
			"Unicode vs. ASCII: visually similar but different bytes",
			func() Block {
				blk := cloneBlock(base)
				blk.Data = "é" // U+0065 U+0301
				return blk
			}(),
			func() Block {
				blk := cloneBlock(base)
				blk.Data = "\u00e9" // U+00E9
				return blk
			}(),
		},
		{
			"All Fields Zero vs. Empty",
			func() Block {
				return Block{}
			}(),
			func() Block {
				return Block{
					Index:              0,
					Timestamp:          0,
					Data:              "",
					PrevHash:          "",
					Nonce:             0,
					explicitlyInitialized: true,
				}
			}(),
		},
		{
			"Identical After Stripping Non-printables",
			func() Block {
				blk := cloneBlock(base)
				blk.Data = "foo\nbar"
				return blk
			}(),
			func() Block {
				blk := cloneBlock(base)
				blk.Data = "foo\rbar"
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
		if hashA == hashB {
			var buf bytes.Buffer
			buf.WriteString("Hash collision detected for case '" + tc.name + "':\n")
			buf.WriteString("Block A: %+v\n")
			buf.WriteString("Block B: %+v\n")
			buf.WriteString("Hash: " + hashA + "\n")
			buf.WriteString("BlockA bytes: " + hex.EncodeToString([]byte(tc.a.Data)) + "\n")
			buf.WriteString("BlockB bytes: " + hex.EncodeToString([]byte(tc.b.Data)) + "\n")
			t.Error(buf.String())
		}
	}
}
