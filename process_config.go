package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
	"text/template"
)

type ProcessConfig struct {
	Datasource     string          `json:"datasource"`
	Agent          *AgentConfig    `json:"agent,omitempty"`
	Subscriptions  []*Subscription `json:"subscriptions,omitempty"`
	MessagePerPull int64           `json:"message-per-pull"`
	Interval       int             `json:"interval"`
	LogLevel       string          `json:"log-level"`
	Patterns       []*Pattern      `json:"patterns"`
}

func LoadProcessConfig(path string) (*ProcessConfig, error) {
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	funcMap := template.FuncMap{"env": os.Getenv}
	t, err := template.New("config").Funcs(funcMap).Parse(string(raw))
	if err != nil {
		return nil, err
	}

	env := map[string]string{}
	for _, s := range os.Environ() {
		parts := strings.SplitN(s, "=", 2)
		env[parts[0]] = parts[1]
	}

	buf := new(bytes.Buffer)
	t.Execute(buf, env)

	var res ProcessConfig
	err = json.Unmarshal(buf.Bytes(), &res)
	if err != nil {
		return nil, err
	}

	if res.Subscriptions != nil {
		for _, sub := range res.Subscriptions {
			sub.isOpened = func() (bool, error) {
				return true, nil
			}
		}
	}

	if res.LogLevel == "" {
		res.LogLevel = "info"
	}

	if res.MessagePerPull == 0 {
		res.MessagePerPull = 10
	}

	return &res, nil
}
