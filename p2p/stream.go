package p2p

//0xe2f4CdCc8CFD336bCa171fD00B68C81D7fbe29D3 asd
//0xD0b302a4c4164f8b1403Dc9b0AA1Ef3b6F073d00 lol
//0xbAF9233923227D82B70bEa3B477Bd9E980111667 lp
//0x167d8f1111bBda6Dd106F801c977d74B2A86bef9 asd
//0x504dD42C56F274687f29F88e71DaeCe478e1fe03
//0xA1F3eB8c17a8D3b78B297837B1F28D8E251b275B
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
