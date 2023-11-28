package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"

	e "blockchain/entities"

	"github.com/pkg/errors"
)

var currentBlock e.Block

func main() {

	// hostIP := flag.String("host-address", "0.0.0.0", "Default address to run node")
	// port := flag.String("port", "4000", "Port to enable network connection")
	// flag.Parse()

	// node, err := p2p.NewNode(&p2p.NodeConfig{
	// 	IP:   *hostIP,
	// 	Port: *port,
	// })

	// if err != nil {
	// 	errors.Wrap(err, "main: p2p.NewNode error")
	// }

	// blockdb, err := NewStore(fmt.Sprintf("node-%s", node.NetworkHost.ID().String()))
	// if err != nil {
	// 	errors.Wrap(err, "main: NewStore error")
	// }

	// defer blockdb.Close()

	// node.MdnsService.Start()
	// node.Start()

	// node.ConnectWithPeers(*peerAddresses)

	// generar datos para la transaccion
	// privKey, publicKey, address := GeneraLlavesYAddress()

	// data := &e.User{
	// 	PrivateKey:    privKey,
	// 	PublicKey:     publicKey,
	// 	Address:       address,
	// 	AccuntBalence: 1000,
	// 	Nonce:         0,
	// }

	userdb, err := NewStore("userdb")
	if err != nil {
		errors.Wrap(err, "main: NewStore error userdb")
	}

	// userdb.Put(node.NetworkHost.ID().String(), data)
	// defer userdb.Close()

	// err = blockdb.Put(address, data)
	// if err != nil {
	// 	errors.Wrap(err, "main: blockdb.Put error")
	// }

	// peerAddresses := node.MdnsService.PeerChan()

	// if *peerAddresses != "" {
	// 	node.ConnectWithPeers(strings.Split(*peerAddresses, ","))
	// }

	// libp2p.Option( libp2p.Identity(privKey))

	// userdb, err := NewStore("userdb")
	// if err != nil {
	// 	errors.Wrap(err, "NewStore error userdb")
	// }
	// defer userdb.Close()

	blockdb, err := NewStore("blockchain")
	if err != nil {
		errors.Wrap(err, "NewStore error blockchain")
	}

	blockNodedb, err := NewStore(fmt.Sprintf("node-%s", node.NetworkHost.ID().String()))
	if err != nil {
		errors.Wrap(err, "NewStore error blockchain")
	}

	go menu(blockNodedb, blockdb, userdb)

	for {
		isEmpty, err := blockdb.IsEmpty()
		if err != nil {
			errors.Wrap(err, "blockdb.IsEmpty error")
		}
		if isEmpty {
			currentBlock = CreateMainBlock()

			key := fmt.Sprintf("%05d", currentBlock.Index)
			time.Sleep(60 * time.Second)

			err = blockdb.Put(key, currentBlock)
			if err != nil {
				errors.Wrap(err, "blockdb.Put error")
			}
		} else {

			lastValues := blockdb.GetLastKey()
			if err != nil {
				errors.Wrap(err, "blockdb.GetLastKey error")
			}

			var result e.Block
			err = json.Unmarshal(lastValues, &result)
			if err != nil {
				errors.Wrap(err, "json.unmarshal error")
			}

			currentBlock = GenerateBlock(result.Index+1, result.Hash)
			key := fmt.Sprintf("%05d", currentBlock.Index)
			time.Sleep(60 * time.Second)

			err = blockdb.Put(key, currentBlock)
			if err != nil {
				errors.Wrap(err, "blockdb.Put error")
			}

		}
	}
}

func menu(blockNodedb *Store, blockdb *Store, userdb *Store) {
	var inputUser, inputPass string
	var resultUser e.User

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

				data := &e.User{
					PrivateKey:    privKey,
					PublicKey:     publicKey,
					Nombre:        userRegister,
					Password:      passRegister,
					Nonce:         0,
					AccuntBalence: 1000,
				}

				err := userdb.Put(address, data)

				if err != nil {
					errors.Wrap(err, "userdb.Put error")
				}

				fmt.Println("Datos guardados de " + userRegister + ", Address: " + address)
				fmt.Print("Presiona enter para continuar...")
				fmt.Scanln()
				ClearScreen()
				bandera = true

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
			fmt.Println("1. Ver Contactos")
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

				fmt.Println("Caso 1")
				users, err := userdb.GetAllUser()
				if err != nil {
					errors.Wrap(err, "case 1 blockdb.Get error")
				}

				fmt.Println("Contactos: ")

				fmt.Println("------------------------------------------------------------------")
				for i, user := range users {
					fmt.Println(fmt.Sprintf("%d. ", i+1) + user.Data.Nombre + " | Addres : " + fmt.Sprintf(user.Key))
				}
				fmt.Println("------------------------------------------------------------------")

				fmt.Print("Presiona enter para continuar...")
				fmt.Scanln()
				ClearScreen()

			case 2:
				var recipient string
				var amount float64

				fmt.Printf("Tienes un saldo de : %f", resultUser.AccuntBalence)
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
					errors.Wrap(err, "case 2 userdb.Put error")
				}

				recipientPut, err := userdb.Get(recipient)
				if err != nil {
					errors.Wrap(err, "No existe o Error")
				}

				var recipientResult e.User
				err = json.Unmarshal(recipientPut, &recipientResult)
				if err != nil {
					errors.Wrap(err, "case 2 json.Unmarshal error")
				}

				recipientResult.AccuntBalence = recipientResult.AccuntBalence + amount

				err = userdb.Put(recipient, recipientResult)
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

				time.Sleep(2 * time.Second)
				ClearScreen()
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
			//fmt.Println("5. Buscar un bloque")
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
