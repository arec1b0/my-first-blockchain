# 🧱 Go Blockchain with Proof-of-Work

A simple blockchain implementation written in Go. This project introduces basic blockchain principles such as immutability, hashing, and proof-of-work (PoW), using efficient binary encoding and timestamping.


## 🔧 Features

- SHA-256 hashing using `bytes.Buffer` and `binary.Write`
- Unix timestamps for consistent, secure time representation
- 'Nonce' field with real-time proof-of-work mining
- Block and chain validation ('IsBlockValid', 'IsChainValid')
- Adjustable mining difficulty
- Simple and readable Go code structure

## 📦 Block Structure

```go
type Block struct {
    Index     int
    Timestamp int64
    Data      string
    PrevHash  string
    Hash      string
    Nonce     int
}

## ⛏️ Proof-of-Work
- Each block requires a valid hash with a configurable number of leading zeros (difficulty).
- The ProofOfWork function increments the Nonce until a valid hash is found.

```go
  func ProofOfWork(block Block, difficulty int) (string, int)

- This simulates real-world mining logic like Bitcoin's (simplified).


## 🧪 Chain Validation

The blockchain is validated by checking:
- Index increment
- Hash linkage (PrevHash)
- Hash correctness
- Optional: PoW consistency

func IsBlockValid(newBlock, prevBlock Block) bool
func IsChainValid(chain []Block) bool


🚀 How to run:
1. Make sure you have Go installed

2. Clone the repository
 git clone https://github.com/ITDan16/my-first-blockchain.git
 cd my-first-blockchain


3. Run the rpoject:
- "go run main.go"


Expexted output:
Blockchain:
Index: 0, Data: Genesis, Hash: 0000e91e...
Index: 1, Data: Second Block, Hash: 0000c9a4...
Index: 2, Data: Third Block, Hash: 00002bb1...

Is blockchain valid? true


🧠 Learning Goals

This project is perfect for those who want to learn:

- How blockchain works under the hood
- How hashing and mining are implemented
- Basic Go syntax and struct usage
- Chain validation logic

Author:
Created by Danylo Mozhaiev.
Inspired by Arec1b0, Go learning projects and blockchain principles.

