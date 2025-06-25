# ⛓️ Simple Blockchain in Go

A lightweight and educational blockchain written in Go.

## 🚀 Features

- 🧱 Block structure with binary-safe serialization  
- 💡 Proof-of-Work mining mechanism  
- 🔐 SHA-256 hashing using `[]byte` for collision safety  
- 🧪 Unit tests with adversarial cases  
- ⚡ Chain validation with O(n) performance using cached hashes  
- 📊 Benchmark support for large chains

---

## 📦 Getting Started

### Run the project:

```bash
go run main.go
```

### Run the tests:

```bash
go test
```

### Benchmark perfomance:

```bash
go test -bench=.
```

## 🧬 Block Structure

### Each block contains:
	•	Index, Timestamp
	•	Data as []byte
	•	PrevHash and Hash as []byte
	•	Nonce for PoW

The chain uses safe serialization via serializeBlock().

## 🔁 Chain Validation

Implemented via:
```bash
ok := isChainValidCached(chain)
```
Cached validation reduces hash recomputation.

## 🧪 Tests & Collision Checks

File main_test.go includes edge case tests:
	•	Length prefix mismatches
	•	Unicode vs ASCII
	•	Null bytes
	•	Prefix injection
	•	Field swapping

 ## 📈 Benchmarks

Validated on chains of size: 100, 1,000, 5,000, 10,000.

Run:

```bash
go test -bench=.
```

## 👤 Author
Created by Danylo Mozhaiev.
Inspired by [Arec1b0](https://gist.github.com/arec1b0), 
Go learning projects and blockchain principles.

## 📜 License

MIT — free to use, fork, and improve.
