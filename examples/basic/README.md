# blocks-concurrent-subscriber simple example

## Overview

1. Setup [blocks-batch-agent](https://github.com/groovenauts/blocks-concurrent-batch-agent)
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


## Setup [blocks-batch-agent](https://github.com/groovenauts/blocks-concurrent-batch-agent)

1. Deploy [blocks-batch-agent](https://github.com/groovenauts/blocks-concurrent-batch-agent) to appengine by following https://github.com/groovenauts/blocks-concurrent-batch-agent#deploy-to-appengine

## Create a new pipeline

### Create a new `pipeline.json`

```json
{
  "name":"[Your Pipeline name]",
  "project_id":"[Your GCP Project ID]",
  "zone":"us-central1-f",
  "source_image":"https://www.googleapis.com/compute/v1/projects/google-containers/global/images/gci-stable-55-8872-76-0",
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
$ export TOKEN="[the token you got before]"
$ export AEHOST="[the host name you deployed]"
$ curl -v -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -X POST http://$AEHOST/pipelines --data @pipeline.json
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

### Create a new config.json

```json
{
  "datasource": "root:@/blocks_subscriber_example1",
  "agent-root-url": "[the token you got before]",
  "agent-root-token": "[the host name you deployed]",
  "interval": 10,
  "patterns": [
    {
      "completed": "true",
      "command": ["examples/basic/recv.sh", "COMPLETED", "%{job_message_id}"]
    },
    {
      "level": "fatal",
      "completed": "false",
      "command": ["examples/basic/recv.sh", "FATAL", "job_message_id: %{job_message_id}, msg: %{data}"]
    },
    {
      "level": "error",
      "command": ["examples/basic/recv.sh", "ERROR", "job_message_id: %{job_message_id}, msg: %{data}"]
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
$ examples/basic/kick.sh "" gs://yourbucket/somewhere
```
