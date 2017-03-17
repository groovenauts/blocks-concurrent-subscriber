package main

import (
	"database/sql"
	"os"
	"os/exec"
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"golang.org/x/net/context"
)

const (
	TEST_DATASOURCE = "root:@/blocks_subscriber_test"
)

func assertCount(t *testing.T, db *sql.DB, expected int, sql string, args ...interface{}) bool {
	var cnt int
	err := db.QueryRow(sql, args...).Scan(&cnt)
	assert.NoError(t, err)
	return assert.Equal(t, expected, cnt)
}

type PipelineJob struct {
	id             int
	pipeline       string
	job_message_id string
	progress       int
}

const PIPELINE_JOBS_QUERY_BASE = "SELECT id, pipeline, job_message_id, progress FROM pipeline_jobs "

func queryPipelineJob(db *sql.DB, where string, args ...interface{}) (*PipelineJob, error) {
	r := PipelineJob{}
	sql := PIPELINE_JOBS_QUERY_BASE + where
	err := db.QueryRow(sql, args...).Scan(&r.id, &r.pipeline, &r.job_message_id, &r.progress)
	return &r, err
}

func TestProgressStoreSave(t *testing.T) {
	cmd := exec.Command("make", "testfixtures")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		log.Fatalf("Failed to load fixtures`\n")
		return
	}

	ctx := context.Background()

	store := &SqlStore{}
	cb, err := store.setup(ctx, "mysql", TEST_DATASOURCE)
	if err != nil {
		log.Fatalf("Failed to connect DB: %q. Please run `make testsetup`\n", TEST_DATASOURCE)
		return
	}
	defer cb()

	db := store.db

	extraCalled := false
	extra := func() error {
		extraCalled = true
		return nil
	}
	msg := &Message{
		msg_id:      "jm01",
		progress:    2,
		publishTime: time.Now(),
		completed:   "false",
		level:       "info",
		data:        "",
	}

	pl, err := queryPipelineJob(db, "WHERE pipeline='pipeline01' AND job_message_id=?", msg.msg_id)
	assert.NoError(t, err)
	assert.Equal(t, 1, pl.progress)

	store.save(ctx, "pipeline01", msg, extra)

	pl, err = queryPipelineJob(db, "WHERE pipeline='pipeline01' AND job_message_id=?", msg.msg_id)
	assert.NoError(t, err)
	assert.Equal(t, 2, pl.progress)

	msg = &Message{
		msg_id:      "jm04",
		progress:    2,
		publishTime: time.Now(),
		completed:   "false",
		level:       "info",
		data:        "",
	}

	pl, err = queryPipelineJob(db, "WHERE pipeline='pipeline01' AND job_message_id=?", msg.msg_id)
	assert.NoError(t, err)
	assert.Equal(t, 4, pl.progress)

	store.save(ctx, "pipeline01", msg, extra)

	pl, err = queryPipelineJob(db, "WHERE pipeline='pipeline01' AND job_message_id=?", msg.msg_id)
	assert.NoError(t, err)
	assert.Equal(t, 4, pl.progress)
}
