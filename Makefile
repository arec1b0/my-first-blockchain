# Makefile for optimized blockchain build

.PHONY: build build-optimized bench bench-optimized test clean profile

# Default build
build:
	go build -o blockchain main.go

# Optimized build with performance flags
build-optimized:
	go build -ldflags "-s -w" -gcflags "-B -C" -o blockchain-optimized main.go

# Development build with debug info
build-debug:
	go build -gcflags "-N -l" -o blockchain-debug main.go

# Run benchmarks
bench:
	go test -bench=. -benchmem -timeout=10m

# Run benchmarks with optimized binary
bench-optimized: build-optimized
	go test -bench=. -benchmem -timeout=10m -gcflags "-B -C"

# Run all tests
test:
	go test -v

# Run tests with race detection
test-race:
	go test -race -v

# Clean build artifacts
clean:
	rm -f blockchain blockchain-optimized blockchain-debug
	rm -f *.prof *.test

# CPU profiling
profile-cpu:
	go test -bench=BenchmarkStressGenerateBlockDifficulty4 -cpuprofile=cpu.prof -benchtime=30s
	go tool pprof cpu.prof

# Memory profiling
profile-mem:
	go test -bench=BenchmarkStressSerializeLargeBlock -memprofile=mem.prof -benchtime=30s
	go tool pprof mem.prof

# Performance comparison
compare-performance:
	@echo "Running performance comparison..."
	@echo "=== BEFORE OPTIMIZATIONS ==="
	@git stash push -m "temp stash for benchmark"
	@git checkout HEAD~1 -- main.go 2>/dev/null || echo "No previous version found"
	@go test -bench=BenchmarkStressGenerateBlockDifficulty4 -benchtime=10s -count=3 | tee before.txt
	@git checkout HEAD -- main.go
	@git stash pop 2>/dev/null || echo "No stash to pop"
	@echo "=== AFTER OPTIMIZATIONS ==="
	@go test -bench=BenchmarkStressGenerateBlockDifficulty4 -benchtime=10s -count=3 | tee after.txt
	@echo "=== COMPARISON ==="
	@go install golang.org/x/perf/cmd/benchcmp@latest 2>/dev/null || echo "Install benchcmp with: go install golang.org/x/perf/cmd/benchcmp@latest"
	@benchcmp before.txt after.txt 2>/dev/null || echo "Could not compare - install benchcmp"

# Install tools
install-tools:
	go install golang.org/x/perf/cmd/benchcmp@latest
	go install github.com/google/pprof@latest

# Lint and format
lint:
	go fmt ./...
	go vet ./...
	golangci-lint run 2>/dev/null || echo "Install golangci-lint for additional linting"

# Release build
release: lint test build-optimized
	@echo "Release build complete: blockchain-optimized"

# Development workflow
dev: lint test build
	@echo "Development build complete"

# Help
help:
	@echo "Available targets:"
	@echo "  build              - Build standard binary"
	@echo "  build-optimized    - Build with optimization flags"
	@echo "  build-debug        - Build with debug information"
	@echo "  bench              - Run all benchmarks"
	@echo "  bench-optimized    - Run benchmarks with optimizations"
	@echo "  test               - Run all tests"
	@echo "  test-race          - Run tests with race detection"
	@echo "  profile-cpu        - CPU profiling"
	@echo "  profile-mem        - Memory profiling"
	@echo "  compare-performance- Compare before/after performance"
	@echo "  clean              - Clean build artifacts"
	@echo "  lint               - Format and lint code"
	@echo "  release            - Complete release build"
	@echo "  dev                - Development workflow"
	@echo "  install-tools      - Install performance tools"
	@echo "  help               - Show this help"