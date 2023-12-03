package pubSub

import (
	p2p "blockchain/p2p"
	"context"
	"fmt"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

func NodePub(n *p2p.Node) {
	// create a new PubSub service using the GossipSub router
	gossipSub, err := pubsub.NewGossipSub(context.Background(), n.NetworkHost)
	if err != nil {
		panic(err)
	}
	room := "data-nodes"
	topic, err := gossipSub.Join(room)
	if err != nil {
		panic(err)
	}
	message := "Hello World!"

	publish(context.Background(), topic, message)
}

func publish(ctx context.Context, topic *pubsub.Topic, message string) {
	fmt.Printf("publishing message: %s\n", message)
	fmt.Print("to topic: " + topic.String() + "\n")
	// publish message to topic
	bytes := []byte(message)
	topic.Publish(ctx, bytes)
}
