package main

import (
	"fmt"
	"net/http"
	"testing"

	pubsub "google.golang.org/api/pubsub/v1"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

type DummyPuller struct {
	Error   error
	PullRes *pubsub.PullResponse
	AckRes  *pubsub.Empty
}

func (dp *DummyPuller) Pull(subscription string, pullrequest *pubsub.PullRequest) (*pubsub.PullResponse, error) {
	if dp.Error != nil {
		return nil, dp.Error
	}
	res := dp.PullRes
	if res == nil {
		res = &pubsub.PullResponse{}
	}
	return dp.PullRes, nil
}

func (dp *DummyPuller) Acknowledge(subscription, ackId string) (*pubsub.Empty, error) {
	if dp.Error != nil {
		return nil, dp.Error
	}
	res := dp.AckRes
	if res == nil {
		res = &pubsub.Empty{}
	}
	return dp.AckRes, nil
}

func TestProcessProgressNotification(t *testing.T) {
	ctx := context.Background()

	dp := &DummyPuller{}

	ps := PubsubSubscriber{
		MessagePerPull: 1,
		puller:         dp,
	}

	subscription := &Subscription{
		PipelineID: "pipeline0123456789",
		Pipeline:   "dummy-pipeline01",
		Name:       "dummy-pipeline01-progress-subscription",
	}

	recvMsg := &pubsub.ReceivedMessage{
		AckId: "dummy-ack-id",
	}

	returnNil := func(msg *pubsub.ReceivedMessage) error {
		return nil
	}

	returnError := func(msg *pubsub.ReceivedMessage) error {
		return fmt.Errorf("Dummy Error")
	}

	// Normal pattern
	err := ps.processProgressNotification(ctx, subscription, recvMsg, returnNil)
	assert.NoError(t, err)

	//  f returns an error
	err = ps.processProgressNotification(ctx, subscription, recvMsg, returnError)
	assert.Error(t, err)
	assert.Equal(t, "Dummy Error", err.Error())

	// Ack error and fail to get pipeline status
	dp.Error = fmt.Errorf("ack-error")
	subscription.isOpened = func() (bool, error) {
		return false, fmt.Errorf("pipeline status error")
	}
	err = ps.processProgressNotification(ctx, subscription, recvMsg, returnNil)
	assert.Error(t, err)

	// Ack error and fail to get pipeline status
	dp.Error = fmt.Errorf("ack-error")
	subscription.isOpened = func() (bool, error) {
		return true, nil // opened
	}
	err = ps.processProgressNotification(ctx, subscription, recvMsg, returnNil)
	assert.Error(t, err)

	// Ack error and fail to get pipeline status
	dp.Error = fmt.Errorf("ack-error")
	subscription.isOpened = func() (bool, error) {
		return false, nil // closing
	}
	err = ps.processProgressNotification(ctx, subscription, recvMsg, returnNil)
	assert.NoError(t, err)

	// Ack error and fail to get pipeline status because the pipeline is already deleted
	dp.Error = fmt.Errorf("ack-error")
	subscription.isOpened = func() (bool, error) {
		return false, &InvalidHttpResponse{StatusCode: http.StatusNotFound, Msg: "Already removed"}
	}
	err = ps.processProgressNotification(ctx, subscription, recvMsg, returnNil)
	assert.NoError(t, err)
}
