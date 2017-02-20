package main

import (
	"log"

	pubsub "google.golang.org/api/pubsub/v1"

	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
)

type (
	Puller interface {
		Pull(subscription string, pullrequest *pubsub.PullRequest) (*pubsub.PullResponse, error)
		Acknowledge(subscription, ackId string) (*pubsub.Empty, error)
	}

	pubsubPuller struct {
		subscriptionsService *pubsub.ProjectsSubscriptionsService
	}

	PubsubSubscriber struct {
		puller Puller
	}
)

func (pp *pubsubPuller) Pull(subscription string, pullrequest *pubsub.PullRequest) (*pubsub.PullResponse, error) {
	return pp.subscriptionsService.Pull(subscription, pullrequest).Do()
}

func (pp *pubsubPuller) Acknowledge(subscription, ackId string) (*pubsub.Empty, error) {
	ackRequest := &pubsub.AcknowledgeRequest{
		AckIds: []string{ackId},
	}
	return pp.subscriptionsService.Acknowledge(subscription, ackRequest).Do()
}

func (ps *PubsubSubscriber) setup(ctx context.Context) error {
	// https://github.com/google/google-api-go-client#application-default-credentials-example
	client, err := google.DefaultClient(ctx, pubsub.PubsubScope)
	if err != nil {
		log.Printf("Failed to create DefaultClient\n")
		return err
	}

	// Creates a pubsubClient
	service, err := pubsub.New(client)
	if err != nil {
		log.Printf("Failed to create pubsub.Service with %v: %v\n", client, err)
		return err
	}

	ps.puller = &pubsubPuller{service.Projects.Subscriptions}
	return nil
}

func (ps *PubsubSubscriber) subscribe(ctx context.Context, subscription *Subscription, f func(msg *pubsub.ReceivedMessage) error) error {
	pullRequest := &pubsub.PullRequest{
		ReturnImmediately: false,
		// MaxMessages: 1,
	}
	res, err := ps.puller.Pull(subscription.Name, pullRequest)
	if err != nil {
		log.Printf("Failed to pull: %v\n", err)
		return err
	}
	for _, receivedMessage := range res.ReceivedMessages {
		err := f(receivedMessage)
		if err == nil {
			if _, err = ps.puller.Acknowledge(subscription.Name, receivedMessage.AckId); err != nil {
				log.Fatalf("Failed to acknowledge for message: %v cause of %v", receivedMessage, err)
			}
		} else {
			log.Printf("the received request process returns error: %v", err)
		}
	}
	return nil
}
