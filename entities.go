package main

type Transaction struct {
    Sender    string  `json:"sender"`
    Recipient string  `json:"recipient"`
    Amount    float64 `json:"amount"`
    Signature string  `json:"signature"`
    PubKey    string  `json:"pub_key"`
	Nonce        int           `json:"nonce"`
}

type Block struct {
	Index        int           `json:"index"`
	Timestamp    int64         `json:"timestamp"`
	Transactions []Transaction `json:"transaction"`
	PreviousHash string        `json:"previous_hash"`
	Hash         string        `json:"hash"`
}