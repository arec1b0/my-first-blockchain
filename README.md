# 🧱 Simple Blockchain in Go

This is a basic blockchain implementation written in Go, created for educational purposes. It includes not only block creation and hashing, but also full-chain validation logic to ensure integrity.

## 📌 Overview

Each block contains:
- `Index` – the block number in the chain
- `Timestamp` – the time of block creation
- `Data` – custom string payload
- `PrevHash` – hash of the previous block
- `Hash` – the current block's hash (SHA-256)

## ⚙️ Features

- ✅ Genesis block creation
- 🔗 Hash chaining between blocks
- 🔒 SHA-256 hash calculation
- ✅ Block and chain validation functions
- 📜 Blockchain printing with shortened hashes

## 📂 Structure

```go
type Block struct {
    Index     int
    Timestamp string
    Data      string
    PrevHash  string
    Hash      string
}

How to run:
- Clone the repository
"git clone https://github.com/YOUR_USERNAME/YOUR_REPOSITORY.git
cd YOUR_REPOSITORY"

Run the code:
- "go run main.go"

Expexted output:
-"Blockchain:
Index: 0, Data: Genesis, Hash: 9c53fa1d...
Index: 1, Data: Second Block, Hash: 78b12cc0...
Index: 2, Data: Third Block, Hash: c7a8e3e9...

Is blockchain valid? true"


Validation logic:
- "IsBlockValid(newBlock, prevBlock)" checks if a single block is valid in relation to the previous block.
- "IsChainValid(chain)" loops through the entire chain to verify consistency and integrity.

Author:
Created by Danylo Mozhaiev.
Inspired by Arec1b0, Go learning projects and blockchain principles.

