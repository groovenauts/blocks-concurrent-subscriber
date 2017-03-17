package main

import (
	"fmt"
	"os"
	"time"

	"github.com/urfave/cli"

	"golang.org/x/net/context"

	log "github.com/Sirupsen/logrus"
)

func main() {
	app := cli.NewApp()
	app.Name = "blocks-concurrent-subscriber"
	app.Usage = "github.com/groovenauts/blocks-concurrent-subscriber"
	app.Version = Version

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config, c",
			Usage: "Load configuration from `FILE`",
		},
	}

	app.Action = executeCommand

	app.Run(os.Args)
}

func executeCommand(c *cli.Context) error {

	config_path := c.String("config")
	config, err := LoadProcessConfig(config_path)
	if err != nil {
		fmt.Printf("Failed to load %v cause of %v\n", config_path, err)
		os.Exit(1)
	}

	level, err := log.ParseLevel(config.LogLevel)
	if err != nil {
		fmt.Printf("Invalid log level %v cause of %v\n", config.LogLevel, err)
		os.Exit(1)
	}
	log.SetLevel(level)

	ctx := context.Background()

	pubsubSubscriber := &PubsubSubscriber{}
	err = pubsubSubscriber.setup(ctx)
	if err != nil {
		os.Exit(1)
	}

	store := &SqlStore{}
	cb, err := store.setup(ctx, "mysql", config.Datasource)
	if err != nil {
		os.Exit(1)
	}
	defer cb()

	for {
		p := &Process{
			agentApi: &DefaultAgentClient{
				httpUrl:   config.AgentRootUrl,
				httpToken: config.AgentRootToken,
			},
			subscriber:   pubsubSubscriber,
			messageStore: store,
			patterns:     config.Patterns,
		}
		p.execute(ctx)

		time.Sleep(time.Duration(config.Interval) * time.Second)
	}
}
