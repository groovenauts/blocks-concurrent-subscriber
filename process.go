package main

import (
	"fmt"

	"golang.org/x/net/context"

	pubsub "google.golang.org/api/pubsub/v1"
)

type AgentApi interface {
	getSubscriptions(ctx context.Context) ([]*Subscription, error)
}

type MessageSubscriber interface {
	subscribe(ctx context.Context, subscription *Subscription, f func(msg *pubsub.ReceivedMessage) error) error
}

type MessageStore interface {
	save(ctx context.Context, pipeline string, msg *Message, f func() error) error
}

type Process struct {
	agentApi     AgentApi
	subscriber   MessageSubscriber
	messageStore MessageStore
	patterns     Patterns
}

type Subscription struct {
	Pipeline string `json:"pipeline"`
	Name     string `json:"subscription"`
}

func (p *Process) execute(ctx context.Context) error {
	subscriptions, err := p.agentApi.getSubscriptions(ctx)
	if err != nil {
		fmt.Println("Process.execute() err: ", err)
		return err
	}
	for _, sub := range subscriptions {
		p.pullAndSave(ctx, sub)
	}
	return nil
}

func (p *Process) pullAndSave(ctx context.Context, subscription *Subscription) error {
	err := p.subscriber.subscribe(ctx, subscription, func(recvMsg *pubsub.ReceivedMessage) error {
		m := recvMsg.Message

		msg := &Message{data: m.Data}
		err := msg.load(m.Attributes)
		if err != nil {
			return err
		}
		err = msg.parse(m.PublishTime)
		if err != nil {
			return err
		}
		err = p.messageStore.save(ctx, subscription.Pipeline, msg, func() error {
			return p.patterns.execute(msg)
		})
		return err
	})
	return err
}
