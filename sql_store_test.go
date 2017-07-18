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
	TEST_DATASOURCE = "root:@/blocks_subscriber_test?parseTime=true"
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
	created_at     *time.Time
	updated_at     *time.Time
}

const PIPELINE_JOBS_QUERY_BASE = "SELECT id, pipeline, job_message_id, progress, created_at, updated_at FROM pipeline_jobs "

func queryPipelineJob(db *sql.DB, where string, args ...interface{}) (*PipelineJob, error) {
	r := PipelineJob{}
	sql := PIPELINE_JOBS_QUERY_BASE + where
	err := db.QueryRow(sql, args...).Scan(&r.id, &r.pipeline, &r.job_message_id, &r.progress, &r.created_at, &r.updated_at)
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

	store := &SqlStore{
		updateTemplate: &SqlTemplate{
			Body:       "UPDATE pipeline_jobs SET progress = ?, updated_at = ? WHERE job_message_id = ? AND progress < ?",
			Parameters: []string{"progress", "now", "job_message_id", "progress"},
		},
	}
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
		progress:    2,
		publishTime: time.Now(),
		completed:   "false",
		level:       "info",
		data:        "",
		attributes: map[string]string{
			"job_message_id": "jm01",
		},
	}

	pl, err := queryPipelineJob(db, "WHERE pipeline='pipeline01' AND job_message_id=?", msg.attributes["job_message_id"])
	assert.NoError(t, err)
	assert.Equal(t, 1, pl.progress)
	assert.Equal(t, pl.created_at.UnixNano(), pl.updated_at.UnixNano())

	time.Sleep(1 * time.Second) // To make difference between updated_at and created_at
	store.save(ctx, msg, extra)

	pl, err = queryPipelineJob(db, "WHERE pipeline='pipeline01' AND job_message_id=?", msg.attributes["job_message_id"])
	assert.NoError(t, err)
	assert.Equal(t, 2, pl.progress)
	assert.NotEqual(t, pl.created_at.UnixNano(), pl.updated_at.UnixNano())

	msg = &Message{
		progress:    2,
		publishTime: time.Now(),
		completed:   "false",
		level:       "info",
		data:        "",
		attributes: map[string]string{
			"job_message_id": "jm04",
		},
	}

	pl, err = queryPipelineJob(db, "WHERE pipeline='pipeline01' AND job_message_id=?", msg.attributes["job_message_id"])
	assert.NoError(t, err)
	assert.Equal(t, 4, pl.progress)

	store.save(ctx, msg, extra)

	pl, err = queryPipelineJob(db, "WHERE pipeline='pipeline01' AND job_message_id=?", msg.attributes["job_message_id"])
	assert.NoError(t, err)
	assert.Equal(t, 4, pl.progress)
}
