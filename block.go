package main

import (
	"crypto/sha256"
	"encoding/hex"

	"fmt"
	"time"
)

func CreateMainBlock() Block {

	mainBlock := Block{
		Index:        1,
		Timestamp:    time.Now().Unix(),
		PreviousHash: "0000000000000000000000000000000000000000000000000000000000000000",
	}

	mainBlock.Hash = CalculateHash(mainBlock)

	return mainBlock
}

func GenerateBlock(index int, previousHash string) Block {
	block := Block{
		Index:        index,
		Timestamp:    time.Now().Unix(),
		PreviousHash: previousHash,
	}

	block.Hash = CalculateHash(block)

	return block
}

func CalculateHash(b Block) string {
	data := fmt.Sprintf("%d%d%s", b.Index, b.Timestamp, b.PreviousHash)
	for _, tx := range b.Transactions {
		data += fmt.Sprintf("%s%s%f", tx.Sender, tx.Recipient, tx.Amount)
	}

	hash := sha256.New()
	hash.Write([]byte(data))

	return hex.EncodeToString(hash.Sum(nil))
}

// func SaveBlock(db *leveldb.DB, block *Block) error {
// 	blockData, err := json.Marshal(block)
// 	if err != nil {
// 		return errors.Wrap(err, "storage: SaveBlock json.Marshal error")
// 	}

// 	err = db.Put([]byte(fmt.Sprintf("block-%d", block.Index)), blockData, nil)
// 	if err != nil {
// 		return errors.Wrap(err, "storage: SaveBlock db.Put error")
// 	}

// 	return nil
// }

// func LoadBlock(db *leveldb.DB, index int) (*Block, error) {
// 	blockData, err := db.Get([]byte(fmt.Sprintf("block-%d", index)), nil)

// 	if err != nil {
// 		return nil, errors.Wrap(err, "storage: LoadBlock db.Get error")
// 	}

// 	var block Block
// 	err = json.Unmarshal(blockData, &block)
// 	if err != nil {
// 		return nil, errors.Wrap(err, "storage: LoadBlock json.Unmarshal error")
// 	}
// 	return &block, nil
// }