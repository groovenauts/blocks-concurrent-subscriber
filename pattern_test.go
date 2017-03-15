package main

import (
	"bytes"
	// "fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPatternsExecute(t *testing.T) {
	patterns := Patterns{
		&Pattern{
			Completed: "true",
			Command: []string{"echo", "COMPLETED %{data}"},
		},
		&Pattern{
			Level: "fatal",
			Command: []string{"echo", "FATAL %{data}"},
		},
		&Pattern{
			Level: "error",
			Command: []string{"grep", "--invalid", "option"},
		},
	}

	testPatterns := []struct {
		msg *Message
		err string
		stdout string
		stderr string
	}{
		{
			msg: &Message{
				msg_id: "test-msg1",
				progress: 5,
				publishTime: time.Now(),
				completed: "true",
				level: "info",
				data: "SUCCESS",
			},
			stdout: "COMPLETED SUCCESS",
			stderr: "",
		},
		{
			msg: &Message{
				msg_id: "test-msg2",
				progress: 2,
				publishTime: time.Now(),
				completed: "false",
				level: "fatal",
				data: "panic!",
			},
			stdout: "FATAL panic!",
			stderr: "",
		},
		{
			msg: &Message{
				msg_id: "test-msg3",
				progress: 3,
				publishTime: time.Now(),
				completed: "false",
				level: "error",
				data: "",
			},
			err: "exit status 2",
			stdout: "",
			stderr: "unrecognized option",
		},
	}

	for _, testPattern := range testPatterns {
		bufStdout := &bytes.Buffer{}
		bufStderr := &bytes.Buffer{}
		CommandStdout = bufStdout
		CommandStderr = bufStderr
		err := patterns.executeOne(testPattern.msg)
		if testPattern.err != "" {
			assert.Regexp(t, testPattern.err, err.Error())
		} else {
			assert.NoError(t, err)
		}
		assert.Regexp(t, testPattern.stdout, bufStdout.String())
		assert.Regexp(t, testPattern.stderr, bufStderr.String())
	}
}
