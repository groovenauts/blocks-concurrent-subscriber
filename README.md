# blocks-concurrent-batch-subscriber

[![Build Status](https://secure.travis-ci.org/groovenauts/blocks-concurrent-subscriber.png)](https://travis-ci.org/groovenauts/blocks-concurrent-subscriber)

## Overview

`blocks-concurrent-subscriber` subscribes the progresses of jobs with `blocks concurrent batch board` of [magellan-blocks](https://www.magellanic-clouds.com/blocks/).
When the progresses are notified, `blocks-concurrent-subscriber` updates the status of the job and inserts the logs on mysql.

`blocks-concurrent-subscriber` access to `blocks-concurrent-batch-agent` to get subscriptions to pull messages.

## Install

Download the file from https://github.com/groovenauts/blocks-concurrent-subscriber/releases
and put it somewhere on PATH.

## Usage

```bash
$ blocks-concurrent-subscriber \
  --datasource [Datasource string to Your MySQL] \
  --agent-root-url [URL to your blocks-concurrent-batch-agent] \
  --agent-token [Token of your blocks-concurrent-batch-agent]
```

### `--datasource`

`datasource` must be a string to connect your MySQL database like this:

```
"user:password@/dbname"
```

See https://github.com/go-sql-driver/mysql#usage for more detail.

### `--agent-root-url`

This is an URL to the blocks-concurrent-batch-agent you deploy to GAE.

For example:

```
https://concurrent-batch-agent-dot-your-gcp-project.appspot.com
```

### `--agent-token`

After you deploy blocks-concurrent-batch-agent, you can get tokens for authorization.
See https://github.com/groovenauts/blocks-concurrent-batch-agent#get-token-on-browser for more detail.
