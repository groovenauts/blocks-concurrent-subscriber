package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"

	"golang.org/x/net/context"

	pubsub "google.golang.org/api/pubsub/v1"
)

type AgentApi interface {
	getSubscriptions(ctx context.Context) ([]*Subscription, error)
}

type MessageSubscriber interface {
	subscribe(ctx context.Context, subscription *Subscription, f func(msg *pubsub.ReceivedMessage) error ) error
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
	for _, sub := range subscriptions {
		p.pullAndSave(ctx, sub)
	}
	return nil
}

func (p *Process) pullAndSave(ctx context.Context, subscription *Subscription) error {
	err := p.subscriber.subscribe(ctx, subscription, func(recvMsg *pubsub.ReceivedMessage) error {
		m := recvMsg.Message
		// https://github.com/groovenauts/magellan-gcs-proxy/blob/master/lib/magellan/gcs/proxy/progress_notification.rb#L24
		msg_id := m.Attributes["job_message_id"]
		progress, err := strconv.Atoi(m.Attributes["progress"])
		if err != nil {
			fmt.Printf("Failed to convert %v to int message_id: %v cause of %v", m.Attributes["progress"], msg_id, err)
			return err
		}
		// https://cloud.google.com/pubsub/docs/reference/rest/v1/PubsubMessage
		// A timestamp in RFC3339 UTC "Zulu" format, accurate to nanoseconds. Example: "2014-10-02T15:01:23.045123456Z".
		time, err := time.Parse(time.RFC3339, m.PublishTime)

		err = p.messageStore.save(ctx, subscription.Pipeline, msg_id, progress, time, func() error{
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
