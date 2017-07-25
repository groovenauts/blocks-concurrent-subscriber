package main

import (
	"regexp"
	"strings"
)

// Examples:
//   INSERT INTO pipeline_job_logs \
//       (pipeline, publish_time, progress, completed, log_level, log_message) \
//       VALUES ($pipeline, $publish_time, $progress, $completed, $log_level, $log_message)
//
// 	UPDATE pipeline_jobs \
//     SET progress = $progress, updated_at = $now \
//   WHERE id = $app_id AND progress < $progress
//

type SqlTemplate struct {
	Source     string
	Body       string
	Parameters []string
}

var ParamRegexp = regexp.MustCompile(`\$[^\s\,\=\>\<\(\)]+`)

func BuildSqlTemplate(src string) *SqlTemplate {
	if src == "" {
		return nil
	}
	t := &SqlTemplate{Source: src}
	t.Setup()
	return t
}

func (t *SqlTemplate) Setup() {
	t.Parameters = []string{}
	t.Body = t.Source
	words := ParamRegexp.FindAllString(t.Source, -1)
	for _, w := range words {
		name := w[1:len(w)]
		t.Parameters = append(t.Parameters, name)
		t.Body = strings.Replace(t.Body, w, "?", -1)
	}
}
