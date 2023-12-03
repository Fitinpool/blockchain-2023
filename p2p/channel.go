package p2p

import (
	"context"
	"fmt"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

const (
	ChannelGeneral  = "data-sharing"
	ChannelFullNode = "full-node"
)

type Channel struct {
	Topic      *pubsub.Topic
	Subscriber *pubsub.Subscription
}

func JoinChannel(n *Node, pubs *pubsub.PubSub, subscribe bool, channel string) *Channel {

	topic, err := pubs.Join(channel)
	if err != nil {
		panic(err)
	}
	if subscribe {
		subscriber, err := topic.Subscribe()
		if err != nil {
			panic(err)
		}
		fmt.Printf("subscribed to topic: %s\n", subscriber.Topic())
		return &Channel{
			Topic:      topic,
			Subscriber: subscriber,
		}
	}
	return &Channel{
		Topic:      topic,
		Subscriber: nil,
	}
}

func (ch *Channel) PublishChannelMessage(ctx context.Context, topic *pubsub.Topic, message string) {

	fmt.Printf("publishing message: %s\n", message)
	fmt.Print("to topic: " + topic.String() + "\n")
	bytes := []byte(message)
	topic.Publish(ctx, bytes)
}

func (ch *Channel) SubscribeChannelMessage(subscriber *pubsub.Subscription, fullNode *Channel, ctx context.Context, n *Node) {
	if n.fullNode {
		msg, err := subscriber.Next(ctx)
		if err != nil {
			panic(err)
		}

		if msg.ReceivedFrom == n.NetworkHost.ID() {
			return
		}

		fmt.Printf("got message: %s, from: %s\n", string(msg.Data), msg.ReceivedFrom.String())
		ch.PublishChannelMessage(ctx, fullNode.Topic, string(msg.Data))
	}

}

func (ch *Channel) CloseChannel(ctx context.Context, subscriber *pubsub.Subscription) {
	if subscriber != nil {
		subscriber.Cancel()
	}
}
