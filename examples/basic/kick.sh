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

cmd="gcloud beta pubsub topics publish ${PIPELINE}-job-topic '' --attribute=${attribute}"
echo $cmd
msgid=`${cmd} | cut -f2 -d\'`

cmd="mysql -u ${DB_USER} ${DB_NAME} -e \"INSERT INTO pipeline_jobs (pipeline, job_message_id, progress, created_at, updated_at) VALUES ('${PIPELINE}', '${msgid}', 0, NOW(), NOW());\""
echo $cmd
eval ${cmd}
