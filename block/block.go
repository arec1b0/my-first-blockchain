package block

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
)

// Block represents a single record in the blockchain.
type Block struct {
	Index     int    `json:"index"`
	Timestamp int64  `json:"timestamp"`
	Data      []byte `json:"data"`
	PrevHash  []byte `json:"prev_hash"`
	Hash      []byte `json:"hash"`
	Nonce     int    `json:"nonce"`

	explicitlyInitialized bool `json:"-"`
}

// NewExplicitBlock returns a Block with the explicitlyInitialized flag set.
func NewExplicitBlock() Block {
	return Block{explicitlyInitialized: true}
}

// Serialize converts a block into a deterministic byte slice.
func Serialize(block *Block) []byte {
	var buf bytes.Buffer

	buf.WriteByte(0x01)
	if block.explicitlyInitialized {
		buf.WriteByte(0xFF)
	} else {
		buf.WriteByte(0x00)
	}

	binary.Write(&buf, binary.LittleEndian, int64(block.Index))
	binary.Write(&buf, binary.LittleEndian, int64(block.Timestamp))
	binary.Write(&buf, binary.LittleEndian, int64(block.Nonce))

	binary.Write(&buf, binary.LittleEndian, int32(len(block.Data)))
	buf.Write(block.Data)

	binary.Write(&buf, binary.LittleEndian, int32(len(block.PrevHash)))
	buf.Write(block.PrevHash)

	return buf.Bytes()
}

// CalculateHash returns a SHA-256 hash of the serialized block.
func CalculateHash(block *Block) []byte {
	bytes := Serialize(block)
	hash := sha256.Sum256(bytes)
	return hash[:]
}
