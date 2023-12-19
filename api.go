// api.go
package main

import (
	e "blockchain/entities"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"

	"github.com/gorilla/mux"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/pkg/errors"
)

func StartServer(userdb *Store, blockdb *Store, topicFullNode *pubsub.Topic) {
	router := mux.NewRouter()

	// Definir rutas para la API
	router.HandleFunc("/transacciones", func(w http.ResponseWriter, r *http.Request) {
		CrearNuevaTransaccion(w, r, blockdb, userdb, topicFullNode)
	}).Methods("POST")
	router.HandleFunc("/bloque/{index}", func(w http.ResponseWriter, r *http.Request) {
		BuscarBloqueHandler(w, r, blockdb)
	}).Methods("GET")
	router.HandleFunc("/buscar_ultima_transaccion", func(w http.ResponseWriter, r *http.Request) {
		BuscarUltimaTransaccionHandler(w, r, blockdb)
	}).Methods("GET")
	// Iniciar el servidor en el puerto 3000
	// fmt.Println("Servidor iniciado en el puerto 3000")
	http.ListenAndServe(":3000", router)
}

func CrearNuevaTransaccion(w http.ResponseWriter, r *http.Request, blockdb *Store, userdb *Store, topicFullNode *pubsub.Topic) {
	var params struct {
		Remitente    string
		Destinatario string
		Monto        float64
	}
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		http.Error(w, "Error decodificando parámetros", http.StatusBadRequest)
		return
	}

	var resultUser e.User
	retrievedData, err := userdb.Get(params.Remitente)
	if err != nil {
		errors.Wrap(err, "userdb.Get error")
	}

	err = json.Unmarshal(retrievedData, &resultUser)
	if err != nil {
		errors.Wrap(err, "user json.Unmarshal error")
	}

	txShare := fmt.Sprintf(`{
		"sender":    "%s",
		"recipient": "%s",
		"amount":    %f,
		"nonce":     %d
	}`, params.Remitente, params.Destinatario, params.Monto, resultUser.Nonce+1)

	userString := ""
	if reflect.TypeOf(resultUser.PrivateKey).Kind() == reflect.String {
		userString = fmt.Sprintf(`{
			"private_key":        "%v",
			"public_key":         "%v",
			"nombre":            "%s",
			"password":          "%s",
			"nonce":             %d,
			"accunt_balence":    %f
		}`, resultUser.PrivateKey, resultUser.PublicKey, resultUser.Nombre, resultUser.Password, resultUser.Nonce, resultUser.AccuntBalence)
	} else {
		userString = fmt.Sprintf(`{
			"private_key":        "%v",
			"public_key":         "%v",
			"nombre":            "%s",
			"password":          "%s",
			"nonce":             %d,
			"accunt_balence":    %f
		}`, fmt.Sprintf("%v", resultUser.PrivateKey), fmt.Sprintf("%v", resultUser.PublicKey), resultUser.Nombre, resultUser.Password, resultUser.Nonce, resultUser.AccuntBalence)
	}

	err = topicFullNode.Publish(context.Background(), []byte("nueva-transaccion;"+txShare+";"+userString+";"+resultUser.Nombre))
	if err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusCreated)

}

func BuscarBloqueHandler(w http.ResponseWriter, r *http.Request, blockdb *Store) {
	params := mux.Vars(r)
	index, err := strconv.Atoi(params["index"])
	if err != nil {
		http.Error(w, "El índice del bloque debe ser un número entero", http.StatusBadRequest)
		return
	}

	key := fmt.Sprintf("%05d", index)

	retrievedData, err := blockdb.Get(key)
	if err != nil {
		http.Error(w, "Error al buscar el bloque", http.StatusInternalServerError)
	}

	var resultBlock e.Block
	err = json.Unmarshal(retrievedData, &resultBlock)
	if err != nil {
		http.Error(w, "Error al decodificar el bloque", http.StatusInternalServerError)
		return
	}

	// Devolver el bloque como respuesta
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resultBlock)
}

func BuscarUltimaTransaccionHandler(w http.ResponseWriter, r *http.Request, blockdb *Store) {
	lastValue := blockdb.GetLastKey()
	if lastValue == nil {
		http.Error(w, "No hay transacciones", http.StatusNotFound)
		return
	}

	// Obtener la última transacción y devolverla como respuesta
	w.Header().Set("Content-Type", "application/json")
	w.Write(lastValue)
}
