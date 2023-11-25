package p2p

import (
	"context"
	"fmt"

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
	Port           string
}

type NodeConfig struct {
	IP   string
	Port string
}

func NewNode(config *NodeConfig) (*Node, error) {
	listenAddress := fmt.Sprintf("/ip4/%s/tcp/%s", config.IP, config.Port)
	address := libp2p.ListenAddrStrings(listenAddress)
	host, err := libp2p.New(address)
	if err != nil {
		return nil, errors.Wrap(err, "network: NewNode libp2p.New error")
	}

	peerInfo := peer.AddrInfo{
		ID:    host.ID(),
		Addrs: host.Addrs(),
	}

	addrs, err := peer.AddrInfoToP2pAddrs(&peerInfo)
	if err != nil {
		return nil, errors.Wrap(err, "network: NewNode peer.AddrInfoToP2pAddrs error")
	}

	fmt.Printf("Node address: %s\n", addrs[0])

	mdnsService := mdns.NewMdnsService(
		host,
		namespace,
		&discoveryveryNotifee{},
	)

	mdnsService.Start()

	return &Node{
		NetworkHost:    host,
		MdnsService:    mdnsService,
		ConnectedPeers: make(map[string]*peer.AddrInfo),
		Port:           config.Port,
	}, nil
}

func (n *Node) ConnectWithPeers(peersAddresses []string) error {

	for _, peerAddr := range peersAddresses {
		peerMultiAddr, err := multiaddr.NewMultiaddr(peerAddr)
		if err != nil {
			err = errors.Wrap(err, "network: Node.ConnectWithPeers multiaddr.NewMultiaddr error")
			fmt.Printf("%s with peer %s", err.Error(), peerAddr)
			return err
		}

		peerAddrInfo, err := peer.AddrInfoFromP2pAddr(peerMultiAddr)
		if err != nil {
			err = errors.Wrap(err, "network: Node.ConnectWithPeers peer.AddrInfoFromP2pAddr error")
			fmt.Printf("%s with peer %s", err.Error(), peerAddr)
			return err
		}

		err = n.NetworkHost.Connect(context.Background(), *peerAddrInfo)
		if err != nil {
			err = errors.Wrap(err, "network: Node.ConnectWithPeers n.NetworkHost.Connect error")
			fmt.Printf("%s with peer %s", err.Error(), peerAddr)
			return err
		}
		fmt.Printf("Connected to peer %s\n", peerAddrInfo.String())

		n.ConnectedPeers[peerAddrInfo.String()] = peerAddrInfo
	}

	return nil
}

type discoveryveryNotifee struct {
}

func (n *discoveryveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	fmt.Printf("Found peer: %s\n", pi.ID)
}
