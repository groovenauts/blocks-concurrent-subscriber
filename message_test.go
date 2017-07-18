package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMessageLoad(t *testing.T) {
	raw1 := map[string]string{
		"job_message_id": "88047337842272",
		"level":          "info",
		"progress":       "2",
		"completed":      "false",
	}
	msg := &Message{}
	err := msg.load("pipelin1", raw1)
	assert.NoError(t, err)
	assert.Equal(t, raw1["job_message_id"], msg.attributes["job_message_id"])
	assert.Equal(t, raw1["level"], msg.level)
	assert.Equal(t, 2, msg.progress)
	assert.Equal(t, raw1["completed"], msg.completed)

	raw2 := map[string]string{
		"job_message_id": "88047337842272",
		"level":          "info",
		"progress":       "5",
		"completed":      "true",
	}
	msg = &Message{}
	err = msg.load("pipelin1", raw2)
	assert.NoError(t, err)
	assert.Equal(t, raw2["job_message_id"], msg.attributes["job_message_id"])
	assert.Equal(t, raw2["level"], msg.level)
	assert.Equal(t, 5, msg.progress)
	assert.Equal(t, raw2["completed"], msg.completed)
}

func TestMessageParamValues(t *testing.T) {
	raw1 := map[string]string{
		"app_id":         "123456",
		"job_message_id": "88047337842272",
		"level":          "info",
		"progress":       "2",
		"completed":      "false",
	}
	msg := &Message{}
	err := msg.load("pipelin1", raw1)
	assert.NoError(t, err)
	values := msg.paramValues([]string{"progress", "app_id"})
	assert.Equal(t, []interface{}{2, "123456"}, values)
}
