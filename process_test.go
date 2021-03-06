package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"golang.org/x/net/context"

	pubsub "google.golang.org/api/pubsub/v1"
)

type DummyAgentApi struct{}

func (dh *DummyAgentApi) getSubscriptions(ctx context.Context) ([]*Subscription, error) {
	return []*Subscription{
		&Subscription{Pipeline: "pipeline01", Name: "pipeline01-progress-subscription"},
	}, nil
}

type DummySubscriber struct{}

func (ds *DummySubscriber) subscribe(ctx context.Context, subscription *Subscription, f func(msg *pubsub.ReceivedMessage) error) error {
	msg := &pubsub.ReceivedMessage{
		Message: &pubsub.PubsubMessage{
			Attributes: map[string]string{
				"app_id":   "0123456789",
				"progress": "14",
			},
			PublishTime: time.Now().Format(time.RFC3339),
		},
	}
	return f(msg)
}

type DummyStore struct{}

func (ds *DummyStore) save(ctx context.Context, msg *Message, f func() error) error {
	// if "pipeline01" != pipeline {
	// 	return fmt.Errorf("pipeline should be pipeline01 but was %v", pipeline)
	// }
	if "0123456789" != msg.attributes["app_id"] {
		return fmt.Errorf("app_id should be 0123456789 but was %v", msg.attributes["app_id"])
	}
	if 14 != msg.progress {
		return fmt.Errorf("progress should be 14 but was %v", msg.progress)
	}
	return nil
}

func TestExecute(t *testing.T) {
	pr := &Process{
		agentApi:     &DummyAgentApi{},
		subscriber:   &DummySubscriber{},
		messageStore: &DummyStore{},
		patterns:     Patterns{},
	}

	ctx := context.Background()
	err := pr.execute(ctx)
	assert.NoError(t, err)
}
