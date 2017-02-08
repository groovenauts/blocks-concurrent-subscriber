package main

import (
	"os"
	"time"

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
