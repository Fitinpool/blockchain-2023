package p2p

//0x0731eA28Ff6Ff1D6c2cdc3B03133A3D5dcD743Ee    /ip4/127.0.0.1/tcp/4000/p2p/12D3KooWBqiZcXfhBv1TveGTm45LisxcB7hqSMmgoqnhUZokZEY8
// 0x75c0e3e50994834E368E053C38560072D785B8Bc    /ip4/127.0.0.1/tcp/4001/p2p/12D3KooWKeGkjwnAwfWTH45bmfS51kyykxPiacRPCiefZMbBBiXe
// 0x7CDa093BF50e5db8120a4396bd9CF981206A80aA		/ip4/127.0.0.1/tcp/4002/p2p/12D3KooWGzA77dES5rPbNSVVEyw1xDvySC517zm1gawCLQN4pkWt
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
