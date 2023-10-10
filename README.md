# Blockchain 2023

## Descripción
Este repositorio contiene una implementación simple de una blockchain en el lenguaje Go. La blockchain está diseñada para manejar transacciones y usuarios, y cuenta con una estructura básica de bloques para almacenar transacciones.

## Componentes principales

### 1. Bloque (`block.go`)
Define la estructura básica de un bloque en la blockchain, incluyendo el hash anterior, el hash actual, las transacciones y la marca de tiempo.

### 2. Entidades (`entities.go`)
Define varias estructuras y tipos que son esenciales para el funcionamiento de la blockchain, como `Transaction`, `User`, y otros.

### 3. Punto de entrada (`main.go`)
Es el punto de entrada del programa. Aquí es donde se inicializa y se ejecuta la blockchain.

### 4. Almacenamiento (`store.go`)
Maneja el almacenamiento y la recuperación de bloques en la blockchain. Esto permite que la blockchain persista entre ejecuciones.

### 5. Transacción (`transaction.go`)
Define cómo se crean y se manejan las transacciones en la blockchain.

## Cómo empezar

1. Clona el repositorio.
2. Asegúrate de tener Go instalado en tu máquina.
3. Ejecuta `go run .` para iniciar la blockchain.

## Usuarios
Para ingresar más usuarios se debe añadir nombres a la cadena de la linea 30 del archivo main.go, todas las contraseñas se establecen como "asd"
