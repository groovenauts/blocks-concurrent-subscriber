package main

import (
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"
)

type Message struct {
	pipeline    string
	msg_id      string
	progress    int
	publishTime time.Time
	completed   string
	level       string
	data        string
	attributes  map[string]string
}

func (m *Message) load(pipeline string, attrs map[string]string) error {
	m.attributes = attrs
	m.pipeline = pipeline
	m.msg_id = attrs["job_message_id"]
	m.level = attrs["level"]
	m.completed = attrs["completed"]
	progress, err := strconv.Atoi(attrs["progress"])
	if err != nil {
		logAttrs := log.Fields(m.buildMap())
		logAttrs["error"] = err
		logAttrs["progress"] = attrs["progress"]
		log.WithFields(logAttrs).Errorf("Failed to convert %q to int", attrs["progress"])
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
		logAttrs := log.Fields(m.buildMap())
		logAttrs["error"] = err
		log.WithFields(logAttrs).Errorf("Failed to time.Parse %q\n", publishTime)
		return err
	}
	m.publishTime = t
	return nil
}

func (m *Message) completedBool() bool {
	return m.completed == "true"
}

func (m *Message) completedInt() int {
	if m.completedBool() {
		return 1
	} else {
		return 0
	}
}

func (m *Message) buildMap() map[string]interface{} {
	r := map[string]interface{}{}
	for k, v := range m.attributes {
		r[k] = v
	}
	r["job_message_id"] = m.msg_id
	r["progress"] = m.progress
	r["publishTime"] = m.publishTime
	r["completed"] = m.completed
	r["level"] = m.level
	r["data"] = m.data
	return r
}

func (m *Message) paramValues(names []string) []interface{} {
	r := []interface{}{}
	for _, name := range names {
		r = append(r, m.paramValue(name))
	}
	return r
}

func (m *Message) paramValue(name string) interface{} {
	switch name {
	case "pipeline":
		return m.pipeline
	case "job_message_id":
		return m.msg_id
	case "progress":
		return m.progress
	case "publishTime":
		return m.publishTime
	case "completed":
		return m.completed
	case "completedInt":
		return m.completedInt()
	case "level":
		return m.level
	case "data":
		return m.data
	case "now":
		return time.Now()
	default:
		return m.attributes[name]
	}
}
