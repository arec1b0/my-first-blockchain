package pow

import (
	"fmt"
	"strings"

	"my-first-blockchain/block"
)

// ProofOfWork finds a valid hash that satisfies the difficulty constraint.
// It returns the discovered hash and the nonce used to generate it.
func ProofOfWork(b *block.Block, difficulty int) ([]byte, int) {
	prefix := strings.Repeat("0", difficulty)
	nonce := 0
	var hash []byte
	for {
		b.Nonce = nonce
		hash = block.CalculateHash(b)
		if strings.HasPrefix(fmt.Sprintf("%x", hash), prefix) {
			break
		}
		nonce++
	}
	return hash, nonce
}
