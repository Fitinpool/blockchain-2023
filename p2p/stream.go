package p2p

//0x2ECE5812Ec971C14EFc5F868f5c25799e5a047fd
//0x08EDFf09Af7C456D9f03E62002e09A66052bb2E9
//[80 245 236 18 22 121 225 174 222 214 11 245 227 28 141 70 66 21 103 26 208 205 151 110 34 234 108 115 111 6 137 160]

import (
	"bufio"
	"context"
	"fmt"

	"github.com/libp2p/go-libp2p/core/network"
)

const (
	ProtocolDataSharing = "xulo-sharing"
)

func (n *Node) Start() {
	n.NetworkHost.SetStreamHandler((Protocol), func(stream network.Stream) {
		fmt.Printf("Received incoming connection on port %s [%s] \n", n.Port, stream.Conn().RemotePeer().String())

		err := stream.Close()
		if err != nil {
			fmt.Println("Error closing the stream:", err)
		}
	})
}

func (n *Node) HandleStream(s network.Stream) {
	fmt.Printf("Stream received by peer %s \n", s.ID())
}

func writeData(stream string, rw *bufio.ReadWriter) {
	rw.WriteString(fmt.Sprintf("%s\n", stream))
	rw.Flush()
}

func readData(rw *bufio.ReadWriter) {

	for {
		str, _ := rw.ReadString('\n')

		if str == "" {
			return
		}
		if str != "\n" {
			fmt.Printf("\x1b[32m%s\x1b[0m> ", str)
		}

	}
}

func (n *Node) UserPutShareData(data string, rw *bufio.ReadWriter) {
	fmt.Printf("UserPutShareData %s \n", data)
	go writeData(data, rw)
	go readData(rw)
}

func (n *Node) NuevoNodo(stream string, rw *bufio.ReadWriter) {
	go writeData(stream, rw)
	go readData(rw)
}

func (n *Node) SetupStreamHandler(ctx context.Context, handler network.StreamHandler) {
	n.NetworkHost.SetStreamHandler(ProtocolDataSharing, handler)
}
