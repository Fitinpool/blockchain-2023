package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"

	e "blockchain/entities"

	"sync"

	"github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"

	"github.com/pkg/errors"
)

var currentBlock e.Block
var state bool = false
var isPublisher bool = false

const (
	Protocol = "/xulo/1.0.0"
)

func main() {

	port := flag.Int("port", 4000, "Puerto de escucha para el nodo")
	protocol := flag.String("protocol", Protocol, "Protocol to enable network connection")
	connectTo := flag.String("connect", "", "Dirección multiaddr del nodo a conectar")
	flagPublisher := flag.Bool("publish", false, "Define si el nodo publicará mensajes")
	flag.Parse()

	isPublisher = *flagPublisher
	mutex := &sync.Mutex{}

	ctx := context.Background()

	// Crear un host libp2p
	h, err := libp2p.New(
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", *port)),
	)
	if err != nil {
		panic(err)
	}
	defer h.Close()

	blockdb, err := NewStore("blockchain")
	if err != nil {
		errors.Wrap(err, "main: block NewStore error")
	}
	defer blockdb.Close()

	userdb, err := NewStore("userdb")
	if err != nil {
		errors.Wrap(err, "main: user NewStore error")
	}
	defer userdb.Close()

	blockNodedb, err := NewStore(fmt.Sprintf("node-block-%s", h.ID().String()))
	if err != nil {
		errors.Wrap(err, "main: blockNodedb NewStore error")
	}
	defer blockNodedb.Close()

	userNodedb, err := NewStore(fmt.Sprintf("node-user-%s", h.ID().String()))
	if err != nil {
		errors.Wrap(err, "main: userNodedb NewStore error")
	}
	defer userNodedb.Close()

	// Conectar a otro nodo si se proporciona una dirección
	if *connectTo != "" {
		ma, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/127.0.0.1/tcp/4000/p2p/%s", *connectTo))
		if err != nil {
			panic(err)
		}

		addrInfo, err := peer.AddrInfoFromP2pAddr(ma)
		if err != nil {
			panic(err)
		}

		if err := h.Connect(ctx, *addrInfo); err != nil {
			panic(err)
		}
		fmt.Println("Conectado a", *connectTo)
	}

	// Configurar pubsub
	ps, err := pubsub.NewGossipSub(ctx, h)
	if err != nil {
		panic(err)
	}

	// Suscribirse a un tópico
	topicBroadcast, err := ps.Join("broadcast")
	if err != nil {
		panic(err)
	}

	subBroadcast, err := topicBroadcast.Subscribe()
	if err != nil {
		panic(err)
	}

	topicFullNode, err := ps.Join("full-node")
	if err != nil {
		panic(err)
	}

	subFullNode, err := topicFullNode.Subscribe()
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			if isPublisher {
				msg, err := subFullNode.Next(ctx)
				if err != nil {
					panic(err)
				}

				protocoloFullNode := strings.Split(string(msg.Data), ";")

				if protocoloFullNode[0] == "nuevo-user" {
					var jsonTemp map[string]interface{}

					err = json.Unmarshal([]byte(protocoloFullNode[1]), &jsonTemp)
					if err != nil {
						errors.Wrap(err, "main: json.Unmarshal error")
					}

					priv_key_temp, err := stringToByteSlice(jsonTemp["private_key"].(string))
					if err != nil {
						errors.Wrap(err, "main: stringToByteSlice error")
					}
					public_key_temp, err := stringToByteSlice(jsonTemp["public_key"].(string))
					if err != nil {
						errors.Wrap(err, "main: stringToByteSlice error")
					}
					userTemp := e.User{
						PrivateKey:    priv_key_temp,
						PublicKey:     public_key_temp,
						Nombre:        jsonTemp["nombre"].(string),
						Password:      jsonTemp["password"].(string),
						Nonce:         int(jsonTemp["nonce"].(float64)),
						AccuntBalence: jsonTemp["accunt_balance"].(float64),
					}

					err = userNodedb.Put(string(protocoloFullNode[2]), userTemp)

					if err != nil {
						errors.Wrap(err, "main: userNodedb.Put error")
					}

					mutex.Lock()
					err = userdb.Put(string(protocoloFullNode[2]), userTemp)
					mutex.Unlock()
					if err != nil {
						errors.Wrap(err, "main: userdb.Put error")
					}

					err = topicBroadcast.Publish(ctx, []byte("aprobado-user;"+protocoloFullNode[1]+";"+protocoloFullNode[2]))

					if err != nil {
						panic(err)
					}

					state = true
					ExportarLevelDB(blockdb, "blockchain")
					ExportarLevelDB(userdb, "userdb")
				} else if protocoloFullNode[0] == "nueva-transaccion" {
					var jsonTrancs map[string]interface{}
					var jsonUser map[string]interface{}

					err = json.Unmarshal([]byte(protocoloFullNode[1]), &jsonTrancs)
					if err != nil {
						errors.Wrap(err, "main: json.Unmarshal error")
					}

					err = json.Unmarshal([]byte(protocoloFullNode[2]), &jsonUser)
					if err != nil {
						errors.Wrap(err, "main: json.Unmarshal error")
					}

					txTemp := &e.Transaction{
						Sender:    jsonTrancs["sender"].(string),
						Recipient: jsonTrancs["recipient"].(string),
						Amount:    jsonTrancs["amount"].(float64),
						Nonce:     int(jsonTrancs["nonce"].(float64)),
					}

					priv_key_temp, err := stringToByteSlice(jsonUser["private_key"].(string))
					if err != nil {
						errors.Wrap(err, "main: stringToByteSlice error")
					}
					public_key_temp, err := stringToByteSlice(jsonUser["public_key"].(string))
					if err != nil {
						errors.Wrap(err, "main: stringToByteSlice error")
					}

					userTemp := e.User{
						PrivateKey:    priv_key_temp,
						PublicKey:     public_key_temp,
						Nombre:        jsonUser["nombre"].(string),
						Password:      jsonUser["password"].(string),
						Nonce:         int(jsonUser["nonce"].(float64)),
						AccuntBalence: jsonUser["accunt_balence"].(float64),
					}

					mutex.Lock()
					FirmaTransaccion(txTemp, userTemp.PrivateKey)
					mutex.Unlock()

					currentBlock.Transactions = append(currentBlock.Transactions, *txTemp)

					userTemp.Nonce = userTemp.Nonce + 1
					userTemp.AccuntBalence = userTemp.AccuntBalence - jsonTrancs["amount"].(float64)

					mutex.Lock()
					err = userdb.Put(protocoloFullNode[3], &userTemp)
					if err != nil {
						errors.Wrap(err, "case 2 userdb.Put error")
					}

					err = userNodedb.Put(protocoloFullNode[3], &userTemp)
					if err != nil {
						errors.Wrap(err, "case 2 userdb.Put error")
					}

					recipientPut, err := userdb.Get(jsonTrancs["recipient"].(string))
					if err != nil {
						state = true
						errors.Wrap(err, "No existe o Error")
					}
					mutex.Unlock()

					var recipientResult e.User
					err = json.Unmarshal(recipientPut, &recipientResult)
					if err != nil {
						errors.Wrap(err, "case 2 json.Unmarshal error")
					}

					recipientResult.AccuntBalence = recipientResult.AccuntBalence + jsonTrancs["amount"].(float64)

					mutex.Lock()
					err = userdb.Put(jsonTrancs["recipient"].(string), recipientResult)
					if err != nil {
						errors.Wrap(err, "case 2 userdb.Put error")
					}

					err = userNodedb.Put(jsonTrancs["recipient"].(string), recipientResult)
					if err != nil {
						errors.Wrap(err, "case 2 userdb.Put error")
					}

					newResult, err := userdb.Get(protocoloFullNode[3])
					if err != nil {
						errors.Wrap(err, "userdb.Get error")
					}

					err = json.Unmarshal(newResult, &userTemp)
					if err != nil {
						errors.Wrap(err, "case 2 json.Unmarshal error")
					}
					mutex.Unlock()

					err = topicBroadcast.Publish(ctx, []byte("aprobado-block;"+protocoloFullNode[1]+";"+protocoloFullNode[2]+";"+protocoloFullNode[3]))
					if err != nil {
						panic(err)
					}

					state = true
					ExportarLevelDB(blockdb, "blockchain")
					ExportarLevelDB(userdb, "userdb")
				}

			} else {
				msg, err := subBroadcast.Next(ctx)
				if err != nil {
					panic(err)
				}

				protocoloBroadcast := strings.Split(string(msg.Data), ";")

				if protocoloBroadcast[0] == "aprobado-user" {
					var jsonTemp map[string]interface{}

					err = json.Unmarshal([]byte(protocoloBroadcast[1]), &jsonTemp)
					if err != nil {
						errors.Wrap(err, "main: json.Unmarshal error")
					}

					priv_key_temp, err := stringToByteSlice(jsonTemp["private_key"].(string))
					if err != nil {
						errors.Wrap(err, "main: stringToByteSlice error")
					}
					public_key_temp, err := stringToByteSlice(jsonTemp["public_key"].(string))
					if err != nil {
						errors.Wrap(err, "main: stringToByteSlice error")
					}

					userTemp := e.User{
						PrivateKey:    priv_key_temp,
						PublicKey:     public_key_temp,
						Nombre:        jsonTemp["nombre"].(string),
						Password:      jsonTemp["password"].(string),
						Nonce:         int(jsonTemp["nonce"].(float64)),
						AccuntBalence: jsonTemp["accunt_balance"].(float64),
					}

					mutex.Lock()
					err = userNodedb.Put(string(protocoloBroadcast[2]), userTemp)
					mutex.Unlock()

					if err != nil {
						errors.Wrap(err, "main: userNodedb.Put error")
					}

					state = true
				} else if protocoloBroadcast[0] == "aprobado-block" {

					var jsonTrancs map[string]interface{}
					var jsonUser map[string]interface{}

					err = json.Unmarshal([]byte(protocoloBroadcast[1]), &jsonTrancs)
					if err != nil {
						errors.Wrap(err, "main: json.Unmarshal error")
					}

					err = json.Unmarshal([]byte(protocoloBroadcast[2]), &jsonUser)
					if err != nil {
						errors.Wrap(err, "main: json.Unmarshal error")
					}

					priv_key_temp, err := stringToByteSlice(jsonUser["private_key"].(string))
					if err != nil {
						errors.Wrap(err, "main: stringToByteSlice error")
					}
					public_key_temp, err := stringToByteSlice(jsonUser["public_key"].(string))
					if err != nil {
						errors.Wrap(err, "main: stringToByteSlice error")
					}

					userTemp := e.User{
						PrivateKey:    priv_key_temp,
						PublicKey:     public_key_temp,
						Nombre:        jsonUser["nombre"].(string),
						Password:      jsonUser["password"].(string),
						Nonce:         int(jsonUser["nonce"].(float64)),
						AccuntBalence: jsonUser["accunt_balence"].(float64),
					}

					userTemp.Nonce = userTemp.Nonce + 1
					userTemp.AccuntBalence = userTemp.AccuntBalence - jsonTrancs["amount"].(float64)

					mutex.Lock()

					err = userNodedb.Put(protocoloBroadcast[3], &userTemp)
					if err != nil {
						errors.Wrap(err, "case 2 userdb.Put error")
					}

					recipientPut, err := userNodedb.Get(jsonTrancs["recipient"].(string))
					if err != nil {
						state = true
						errors.Wrap(err, "No existe o Error")
					}
					mutex.Unlock()

					var recipientResult e.User
					err = json.Unmarshal(recipientPut, &recipientResult)
					if err != nil {
						errors.Wrap(err, "case 2 json.Unmarshal error")
					}

					recipientResult.AccuntBalence = recipientResult.AccuntBalence + jsonTrancs["amount"].(float64)

					mutex.Lock()

					err = userNodedb.Put(jsonTrancs["recipient"].(string), recipientResult)
					if err != nil {
						errors.Wrap(err, "case 2 userdb.Put error")
					}
					mutex.Unlock()

					state = true

				} else if protocoloBroadcast[0] == "agrega-bloque" {

					var jsonBlock e.Block

					err = json.Unmarshal([]byte(protocoloBroadcast[2]), &jsonBlock)
					if err != nil {
						errors.Wrap(err, "main: json.Unmarshal error")
					}

					mutex.Lock()
					err = blockNodedb.Put(protocoloBroadcast[1], jsonBlock)
					mutex.Unlock()
					if err != nil {
						errors.Wrap(err, "case 2 userdb.Put error")
					}
				}
			}
		}
	}()

	if *protocol == Protocol {

		if isPublisher {
			ExportarLevelDB(blockdb, "blockchain")
			ExportarLevelDB(userdb, "userdb")
			ImportarLevelDB(blockNodedb, "blockchain")
			ImportarLevelDB(userNodedb, "userdb")
			go StartServer(userdb, blockdb, topicFullNode)
		} else {
			ImportarLevelDB(blockNodedb, "blockchain")
			ImportarLevelDB(userNodedb, "userdb")
		}

		go menu(blockNodedb, userNodedb, h, topicFullNode, subBroadcast)

		for {
			if isPublisher {
				isEmpty, err := blockdb.IsEmpty()
				if err != nil {
					errors.Wrap(err, "main: blockdb.IsEmpty error")
				}
				if isEmpty {
					currentBlock = CreateMainBlock()

					key := fmt.Sprintf("%05d", currentBlock.Index)
					time.Sleep(10 * time.Second)

					mutex.Lock()
					err = blockNodedb.Put(key, currentBlock)
					if err != nil {
						errors.Wrap(err, "main: blockNodedb.Put error")
					}

					err = blockdb.Put(key, currentBlock)
					if err != nil {
						errors.Wrap(err, "main: blockdb.Put error")
					}

					blockString, err := json.Marshal(currentBlock)
					if err != nil {
						errors.Wrap(err, "main: json.Marshal error")
					}
					mutex.Unlock()

					err = topicBroadcast.Publish(ctx, []byte("agrega-bloque;"+key+";"+string(blockString)))

					if err != nil {
						panic(err)
					}

				} else {
					if isPublisher {
						lastValues := blockdb.GetLastKey()

						if err != nil {
							errors.Wrap(err, "main: blockNodedb.GetLastKey error")
						}
						var result e.Block
						err = json.Unmarshal(lastValues, &result)
						if err != nil {
							errors.Wrap(err, "main: json.unmarshal error")
						}

						currentBlock = GenerateBlock(result.Index+1, result.Hash)
						key := fmt.Sprintf("%05d", currentBlock.Index)
						time.Sleep(10 * time.Second)

						mutex.Lock()
						err = blockNodedb.Put(key, currentBlock)
						if err != nil {
							errors.Wrap(err, "main: blockdb.Put error")
						}

						err = blockdb.Put(key, currentBlock)
						if err != nil {
							errors.Wrap(err, "main: blockdb.Put error")
						}

						blockString, err := json.Marshal(currentBlock)
						if err != nil {
							errors.Wrap(err, "main: json.Marshal error")
						}
						mutex.Unlock()

						err = topicBroadcast.Publish(ctx, []byte("agrega-bloque;"+key+";"+string(blockString)))

						if err != nil {
							panic(err)
						}
					}
				}
			}

		}

	} else {
		fmt.Println("Protocolo no valido")
	}

}

func menu(blockdb *Store, userdb *Store, h host.Host, topicFullNode *pubsub.Topic, subBroadcast *pubsub.Subscription) {
	var inputPass, inputUser string
	var resultUser e.User
	var p_address string

	for {
		for {

			var option int
			var bandera bool = false
			var userRegister, passRegister string
			fmt.Println("-----------------------------")
			fmt.Printf("Nodo ID: %s\n", h.ID().String())
			fmt.Printf("Address: %s \n", p_address)
			fmt.Println("-----------------------------")
			fmt.Println("\n---------- Menú ----------")
			fmt.Println("1. Ingresar")
			fmt.Println("2. Registrar")
			fmt.Print("Elige una opción: ")
			fmt.Scan(&option)

			switch option {
			case 1:
				fmt.Print("Introduce Address: ")
				fmt.Scanln(&inputUser)

				fmt.Print("Introduce la contraseña: ")
				fmt.Scanln(&inputPass)

				retrievedData, err := userdb.Get(inputUser)
				if err != nil {
					errors.Wrap(err, "userdb.Get error")
				}

				err = json.Unmarshal(retrievedData, &resultUser)
				if err != nil {
					errors.Wrap(err, "user json.Unmarshal error")
				}

				if resultUser.Password != "" {
					if resultUser.Password == inputPass {
						fmt.Println("Credenciales Correctas.")
						p_address = inputUser
						time.Sleep(2 * time.Second)
						ClearScreen()
						bandera = true
					} else {
						fmt.Println("Usuario o contraseña incorrectos.")
						time.Sleep(2 * time.Second)
						ClearScreen()
					}

				} else {
					fmt.Println("Usuario o contraseña incorrectos.")
					time.Sleep(2 * time.Second)
					ClearScreen()
				}
			case 2:
				fmt.Print("Introduce Usuario: ")
				fmt.Scanln(&userRegister)

				fmt.Print("Introduce la contraseña: ")
				fmt.Scanln(&passRegister)

				privKey, publicKey, address := GeneraLlavesYAddress()
				p_address = address
				dataString := fmt.Sprintf(`{
					"private_key":    "%v",
					"public_key":     "%v",
					"nombre":        "%s",
					"password":      "%s",
					"nonce":         %d,
					"accunt_balance": %d
				};%s`, fmt.Sprintf("%v", privKey), fmt.Sprintf("%v", publicKey), userRegister, passRegister, 0, 1000, address)

				topicFullNode.Publish(context.Background(), []byte("nuevo-user;"+dataString))

				for {
					if state {
						break
					}
					fmt.Print("Esperando respuesta...")
					time.Sleep(2 * time.Second)
					// ClearScreen()
				}

				fmt.Println("Datos guardados de " + userRegister + ", Address: " + address)
				fmt.Print("Presiona enter para continuar...")
				fmt.Scanln()
				ClearScreen()
				state = false

			default:
				fmt.Println("Opcion no valida")
			}

			if bandera {
				break
			}
		}

		for {
			lastKey := blockdb.GetLastKey()
			err := json.Unmarshal(lastKey, &currentBlock)
			if err != nil {
				errors.Wrap(err, "main: json.unmarshal error")
			}

			retrievedData, err := userdb.Get(p_address)
			if err != nil {
				errors.Wrap(err, "userdb.Get error")
			}

			err = json.Unmarshal(retrievedData, &resultUser)
			if err != nil {
				errors.Wrap(err, "user json.Unmarshal error")
			}

			var option int
			var bandera bool = false
			fmt.Println("-----------------------------")
			fmt.Printf("Nodo ID: %s\n", h.ID().String())
			fmt.Printf("Address: %s \n", p_address)
			fmt.Println("-----------------------------")
			fmt.Printf("Saldo : %f", resultUser.AccuntBalence)
			fmt.Println("\n---------- Menú ----------")
			fmt.Println("1. Ver contactos")
			fmt.Println("2. Realizar transacción")
			fmt.Println("3. Mostrar transacciones")
			fmt.Println("4. Buscar transacción específica")
			fmt.Println("5. Buscar un bloque")
			fmt.Println("6. Mostrar todos los bloques")
			fmt.Println("7. Cerrar Sesión")
			fmt.Println("8. Salir")
			fmt.Print("Elige una opción: ")
			fmt.Scan(&option)

			switch option {
			case 1:
				users, err := userdb.GetAllUser()
				if err != nil {
					errors.Wrap(err, "case 1 blockdb.Get error")
				}

				fmt.Println("Addresses: ")

				fmt.Println("------------------------------------------------------------------")
				for i, user := range users {
					if fmt.Sprintf(user.Key) != inputUser {
						fmt.Println(fmt.Sprintf("%d. ", i+1) + user.Data.Nombre + " | Address : " + fmt.Sprintf(user.Key))
					}
				}
				fmt.Println("------------------------------------------------------------------")

				fmt.Print("Presiona enter para continuar...")
				fmt.Scanln()
				ClearScreen()

			case 2:
				var recipient string
				var amountStr string
				var amount float64

				fmt.Print("Introduce el destinatario: ")
				fmt.Scanln(&recipient)

				fmt.Print("Introduce la cantidad: ")
				fmt.Scanln(&amountStr)

				amount, err := strconv.ParseFloat(amountStr, 64)

				if err != nil {

					fmt.Println("Por favor, introduce un número válido.")
				} else if amount > resultUser.AccuntBalence {
					fmt.Println("No tienes saldo suficiente")
					time.Sleep(2 * time.Second)
					ClearScreen()
					break
				} else if amount <= 0 {
					fmt.Println("Cantidad no válida")
					time.Sleep(2 * time.Second)
					ClearScreen()
					break
				} else {
					txShare := fmt.Sprintf(`{
						"sender":    "%s",
						"recipient": "%s",
						"amount":    %f,
						"nonce":     %d
					}`, inputUser, recipient, amount, resultUser.Nonce+1)

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

					err := topicFullNode.Publish(context.Background(), []byte("nueva-transaccion;"+txShare+";"+userString+";"+inputUser))
					if err != nil {
						panic(err)
					}

					fmt.Println("Transaccion agregada en el bloque: " + fmt.Sprintf("%d", currentBlock.Index+1))
					fmt.Println("Nonce:" + fmt.Sprint(resultUser.Nonce))

					for {
						if state {
							break
						}
						fmt.Print("Esperando respuesta...")
						time.Sleep(2 * time.Second)
						ClearScreen()
					}

					fmt.Print("Presiona enter para continuar...")
					fmt.Scanln()
					ClearScreen()
					state = false
				}

			case 3:
				blocks, err := blockdb.GetAllBlocks()
				if len(blocks) == 0 {
					fmt.Println("No hay transacciones")
					fmt.Print("Presiona enter para continuar...")
					fmt.Scanln()
					ClearScreen()
					break
				}
				if err != nil {
					errors.Wrap(err, "case 3 blockdb.GetAllBlocks error")
				}
				fmt.Println("------------------------------------------------------------------")
				for _, block := range blocks {
					for _, tx := range block.Transactions {
						if tx.Sender == inputUser || tx.Recipient == inputUser {
							fmt.Println("+++++++++++++++++++++++++++++++++++++++++++++++++++")
							fmt.Println("Sender: " + fmt.Sprint(tx.Sender))
							fmt.Println("Recipient: " + fmt.Sprint(tx.Recipient))
							fmt.Println("Amount:" + fmt.Sprint(tx.Amount))
							fmt.Println("Nonce:" + fmt.Sprint(tx.Nonce))
							fmt.Println("Bloque:" + fmt.Sprint(block.Index))
						}
					}
				}
				fmt.Println("+++++++++++++++++++++++++++++++++++++++++++++++++++")
				fmt.Print("Presiona enter para continuar...")
				fmt.Scanln()
				ClearScreen()
			case 4:
				var index, nonce int
				fmt.Print("Introduce el numero del bloque: ")
				fmt.Scanln(&index)
				fmt.Print("Introduce el nonce de la transaccion: ")
				fmt.Scanln(&nonce)

				key := fmt.Sprintf("%05d", index)
				nonceString := fmt.Sprintf("%d", nonce)

				if index > currentBlock.Index {
					fmt.Println("El bloque no existe")
					time.Sleep(2 * time.Second)
					ClearScreen()
				} else if index == currentBlock.Index {
					fmt.Println("------------------------------------------------------------------")
					bandera := false
					for _, tx := range currentBlock.Transactions {
						if fmt.Sprint(tx.Nonce) == nonceString {
							fmt.Println("    Sender: " + tx.Sender)
							fmt.Println("    Recipient: " + tx.Recipient)
							fmt.Println("    Amount: " + fmt.Sprintf("%f", tx.Amount))
							fmt.Println("    Nonce: " + fmt.Sprintf("%d", tx.Nonce))
							bandera = true
						}
					}
					if !bandera {
						fmt.Println("No se encontro la transaccion")
					}
					fmt.Println("------------------------------------------------------------------")
					fmt.Print("Presiona enter para continuar...")
					fmt.Scanln()
					ClearScreen()
				} else {
					retrievedData, err := blockdb.Get(key)
					if err != nil {
						errors.Wrap(err, "case 4 blockdb.Get error")
					}

					var resultBlock e.Block
					err = json.Unmarshal(retrievedData, &resultBlock)
					if err != nil {
						errors.Wrap(err, "case 4 json.Unmarshal error")
					}

					fmt.Println("------------------------------------------------------------------")
					bandera := false
					for _, tx := range resultBlock.Transactions {
						if fmt.Sprint(tx.Nonce) == nonceString {
							fmt.Println("    Sender: " + tx.Sender)
							fmt.Println("    Recipient: " + tx.Recipient)
							fmt.Println("    Amount: " + fmt.Sprintf("%f", tx.Amount))
							fmt.Println("    Nonce: " + fmt.Sprintf("%d", tx.Nonce))
							bandera = true
						}
					}
					if !bandera {
						fmt.Println("No se encontro la transaccion")
					}
					fmt.Println("------------------------------------------------------------------")
					fmt.Print("Presiona enter para continuar...")
					fmt.Scanln()
					ClearScreen()
				}

			case 5:

				var index int
				fmt.Print("Introduce el numero del bloque: ")
				fmt.Scanln(&index)

				key := fmt.Sprintf("%05d", index)

				if index > currentBlock.Index || index < 0 {
					fmt.Println("El bloque no existe")
					time.Sleep(2 * time.Second)
					ClearScreen()
				}

				retrievedData, err := blockdb.Get(key)
				if err != nil {
					errors.Wrap(err, "case 5 blockdb.Get error")
				}

				var resultBlock e.Block
				err = json.Unmarshal(retrievedData, &resultBlock)
				if err != nil {
					errors.Wrap(err, "case 5 json.Unmarshal error")
				}

				fmt.Println("------------------------------------------------------------------")
				fmt.Println("Index: " + fmt.Sprint(resultBlock.Index))
				fmt.Println("Timestamp: " + fmt.Sprint(resultBlock.Timestamp))
				fmt.Println("PreviousHash: " + fmt.Sprint(resultBlock.PreviousHash))
				fmt.Println("Hash: " + fmt.Sprint(resultBlock.Hash))
				fmt.Println("Transactions: ")
				for _, tx := range resultBlock.Transactions {
					if tx.Sender == inputUser || tx.Recipient == inputUser {
						fmt.Println("----------------")
						fmt.Println("Sender: " + fmt.Sprint(tx.Sender))
						fmt.Println("Recipient: " + fmt.Sprint(tx.Recipient))
						fmt.Println("Amount:" + fmt.Sprint(tx.Amount))
						fmt.Println("Nonce:" + fmt.Sprint(tx.Nonce))
						fmt.Println("----------------")
					}
				}
				fmt.Println("+++++++++++++++++++++++++++++++++++++++++++++++++++")
				fmt.Println("------------------------------------------------------------------")
				fmt.Print("Presiona enter para continuar...")
				fmt.Scanln()
				ClearScreen()

			case 6:
				blocks, err := blockdb.GetAllBlocks()
				if len(blocks) == 0 {
					fmt.Println("No hay Bloques")
					fmt.Print("Presiona enter para continuar...")
					fmt.Scanln()
					ClearScreen()
					break
				}
				if err != nil {
					errors.Wrap(err, "case 5 blockdb.GetAllBlocks error")
				}
				fmt.Println("------------------------------------------------------------------")
				for _, block := range blocks {
					fmt.Println("+++++++++++++++++++++++++++++++++++++++++++++++++++")
					fmt.Println("Index: " + fmt.Sprint(block.Index))
					fmt.Println("Timestamp: " + fmt.Sprint(block.Timestamp))
					fmt.Println("PreviousHash: " + fmt.Sprint(block.PreviousHash))
					fmt.Println("Hash: " + fmt.Sprint(block.Hash))
					fmt.Println("Transactions: ")
					for _, tx := range block.Transactions {
						if tx.Sender == inputUser || tx.Recipient == inputUser {
							fmt.Println("----------------")
							fmt.Println("Sender: " + fmt.Sprint(tx.Sender))
							fmt.Println("Recipient: " + fmt.Sprint(tx.Recipient))
							fmt.Println("Amount:" + fmt.Sprint(tx.Amount))
							fmt.Println("Nonce:" + fmt.Sprint(tx.Nonce))
							fmt.Println("----------------")
						}
					}
					fmt.Println("+++++++++++++++++++++++++++++++++++++++++++++++++++")
				}

				fmt.Print("Presiona enter para continuar...")
				fmt.Scanln()
				ClearScreen()
			case 7:
				fmt.Println("Cerrar Sesión...")
				time.Sleep(2 * time.Second)
				ClearScreen()
				bandera = true
			case 8:
				fmt.Println("Saliendo...")
				os.Exit(0)

			default:
				fmt.Println("Opción no válida")
			}

			if bandera {
				break
			}
		}
	}
}

func ClearScreen() {
	system := runtime.GOOS
	switch system {
	case "windows":
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	case "darwin", "linux":
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}

func stringToByteSlice(str string) ([]byte, error) {
	trimmedString := strings.Trim(str, "[]")

	byteStrings := strings.Fields(trimmedString)

	byteSlice := make([]byte, len(byteStrings))

	for i, s := range byteStrings {
		val, err := strconv.Atoi(s)
		if err != nil {
			return nil, err
		}
		byteSlice[i] = byte(val)
	}

	return byteSlice, nil
}
