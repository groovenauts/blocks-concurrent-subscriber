package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"

	"cloud.google.com/go/pubsub"

	"github.com/urfave/cli"

	"golang.org/x/net/context"
)

func main() {
	app := cli.NewApp()
	app.Name = "blocks-concurrent-subscriber"
	app.Usage = "github.com/groovenauts/blocks-concurrent-subscriber"
	app.Version = Version

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "project",
			Usage:  "GCS Project ID",
			EnvVar: "GCP_PROJECT,PROJECT",
		},
		cli.StringFlag{
			Name:   "datasource",
			Usage:  "Data source name to your database",
			EnvVar: "DATASOURCE",
		},
		cli.StringFlag{
			Name:   "agent-root-url, A",
			Usage:  "URL to your blocks-concurrent-batch-agent root",
			EnvVar: "AGENT_URL",
		},
		cli.StringFlag{
			Name:   "agent-token, T",
			Usage:  "Authorization token for blocks-concurrent-batch-agent",
			EnvVar: "AGENT_TOKEN",
		},
		cli.UintFlag{
			Name:  "interval",
			Value: 10,
			Usage: "Interval to pull",
		},
	}

	app.Action = executeCommand

	app.Run(os.Args)
}

func executeCommand(c *cli.Context) error {

	proj := c.String("project")
	if proj == "" {
		cli.ShowAppHelp(c)
		os.Exit(1)
	}
	interval := c.Uint("interval")

	ctx := context.Background()

	pubsubClient := &PubsubClient{}
	err := pubsubClient.setup(ctx, proj)
	if err != nil {
		os.Exit(1)
	}

	store := &SqlStore{}
	cb, err := store.setup(ctx, "mysql", c.String("datasource"))
	if err != nil {
		os.Exit(1)
	}
	defer cb()

	for {
		p := &Process{
			agentApi:     &DefaultAgentClient{
				httpUrl:   c.String("agent-url"),
				httpToken: c.String("agent-token"),
			},
			subscriber:   pubsubClient,
			messageStore: store,
			command_args: c.Args(),
		}
		p.execute(ctx)

		time.Sleep(time.Duration(interval) * time.Second)
	}

	return nil
}

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
		return err
	}
	for _, sub := range subscriptions {
		p.pullAndSave(ctx, sub)
	}
	return nil
}

func (p *Process) pullAndSave(ctx context.Context, subscription *Subscription) error {
	err := p.subscriber.subscribe(ctx, subscription, func(m *pubsub.Message) error {
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
