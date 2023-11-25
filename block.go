package main

import (
	"crypto/sha256"
	"encoding/hex"

	e "blockchain/entities"
	"fmt"
	"time"
)

func CreateMainBlock() e.Block {

	mainBlock := e.Block{
		Index:        1,
		Timestamp:    time.Now().Unix(),
		PreviousHash: "0000000000000000000000000000000000000000000000000000000000000000",
	}

	mainBlock.Hash = CalculateHash(mainBlock)

	return mainBlock
}

func GenerateBlock(index int, previousHash string) e.Block {
	block := e.Block{
		Index:        index,
		Timestamp:    time.Now().Unix(),
		PreviousHash: previousHash,
	}

	block.Hash = CalculateHash(block)

	return block
}

func CalculateHash(b e.Block) string {
	data := fmt.Sprintf("%d%d%s", b.Index, b.Timestamp, b.PreviousHash)
	for _, tx := range b.Transactions {
		data += fmt.Sprintf("%s%s%f", tx.Sender, tx.Recipient, tx.Amount)
	}

	hash := sha256.New()
	hash.Write([]byte(data))

	return hex.EncodeToString(hash.Sum(nil))
}
