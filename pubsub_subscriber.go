package main

import (
	pubsub "google.golang.org/api/pubsub/v1"

	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"

	log "github.com/Sirupsen/logrus"
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
		MessagePerPull int64
		puller         Puller
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
		log.Errorf("Failed to create DefaultClient\n")
		return err
	}

	// Creates a pubsubClient
	service, err := pubsub.New(client)
	if err != nil {
		log.Errorf("Failed to create pubsub.Service with %v: %v\n", client, err)
		return err
	}

	ps.puller = &pubsubPuller{service.Projects.Subscriptions}
	return nil
}

func (ps *PubsubSubscriber) subscribe(ctx context.Context, subscription *Subscription, f func(msg *pubsub.ReceivedMessage) error) error {
	pullRequest := &pubsub.PullRequest{
		ReturnImmediately: true,
		MaxMessages:       ps.MessagePerPull,
	}
	log.WithFields(log.Fields{"subscription": subscription.Name}).Debugln("Pulling")
	res, err := ps.puller.Pull(subscription.Name, pullRequest)
	if err != nil {
		log.Errorf("Failed to pull: [%T] %v\n", err, err)
		return err
	}
	log.WithFields(log.Fields{"subscription": subscription.Name}).Debugln("Pulled successfully")
	for _, receivedMessage := range res.ReceivedMessages {
		err := ps.processProgressNotification(ctx, subscription, receivedMessage, f)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ps *PubsubSubscriber) processProgressNotification(ctx context.Context, subscription *Subscription, receivedMessage *pubsub.ReceivedMessage, f func(msg *pubsub.ReceivedMessage) error) error {
	err := f(receivedMessage)
	if err != nil {
		log.Errorf("the received request process returns error: [%T] %v", err, err)
		return err
	}
	_, err = ps.puller.Acknowledge(subscription.Name, receivedMessage.AckId)
	if err != nil {
		log.Infof("Failed to acknowledge for message: %v cause of [%T] %v", receivedMessage, err, err)
		opened, err2 := subscription.isOpened()
		if err2 != nil {
			switch err2.(type) {
			case *InvalidHttpResponse:
				e := err2.(*InvalidHttpResponse)
				if e.StatusCode == 404 {
					log.Infof("Skipping acknowledgement to pipeline: %v because the pipeline must be removed", subscription.Pipeline)
					return nil
				}
			}
			log.Errorf("Failed to check if the pipeline is opened because of [%T] %v for pipeline: %v", err2, err2, subscription.Pipeline)
			return err2
		} else if opened {
			log.Errorf("Failed to acknowledge for message: %v cause of [%T] %v", receivedMessage, err, err)
			return err
		} else {
			log.Infof("Skipping acknowledgement to pipeline: %v because the pipeline isn't opened.", subscription.Pipeline)
			return nil
		}
	}
	return nil
}
