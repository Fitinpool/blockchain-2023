package main

import (
	"crypto/ecdsa"
	"encoding/json"
	"log"

	"github.com/ethereum/go-ethereum/crypto"
)

func GeneraLlavesYAddress() ([]byte, []byte, string) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		log.Fatal(err)
	}

	privateKeyBytes := crypto.FromECDSA(privateKey)

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	publicKeyBytes := crypto.FromECDSAPub(publicKeyECDSA)

	address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()

	return privateKeyBytes, publicKeyBytes, address
}

func FirmaTransaccion(tx *Transaction, privateKeyBytes []byte) error {
	txCopy := *tx
	txCopy.Signature = nil

	privateKey, err := crypto.ToECDSA(privateKeyBytes)
	if err != nil {
		log.Fatalf("Error convirtiendo los bytes a clave privada: %v", err)
	}

	data, err := json.Marshal(txCopy)
	if err != nil {
		return err
	}

	hash := crypto.Keccak256Hash(data)
	sig, err := crypto.Sign(hash.Bytes(), privateKey)
	if err != nil {
		return err
	}

	tx.Signature = sig
	return nil
}
