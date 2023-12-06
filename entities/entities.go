package entitites

type Transaction struct {
	Sender    string  `json:"sender"`
	Recipient string  `json:"recipient"`
	Amount    float64 `json:"amount"`
	Signature []byte  `json:"signature"`
	Nonce     int     `json:"nonce"`
}

type Block struct {
	Index        int           `json:"index"`
	Timestamp    int64         `json:"timestamp"`
	Transactions []Transaction `json:"transaction"`
	PreviousHash string        `json:"previous_hash"`
	Hash         string        `json:"hash"`
}

type User struct {
	PrivateKey    []byte  `json:"private_key"`
	PublicKey     []byte  `json:"public_key"`
	Nombre        string  `json:"nombre"`
	Password      string  `json:"password"`
	Nonce         int     `json:"nonce"`
	AccuntBalence float64 `json:"accunt_balence"`
}
