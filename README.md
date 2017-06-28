# blocks-concurrent-subscriber

[![Build Status](https://secure.travis-ci.org/groovenauts/blocks-concurrent-subscriber.png)](https://travis-ci.org/groovenauts/blocks-concurrent-subscriber)

## Overview

`blocks-concurrent-subscriber` subscribes the progresses of jobs with `blocks concurrent batch board` of [magellan-blocks](https://www.magellanic-clouds.com/blocks/).
When the progresses are notified, `blocks-concurrent-subscriber` updates the status of the job and inserts the logs on mysql.


## Install

Download the file from https://github.com/groovenauts/blocks-concurrent-subscriber/releases
and put it somewhere on PATH.

## Usage

```bash
$ blocks-concurrent-subscriber -c config.json
```

## Configuration file

### Fields

| Field      | Type   | Required | Description |
|------------|--------|----------|---------------|
| datasource | string | True     | String to connect your MySQL database like `root:@/database1?parseTime=true` |
| agent      | map[string]string | False | Settings to work with [blocks-concurrent-batch-agent](https://github.com/groovenauts/blocks-concurrent-batch-agent) |
| agent.root-url | string | True | The root URL to the blocks-concurrent-batch-agent to launch pipelines |
| agent.organization | string | True | The organization ID on the blocks-concurrent-batch-agent to launch pipelines. See https://github.com/groovenauts/blocks-concurrent-batch-agent#get-token-on-browser for more detail. |
| agent.token    | string | True | The access token to the blocks-concurrent-batch-agent to launch pipelines. See https://github.com/groovenauts/blocks-concurrent-batch-agent#get-token-on-browser for more detail. |
| subscriptions | []map[string]string | False | Array of subscription setting |
| subscriptions[] | map[string]string | False | subscription setting |
| subscriptions[].pipeline | string | True | The pipeline name. You can set any name you like |
| subscriptions[].subscription | string | True | The full qualified subscription name |
| message-per-pull | int | False | The number of messages per one pulling. Default: 10|
| interval         | int | False | The interval time in second to next pulling. Default: 10 |
| log-level        | string | False | The one of `debug`, `info`, `warn`, `error`, `fatal`, `panic`. Default: `info` |
| patterns         | []map[string]string | False | Array of pattern setting |
| patterns[]       | map[string]string | False | pattern setting |
| patterns[].completed | string | False | The condition to match with message `completed` attribute. Match any message `completed` attribute if blank |
| patterns[].level | string | False | The condition to match with message `level` attribute. Match any message `level` attribute if blank |
| patterns[].command | []string | The command and arguments when the conditions matches |

#### Attention!

`datasource` must have `parseTime=true` option to parse datetime column value.

### Example

```
{
  "datasource": "root:@/database1?parseTime=true",
  "agent": {
    "root-url": "https://blocks-concurrent-batch-agent-somewhere.com",
    "organization": "organization1",
    "token": "password1"
  },
  "subscriptions": [
    {
      "pipeline": "pipeline1",
      "subscription": "projects/proj-dummy-999/subscriptions/pipeline1-progress-subscription"
    },
    {
      "pipeline": "pipeline2",
      "subscription": "projects/proj-dummy-999/subscriptions/pipeline2-progress-subscription"
    }
  ],
  "message-per-pull": 100,
  "interval": 10,
  "log-level": "debug",
  "patterns": [
    {
      "completed": "true",
      "command": ["bin/rails", "r", "Model.complete('%{job_message_id}')"]
    },
    {
      "level": "fatal",
      "completed": "false",
      "command": ["bin/rails", "r", "Model.fatalError(%{job_message_id}, '%{data}')"]
    }
  ]
}
```

## Docker image

Docker image [groovenauts/blocks-concurrent-subscriber](https://hub.docker.com/r/groovenauts/blocks-concurrent-subscriber) is available.

```shell
docker pull groovenauts/blocks-concurrent-subscriber:${TAG}
docker run -v /path/to/config.json:/config.json groovenauts/blocks-concurrent-subscriber:${TAG} /blocks-concurrent-subscriber -c /config.json
```

About available `${TAG}`, see https://hub.docker.com/r/groovenauts/blocks-concurrent-subscriber/tags/


