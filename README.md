# Simple Blockchain in Go
A lightweight and educational blockchain written in Go, featuring:
- Block structure with binary-safe serialization
- Proof-of-Work mining mechanism
- SHA-256 hashing using `[]byte` for collision safety
- Advanced unit tests for hash uniqueness
- Chain validation with caching (O(n) performance)
- Benchmark tests for large chains
---
## Getting Started
Clone the repo and run:
go run main.go
To run tests:
go test
To benchmark chain validation:
go test -bench=.
---
## Structure
### Block
Each block stores:
- `Index`, `Timestamp`
- `Data` as `[]byte`
- `PrevHash` and `Hash` as `[]byte`
- `Nonce` (for Proof-of-Work)
### Proof-of-Work
format).
Implemented via `proofOfWork()` with difficulty level (`n` leading zeroes in hex
### Chain Validation
- `isChainValidCached()` validates the blockchain using cached hashes to avoid redundant
computation.
---
## Test Coverage
- `main_test.go` includes adversarial hash collision cases.
- Verifies that different field combinations do not produce same hashes.
---
## Benchmarks
Tested on chains with:
- 100 blocks
- 1,000 blocks
- 5,000 blocks
- 10,000 blocks
Use `go test -bench=.` to see performance.
---
## Author
Created by Danylo Mozhaiev.
Inspired by Arec1b0, Go learning projects and blockchain principles.
---
## License
MIT -- free to use, fork, and improve!
