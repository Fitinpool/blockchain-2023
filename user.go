package main

import (
	"fmt"

	"github.com/pkg/errors"
)

func CreateUser(db *Store, username string, password string) {
	err := db.Put(username, password)
	if err != nil {
		errors.Wrap(err, "CreateUser db.Put error")
	}

	fmt.Println("Usuario creado con éxito!")
}

func VerifyUser(db *Store, username string, password string) (bool, error) {
	storedPassword, err := db.Get(username)
	if err != nil {
		return false, errors.Wrap(err, "VerifyUser db.Get error")
	}

	if password == string(storedPassword) {
		fmt.Println("¡Acceso concedido!")
		return true, nil
	} else {
		fmt.Println("Contraseña incorrecta.")
		return false, nil
	}
}
