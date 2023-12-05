package p2p

import (
	"context"
	"fmt"
	"sync"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
)

const (
	Protocol  = "/xulo/1.0.0"
	Namespace = "NETWORK"

	namespace = "PROTOCOL-XULO-PEERS"
)

type Node struct {
	NetworkHost    host.Host
	MdnsService    mdns.Service
	ConnectedPeers map[string]*peer.AddrInfo
	Port           int
	mu             sync.Mutex
	fullNode       bool
}

type NodeConfig struct {
	IP       string
	Port     int
	FullNode bool
}

func NewNode(config *NodeConfig) (*Node, error) {
	listenAddress := fmt.Sprintf("/ip4/%s/tcp/%d", config.IP, config.Port)
	address := libp2p.ListenAddrStrings(listenAddress)
	host, err := libp2p.New(address)
	if err != nil {
		return nil, errors.Wrap(err, "p2p: NewNode libp2p.New error")
	}
	peerInfo := peer.AddrInfo{
		ID:    host.ID(),
		Addrs: host.Addrs(),
	}

	addrs, err := peer.AddrInfoToP2pAddrs(&peerInfo)
	if err != nil {
		return nil, errors.Wrap(err, "p2p: NewNode peer.AddrInfoToP2pAddrs error")
	}

	fmt.Println("--------------------------------------------------")
	fmt.Printf("Node address: %s\n", addrs[0])
	fmt.Println("--------------------------------------------------")

	notifee := &discoveryveryNotifee{}

	mdnsService := mdns.NewMdnsService(
		host,
		namespace,
		notifee,
	)

	notifee.node = &Node{
		NetworkHost:    host,
		MdnsService:    mdnsService,
		ConnectedPeers: make(map[string]*peer.AddrInfo),
		Port:           config.Port,
		fullNode:       config.FullNode,
	}

	return notifee.node, nil
}

func (n *Node) ConnectWithPeers(peerAddress string) error {

	peerMultiAddr, err := multiaddr.NewMultiaddr(peerAddress)
	if err != nil {
		err = errors.Wrap(err, "p2p: Node.ConnectWithPeers multiaddr.NewMultiaddr error")
		fmt.Printf("%s with peer %s", err.Error(), peerAddress)
		return err
	}

	peerAddrInfo, err := peer.AddrInfoFromP2pAddr(peerMultiAddr)
	if err != nil {
		err = errors.Wrap(err, "p2p: Node.ConnectWithPeers peer.AddrInfoFromP2pAddr error")
		fmt.Printf("%s with peer %s", err.Error(), peerAddress)
		return err
	}

	n.mu.Lock()
	err = n.NetworkHost.Connect(context.Background(), *peerAddrInfo)
	if err != nil {
		err = errors.Wrap(err, "p2p: Node.ConnectWithPeers n.NetworkHost.Connect error")
		fmt.Printf("%s with peer %s", err.Error(), peerAddress)
		return err
	}
	fmt.Println("--------------------------------------------------")
	fmt.Printf("Conectado con %s\n", peerAddrInfo.ID)
	fmt.Println("--------------------------------------------------")
	n.mu.Unlock()

	n.ConnectedPeers[peerAddrInfo.String()] = peerAddrInfo
	return nil
}

type discoveryveryNotifee struct {
	node *Node
}

func (n *discoveryveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	peerAddress := fmt.Sprintf("%s/p2p/%s", pi.Addrs[0].String(), pi.ID.String())
	fmt.Println("--------------------------------------------------")
	fmt.Printf("Peer encontrado: %s\n", peerAddress)
	fmt.Println("--------------------------------------------------")
}
