package main

import (
	"fmt"
	"testing"
	"time"

	"cloud.google.com/go/pubsub"

	"github.com/stretchr/testify/assert"

	"golang.org/x/net/context"
)

type DummyAgentApi struct {}
func (dh *DummyAgentApi) getSubscriptions(ctx context.Context) ([]*Subscription, error) {
	return []*Subscription{
		&Subscription{Pipeline: "pipeline01", Name: "pipeline01-progress-subscription"},
	}, nil
}

type DummySubscriber struct{}
func (ds *DummySubscriber) subscribe(ctx context.Context, subscription *Subscription, f func(msg *pubsub.Message) error ) error {
	msg := &pubsub.Message{
		Attributes: map[string]string{
			"job_message_id": "0123456789",
			"progress": "14",
		},
		PublishTime: time.Now(),
	}
	return f(msg)
}

type DummyStore struct{}
func (ds *DummyStore) save(ctx context.Context, pipeline, msg_id string, progress int, publishTime time.Time, f func() error ) error {
	if "pipeline01" != pipeline {
		return fmt.Errorf("pipeline should be pipeline01 but was %v", pipeline)
	}
	if "0123456789" != msg_id {
		return fmt.Errorf("msg_id should be 0123456789 but was %v", msg_id)
	}
	if 14 != progress {
		return fmt.Errorf("progress should be 14 but was %v", progress)
	}
	return nil
}

func TestExecute(t *testing.T) {
	pr := &Process{
		agentApi: &DummyAgentApi{},
		subscriber: &DummySubscriber{},
		messageStore: &DummyStore{},
		command_args: []string{},
	}

	ctx := context.Background()
	err := pr.execute(ctx)
	assert.NoError(t, err)
}
