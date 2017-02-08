package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"

	"cloud.google.com/go/pubsub"

	"golang.org/x/net/context"
)

type AgentApi interface {
	getSubscriptions(ctx context.Context) ([]*Subscription, error)
}

type MessageSubscriber interface {
	subscribe(ctx context.Context, subscription *Subscription, f func(msg *pubsub.Message) error ) error
}

type MessageStore interface {
	save(ctx context.Context, pipeline, msg_id string, progress int, publishTime time.Time, f func() error ) error
}


type Process struct {
	agentApi     AgentApi
	subscriber   MessageSubscriber
	messageStore MessageStore
	command_args []string
}

type Subscription struct {
	Pipeline string `json:"pipeline"`
	Name     string `json:"subscription"`
}

func (p *Process) execute(ctx context.Context) error {
	subscriptions, err := p.agentApi.getSubscriptions(ctx)
	if err != nil {
		fmt.Println("Process.execute() err: %v", err)
		return err
	}
	fmt.Println("Process.execute() subscriptions: %v", subscriptions)
	for _, sub := range subscriptions {
		p.pullAndSave(ctx, sub)
	}
	return nil
}

func (p *Process) pullAndSave(ctx context.Context, subscription *Subscription) error {
	err := p.subscriber.subscribe(ctx, subscription, func(m *pubsub.Message) error {
		fmt.Println("Process.pullAndSave message: %v", m)

		// https://github.com/groovenauts/magellan-gcs-proxy/blob/master/lib/magellan/gcs/proxy/progress_notification.rb#L24
		msg_id := m.Attributes["job_message_id"]
		progress, err := strconv.Atoi(m.Attributes["progress"])
		if err != nil {
			fmt.Println("Failed to convert %v to int message_id: %v cause of %v", m.Attributes["progress"], msg_id, err)
			return err
		}

		err = p.messageStore.save(ctx, subscription.Pipeline, msg_id, progress, m.PublishTime, func() error{
			// Execute command to notify
			if len(p.command_args) > 0 {
				name := p.command_args[0]
				args := p.command_args[1:] // TODO how should the parameters be passed
				cmd := exec.Command(name, args...)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				err = cmd.Run()
				return err
			}
			return nil
		})
		return err
	})
	return err
}
