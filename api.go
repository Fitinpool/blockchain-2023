// api.go
package main

import (
	e "blockchain/entities"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

var userdb Store
var blockdb Store

func StartServer() {
	var err error
	userdb, err := NewStore("userdb")
	if err != nil {
		errors.Wrap(err, "NewStore error userdb")
	}
	defer userdb.Close()

	blockdb, err := NewStore("blockchain")
	if err != nil {
		errors.Wrap(err, "NewStore error blockchain")
	}
	defer blockdb.Close()

	router := mux.NewRouter()

	// Definir rutas para la API
	router.HandleFunc("/transacciones", CrearNuevaTransaccion).Methods("POST")
	router.HandleFunc("/bloque/{index}", ConsultarPorBloque).Methods("GET")
	router.HandleFunc("/ultima-transaccion", ConsultarUltimaTransaccion).Methods("GET")

	// Iniciar el servidor en el puerto 3000
	fmt.Println("Servidor iniciado en el puerto 3000")
	http.ListenAndServe(":3000", router)
}

func CrearNuevaTransaccion(w http.ResponseWriter, r *http.Request) {
	var tx e.Transaction
	err := json.NewDecoder(r.Body).Decode(&tx)
	if err != nil {
		http.Error(w, "Error al decodificar la transacción", http.StatusBadRequest)
		return
	}

	// Obtener el usuario actual desde la base de datos
	retrievedData, err := userdb.Get(tx.Sender)
	if err != nil {
		http.Error(w, "Error al obtener el usuario de la base de datos", http.StatusInternalServerError)
		return
	}

	var resultUser e.User
	err = json.Unmarshal(retrievedData, &resultUser)
	if err != nil {
		http.Error(w, "Error al decodificar el usuario", http.StatusInternalServerError)
		return
	}

	// Verificar si el usuario tiene saldo suficiente
	if tx.Amount > resultUser.AccuntBalence {
		http.Error(w, "No tienes saldo suficiente", http.StatusForbidden)
		return
	}

	// Firmar la transacción
	err = FirmaTransaccion(&tx, resultUser.PrivateKey)
	if err != nil {
		http.Error(w, "Error al firmar la transacción", http.StatusInternalServerError)
		return
	}

	// Agregar la transacción al bloque actual
	currentBlock.Transactions = append(currentBlock.Transactions, tx)

	// Actualizar el nonce y el saldo del usuario
	resultUser.Nonce++
	resultUser.AccuntBalence -= tx.Amount

	err = userdb.Put(tx.Sender, resultUser)
	if err != nil {
		http.Error(w, "Error al actualizar el usuario en la base de datos", http.StatusInternalServerError)
		return
	}

	// Actualizar el saldo del destinatario
	recipientData, err := userdb.Get(tx.Recipient)
	if err != nil {
		http.Error(w, "Error al obtener el destinatario de la base de datos", http.StatusInternalServerError)
		return
	}

	var recipientResult e.User
	err = json.Unmarshal(recipientData, &recipientResult)
	if err != nil {
		http.Error(w, "Error al decodificar el destinatario", http.StatusInternalServerError)
		return
	}

	recipientResult.AccuntBalence += tx.Amount

	err = userdb.Put(tx.Recipient, recipientResult)
	if err != nil {
		http.Error(w, "Error al actualizar el destinatario en la base de datos", http.StatusInternalServerError)
		return
	}

	// Guardar el bloque actualizado en la base de datos
	key := fmt.Sprintf("%05d", currentBlock.Index)
	err = blockdb.Put(key, currentBlock)
	if err != nil {
		http.Error(w, "Error al guardar el bloque en la base de datos", http.StatusInternalServerError)
		return
	}

	// Responder con el resultado de la transacción
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(tx)
}

func ConsultarPorBloque(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	index, err := strconv.Atoi(params["index"])
	if err != nil {
		http.Error(w, "Error al convertir el índice del bloque", http.StatusBadRequest)
		return
	}

	// Obtener el bloque correspondiente al índice
	var retrievedData []byte
	if index == currentBlock.Index {
		retrievedData, err = json.Marshal(currentBlock)
	} else {
		key := fmt.Sprintf("%05d", index)
		retrievedData, err = blockdb.Get(key)
	}

	if err != nil {
		http.Error(w, "Error al obtener el bloque de la base de datos", http.StatusInternalServerError)
		return
	}

	// Enviar el bloque como respuesta
	w.WriteHeader(http.StatusOK)
	w.Write(retrievedData)
}

func ConsultarUltimaTransaccion(w http.ResponseWriter, r *http.Request) {
	// Obtener la última transacción del bloque actual
	if len(currentBlock.Transactions) == 0 {
		http.Error(w, "No hay transacciones en el bloque actual", http.StatusNotFound)
		return
	}

	lastTransaction := currentBlock.Transactions[len(currentBlock.Transactions)-1]

	// Enviar la última transacción como respuesta
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(lastTransaction)
}
