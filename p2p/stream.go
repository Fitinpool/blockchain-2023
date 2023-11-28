package p2p

import (
    "bufio"
    "context"
    "fmt"

    "github.com/libp2p/go-libp2p/core/network"
)

const (
    ProtocolDataSharing = "xulo-sharing"
)

func (n Node) Start() {
    n.NetworkHost.SetStreamHandler((Protocol), func(stream network.Stream) {
        fmt.Printf("Received incoming connection on port %s [%s] \n", n.Port, stream.Conn().RemotePeer().String())

        err := stream.Close()
        if err != nil {
            fmt.Println("Error closing the stream:", err)
        }
    })
}

func (nNode) HandleStream(s network.Stream) {
    fmt.Printf("Stream received by peer %s \n", s.ID())
    rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

    n.NuevoNodo("Hola esto es un broadcast \n", rw)

}

func writeData(stream string, rw bufio.ReadWriter) {
    rw.WriteString(fmt.Sprintf("%s\n", stream))
    rw.Flush()
}

func readData(rwbufio.ReadWriter) {

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

func (n Node) NuevoNodo(stream string, rwbufio.ReadWriter) {
    go writeData(stream, rw)
    go readData(rw)
}

func (n *Node) SetupStreamHandler(ctx context.Context, handler network.StreamHandler) {
    n.NetworkHost.SetStreamHandler(ProtocolDataSharing, handler)
}