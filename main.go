package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"

	"cloud.google.com/go/pubsub"

	_ "github.com/go-sql-driver/mysql"
	"github.com/urfave/cli"

	"golang.org/x/net/context"
	"google.golang.org/api/iterator"
)

func main() {
	app := cli.NewApp()
	app.Name = "pubsub-devsub"
	app.Usage = "github.com/akm/pubsub-devsub"
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

	httpClient := new(http.Client)

	ctx := context.Background()
	pubsubClient, err := pubsub.NewClient(ctx, proj)
	if err != nil {
		fmt.Println("Failed to get new pubsubClient for ", proj, " cause of ", err)
		os.Exit(1)
	}

	db, err := sql.Open("mysql", c.String("datasource"))
	if err != nil {
		fmt.Println("Failed to get database connection for ", c.String("datasource"), " cause of ", err)
		os.Exit(1)
	}
	defer db.Close()

	for {
		p := &Process{
			httpClient:   httpClient,
			httpUrl:      c.String("agent-url"),
			httpToken:    c.String("agent-token"),
			pubsubClient: pubsubClient,
			db:           db,
			command_args: c.Args(),
		}
		p.execute(ctx)

		time.Sleep(time.Duration(interval) * time.Second)
	}

	return nil
}

type Process struct {
	httpClient   *http.Client
	httpUrl      string
	httpToken    string
	pubsubClient *pubsub.Client
	db           *sql.DB
	command_args []string
}

type Subscription struct {
	Pipeline string `json:"pipeline"`
	Name     string `json:"subscription"`
}

func (p *Process) execute(ctx context.Context) error {
	subscriptions, err := p.getSubscriptions(ctx)
	if err != nil {
		return err
	}
	for _, sub := range subscriptions {
		p.pullAndSave(ctx, sub)
	}
	return nil
}

func (p *Process) getSubscriptions(ctx context.Context) ([]*Subscription, error) {
	req, err := http.NewRequest("GET", p.httpUrl+"/pipelines/subscriptions", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+p.httpToken)
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	byteArray, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var subscriptions []Subscription
	err = json.Unmarshal(byteArray, &subscriptions)
	if err != nil {
		return nil, err
	}
	result := []*Subscription{}
	for _, subscription := range subscriptions {
		result = append(result, &subscription)
	}
	return result, nil
}

const (
	SQL_UPDATE_JOBS = "UPDATE pipeline_jobs SET status = ? WHERE message_id = ? AND status < ?"
	SQL_INSERT_LOGS = "INSERT INTO pipeline_job_logs (pipeline, message_id, status, publish_time) VALUES (?, ?, ?, ?)"
)

func (p *Process) pullAndSave(ctx context.Context, subscription *Subscription) error {
	sub := p.pubsubClient.Subscription(subscription.Name)
	it, err := sub.Pull(ctx)
	if err != nil {
		fmt.Println("Failed to pull from ", subscription, " cause of ", err)
		return err
	}
	// Ensure that the iterator is closed down cleanly.
	defer it.Stop()

	for {
		m, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			fmt.Println("Failed to get pulled message from ", subscription, " cause of ", err)
			return err
		}

		// https://github.com/groovenauts/magellan-gcs-proxy/blob/master/lib/magellan/gcs/proxy/progress_notification.rb#L24
		msg_id := m.Attributes["job_message_id"]
		progress, err := strconv.Atoi(m.Attributes["progress"])
		if err != nil {
			fmt.Println("Failed to convert %v to int message_id: %v cause of %v", m.Attributes["progress"], msg_id, err)
			return err
		}

		err = p.transaction(func(tx *sql.Tx) error {
			_, err = tx.Exec(SQL_UPDATE_JOBS, progress, msg_id, progress)
			if err != nil {
				fmt.Println("Failed to update pipeline_jobs message_id: %v to status: %v cause of %v", msg_id, progress, err)
				return err
			}

			_, err = tx.Exec(SQL_INSERT_LOGS, subscription.Pipeline, msg_id, progress, m.PublishTime)
			if err != nil {
				fmt.Println("Failed to insert pipeline_job_logs pipeline: %v, message_id: %v, status: %v cause of %v", subscription.Pipeline, msg_id, progress, err)
				return err
			}

			// Execute command to notify
			if len(p.command_args) > 0 {
				name := p.command_args[0]
				args := p.command_args[1:] // TODO how should the parameters be passed
				cmd := exec.Command(name, args...)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				err = cmd.Run()
				if err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			fmt.Println("Failed to begin a transaction message_id: %v to status: %v cause of %v", msg_id, progress, err)
			return err
		}

		m.Done(true)
	}
	return nil
}

func (p *Process) transaction(impl func(tx *sql.Tx) error) error {
	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err := recover(); err != nil {
			tx.Rollback()
		}
	}()
	err = impl(tx)
	if err == nil {
		tx.Commit()
		return nil
	} else {
		tx.Rollback()
		return err
	}
}
