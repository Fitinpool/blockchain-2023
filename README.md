# Blockchain 2023

## Descripción
Este repositorio contiene una implementación simple de una blockchain en el lenguaje Go. La blockchain está diseñada para manejar transacciones y usuarios, y cuenta con una estructura básica de bloques para almacenar transacciones. Además, esta permite el uso de múltiples nodos conectados a un Nodo Full.

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

## Usuarios
El manejo de usuarios está definido en el menú de inicio de cualquier nodo, en el cual se pueden crear en cualquier nodo siempre y cuando estos se encuentren conectados al Full Node.
Para el inicio de sesión, este utiliza la Address del usuario y su contraseña.

## Cómo empezar

1. Clona el repositorio.
2. Asegúrate de tener Go instalado en tu máquina.
3. Ejecuta `go run . -publish` para iniciar el Full Node.
4. Ejecuta `go run . -port "Puerto" -connect "Node Address"` para iniciar un Light Node conectado al Full Node, con `Puerto` distinto a 3000 y 4000 (Y también a otros Light Node), y con `Node Address` al Node Address (Representado en el menú) del Full Node.

## Consideraciones

* Cada bloque en la blockchain se genera cada 1 minuto para modo de ver su uso de manera más rápida, este valor es modificable en la línea 447 y 487 del archivo (`main.go`), si no hay transacciones dentro de ese tiempo la blockchain introducirá el bloque vacío.

* La API implementada va a conectada al nodo FULL

* Cada vez que se ejecuta el `go run . -publish` se crearán nuevos archivos asociado a las bases y modelo de datos, por lo que es importante eliminar los archivos `blockchain`, `userdb` y la carpeta `data` en caso de realizar cualquier cambio a la implementacion.

* Todo el código está probado en Linux y MacOS.
