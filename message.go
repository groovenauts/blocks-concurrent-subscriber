package main

import (
	"fmt"
	"strconv"
	"time"
)

type Message struct {
	msg_id      string
	progress    int
	publishTime time.Time
	completed   string
	level       string
	data        string
}

func (m *Message) load(attrs map[string]string) error {
	m.msg_id = attrs["job_message_id"]
	progress, err := strconv.Atoi(attrs["progress"])
	if err != nil {
		fmt.Printf("Failed to convert %v to int message_id: %v cause of %v", attrs["progress"], m.msg_id, err)
		return err
	}
	m.progress = progress
	return nil
}

func (m *Message) parse(publishTime string) error {
	// https://cloud.google.com/pubsub/docs/reference/rest/v1/PubsubMessage
	// A timestamp in RFC3339 UTC "Zulu" format, accurate to nanoseconds. Example: "2014-10-02T15:01:23.045123456Z".
	t, err := time.Parse(time.RFC3339, publishTime)
	if err != nil {
		fmt.Printf("Failed to parse publishTime(%v) message_id: %v cause of %v", m.publishTime, m.msg_id, err)
		return err
	}
	m.publishTime = t
	return nil
}

func (m *Message) completedInt() int {
	if m.completed == "true" {
		return 1
	} else {
		return 0
	}
}

func (m *Message) buildMap() map[string]interface{} {
	return map[string]interface{}{
		"job_message_id": m.msg_id,
		"progress":       m.progress,
		"publishTime":    m.publishTime,
		"completed":      m.completed,
		"level":          m.level,
		"data":           m.data,
	}
}
