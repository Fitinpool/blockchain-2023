package pubSub

import (
	p2p "blockchain/p2p"
	"context"
	"fmt"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
)

func NodeSub(n *p2p.Node) {

	gossipSub, err := pubsub.NewGossipSub(context.Background(), n.NetworkHost)
	if err != nil {
		panic(err)
	}

	room := "data-nodes"
	topic, err := gossipSub.Join(room)
	if err != nil {
		panic(err)
	}

	subscriber, err := topic.Subscribe()
	if err != nil {
		panic(err)
	}
	fmt.Printf("subscribed to topic: %s\n", subscriber.Topic())
	subscribe(subscriber, context.Background(), n.NetworkHost.ID())
}

func subscribe(subscriber *pubsub.Subscription, ctx context.Context, hostID peer.ID) {
	for {
		msg, err := subscriber.Next(ctx)
		if err != nil {
			panic(err)
		}
		fmt.Println(msg)

		// only consider messages delivered by other peers
		if msg.ReceivedFrom == hostID {
			continue
		}

		fmt.Printf("got message: %s, from: %s\n", string(msg.Data), msg.ReceivedFrom.String())
	}
}
