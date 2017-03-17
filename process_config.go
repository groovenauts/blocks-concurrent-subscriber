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
	Datasource     string     `json:"datasource"`
	AgentRootUrl   string     `json:"agent-root-url"`
	AgentRootToken string     `json:"agent-root-token"`
	Interval       int        `json:"interval"`
	LogLevel       string     `json:"log-level"`
	Patterns       []*Pattern `json:"patterns"`
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

	if res.LogLevel == "" {
		res.LogLevel = "info"
	}

	return &res, nil
}
