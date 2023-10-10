package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/pkg/errors"
)

var currentBlock Block

func main() {

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

	privKey, publicKey, address := GeneraLlavesYAddress()

	data := &User{
		PrivateKey:    privKey,
		PublicKey:     publicKey,
		Nombre:        "Julio",
		Password:      "asd",
		Nonce:         0,
		AccuntBalence: 1000,
	}

	err = userdb.Put(address, data)
	if err != nil {
		errors.Wrap(err, "userdb.Put error")
	}
	fmt.Println("Datos guardados de Julio, Address: " + address)

	privKey, publicKey, address = GeneraLlavesYAddress()

	data = &User{
		PrivateKey:    privKey,
		PublicKey:     publicKey,
		Nombre:        "Vania",
		Password:      "asd",
		Nonce:         0,
		AccuntBalence: 1000,
	}

	err = userdb.Put(address, data)
	if err != nil {
		errors.Wrap(err, "userdb.Put error")
	}
	fmt.Println("Datos guardados de Vania, Address: " + address)

	fmt.Print("Presiona enter para continuar...")
	fmt.Scanln()
	ClearScreen()

	go menu(userdb, blockdb)

	for {
		isEmpty, err := blockdb.IsEmpty()
		if err != nil {
			errors.Wrap(err, "blockdb.IsEmpty error")
		}
		if isEmpty {
			currentBlock = CreateMainBlock()

			key := fmt.Sprintf("%d", currentBlock.Index)
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

			var result Block
			err = json.Unmarshal(lastValues, &result)
			if err != nil {
				errors.Wrap(err, "json.unmarshal error")
			}

			currentBlock = GenerateBlock(result.Index+1, result.Hash)
			key := fmt.Sprintf("%d", currentBlock.Index)
			time.Sleep(60 * time.Second)

			err = blockdb.Put(key, currentBlock)
			if err != nil {
				errors.Wrap(err, "blockdb.Put error")
			}

		}
	}
}

func menu(userdb *Store, blockdb *Store) {
	var inputUser, inputPass string
	var result User

	for {
		for {
			fmt.Print("Introduce el usuario: ")
			fmt.Scanln(&inputUser)

			fmt.Print("Introduce la contraseña: ")
			fmt.Scanln(&inputPass)

			retrievedData, err := userdb.Get(inputUser)
			if err != nil {
				errors.Wrap(err, "userdb.Get error")
			}

			err = json.Unmarshal(retrievedData, &result)
			if err != nil {
				errors.Wrap(err, "user json.Unmarshal error")
			}

			if result.Password != inputPass {
				fmt.Println("Usuario o contraseña incorrectos.")
				time.Sleep(2 * time.Second)
				ClearScreen()
			} else {
				fmt.Println("Credenciales Correctas.")
				time.Sleep(2 * time.Second)
				ClearScreen()
				break
			}
		}

		for {
			var option int
			var bandera bool = false
			fmt.Println("\n---------- Menú ----------")
			fmt.Println("Bienvenid@ " + result.Nombre + "! 🎉")
			fmt.Println("1. Ver Contactos")
			fmt.Println("2. Realizar transacción")
			fmt.Println("3. Mostrar transacciones")
			fmt.Println("4. Buscar un bloque")
			fmt.Println("5. Buscar transacción específica")
			fmt.Println("6. Mostrar todos los bloques")
			fmt.Println("7. Cerar Sesión")
			fmt.Println("8. Salir")
			fmt.Print("Elige una opción: ")
			fmt.Scan(&option)

			switch option {
			case 1:

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

				fmt.Printf("Tienes un saldo de : %f", result.AccuntBalence)
				fmt.Println("")
				fmt.Print("Introduce el destinatario: ")
				fmt.Scanln(&recipient)

				fmt.Print("Introduce la cantidad: ")
				fmt.Scanln(&amount)

				if amount > result.AccuntBalence {
					fmt.Println("No tienes saldo suficiente")
					time.Sleep(2 * time.Second)
					ClearScreen()
					break
				}

				tx := &Transaction{
					Sender:    inputUser,
					Recipient: recipient,
					Amount:    amount,
					Nonce:     result.Nonce + 1,
				}

				FirmaTransaccion(tx, result.PrivateKey)

				currentBlock.Transactions = append(currentBlock.Transactions, *tx)

				fmt.Println("Transaccion agregada en el bloque: " + fmt.Sprintf("%d", currentBlock.Index))

				result.Nonce = result.Nonce + 1
				result.AccuntBalence = result.AccuntBalence - amount

				err := userdb.Put(inputUser, result)
				if err != nil {
					errors.Wrap(err, "case 2 userdb.Put error")
				}

				recipientPut, err := userdb.Get(recipient)
				if err != nil {
					errors.Wrap(err, "No existe o Error")
				}

				var recipientResult User
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

				err = json.Unmarshal(newResult, &result)
				if err != nil {
					errors.Wrap(err, "case 2 json.Unmarshal error")
				}

				time.Sleep(2 * time.Second)
				ClearScreen()
			// case 2:
			// 	fmt.Println("Transactions: ")
			// 	for _, tx := range result.Transaccion {
			// 		fmt.Println("+++++++++++++++++++++++++++++++++++++++++++++++++++")
			// 		fmt.Println("+ " + fmt.Sprint(tx))
			// 	}
			// 	fmt.Println("+++++++++++++++++++++++++++++++++++++++++++++++++++")
			// 	fmt.Print("Presiona enter para continuar...")
			// 	fmt.Scanln()
			// 	ClearScreen()
			// case 3:
			// 	var index int
			// 	fmt.Print("Introduce el numero del bloque: ")
			// 	fmt.Scanln(&index)

			// 	key := fmt.Sprintf("%d", index)

			// 	if index > currentBlock.Index {
			// 		fmt.Println("El bloque no existe")
			// 		time.Sleep(2 * time.Second)
			// 		ClearScreen()

			// 	} else if index == currentBlock.Index {
			// 		fmt.Println("------------------------------------------------------------------")
			// 		fmt.Println("Block: " + fmt.Sprintf("%d", currentBlock.Index))
			// 		fmt.Println("Timestamp: " + fmt.Sprintf("%d", currentBlock.Timestamp))
			// 		fmt.Println("PreviousHash: " + currentBlock.PreviousHash)
			// 		fmt.Println("Hash: " + currentBlock.Hash)
			// 		fmt.Println("Transactions: ")
			// 		for _, tx := range currentBlock.Transactions {
			// 			fmt.Println("+++++++++++++++++++++++++++++++++++++++++++++++++++")
			// 			fmt.Println("    Sender: " + tx.Sender)
			// 			fmt.Println("    Recipient: " + tx.Recipient)
			// 			fmt.Println("    Amount: " + fmt.Sprintf("%f", tx.Amount))
			// 			fmt.Println("    Signature: " + tx.Signature)
			// 			fmt.Println("    PubKey: " + tx.PubKey)
			// 			fmt.Println("    Nonce: " + fmt.Sprintf("%d", tx.Nonce))
			// 		}
			// 		fmt.Println("+++++++++++++++++++++++++++++++++++++++++++++++++++")
			// 		fmt.Println("------------------------------------------------------------------")
			// 		fmt.Print("Presiona enter para continuar...")
			// 		fmt.Scanln()
			// 		ClearScreen()
			// 	} else {
			// 		retrievedData, err := blockdb.Get(key)
			// 		if err != nil {
			// 			errors.Wrap(err, "case 3 blockdb.Get error")
			// 		}

			// 		var resultBlock Block
			// 		err = json.Unmarshal(retrievedData, &resultBlock)
			// 		if err != nil {
			// 			errors.Wrap(err, " case 3 json.Unmarshal error")
			// 		}
			// 		//imprimir el bloque y todos sus datos
			// 		fmt.Println("------------------------------------------------------------------")
			// 		fmt.Println("Block: " + fmt.Sprintf("%d", resultBlock.Index))
			// 		fmt.Println("Timestamp: " + fmt.Sprintf("%d", resultBlock.Timestamp))
			// 		fmt.Println("PreviousHash: " + resultBlock.PreviousHash)
			// 		fmt.Println("Hash: " + resultBlock.Hash)
			// 		fmt.Println("Transactions: ")
			// 		for _, tx := range resultBlock.Transactions {
			// 			fmt.Println("+++++++++++++++++++++++++++++++++++++++++++++++++++")
			// 			fmt.Println("    Sender: " + tx.Sender)
			// 			fmt.Println("    Recipient: " + tx.Recipient)
			// 			fmt.Println("    Amount: " + fmt.Sprintf("%f", tx.Amount))
			// 			fmt.Println("    Signature: " + tx.Signature)
			// 			fmt.Println("    PubKey: " + tx.PubKey)
			// 			fmt.Println("    Nonce: " + fmt.Sprintf("%d", tx.Nonce))
			// 		}
			// 		fmt.Println("+++++++++++++++++++++++++++++++++++++++++++++++++++")
			// 		fmt.Println("------------------------------------------------------------------")
			// 		fmt.Print("Presiona enter para continuar...")
			// 		fmt.Scanln()
			// 		ClearScreen()
			// 	}
			// case 4:
			// 	var index, nonce int
			// 	fmt.Print("Introduce el numero del bloque: ")
			// 	fmt.Scanln(&index)
			// 	fmt.Print("Introduce el nonce de la transaccion: ")
			// 	fmt.Scanln(&nonce)

			// 	key := fmt.Sprintf("%d", index)
			// 	nonceString := fmt.Sprintf("%d", nonce)

			// 	if index > currentBlock.Index {
			// 		fmt.Println("El bloque no existe")
			// 		time.Sleep(2 * time.Second)
			// 		ClearScreen()
			// 	} else if index == currentBlock.Index {
			// 		fmt.Println("------------------------------------------------------------------")
			// 		bandera := false
			// 		for _, tx := range currentBlock.Transactions {
			// 			if fmt.Sprint(tx.Nonce) == nonceString {
			// 				fmt.Println("    Sender: " + tx.Sender)
			// 				fmt.Println("    Recipient: " + tx.Recipient)
			// 				fmt.Println("    Amount: " + fmt.Sprintf("%f", tx.Amount))
			// 				fmt.Println("    Signature: " + tx.Signature)
			// 				fmt.Println("    PubKey: " + tx.PubKey)
			// 				fmt.Println("    Nonce: " + fmt.Sprintf("%d", tx.Nonce))
			// 				bandera = true
			// 			}
			// 		}
			// 		if !bandera {
			// 			fmt.Println("No se encontro la transaccion")
			// 		}
			// 		fmt.Println("------------------------------------------------------------------")
			// 		fmt.Print("Presiona enter para continuar...")
			// 		fmt.Scanln()
			// 		ClearScreen()
			// 	} else {
			// 		retrievedData, err := blockdb.Get(key)
			// 		if err != nil {
			// 			errors.Wrap(err, "case 4 blockdb.Get error")
			// 		}

			// 		var resultBlock Block
			// 		err = json.Unmarshal(retrievedData, &resultBlock)
			// 		if err != nil {
			// 			errors.Wrap(err, "case 4 json.Unmarshal error")
			// 		}

			// 		//imprimir el bloque y todos sus datos
			// 		fmt.Println("------------------------------------------------------------------")
			// 		bandera := false
			// 		for _, tx := range resultBlock.Transactions {
			// 			if fmt.Sprint(tx.Nonce) == nonceString {
			// 				fmt.Println("    Sender: " + tx.Sender)
			// 				fmt.Println("    Recipient: " + tx.Recipient)
			// 				fmt.Println("    Amount: " + fmt.Sprintf("%f", tx.Amount))
			// 				fmt.Println("    Signature: " + tx.Signature)
			// 				fmt.Println("    PubKey: " + tx.PubKey)
			// 				fmt.Println("    Nonce: " + fmt.Sprintf("%d", tx.Nonce))
			// 				bandera = true
			// 			}
			// 		}
			// 		if !bandera {
			// 			fmt.Println("No se encontro la transaccion")
			// 		}
			// 		fmt.Println("------------------------------------------------------------------")
			// 		fmt.Print("Presiona enter para continuar...")
			// 		fmt.Scanln()
			// 		ClearScreen()
			// 	}

			// case 5:
			// 	blocks, err := blockdb.GetAllBlocks()
			// 	if len(blocks) == 0 {
			// 		fmt.Println("No hay bloques")
			// 		fmt.Print("Presiona enter para continuar...")
			// 		fmt.Scanln()
			// 		ClearScreen()
			// 		break
			// 	}
			// 	if err != nil {
			// 		errors.Wrap(err, "case 7 blockdb.GetAllBlocks error")
			// 	}
			// 	for _, block := range blocks {
			// 		fmt.Println("------------------------------------------------------------------")
			// 		fmt.Println("Block " + fmt.Sprintf("%d", block.Index))
			// 		fmt.Println("Timestamp: " + fmt.Sprintf("%d", block.Timestamp))
			// 		fmt.Println("PreviousHash: " + block.PreviousHash)
			// 		fmt.Println("Hash: " + block.Hash)
			// 		fmt.Println("Transactions: ")
			// 		for _, tx := range block.Transactions {
			// 			fmt.Println("+++++++++++++++++++++++++++++++++++++++++++++++++++")
			// 			fmt.Println("    Sender: " + tx.Sender)
			// 			fmt.Println("    Recipient: " + tx.Recipient)
			// 			fmt.Println("    Amount: " + fmt.Sprintf("%f", tx.Amount))
			// 			fmt.Println("    Signature: " + tx.Signature)
			// 			fmt.Println("    PubKey: " + tx.PubKey)
			// 			fmt.Println("    Nonce: " + fmt.Sprintf("%d", tx.Nonce))
			// 		}
			// 		fmt.Println("------------------------------------------------------------------")
			// 	}
			// 	fmt.Print("Presiona enter para continuar...")
			// 	fmt.Scanln()
			// 	ClearScreen()

			case 6:
				fmt.Println("Cerrar Sesión...")
				time.Sleep(2 * time.Second)
				ClearScreen()
				bandera = true
			case 7:
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
