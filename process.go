package main

import (
	"encoding/base64"

	"golang.org/x/net/context"

	pubsub "google.golang.org/api/pubsub/v1"

	log "github.com/Sirupsen/logrus"
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
	PipelineID string `json:"pipeline_id"`
	Pipeline   string `json:"pipeline"`
	Name       string `json:"subscription"`
	isOpened   func() (bool, error)
}

func (p *Process) execute(ctx context.Context) error {
	targets := []*Subscription{}
	if p.agentApi != nil {
		subsFromAgent, err := p.agentApi.getSubscriptions(ctx)
		if err != nil {
			switch err.(type) {
			case *InvalidHttpResponse:
				return nil
			default:
				return err
			}
		}
		targets = append(targets, subsFromAgent...)
	}
	for _, sub := range targets {
		p.pullAndSave(ctx, sub)
	}
	return nil
}

func (p *Process) pullAndSave(ctx context.Context, subscription *Subscription) error {
	err := p.subscriber.subscribe(ctx, subscription, func(recvMsg *pubsub.ReceivedMessage) error {
		m := recvMsg.Message

		fields := log.Fields{}
		for k, v := range m.Attributes {
			fields[k] = v
		}

		raw, err := base64.StdEncoding.DecodeString(m.Data)
		if err != nil {
			fields["rawdata"] = raw
			log.WithFields(fields).Errorln("Message received but failed to decode", err)
			return err
		}

		data := string(raw)
		fields["data"] = data
		log.WithFields(fields).Debugln("Message received")

		msg := &Message{data: data}
		err = msg.load(m.Attributes)
		if err != nil {
			return err
		}
		err = msg.parse(m.PublishTime)
		if err != nil {
			return err
		}

		log.WithFields(log.Fields(msg.buildMap())).Debugln("Message parsed")

		err = p.messageStore.save(ctx, subscription.Pipeline, msg, func() error {
			return p.patterns.execute(msg)
		})
		return err
	})
	return err
}
