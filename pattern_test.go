package main

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPatternsExecute(t *testing.T) {
	patterns := Patterns{
		&Pattern{
			Completed: "true",
			Command:   []string{"echo", "COMPLETED %{data}"},
		},
		&Pattern{
			Level:   "fatal",
			Command: []string{"echo", "FATAL %{data}"},
		},
		&Pattern{
			Level:   "error",
			Command: []string{"grep", "--invalid", "option"},
		},
	}

	testPatterns := []struct {
		msg    *Message
		err    string
		stdout string
		stderr string
	}{
		{
			msg: &Message{
				progress:    5,
				publishTime: time.Now(),
				completed:   "true",
				level:       "info",
				data:        "SUCCESS",
				attributes: map[string]string{
					"job_message_id": "test-msg1",
				},
			},
			stdout: "COMPLETED SUCCESS",
			stderr: "",
		},
		{
			msg: &Message{
				progress:    2,
				publishTime: time.Now(),
				completed:   "false",
				level:       "fatal",
				data:        "panic!",
				attributes: map[string]string{
					"job_message_id": "test-msg2",
				},
			},
			stdout: "FATAL panic!",
			stderr: "",
		},
		{
			msg: &Message{
				progress:    3,
				publishTime: time.Now(),
				completed:   "false",
				level:       "error",
				data:        "",
				attributes: map[string]string{
					"job_message_id": "test-msg3",
				},
			},
			err:    "exit status 2",
			stdout: "",
			stderr: "unrecognized option",
		},
		{
			msg: &Message{
				progress:    1,
				publishTime: time.Now(),
				completed:   "false",
				level:       "debug",
				data:        "Mismatch",
				attributes: map[string]string{
					"job_message_id": "",
				},
			},
			stdout: "",
			stderr: "",
		},
	}

	type calling func(msg *Message) error
	callings := []calling{
		func(msg *Message) error {
			return patterns.execute(msg)
		},
		func(msg *Message) error {
			return patterns.executeOne(msg)
		},
	}

	for _, calling := range callings {
		for _, testPattern := range testPatterns {
			bufStdout := &bytes.Buffer{}
			bufStderr := &bytes.Buffer{}
			CommandStdout = bufStdout
			CommandStderr = bufStderr
			err := calling(testPattern.msg)
			if testPattern.err != "" {
				assert.Regexp(t, testPattern.err, err.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Regexp(t, testPattern.stdout, bufStdout.String())
			assert.Regexp(t, testPattern.stderr, bufStderr.String())
		}
	}
}
