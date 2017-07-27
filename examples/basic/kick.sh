#!/bin/bash

usage() {
  echo usage: `basename $0` '""' [url]
  echo usage: `basename $0` default [url]
  echo usage: `basename $0` sleep [seconds]
}

if [ $# -ne 2 ]; then
  echo `basename $0`: missing operand 1>&2
  usage
  exit 1
fi

set -ex

cmd=$1

case "$cmd" in
    "" | "default" )
        url=$2
        attribute="download_files=[\"${url}\"]"
        ;;
    "" | "sleep" )
        time=$2
        attribute="sleep=${time}"
        ;;
    * )
        echo "Invalid cmd given" 1>&2
        exit 1
        ;;
esac

DB_USER=${DB_USERNAME:-root}
DB_NAME=${DB_NAME:-blocks_subscriber_example1}
TOPIC=projects/${PROJECT}/topics/${PIPELINE}-job-topic

job_id=`./exec_sql ${DB_USER}:@/${DB_NAME} "INSERT INTO pipeline_jobs (pipeline, progress, created_at, updated_at) VALUES ('pipeline01', 0, NOW(), NOW())"`

msgid=`gcloud beta pubsub topics publish ${TOPIC} '' --attribute=${attribute} --attribute=app_id=${job_id} | cut -f2 -d\'`

rows=`./exec_sql ${DB_USER}:@/${DB_NAME} "UPDATE pipeline_jobs SET job_message_id = '${msgid}' WHERE id = ${job_id}"`
