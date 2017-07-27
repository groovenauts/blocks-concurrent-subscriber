# blocks-concurrent-subscriber simple example

## Overview

1. Setup [blocks-concurrent-batch-agent](https://github.com/groovenauts/blocks-concurrent-batch-agent)
    1. deploy blocks-concurrent-batch-agent
    1. Create your organization on your blocks-concurrent-batch-agent
1. Create a new pipeline
1. Setup MySQL database
1. Launch `blocks-concurrent-subscriber`
1. Run
    1. Publish a job message
    1. Worker executes the job
    1. `blocks-concurrent-subscriber` receives progress messages
        1. Insert the data into `pipeline_job_logs` and update `pipeline_jobs`
        1. Run commands if the progress message matches patterns

## Prerequisite

- [gcloud command](https://cloud.google.com/sdk/gcloud/)
- MySQL

### Clone repository for this

```
$ git clone https://github.com/groovenauts/blocks-concurrent-subscriber.git
$ cd blocks-concurrent-subscriber
```


## Setup [blocks-concurrent-batch-agent](https://github.com/groovenauts/blocks-concurrent-batch-agent)

### Deploy [blocks-concurrent-batch-agent](https://github.com/groovenauts/blocks-concurrent-batch-agent)

1. Follow https://github.com/groovenauts/blocks-concurrent-batch-agent/blob/master/README.md#deploy-to-appengine

### Create your organization on your blocks-concurrent-batch-agent

1. Follow https://github.com/groovenauts/blocks-concurrent-batch-agent/blob/master/README.md#get-token-on-browser
    - Replace `http://localhost:8080` to your URL on GAE

## Create a new pipeline

### Create a new `pipeline.json`

```json
{
  "name":"[Your Pipeline name]",
  "project_id":"[Your GCP Project ID]",
  "zone":"us-central1-f",
  "boot_disk": {
    "source_image":"https://www.googleapis.com/compute/v1/projects/cos-cloud/global/images/family/cos-stable"
  },
  "machine_type":"f1-micro",
  "target_size":1,
  "container_size":1,
  "container_name":"groovenauts/concurrent_batch_command_options_example:0.4.0",
  "command":""
}
```

Don't forget to replace `[Your Pipeline name]` and `[Your GCP Project ID]`

### Send a request to create a new pipeline

```
$ export ORG_ID="[the organization ID you got before]"
$ export TOKEN="[the token you got before]"
$ export AEHOST="[the host name you deployed]"
$ curl -v -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -X POST http://$AEHOST/orgs/$ORG_ID/pipelines --data @pipeline.json
```

## Setup MySQL database

### Launch MySQL server

It depends on your environment.

### Setup Database

```
$ mysql -u root -e "CREATE DATABASE IF NOT EXISTS blocks_subscriber_example1 DEFAULT CHARACTER SET utf8 COLLATE utf8_general_ci;"
$ mysql -u root blocks_subscriber_example1 < migrations/up.sql
```

## Launch `blocks-concurrent-subscriber`

### Change directory

```bash
$ examples/basic
```

### Compile exec_sql.go

```bash
$ go build exec_sql.go
```

Check if `exec_sql` exists


### Create a new config.json

```json
{
  "datasource": "root:@/blocks_subscriber_example1?parseTime=true",
  "sql": {
    "update-jobs": "UPDATE pipeline_jobs SET progress = $progress, updated_at = $now WHERE id = $app_id AND progress < $progress",
    "insert-logs": "INSERT INTO pipeline_job_logs (pipeline, publish_time, progress, completed, log_level, log_message) VALUES ($pipeline, $publishTime, $progress, $completed, $level, $data)"
  },
  "agent": {
    "root-url": "[the token you got before]",
    "organization": "[the organization ID on blocks-concurrent-batch-agent]",
    "token": "[the host name you deployed]"
  },
  "interval": 10,
  "patterns": [
    {
      "completed": "true",
      "command": ["./recv.sh", "COMPLETED", "%{app_id}"]
    },
    {
      "level": "fatal",
      "completed": "false",
      "command": ["./recv.sh", "FATAL", "app_id: %{app_id}, msg: %{data}"]
    },
    {
      "level": "error",
      "command": ["./recv.sh", "ERROR", "app_id: %{app_id}, msg: %{data}"]
    }
  ]
}
```

### Launch `blocks-concurrent-subscriber`

```
$ go build
$ ./blocks-concurrent-subscriber --config config.json
```

### Upload files somewhere

```
$ gsutil cp path/to/some/file gs://yourbucket/somewhere
```


### Publish a job message

```
$ export PIPELINE=[[Your Pipeline name]]
$ ./kick.sh "" gs://yourbucket/somewhere
```
