package main

import (
	"flag"
	"fmt"

	"my-first-blockchain/block"
	"my-first-blockchain/blockchain"
)

// main demonstrates block creation and chain validation.
func main() {
	blocks := flag.Int("blocks", 2, "number of additional blocks to generate")
	difficulty := flag.Int("difficulty", 4, "proof-of-work difficulty")
	output := flag.String("output", "", "optional path to write blockchain as JSON")
	flag.Parse()

	chain := []*block.Block{blockchain.NewGenesisBlock()}

	for i := 1; i <= *blocks; i++ {
		blk := blockchain.GenerateBlock(chain[len(chain)-1], fmt.Sprintf("Block %d", i), *difficulty)
		chain = append(chain, blk)
	}

	fmt.Println("Blockchain:")
	for _, blk := range chain {
		fmt.Printf("Index: %d, Data: %s, Hash: %s\n", blk.Index, string(blk.Data), fmt.Sprintf("%x", blk.Hash)[:10]+"...")
	}

	fmt.Printf("\nIs blockchain valid? %t\n", blockchain.IsChainValidCached(chain))

	if *output != "" {
		if err := blockchain.WriteChainJSON(chain, *output); err != nil {
			fmt.Printf("Error writing JSON: %v\n", err)
		} else {
			fmt.Printf("Blockchain written to %s\n", *output)
		}
	}
}
