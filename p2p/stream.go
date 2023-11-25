package p2p

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"

	"github.com/libp2p/go-libp2p/core/network"
)

const (
	ProtocolDataSharing = "data-sharing"
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

	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

	go writeData(rw)
	go readData(rw)
}

func writeData(rw *bufio.ReadWriter) {
	stdReader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		sendData, err := stdReader.ReadString('\n')
		if err != nil {
			log.Println(err)
			return
		}

		rw.WriteString(fmt.Sprintf("%s\n", sendData))
		rw.Flush()
	}
}
func readData(rw *bufio.ReadWriter) {
	for {
		str, _ := rw.ReadString('\n')

		if str == "" {
			return
		}
		if str != "\n" {
			// Green console colour: 	\x1b[32m
			// Reset console colour: 	\x1b[0m
			fmt.Printf("\x1b[32m%s\x1b[0m> ", str)
		}

	}
}

func (n *Node) SetupStreamHandler(ctx context.Context, handler network.StreamHandler) {
	n.NetworkHost.SetStreamHandler(ProtocolDataSharing, handler)
}
