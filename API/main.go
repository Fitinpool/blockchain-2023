package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

// var currentBlock e.Block

func getBlock(c *gin.Context) {
	//hacer case 5 aqui
	index := c.Param("ID")

	key := fmt.Sprintf("%05d", index)

	retrievedData, err := blockdb.Get(key)

	if err != nil {
		return
	}

	var resultBlock e.Block
	err = json.Unmarshal(retrievedData, &resultBlock)

	//se envia estado y resultado
	c.IndentedJSON(http.StatusOK, resultBlock)
}

func getLastTransaction(c *gin.Context) {
	//hacer last

	//se envia estado
	c.IndentedJSON(http.StatusOK, lastTransaction)
}

func newTransaction(c *gin.Context) {

	var recipient string
	var amount float64

	//hay que ver que pedir

	tx := &e.Transaction{
		Sender:    inputUser,
		Recipient: recipient,
		Amount:    amount,
		Nonce:     resultUser.Nonce + 1,
	}

	FirmaTransaccion(tx, resultUser.PrivateKey)

	currentBlock.Transactions = append(currentBlock.Transactions, *tx)

	fmt.Println("Transaccion agregada en el bloque: " + fmt.Sprintf("%d", currentBlock.Index))
	fmt.Println("Nonce:" + fmt.Sprint(tx.Nonce))

	resultUser.Nonce = resultUser.Nonce + 1
	resultUser.AccuntBalence = resultUser.AccuntBalence - amount

	err := userdb.Put(inputUser, resultUser)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "error put"})
	}

	recipientPut, err := userdb.Get(recipient)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "no existe o error"})
	}

	var recipientResult e.User
	err = json.Unmarshal(recipientPut, &recipientResult)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "json.Unmarshal error"})
	}

	recipientResult.AccuntBalence = recipientResult.AccuntBalence + amount

	err = userdb.Put(recipient, recipientResult)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "juserdb.Put error"})
	}

	newResult, err := userdb.Get(inputUser)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "userdb.Get error"})
	}

	err = json.Unmarshal(newResult, &resultUser)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "json.Unmarshal error"})
	}

	//se envia estado ok
	c.IndentedJSON(http.StatusCreated, tx)
}

func main() {

	blockdb, err := NewStore("blockchain")
	if err != nil {
		errors.Wrap(err, "NewStore error blockchain")
	}

	userdb, err := NewStore("userdb")
	if err != nil {
		errors.Wrap(err, "main: NewStore error userdb")
	}
	//rutas

	router := gin.Default()
	router.GET("/block/:ID", getBlock)
	router.POST("/transaction", getBlock)

	router.Run("localhost:3000")
}
