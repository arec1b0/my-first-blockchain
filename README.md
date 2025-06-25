Simple Blockchain in Go
[Overview] Overview
This is a lightweight and educational blockchain written in Go. It includes features such as:
* - Block structure Block structure with binary-safe serialization
* - Proof-of-Work Proof-of-Work mining mechanism
* - SHA-256 hashing SHA-256 hashing using byte slices
* [Tests] Adversarial unit tests for hash uniqueness
* - Cached validation Cached validation for O(n) performance
* - Benchmarks Performance benchmarks for large blockchains
[Getting Started] Getting Started
Run the project using:
go run main.go
Run tests with:
go test
Benchmark performance with:
go test -bench=.
[Block Structure] Block Structure
* `Index`, `Timestamp`
* `Data` as `[]byte`
* `PrevHash` and `Hash` as `[]byte`
* `Nonce` for Proof-of-Work
The blockchain uses `[]byte` for all hash-related fields and data to ensure binary safety.
[Validation] Chain Validation
The function `isChainValidCached()` ensures the chain is valid by caching hashes of each block.
ok := isChainValidCached(chain)
[Tests] Tests & Collision Checks
The `main_test.go` file includes collision test cases to detect possible hash duplications.
Each test case validates different edge conditions such as prefix collisions, length variations, and character encodings.
[Benchmarks] Benchmarks
Validation tested on blockchains of 100, 1,000, 5,000 and 10,000 blocks.
go test -bench=.
[Author] Author
Created by Danylo Mozhaiev.
Inspired by [Arec1b0] (https://gist.github.com/arec1b0), Go learning projects and blockchain principles.
[License] License
MIT - free to use, fork, and improve!
