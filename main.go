package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	e "blockchain/entities"

	"github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"

	"blockchain/p2p"

	"github.com/pkg/errors"
)

var currentBlock e.Block
var state bool = false
var resultUser e.User
var inputUser string

func main() {

	port := flag.Int("port", 4000, "Puerto de escucha para el nodo")
	protocol := flag.String("protocol", p2p.Protocol, "Protocol to enable network connection")
	connectTo := flag.String("connect", "", "Dirección multiaddr del nodo a conectar")
	isPublisher := flag.Bool("publish", false, "Define si el nodo publicará mensajes")
	flag.Parse()

	ctx := context.Background()

	// Crear un host libp2p
	h, err := libp2p.New(
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", *port)),
	)
	if err != nil {
		panic(err)
	}
	defer h.Close()

	fmt.Printf("Nodo iniciado con ID: %s\n", h.ID().String())

	// blockdb, err := NewStore("blockchain")
	// if err != nil {
	// 	errors.Wrap(err, "main: block NewStore error")
	// }

	userdb, err := NewStore("userdb")
	if err != nil {
		errors.Wrap(err, "main: user NewStore error")
	}

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

	// // Publicar mensajes si es un nodo publicador
	// if *isPublisher {
	// 	go func() {
	// 		for {
	// 			time.Sleep(5 * time.Second)
	// 			err := topic.Publish(ctx, []byte("Hello, world!"))
	// 			if err != nil {
	// 				panic(err)
	// 			}
	// 			fmt.Println("Mensaje publicado")
	// 		}
	// 	}()
	// }

	// go func() {
	// 	for {
	// 		msg, err := sub.Next(ctx)
	// 		if err != nil {
	// 			panic(err)
	// 		}

	// 		fmt.Printf("Mensaje recibido: %s\n", string(msg.Data))
	// 	}
	// }()

	go func() {
		for {
			if *isPublisher {
				msg, err := subFullNode.Next(ctx)
				if err != nil {
					panic(err)
				}

				protocoloFullNode := strings.Split(string(msg.Data), ";")

				// fmt.Println("Mensaje recibido: " + protocoloFullNode[1])

				// retrievedData, err := userdb.GetAllUser()
				// if err != nil {
				// 	errors.Wrap(err, "userdb.Get error")
				// }

				// for _, user := range retrievedData {
				// 	fmt.Println(string(user.Data.Password))
				// }

				if protocoloFullNode[0] == "nuevo-user" {
					var jsonTemp map[string]interface{}

					err = json.Unmarshal([]byte(protocoloFullNode[1]), &jsonTemp)
					if err != nil {
						errors.Wrap(err, "main: json.Unmarshal error")
					}

					userTemp := e.User{
						PrivateKey:    []byte(jsonTemp["private_key"].(string)),
						PublicKey:     []byte(jsonTemp["public_key"].(string)),
						Nombre:        jsonTemp["nombre"].(string),
						Password:      jsonTemp["password"].(string),
						Nonce:         int(jsonTemp["nonce"].(float64)),
						AccuntBalence: jsonTemp["accunt_balance"].(float64),
					}

					// err = json.Unmarshal([]byte(protocoloFullNode[1]), &userTemp)
					// if err != nil {
					// 	errors.Wrap(err, "main: json.Unmarshal error")
					// }

					err = userNodedb.Put(string(protocoloFullNode[2]), userTemp)

					if err != nil {
						errors.Wrap(err, "main: userNodedb.Put error")
					}

					err = userdb.Put(string(protocoloFullNode[2]), userTemp)
					if err != nil {
						errors.Wrap(err, "main: userdb.Put error")
					}

					err := topicBroadcast.Publish(ctx, []byte("aprobado-user;"+protocoloFullNode[1]+";"+protocoloFullNode[2]))
					if err != nil {
						panic(err)
					}

					state = true
				} else if protocoloFullNode[0] == "nueva-transaccion" {

					var jsonTemp map[string]interface{}

					err = json.Unmarshal([]byte(protocoloFullNode[1]), &jsonTemp)
					if err != nil {
						errors.Wrap(err, "main: json.Unmarshal error")
					}

					txTemp := &e.Transaction{
						Sender:    jsonTemp["sender"].(string),
						Recipient: jsonTemp["recipient"].(string),
						Amount:    jsonTemp["amount"].(float64),
						Nonce:     int(jsonTemp["nonce"].(float64)),
					}

					FirmaTransaccion(txTemp, resultUser.PrivateKey)

					currentBlock.Transactions = append(currentBlock.Transactions, *txTemp)

					resultUser.Nonce = resultUser.Nonce + 1
					resultUser.AccuntBalence = resultUser.AccuntBalence - jsonTemp["amount"].(float64)

					err = userdb.Put(inputUser, resultUser)
					if err != nil {
						errors.Wrap(err, "case 2 userdb.Put error")
					}

					err = userNodedb.Put(inputUser, resultUser)
					if err != nil {
						errors.Wrap(err, "case 2 userdb.Put error")
					}

					recipientPut, err := userdb.Get(jsonTemp["recipient"].(string))
					if err != nil {
						errors.Wrap(err, "No existe o Error")
					}

					var recipientResult e.User
					err = json.Unmarshal(recipientPut, &recipientResult)
					if err != nil {
						errors.Wrap(err, "case 2 json.Unmarshal error")
					}

					recipientResult.AccuntBalence = recipientResult.AccuntBalence + jsonTemp["amount"].(float64)

					err = userdb.Put(jsonTemp["recipient"].(string), recipientResult)
					if err != nil {
						errors.Wrap(err, "case 2 userdb.Put error")
					}

					err = userNodedb.Put(jsonTemp["recipient"].(string), recipientResult)
					if err != nil {
						errors.Wrap(err, "case 2 userdb.Put error")
					}

					newResult, err := userdb.Get(inputUser)
					if err != nil {
						errors.Wrap(err, "userdb.Get error")
					}

					err = json.Unmarshal(newResult, &resultUser)
					if err != nil {
						errors.Wrap(err, "case 2 json.Unmarshal error")
					}

					err = topicBroadcast.Publish(ctx, []byte("aprobado-block;"+protocoloFullNode[1]))
					if err != nil {
						panic(err)
					}

					state = true
				}

			} else {
				msg, err := subBroadcast.Next(ctx)
				if err != nil {
					panic(err)
				}

				protocoloBroadcast := strings.Split(string(msg.Data), ";")

				if protocoloBroadcast[0] == "aprobado-user" {
					// var userTemp e.User

					// err = json.Unmarshal([]byte(protocoloBroadcast[1]), &userTemp)
					// if err != nil {
					// 	errors.Wrap(err, "main: json.Unmarshal error")
					// }

					var jsonTemp map[string]interface{}

					err = json.Unmarshal([]byte(protocoloBroadcast[1]), &jsonTemp)
					if err != nil {
						errors.Wrap(err, "main: json.Unmarshal error")
					}

					userTemp := e.User{
						PrivateKey:    []byte(jsonTemp["private_key"].(string)),
						PublicKey:     []byte(jsonTemp["public_key"].(string)),
						Nombre:        jsonTemp["nombre"].(string),
						Password:      jsonTemp["password"].(string),
						Nonce:         int(jsonTemp["nonce"].(float64)),
						AccuntBalence: jsonTemp["accunt_balance"].(float64),
					}

					err = userNodedb.Put(string(protocoloBroadcast[2]), userTemp)

					if err != nil {
						errors.Wrap(err, "main: userNodedb.Put error")
					}

					state = true
				} else if protocoloBroadcast[0] == "aprobado-block" {
					var jsonTemp map[string]interface{}

					err = json.Unmarshal([]byte(protocoloBroadcast[1]), &jsonTemp)
					if err != nil {
						errors.Wrap(err, "main: json.Unmarshal error")
					}

					err = userNodedb.Put(inputUser, resultUser)
					if err != nil {
						errors.Wrap(err, "case 2 userdb.Put error")
					}

					recipientPut, err := userNodedb.Get(jsonTemp["recipient"].(string))
					if err != nil {
						errors.Wrap(err, "No existe o Error")
					}

					var recipientResult e.User
					err = json.Unmarshal(recipientPut, &recipientResult)
					if err != nil {
						errors.Wrap(err, "case 2 json.Unmarshal error")
					}

					recipientResult.AccuntBalence = recipientResult.AccuntBalence + jsonTemp["amount"].(float64)

					err = userNodedb.Put(jsonTemp["recipient"].(string), recipientResult)
					if err != nil {
						errors.Wrap(err, "case 2 userdb.Put error")
					}

					state = true
				}
			}
		}
	}()

	if *protocol == p2p.Protocol {

		CopyStore("blockchain", fmt.Sprintf("node-block-%s", h.ID().String()))
		CopyStore("userdb", fmt.Sprintf("node-user-%s", h.ID().String()))
		go menu(blockNodedb, userNodedb, topicFullNode, subBroadcast)

		for {
			isEmpty, err := blockNodedb.IsEmpty()
			if err != nil {
				errors.Wrap(err, "main: blockdb.IsEmpty error")
			}
			if isEmpty {
				currentBlock = CreateMainBlock()

				key := fmt.Sprintf("%05d", currentBlock.Index)
				time.Sleep(60 * time.Second)

				err = blockNodedb.Put(key, currentBlock)
				if err != nil {
					errors.Wrap(err, "main: blockNodedb.Put error")
				}
			} else {

				lastValues := blockNodedb.GetLastKey()
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
				time.Sleep(60 * time.Second)

				err = blockNodedb.Put(key, currentBlock)
				if err != nil {
					errors.Wrap(err, "main: blockdb.Put error")
				}

			}
		}
	} else {
		fmt.Println("Protocolo no valido")
	}

}

func menu(blockdb *Store, userdb *Store, topicFullNode *pubsub.Topic, subBroadcast *pubsub.Subscription) {
	var inputPass string

	for {
		for {

			var option int
			var bandera bool = false
			var userRegister, passRegister string
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

				dataString := fmt.Sprintf(`{
					"private_key":    "%v",
					"public_key":     "%v",
					"nombre":        "%s",
					"password":      "%s",
					"nonce":         %d,
					"accunt_balance": %d
				};%s`, privKey, publicKey, userRegister, passRegister, 0, 1000, address)

				topicFullNode.Publish(context.Background(), []byte("nuevo-user;"+dataString))

				for {
					if state {
						break
					}
					fmt.Print("Esperando respuesta...")
					time.Sleep(2 * time.Second)
					ClearScreen()
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
			var option int
			var bandera bool = false
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
				var amount float64

				fmt.Printf("Saldo : %f", resultUser.AccuntBalence)
				fmt.Println("")
				fmt.Print("Introduce el destinatario: ")
				fmt.Scanln(&recipient)

				fmt.Print("Introduce la cantidad: ")
				fmt.Scanln(&amount)

				if amount > resultUser.AccuntBalence {
					fmt.Println("No tienes saldo suficiente")
					time.Sleep(2 * time.Second)
					ClearScreen()
					break
				}

				// tx := &e.Transaction{
				// 	Sender:    inputUser,
				// 	Recipient: recipient,
				// 	Amount:    amount,
				// 	Nonce:     resultUser.Nonce + 1,
				// }

				txShare := fmt.Sprintf(`{
					"sender":    %s,
					"recipient": %s,
					"amount":    %f,
					"nonce":     %d,
				)`, inputUser, recipient, amount, resultUser.Nonce+1)

				err := topicFullNode.Publish(context.Background(), []byte("nueva-transaccion;"+txShare))
				if err != nil {
					panic(err)
				}

				// FirmaTransaccion(tx, resultUser.PrivateKey)

				// currentBlock.Transactions = append(currentBlock.Transactions, *tx)

				fmt.Println("Transaccion agregada en el bloque: " + fmt.Sprintf("%d", currentBlock.Index))
				fmt.Println("Nonce:" + fmt.Sprint(resultUser.Nonce))

				// resultUser.Nonce = resultUser.Nonce + 1
				// resultUser.AccuntBalence = resultUser.AccuntBalence - amount

				// err = userdb.Put(inputUser, resultUser)
				// if err != nil {
				// 	errors.Wrap(err, "case 2 userdb.Put error")
				// }

				// recipientPut, err := userdb.Get(recipient)
				// if err != nil {
				// 	errors.Wrap(err, "No existe o Error")
				// }

				// var recipientResult e.User
				// err = json.Unmarshal(recipientPut, &recipientResult)
				// if err != nil {
				// 	errors.Wrap(err, "case 2 json.Unmarshal error")
				// }

				// recipientResult.AccuntBalence = recipientResult.AccuntBalence + amount

				// err = userdb.Put(recipient, recipientResult)
				// if err != nil {
				// 	errors.Wrap(err, "case 2 userdb.Put error")
				// }

				// newResult, err := userdb.Get(inputUser)
				// if err != nil {
				// 	errors.Wrap(err, "userdb.Get error")
				// }

				// err = json.Unmarshal(newResult, &resultUser)
				// if err != nil {
				// 	errors.Wrap(err, "case 2 json.Unmarshal error")
				// }

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
