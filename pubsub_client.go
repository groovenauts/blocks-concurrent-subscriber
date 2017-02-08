package main

import (
	"fmt"

	"cloud.google.com/go/pubsub"

	"golang.org/x/net/context"
	"google.golang.org/api/iterator"
)

type PubsubClient struct{
	impl *pubsub.Client
}

func (pc *PubsubClient) setup(ctx context.Context, proj string) error {
	client, err := pubsub.NewClient(ctx, proj)
	if err != nil {
		fmt.Println("Failed to get new pubsubClient for ", proj, " cause of ", err)
		return err
	}
	pc.impl = client
	return nil
}

func (pc *PubsubClient) subscribe(ctx context.Context, subscription *Subscription, f func(msg *pubsub.Message) error ) error {

	sub := pc.impl.Subscription(subscription.Name)
	it, err := sub.Pull(ctx)
	if err != nil {
		fmt.Println("Failed to pull from ", subscription, " cause of ", err)
		return err
	}
	// Ensure that the iterator is closed down cleanly.
	defer it.Stop()

	for {
		m, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			fmt.Println("Failed to get pulled message from ", subscription, " cause of ", err)
			return err
		}

		err = f(m)

		m.Done(err == nil)
	}
	return nil
}
