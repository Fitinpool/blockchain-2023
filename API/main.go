package main

import (
	"blockchain/p2p"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func api() {
	//copy paste del main

	hostIP := flag.String("host-address", "0.0.0.0", "Default address to run node")
	port := flag.String("port", "4000", "Port to enable network connection")
	protocol := flag.String("protocol", "/xulo/1.0.0", "Protocol to enable network connection")
	flag.Parse()

	node, err := p2p.NewNode(&p2p.NodeConfig{
		IP:   *hostIP,
		Port: *port,
	})

	if err != nil {
		errors.Wrap(err, "main: p2p.NewNode error")
	}

	node.MdnsService.Start()

	blockNodedb, err := NewStore(fmt.Sprintf("node-block-%s", node.NetworkHost.ID().String()))
	if err != nil {
		errors.Wrap(err, "NewStore error blockchain")
	}
	defer blockNodedb.Close()

	userNodedb, err := NewStore(fmt.Sprintf("node-user-%s", node.NetworkHost.ID().String()))
	if err != nil {
		errors.Wrap(err, "NewStore error blockchain")
	}
	defer userNodedb.Close()

	node.SetupStreamHandler(context.Background(), node.HandleStream)
	node.Start()

	if *protocol == p2p.Protocol {

		CopyStore("blockchain", fmt.Sprintf("node-block-%s", node.NetworkHost.ID().String()))
		CopyStore("userdb", fmt.Sprintf("node-user-%s", node.NetworkHost.ID().String()))

		for _, peer := range node.ConnectedPeers {
			stream, error := node.NetworkHost.NewStream(context.Background(), peer.ID, p2p.ProtocolDataSharing)
			if error != nil {
				errors.Wrap(error, "main: node.NetworkHost.NewStream error")
			}
			node.HandleStream(stream)
		}
	}

	//rutas

	router := gin.Default()
	router.GET("/block/:ID", getBlock(blockNodedb))
	//router.POST("/transaction", newTransaction)

	router.Run("localhost:3000")
}

func getBlock(blockdb *Store) gin.HandlerFunc {
	return func(c *gin.Context) {
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
}

/*
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
*/
