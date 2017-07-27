package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSqlTemplateSetupUpdate1(t *testing.T) {
	src := "UPDATE pipeline_jobs SET progress = $progress, updated_at = $now WHERE id = $app_id AND progress < $progress"
	tmpl := &SqlTemplate{Source: src}
	tmpl.Setup()

	assert.Equal(t, "UPDATE pipeline_jobs SET progress = ?, updated_at = ? WHERE id = ? AND progress < ?", tmpl.Body)
	assert.Equal(t, []string{
		"progress",
		"now",
		"app_id",
		"progress",
	}, tmpl.Parameters)
	assert.Equal(t, src, tmpl.Source)
}

func TestSqlTemplateSetupInsert1(t *testing.T) {
	src := "INSERT INTO pipeline_job_logs" +
		" (pipeline, publish_time, progress, completed, log_level, log_message)" +
		" VALUES ($pipeline, $publish_time, $progress, $completed, $log_level, $log_message)"
	tmpl := &SqlTemplate{Source: src}
	tmpl.Setup()

	assert.Equal(t, "INSERT INTO pipeline_job_logs"+
		" (pipeline, publish_time, progress, completed, log_level, log_message)"+
		" VALUES (?, ?, ?, ?, ?, ?)", tmpl.Body)
	assert.Equal(t, []string{
		"pipeline",
		"publish_time",
		"progress",
		"completed",
		"log_level",
		"log_message",
	}, tmpl.Parameters)
	assert.Equal(t, src, tmpl.Source)
}
