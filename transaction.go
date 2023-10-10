package main

import (
	"fmt"
    "crypto/ecdsa"
    "crypto/elliptic"
    "crypto/rand"
    "crypto/sha256"
    "encoding/hex"
    "crypto/x509"
	"log"
    "math/big"
)

type PrivateKeyData struct {
	PrivateKey []byte
	Password  string
    Transaccion []string
}


func GeneraLlavesPrivadas() (*ecdsa.PrivateKey, error) {
    privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
    if err != nil {
        return nil, err
    }
    return privKey, nil
}

func PrivateKeyToBytes(priv *ecdsa.PrivateKey) []byte {
	privBytes, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		log.Fatal(err)
	}
	return privBytes
}

func BytesToPrivateKey(privBytes []byte) *ecdsa.PrivateKey {
	priv, err := x509.ParseECPrivateKey(privBytes)
	if err != nil {
		log.Fatal(err)
	}
	return priv
}

func FirmaConLlavePrivada(tx *Transaction, privKey *ecdsa.PrivateKey) {
    txHash := sha256.Sum256([]byte(tx.Sender + tx.Recipient + fmt.Sprintf("%f", tx.Amount)))
    r, s, err := ecdsa.Sign(rand.Reader, privKey, txHash[:])
    if err != nil {
        panic(err)
    }
    signature := append(r.Bytes(), s.Bytes()...)
    tx.Signature = hex.EncodeToString(signature)
    tx.PubKey = hex.EncodeToString(elliptic.Marshal(elliptic.P256(), privKey.PublicKey.X, privKey.PublicKey.Y))
    tx.Nonce = FindNonce(*tx)
}

func VerificaFirmaTransaccion(tx *Transaction) bool {
    pubKeyBytes, err := hex.DecodeString(tx.PubKey)
    if err != nil {
        return false
    }
    x, y := elliptic.Unmarshal(elliptic.P256(), pubKeyBytes)
    pubKey := ecdsa.PublicKey{Curve: elliptic.P256(), X: x, Y: y}

    txHash := sha256.Sum256([]byte(tx.Sender + tx.Recipient + fmt.Sprintf("%f", tx.Amount)))
    signature, err := hex.DecodeString(tx.Signature)
    if err != nil {
        return false
    }
    r := new(big.Int).SetBytes(signature[:len(signature)/2])
    s := new(big.Int).SetBytes(signature[len(signature)/2:])
    return ecdsa.Verify(&pubKey, txHash[:], r, s)
}

func FindNonce(tx Transaction) int {
	targetPrefix := "0000" // Supongamos que queremos un hash que comience con cuatro ceros
	for nonce := 0; ; nonce++ {
		tx.Nonce = nonce
		hash := calculateHash(tx)
		if hash[:len(targetPrefix)] == targetPrefix {
			return nonce
		}
	}
}

func calculateHash(tx Transaction) string {
	record := tx.Sender + tx.Recipient + fmt.Sprintf("%f", tx.Amount) + tx.Signature + tx.PubKey + fmt.Sprintf("%d", tx.Nonce)
	h := sha256.New()
	h.Write([]byte(record))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}
